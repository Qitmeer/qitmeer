// Copyright (c) 2017-2018 The qitmeer developers

package index

import (
	"bytes"
	"fmt"
	"github.com/Qitmeer/qitmeer/common/hash"
	"github.com/Qitmeer/qitmeer/common/math"
	"github.com/Qitmeer/qitmeer/core/blockchain"
	"github.com/Qitmeer/qitmeer/core/dbnamespace"
	"github.com/Qitmeer/qitmeer/core/types"
	"github.com/Qitmeer/qitmeer/database"
	"github.com/Qitmeer/qitmeer/log"
	"github.com/Qitmeer/qitmeer/params"
	"github.com/Qitmeer/qitmeer/services/common/progresslog"
)

// Manager defines an index manager that manages multiple optional indexes and
// implements the blockchain.IndexManager interface so it can be seamlessly
// plugged into normal chain processing.
type Manager struct {
	params         *params.Params
	db             database.DB
	enabledIndexes []Indexer
}

// Ensure the Manager type implements the blockchain.IndexManager interface.
var _ blockchain.IndexManager = (*Manager)(nil)

// NewManager returns a new index manager with the provided indexes enabled.
//
// The manager returned satisfies the blockchain.IndexManager interface and thus
// cleanly plugs into the normal blockchain processing path.
func NewManager(db database.DB, enabledIndexes []Indexer, params *params.Params) *Manager {
	return &Manager{
		db:             db,
		enabledIndexes: enabledIndexes,
		params:         params,
	}
}

