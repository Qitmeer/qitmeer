// Copyright (c) 2017-2018 The nox developers

package index

import (

	"bytes"
	"errors"
	"fmt"
	"github.com/HalalChain/qitmeer/common/hash"
	"github.com/HalalChain/qitmeer/core/dbnamespace"
	"github.com/HalalChain/qitmeer/core/blockchain"
	"github.com/HalalChain/qitmeer/core/types"
	"github.com/HalalChain/qitmeer/database"
	"github.com/HalalChain/qitmeer/log"
	"github.com/HalalChain/qitmeer/params"
	"github.com/HalalChain/qitmeer/services/common/progresslog"

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
	}

	// Rollback indexes to the main chain if their tip is an orphaned fork.
	// This is fairly unlikely, but it can happen if the chain is
	// reorganized while the index is disabled.  This has to be done in
	// reverse order because later indexes can depend on earlier ones.
	for i := len(m.enabledIndexes); i > 0; i-- {
		indexer := m.enabledIndexes[i-1]

		// Fetch the current tip for the index.
		var height int32
		err := m.db.View(func(dbTx database.Tx) error {
			idxKey := indexer.Key()
			_, height, err = dbFetchIndexerTip(dbTx, idxKey)
			return err
		})
		if err != nil {
			return err
		}

		// Nothing to do if the index does not have any entries yet.
		if height == 0 {
			continue
		}

	}

	// Fetch the current tip heights for each index along with tracking the
	// lowest one so the catchup code only needs to start at the earliest
	// block and is able to skip connecting the block for the indexes that
	// don't need it.
	bestHeight := int32(chain.BestSnapshot().Order)
	lowestHeight := bestHeight
	indexerHeights := make([]int32, len(m.enabledIndexes))
	err = m.db.View(func(dbTx database.Tx) error {
		for i, indexer := range m.enabledIndexes {
			idxKey := indexer.Key()
			h, height, err := dbFetchIndexerTip(dbTx, idxKey)
			if err != nil {
				return err
			}
			log.Debug(fmt.Sprintf("Current %s tip", indexer.Name()),
				"height", height, "hash", h)
			indexerHeights[i] = height
			if height < lowestHeight {
				lowestHeight = height
			}
		}
		return nil
	})
	if err != nil {
		return err
	}

	// Nothing to index if all of the indexes are caught up.
	if lowestHeight == bestHeight {
		return nil
	}

	// Create a progress logger for the indexing process below.
	progressLogger := progresslog.NewBlockProgressLogger("Indexed", log.Root())

	// At this point, one or more indexes are behind the current best chain
	// tip and need to be caught up, so log the details and loop through
	// each block that needs to be indexed.
	log.Info(fmt.Sprintf("Catching up indexes from height %d to %d", lowestHeight,
		bestHeight))

	for height := lowestHeight + 1; height <= bestHeight; height++ {
		if interruptRequested(interrupt) {
			return errInterruptRequested
		}

		var block *types.SerializedBlock
		err = m.db.Update(func(dbTx database.Tx) error {
			// Load the block for the height since it is required to index
			// it.
			block, err = blockchain.DBFetchBlockByOrder(dbTx,
				uint64(height))
			if err != nil {
				return err
			}

			if interruptRequested(interrupt) {
				return errInterruptRequested
			}

			// Connect the block for all indexes that need it.
			var view *blockchain.UtxoViewpoint
			for i, indexer := range m.enabledIndexes {
				// Skip indexes that don't need to be updated with this
				// block.
				if indexerHeights[i] >= height {
					continue
				}

				// When the index requires all of the referenced
				// txouts and they haven't been loaded yet, they
				// need to be retrieved from the transaction
				// index.
				if view == nil && indexNeedsInputs(indexer) {
					var errMakeView error
					view, errMakeView = makeUtxoView(dbTx, block,interrupt)
					if errMakeView != nil {
						return errMakeView
					}
				}
				err = dbIndexConnectBlock(dbTx, indexer, block, view)
				if err != nil {
					return err
				}

				indexerHeights[i] = height
			}

			return nil
		})
		if err != nil {
			return err
		}
		progressLogger.LogBlockHeight(block)
	}

	log.Info(fmt.Sprintf("Indexes caught up to height %d", bestHeight))
	return nil
}