// Init initializes the enabled indexes.  This is called during chain
// initialization and primarily consists of catching up all indexes to the
// current best chain tip.  This is necessary since each index can be disabled
// and re-enabled at any time and attempting to catch-up indexes at the same
// time new blocks are being downloaded would lead to an overall longer time to
// catch up due to the I/O contention.
//
// This is part of the blockchain.IndexManager interface.
func (m *Manager) Init(chain *blockchain.BlockChain, interrupt <-chan struct{}) error {
	// Nothing to do when no indexes are enabled.
	if len(m.enabledIndexes) == 0 {
		return nil
	}

	if interruptRequested(interrupt) {
		return errInterruptRequested
	}

	// Finish any drops that were previously interrupted.
	if err := m.maybeFinishDrops(interrupt); err != nil {
		return err
	}

	// Create the initial state for the indexes as needed.
	err := m.db.Update(func(dbTx database.Tx) error {
		// Create the bucket for the current tips as needed.
		meta := dbTx.Metadata()
		_, err := meta.CreateBucketIfNotExists(dbnamespace.IndexTipsBucketName)
		if err != nil {
			return err
		}

		return m.maybeCreateIndexes(dbTx)
	})
	if err != nil {
		return err
	}

	// Initialize each of the enabled indexes.
	for _, indexer := range m.enabledIndexes {
		if err := indexer.Init(); err != nil {
			return err
		}
		if indexer.Name() == txIndexName {
			indexer.(*TxIndex).chain = chain
			if chain.CacheInvalidTx {
				if indexer.(*TxIndex).curBlockID == 0 {
					m.db.Update(func(dbTx database.Tx) error {
						dbTx.Metadata().Put(dbnamespace.CacheInvalidTxName, []byte{byte(0)})
						return nil
					})
				}
			} else {
				m.db.Update(func(dbTx database.Tx) error {
					dbTx.Metadata().Delete(dbnamespace.CacheInvalidTxName)
					return nil
				})
			}

		}
	}

	bestOrder := uint32(chain.BestSnapshot().GraphState.GetMainOrder())

	// Rollback indexes to the main chain if their tip is an orphaned fork.
	// This is fairly unlikely, but it can happen if the chain is
	// reorganized while the index is disabled.  This has to be done in
	// reverse order because later indexes can depend on earlier ones.
	var spentTxos []blockchain.SpentTxOut
	for i := len(m.enabledIndexes); i > 0; i-- {
		indexer := m.enabledIndexes[i-1]

		// Fetch the current tip for the index.
		var order uint32
		err := m.db.View(func(dbTx database.Tx) error {
			idxKey := indexer.Key()
			_, order, err = dbFetchIndexerTip(dbTx, idxKey)
			return err
		})
		if err != nil {
			return err
		}

		// Nothing to do if the index does not have any entries yet.
		if order == math.MaxUint32 {
			continue
		}
		var block *types.SerializedBlock
		for order > bestOrder {
			err = m.db.Update(func(dbTx database.Tx) error {
				// Load the block for the height since it is required to index
				// it.
				block, err = chain.DBFetchBlockByOrder(dbTx, uint64(order))
				if err != nil {
					return err
				}
				spentTxos = nil
				if indexNeedsInputs(indexer) {
					spentTxos, err = chain.FetchSpendJournal(block)
					if err != nil {
						return err
					}
				}
				err = m.dbIndexDisconnectBlock(dbTx, indexer, block, spentTxos)
				if err != nil {
					return err
				}
				log.Trace(fmt.Sprintf("%s rollback order= %d", indexer.Name(), order))
				order--
				return nil
			})
			if err != nil {
				return err
			}
			if interruptRequested(interrupt) {
				return errInterruptRequested
			}
		}
	}

	// Fetch the current tip heights for each index along with tracking the
	// lowest one so the catchup code only needs to start at the earliest
	// block and is able to skip connecting the block for the indexes that
	// don't need it.

	lowestOrder := int64(bestOrder)
	indexerOrders := make([]int64, len(m.enabledIndexes))
	err = m.db.View(func(dbTx database.Tx) error {
		for i, indexer := range m.enabledIndexes {
			idxKey := indexer.Key()
			h, order, err := dbFetchIndexerTip(dbTx, idxKey)
			if err != nil {
				return err
			}
			orderShow := int64(order)
			if order == math.MaxUint32 {
				lowestOrder = -1
				indexerOrders[i] = -1
				orderShow = -1
			} else if int64(order) < lowestOrder {
				lowestOrder = int64(order)
				indexerOrders[i] = int64(order)
			}
			log.Debug(fmt.Sprintf("Current %s tip", indexer.Name()),
				"order", orderShow, "hash", h)
		}
		return nil
	})
	if err != nil {
		return err
	}

	// Nothing to index if all of the indexes are caught up.
	if lowestOrder == int64(bestOrder) {
		return nil
	}

	// Create a progress logger for the indexing process below.
	progressLogger := progresslog.NewBlockProgressLogger("Indexed", log.Root())

	// At this point, one or more indexes are behind the current best chain
	// tip and need to be caught up, so log the details and loop through
	// each block that needs to be indexed.
	log.Info(fmt.Sprintf("Catching up indexes from order %d to %d", lowestOrder,
		bestOrder))

	for order := lowestOrder + 1; order <= int64(bestOrder); order++ {
		if interruptRequested(interrupt) {
			return errInterruptRequested
		}

		var block *types.SerializedBlock
		err = m.db.Update(func(dbTx database.Tx) error {
			// Load the block for the height since it is required to index
			// it.
			block, err = chain.DBFetchBlockByOrder(dbTx, uint64(order))
			if err != nil {
				return err
			}
			return nil
		})
		if err != nil {
			return err
		}

		if interruptRequested(interrupt) {
			return errInterruptRequested
		}
		chain.CalculateDAGDuplicateTxs(block)
		// Connect the block for all indexes that need it.
		spentTxos = nil
		for i, indexer := range m.enabledIndexes {
			// Skip indexes that don't need to be updated with this
			// block.
			if indexerOrders[i] >= order {
				continue
			}

			// When the index requires all of the referenced
			// txouts and they haven't been loaded yet, they
			// need to be retrieved from the transaction
			// index.
			if spentTxos == nil && indexNeedsInputs(indexer) {
				spentTxos, err = chain.FetchSpendJournal(block)
				if err != nil {
					return err
				}
			}
			err = m.db.Update(func(dbTx database.Tx) error {
				return dbIndexConnectBlock(dbTx, indexer, block, spentTxos)
			})
			if err != nil {
				return err
			}
			indexerOrders[i] = order
		}

		progressLogger.LogBlockHeight(block)
	}

	log.Info(fmt.Sprintf("Indexes caught up to order %d", bestOrder))
	return nil
}

// maybeFinishDrops determines if each of the enabled indexes are in the middle
// of being dropped and finishes dropping them when the are.  This is necessary
// because dropping and index has to be done in several atomic steps rather than
// one big atomic step due to the massive number of entries.
func (m *Manager) maybeFinishDrops(interrupt <-chan struct{}) error {
	indexNeedsDrop := make([]bool, len(m.enabledIndexes))
	err := m.db.View(func(dbTx database.Tx) error {
		// None of the indexes needs to be dropped if the index tips
		// bucket hasn't been created yet.
		indexesBucket := dbTx.Metadata().Bucket(dbnamespace.IndexTipsBucketName)
		if indexesBucket == nil {
			return nil
		}

		// Mark the indexer as requiring a drop if one is already in
		// progress.
		for i, indexer := range m.enabledIndexes {
			dropKey := indexDropKey(indexer.Key())
			if indexesBucket.Get(dropKey) != nil {
				indexNeedsDrop[i] = true
			}
		}

		return nil
	})
	if err != nil {
		return err
	}

	if interruptRequested(interrupt) {
		return errInterruptRequested
	}

	// Finish dropping any of the enabled indexes that are already in the
	// middle of being dropped.
	for i, indexer := range m.enabledIndexes {
		if !indexNeedsDrop[i] {
			continue
		}

		log.Info(fmt.Sprintf("Resuming %s drop", indexer.Name()))
		err := dropIndex(m.db, indexer.Key(), indexer.Name(), interrupt)
		if err != nil {
			return err
		}
	}

	return nil
}

// indexNeedsInputs returns whether or not the index needs access to the txouts
// referenced by the transaction inputs being indexed.
func indexNeedsInputs(index Indexer) bool {
	if idx, ok := index.(NeedsInputser); ok {
		return idx.NeedsInputs()
	}
	return false
}

// maybeCreateIndexes determines if each of the enabled indexes have already
// been created and creates them if not.
func (m *Manager) maybeCreateIndexes(dbTx database.Tx) error {
	indexesBucket := dbTx.Metadata().Bucket(dbnamespace.IndexTipsBucketName)
	for _, indexer := range m.enabledIndexes {
		// Nothing to do if the index tip already exists.
		idxKey := indexer.Key()
		if indexesBucket.Get(idxKey) != nil {
			continue
		}

		// The tip for the index does not exist, so create it and
		// invoke the create callback for the index so it can perform
		// any one-time initialization it requires.
		if err := indexer.Create(dbTx); err != nil {
			return err
		}

		// Set the tip for the index to values which represent an
		// uninitialized index (the genesis block hash and height).
		err := dbPutIndexerTip(dbTx, idxKey, &hash.ZeroHash, math.MaxUint32)
		if err != nil {
			return err
		}
	}

	return nil
}

// ConnectBlock must be invoked when a block is extending the main chain.  It
// keeps track of the state of each index it is managing, performs some sanity
// checks, and invokes each indexer.
//
// This is part of the blockchain.IndexManager interface.
func (m *Manager) ConnectBlock(dbTx database.Tx, block *types.SerializedBlock, stxos []blockchain.SpentTxOut) error {
	// Call each of the currently active optional indexes with the block
	// being connected so they can update accordingly.
	for _, index := range m.enabledIndexes {
		err := dbIndexConnectBlock(dbTx, index, block, stxos)
		if err != nil {
			return err
		}
	}
	return nil
}

// DisconnectBlock must be invoked when a block is being disconnected from the
// end of the main chain.  It keeps track of the state of each index it is
// managing, performs some sanity checks, and invokes each indexer to remove
// the index entries associated with the block.
//
// This is part of the blockchain.IndexManager interface.
func (m *Manager) DisconnectBlock(dbTx database.Tx, block *types.SerializedBlock, stxos []blockchain.SpentTxOut) error {
	// Call each of the currently active optional indexes with the block
	// being disconnected so they can update accordingly.
	for _, index := range m.enabledIndexes {
		err := m.dbIndexDisconnectBlock(dbTx, index, block, stxos)
		if err != nil {
			return err
		}
	}
	return nil
}

// HasTransaction
func (m *Manager) IsDuplicateTx(dbTx database.Tx, txid *hash.Hash, blockHash *hash.Hash) bool {
	blockRegion, err := dbFetchTxIndexEntry(dbTx, txid)
	if err != nil {
		return false
	}
	if blockRegion == nil {
		return false
	}
	if blockRegion.Hash.IsEqual(blockHash) {
		return false
	}
	return true
}

// dbFetchTx looks up the passed transaction hash in the transaction index and
// loads it from the database.
func DBFetchTx(dbTx database.Tx, hash *hash.Hash) (*types.Transaction, error) {
	// Look up the location of the transaction.
	blockRegion, err := dbFetchTxIndexEntry(dbTx, hash)
	if err != nil {
		return nil, err
	}
	if blockRegion == nil {
		return nil, fmt.Errorf("transaction %v not found in the txindex", hash)
	}

	// Load the raw transaction bytes from the database.
	txBytes, err := dbTx.FetchBlockRegion(blockRegion)
	if err != nil {
		return nil, err
	}

	// Deserialize the transaction.
	var tx types.Transaction
	err = tx.Deserialize(bytes.NewReader(txBytes))
	if err != nil {
		return nil, err
	}

	return &tx, nil
}