// makeUtxoView creates a mock unspent transaction output view by using the
// transaction index in order to look up all inputs referenced by the
// transactions in the block.  This is sometimes needed when catching indexes up
// because many of the txouts could actually already be spent however the
// associated scripts are still required to index them.
func makeUtxoView(dbTx database.Tx, block *types.SerializedBlock, interrupt <-chan struct{}) (*blockchain.UtxoViewpoint, error) {
	view := blockchain.NewUtxoViewpoint()
	var parentRegularTxs []*types.Tx = block.Transactions()

	for txIdx, tx := range parentRegularTxs {
		// Coinbases do not reference any inputs.  Since the block is
		// required to have already gone through full validation, it has
		// already been proven on the first transaction in the block is
		// a coinbase.
		if txIdx == 0 {
			continue
		}

		// Use the transaction index to load all of the referenced
		// inputs and add their outputs to the view.
		for _, txIn := range tx.Transaction().TxIn {
			// Skip already fetched outputs.
			originOut := &txIn.PreviousOut
			if view.LookupEntry(&originOut.Hash) != nil {
				continue
			}

			originTx, err := dbFetchTx(dbTx, &originOut.Hash)
			if err != nil {
				return nil, err
			}

			view.AddTxOuts(types.NewTx(originTx),
				int64(types.NullBlockHeight), types.NullTxIndex)
		}

		if interruptRequested(interrupt) {
			return nil, errInterruptRequested
		}
	}
	return view, nil
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

		switch d := indexer.(type) {
		case IndexDropper:
			err := d.DropIndex(m.db, interrupt)
			if err != nil {
				return err
			}
		default:
			err := dropIndex(m.db, indexer.Key(), indexer.Name())
			if err != nil {
				return err
			}
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
		genesisBlockHash := m.params.GenesisBlock.BlockHash()
		err := dbPutIndexerTip(dbTx, idxKey, &genesisBlockHash, 0)
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
func (m *Manager) ConnectBlock(dbTx database.Tx, block *types.SerializedBlock, view *blockchain.UtxoViewpoint) error {
	// Call each of the currently active optional indexes with the block
	// being connected so they can update accordingly.
	for _, index := range m.enabledIndexes {
		err := dbIndexConnectBlock(dbTx, index, block, view)
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
func (m *Manager) DisconnectBlock(dbTx database.Tx, block *types.SerializedBlock, view *blockchain.UtxoViewpoint) error {
	// Call each of the currently active optional indexes with the block
	// being disconnected so they can update accordingly.
	for _, index := range m.enabledIndexes {
		err := dbIndexDisconnectBlock(dbTx, index, block, view)
		if err != nil {
			return err
		}
	}
	return nil
}

// dbFetchTx looks up the passed transaction hash in the transaction index and
// loads it from the database.
func dbFetchTx(dbTx database.Tx, hash *hash.Hash) (*types.Transaction, error) {
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

// dbIndexDisconnectBlock removes all of the index entries associated with the
// given block using the provided indexer and updates the tip of the indexer
// accordingly.  An error will be returned if the current tip for the indexer is
// not the passed block.
func dbIndexDisconnectBlock(dbTx database.Tx, indexer Indexer, block *types.SerializedBlock, view *blockchain.UtxoViewpoint) error {
	// Assert that the block being disconnected is the current tip of the
	// index.
	idxKey := indexer.Key()
	/*curTipHash, _, err := dbFetchIndexerTip(dbTx, idxKey)
	if err != nil {
		return err
	}
	if !curTipHash.IsEqual(block.Hash()) {
		return AssertError(fmt.Sprintf("dbIndexDisconnectBlock must "+
			"be called with the block at the current index tip "+
			"(%s, tip %s, block %s)", indexer.Name(),
			curTipHash, block.Hash()))
	}*/

	// Notify the indexer with the disconnected block so it can remove all
	// of the appropriate entries.
	if err := indexer.DisconnectBlock(dbTx, block, view); err != nil {
		return err
	}

	// Update the current index tip.
	prevHash := &block.Block().Header.ParentRoot
	return dbPutIndexerTip(dbTx, idxKey, prevHash, int32(block.Order()-1))
}

// dbIndexConnectBlock adds all of the index entries associated with the
// given block using the provided indexer and updates the tip of the indexer
// accordingly.  An error will be returned if the current tip for the indexer is
// not the previous block for the passed block.
func dbIndexConnectBlock(dbTx database.Tx, indexer Indexer, block *types.SerializedBlock, view *blockchain.UtxoViewpoint) error {
	// Assert that the block being connected properly connects to the
	// current tip of the index.
	idxKey := indexer.Key()
	/*curTipHash, _, err := dbFetchIndexerTip(dbTx, idxKey)
	if err != nil {
		return err
	}
	if !curTipHash.IsEqual(&block.Block().Header.ParentRoot) {
		return AssertError(fmt.Sprintf("dbIndexConnectBlock must be "+
			"called with a block that extends the current index "+
			"tip (%s, tip %s, block %s)", indexer.Name(),
			curTipHash, block.Hash()))
	}*/

	// Notify the indexer with the connected block so it can index it.
	if err := indexer.ConnectBlock(dbTx, block, view); err != nil {
		return err
	}

	// Update the current index tip.
	return dbPutIndexerTip(dbTx, idxKey, block.Hash(), int32(block.Order()))
}

// dbFetchIndexerTip uses an existing database transaction to retrieve the
// hash and height of the current tip for the provided index.
func dbFetchIndexerTip(dbTx database.Tx, idxKey []byte) (*hash.Hash, int32, error) {
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
	height := int32(byteOrder.Uint32(serialized[hash.HashSize:]))
	return &h, height, nil
}

// -----------------------------------------------------------------------------
// The index manager tracks the current tip of each index by using a parent
// bucket that contains an entry for index.
//
// The serialized format for an index tip is:
//
//   [<block hash><block height>],...
//
//   Field           Type             Size
//   block hash      chainhash.Hash   chainhash.HashSize
//   block height    uint32           4 bytes
// -----------------------------------------------------------------------------

// dbPutIndexerTip uses an existing database transaction to update or add the
// current tip for the given index to the provided values.
func dbPutIndexerTip(dbTx database.Tx, idxKey []byte, h *hash.Hash, height int32) error {
	serialized := make([]byte, hash.HashSize+4)
	copy(serialized, h[:])
	byteOrder.PutUint32(serialized[hash.HashSize:], uint32(height))

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

// dropFlatIndex incrementally drops the passed index from the database.  Since
// indexes can be massive, it deletes the index in multiple database
// transactions in order to keep memory usage to reasonable levels.  For this
// algorithm to work, the index must be "flat" (have no nested buckets).  It
// also marks the drop in progress so the drop can be resumed if it is stopped
// before it is done before the index can be used again.
func dropFlatIndex(db database.DB, idxKey []byte, idxName string, interrupt <-chan struct{}) error {
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

	// Remove the index tip, index bucket, and in-progress drop flag now
	// that all index entries have been removed.
	err = dropIndexMetadata(db, idxKey, idxName)
	if err != nil {
		return err
	}

	log.Info(fmt.Sprintf("Dropped %s", idxName))
	return nil
}

// incrementalFlatDrop uses multiple database updates to remove key/value pairs
// saved to a flat index.
func incrementalFlatDrop(db database.DB, idxKey []byte, idxName string, interrupt <-chan struct{}) error {
	const maxDeletions = 2000000
	var totalDeleted uint64
	for numDeleted := maxDeletions; numDeleted == maxDeletions; {
		numDeleted = 0
		err := db.Update(func(dbTx database.Tx) error {
			bucket := dbTx.Metadata().Bucket(idxKey)
			cursor := bucket.Cursor()
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

		if interruptRequested(interrupt) {
			return errors.New("interrupt requested")
		}

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
func dropIndex(db database.DB, idxKey []byte, idxName string) error {
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

	// Remove the index tip, index bucket, and in-progress drop flag.  Removing
	// the index bucket also recursively removes all values saved to the index.
	err = dropIndexMetadata(db, idxKey, idxName)
	if err != nil {
		return err
	}

	log.Info(fmt.Sprintf("Dropped %s", idxName))
	return nil
}