func DBFetchTxAndBlock(dbTx database.Tx, hash *hash.Hash) (*types.Transaction, *hash.Hash, error) {
	// Look up the location of the transaction.
	blockRegion, err := dbFetchTxIndexEntry(dbTx, hash)
	if err != nil {
		return nil, nil, err
	}
	if blockRegion == nil {
		return nil, nil, fmt.Errorf("transaction %v not found in the txindex", hash)
	}

	// Load the raw transaction bytes from the database.
	txBytes, err := dbTx.FetchBlockRegion(blockRegion)
	if err != nil {
		return nil, nil, err
	}

	// Deserialize the transaction.
	var tx types.Transaction
	err = tx.Deserialize(bytes.NewReader(txBytes))
	if err != nil {
		return nil, nil, err
	}

	return &tx, blockRegion.Hash, nil
}

// dbIndexDisconnectBlock removes all of the index entries associated with the
// given block using the provided indexer and updates the tip of the indexer
// accordingly.  An error will be returned if the current tip for the indexer is
// not the passed block.
func (m *Manager) dbIndexDisconnectBlock(dbTx database.Tx, indexer Indexer, block *types.SerializedBlock, stxos []blockchain.SpentTxOut) error {
	// Assert that the block being disconnected is the current tip of the
	// index.
	idxKey := indexer.Key()
	curTipHash, order, err := dbFetchIndexerTip(dbTx, idxKey)
	if err != nil {
		return err
	}
	if !curTipHash.IsEqual(block.Hash()) {
		log.Warn(fmt.Sprintf("dbIndexDisconnectBlock must "+
			"be called with the block at the current index tip "+
			"(%s, tip %s, block %s)", indexer.Name(),
			curTipHash, block.Hash()))
		return nil
	}
	if order == math.MaxUint32 {
		log.Warn(fmt.Sprintf("Can't disconnect root index tip"))
		return nil
	}
	// Notify the indexer with the disconnected block so it can remove all
	// of the appropriate entries.
	if err := indexer.DisconnectBlock(dbTx, block, stxos); err != nil {
		return err
	}

	// Update the current index tip.
	var prevHash *hash.Hash
	var preOrder uint32
	if order == 0 {
		prevHash = &hash.ZeroHash
		preOrder = math.MaxUint32
	} else {
		prevHash, err = dbFetchBlockHashByID(dbTx, order)
		if err != nil {
			return err
		}
		preOrder = uint32(order - 1)
	}

	return dbPutIndexerTip(dbTx, idxKey, prevHash, preOrder)
}

// dbIndexConnectBlock adds all of the index entries associated with the
// given block using the provided indexer and updates the tip of the indexer
// accordingly.  An error will be returned if the current tip for the indexer is
// not the previous block for the passed block.
func dbIndexConnectBlock(dbTx database.Tx, indexer Indexer, block *types.SerializedBlock, stxos []blockchain.SpentTxOut) error {
	// Assert that the block being connected properly connects to the
	// current tip of the index.
	idxKey := indexer.Key()
	_, order, err := dbFetchIndexerTip(dbTx, idxKey)
	if err != nil {
		return err
	}
	if order != math.MaxUint32 && order+1 != uint32(block.Order()) ||
		order == math.MaxUint32 && block.Order() != 0 {

		log.Warn(fmt.Sprintf("dbIndexConnectBlock must be "+
			"called with a block that extends the current index "+
			"tip (%s, tip %d, block %d)", indexer.Name(),
			order, block.Order()))
		return nil
	}

	// Notify the indexer with the connected block so it can index it.
	if err := indexer.ConnectBlock(dbTx, block, stxos); err != nil {
		return err
	}

	// Update the current index tip.
	return dbPutIndexerTip(dbTx, idxKey, block.Hash(), uint32(block.Order()))
}

// dbFetchIndexerTip uses an existing database transaction to retrieve the
// hash and height of the current tip for the provided index.
func dbFetchIndexerTip(dbTx database.Tx, idxKey []byte) (*hash.Hash, uint32, error) {
	indexesBucket := dbTx.Metadata().Bucket(dbnamespace.IndexTipsBucketName)
	serialized := indexesBucket.Get(idxKey)
	if len(serialized) < hash.HashSize+4 {
		return nil, 0, database.Error{
			ErrorCode: database.ErrCorruption,
			Description: fmt.Sprintf("unexpected end of data for "+
				"index %q tip", string(idxKey)),
		}
	}

	var h hash.Hash
	copy(h[:], serialized[:hash.HashSize])
	order := uint32(byteOrder.Uint32(serialized[hash.HashSize:]))
	return &h, order, nil
}

// -----------------------------------------------------------------------------
// The index manager tracks the current tip of each index by using a parent
// bucket that contains an entry for index.
//
// The serialized format for an index tip is:
//
//   [<block hash><block order>],...
//
//   Field           Type             Size
//   block hash      chainhash.Hash   chainhash.HashSize
//   block order    uint32           4 bytes
// -----------------------------------------------------------------------------

// dbPutIndexerTip uses an existing database transaction to update or add the
// current tip for the given index to the provided values.
func dbPutIndexerTip(dbTx database.Tx, idxKey []byte, h *hash.Hash, order uint32) error {
	serialized := make([]byte, hash.HashSize+4)
	copy(serialized, h[:])
	byteOrder.PutUint32(serialized[hash.HashSize:], uint32(order))

	indexesBucket := dbTx.Metadata().Bucket(dbnamespace.IndexTipsBucketName)
	return indexesBucket.Put(idxKey, serialized)
}

// existsIndex returns whether the index keyed by idxKey exists in the database.
func existsIndex(db database.DB, idxKey []byte, idxName string) (bool, error) {
	var exists bool
	err := db.View(func(dbTx database.Tx) error {
		indexesBucket := dbTx.Metadata().Bucket(dbnamespace.IndexTipsBucketName)
		if indexesBucket != nil && indexesBucket.Get(idxKey) != nil {
			exists = true
		}
		return nil
	})
	return exists, err
}

// markIndexDeletion marks the index identified by idxKey for deletion.  Marking
// an index for deletion allows deletion to resume next startup if an
// incremental deletion was interrupted.
func markIndexDeletion(db database.DB, idxKey []byte) error {
	return db.Update(func(dbTx database.Tx) error {
		indexesBucket := dbTx.Metadata().Bucket(dbnamespace.IndexTipsBucketName)
		return indexesBucket.Put(indexDropKey(idxKey), idxKey)
	})
}

// indexDropKey returns the key for an index which indicates it is in the
// process of being dropped.
func indexDropKey(idxKey []byte) []byte {
	dropKey := make([]byte, len(idxKey)+1)
	dropKey[0] = 'd'
	copy(dropKey[1:], idxKey)
	return dropKey
}

// incrementalFlatDrop uses multiple database updates to remove key/value pairs
// saved to a flat index.
func incrementalFlatDrop(db database.DB, idxKey []byte, idxName string, interrupt <-chan struct{}) error {
	// Since the indexes can be so large, attempting to simply delete
	// the bucket in a single database transaction would result in massive
	// memory usage and likely crash many systems due to ulimits.  In order
	// to avoid this, use a cursor to delete a maximum number of entries out
	// of the bucket at a time. Recurse buckets depth-first to delete any
	// sub-buckets.
	const maxDeletions = 2000000
	var totalDeleted uint64

	// Recurse through all buckets in the index, cataloging each for
	// later deletion.
	var subBuckets [][][]byte
	var subBucketClosure func(database.Tx, []byte, [][]byte) error
	subBucketClosure = func(dbTx database.Tx,
		subBucket []byte, tlBucket [][]byte) error {
		// Get full bucket name and append to subBuckets for later
		// deletion.
		var bucketName [][]byte
		if (tlBucket == nil) || (len(tlBucket) == 0) {
			bucketName = append(bucketName, subBucket)
		} else {
			bucketName = append(tlBucket, subBucket)
		}
		subBuckets = append(subBuckets, bucketName)
		// Recurse sub-buckets to append to subBuckets slice.
		bucket := dbTx.Metadata()
		for _, subBucketName := range bucketName {
			bucket = bucket.Bucket(subBucketName)
			if bucket == nil {
				return database.Error{
					ErrorCode:   database.ErrBucketNotFound,
					Description: fmt.Sprintf("db bucket '%s' not found, your data is corrupted, please clean up your block database by using '--cleanup'", subBucketName),
					Err:         nil}
			}

		}
		return bucket.ForEachBucket(func(k []byte) error {
			return subBucketClosure(dbTx, k, bucketName)
		})
	}

	// Call subBucketClosure with top-level bucket.
	err := db.View(func(dbTx database.Tx) error {
		return subBucketClosure(dbTx, idxKey, nil)
	})
	if err != nil {
		return err
	}

	// Iterate through each sub-bucket in reverse, deepest-first, deleting
	// all keys inside them and then dropping the buckets themselves.
	for i := range subBuckets {
		bucketName := subBuckets[len(subBuckets)-1-i]
		// Delete maxDeletions key/value pairs at a time.
		for numDeleted := maxDeletions; numDeleted == maxDeletions; {
			numDeleted = 0
			err := db.Update(func(dbTx database.Tx) error {
				subBucket := dbTx.Metadata()
				for _, subBucketName := range bucketName {
					subBucket = subBucket.Bucket(subBucketName)
				}
				cursor := subBucket.Cursor()
				for ok := cursor.First(); ok; ok = cursor.Next() &&
					numDeleted < maxDeletions {

					if err := cursor.Delete(); err != nil {
						return err
					}
					numDeleted++
				}
				return nil
			})
			if err != nil {
				return err
			}

			if numDeleted > 0 {
				totalDeleted += uint64(numDeleted)
				log.Info(fmt.Sprintf("Deleted %d keys (%d total) from %s",
					numDeleted, totalDeleted, idxName))
			}
		}

		if interruptRequested(interrupt) {
			return errInterruptRequested
		}

		// Drop the bucket itself.
		db.Update(func(dbTx database.Tx) error {
			bucket := dbTx.Metadata()
			for j := 0; j < len(bucketName)-1; j++ {
				bucket = bucket.Bucket(bucketName[j])
			}
			return bucket.DeleteBucket(bucketName[len(bucketName)-1])
		})
	}
	return nil
}

// dropIndexMetadata drops the passed index from the database by removing the
// top level bucket for the index, the index tip, and any in-progress drop flag.
func dropIndexMetadata(db database.DB, idxKey []byte, idxName string) error {
	return db.Update(func(dbTx database.Tx) error {
		meta := dbTx.Metadata()
		indexesBucket := meta.Bucket(dbnamespace.IndexTipsBucketName)
		err := indexesBucket.Delete(idxKey)
		if err != nil {
			return err
		}

		err = meta.DeleteBucket(idxKey)
		if err != nil && !database.IsError(err, database.ErrBucketNotFound) {
			return err
		}

		return indexesBucket.Delete(indexDropKey(idxKey))
	})
}

// dropIndex drops the passed index from the database without using incremental
// deletion.  This should be used to drop indexes containing nested buckets,
// which can not be deleted with dropFlatIndex.
func dropIndex(db database.DB, idxKey []byte, idxName string, interrupt <-chan struct{}) error {
	// Nothing to do if the index doesn't already exist.
	exists, err := existsIndex(db, idxKey, idxName)
	if err != nil {
		return err
	}
	if !exists {
		log.Info(fmt.Sprintf("Not dropping %s because it does not exist", idxName))
		return nil
	}

	log.Info(fmt.Sprintf("Dropping all %s entries.  This might take a while...",
		idxName))

	// Mark that the index is in the process of being dropped so that it
	// can be resumed on the next start if interrupted before the process is
	// complete.
	err = markIndexDeletion(db, idxKey)
	if err != nil {
		return err
	}

	// Since the indexes can be so large, attempting to simply delete
	// the bucket in a single database transaction would result in massive
	// memory usage and likely crash many systems due to ulimits.  In order
	// to avoid this, use a cursor to delete a maximum number of entries out
	// of the bucket at a time.
	err = incrementalFlatDrop(db, idxKey, idxName, interrupt)
	if err != nil {
		return err
	}

	// Call extra index specific deinitialization for the transaction index.
	if idxName == txIndexName {
		err = dropBlockIDIndex(db)
		if err != nil {
			return err
		}
		err = dropInvalidTx(db)
		if err != nil {
			return err
		}
	}

	// Remove the index tip, index bucket, and in-progress drop flag.  Removing
	// the index bucket also recursively removes all values saved to the index.
	err = dropIndexMetadata(db, idxKey, idxName)
	if err != nil {
		return err
	}

	log.Info(fmt.Sprintf("Dropped %s", idxName))
	return nil
}
