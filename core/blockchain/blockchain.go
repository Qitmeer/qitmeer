// Copyright (c) 2017-2018 The nox developers

package blockchain

import (
	"container/list"
	"encoding/binary"
	"fmt"
	"github.com/noxproject/nox/common/hash"
	"github.com/noxproject/nox/core/dbnamespace"
	"github.com/noxproject/nox/core/types"
	"github.com/noxproject/nox/database"
	"github.com/noxproject/nox/engine/txscript"
	"github.com/noxproject/nox/params"
	"github.com/noxproject/nox/services/common/progresslog"
	"os"
	"sync"
	"time"
)

const (

	// maxOrphanBlocks is the maximum number of orphan blocks that can be
	// queued.
	maxOrphanBlocks = 500

	// minMemoryNodes is the minimum number of consecutive nodes needed
	// in memory in order to perform all necessary validation.  It is used
	// to determine when it's safe to prune nodes from memory without
	// causing constant dynamic reloading.  This value should be larger than
	// that for minMemoryStakeNodes.
	minMemoryNodes = 2880

	// mainchainBlockCacheSize is the number of mainchain blocks to
	// keep in memory, by height from the tip of the mainchain.
	mainchainBlockCacheSize = 12
)

// BlockChain provides functions such as rejecting duplicate blocks, ensuring
// blocks follow all rules, orphan handling, checkpoint handling, and best chain
// selection with reorganization.
type BlockChain struct {

	params         		*params.Params

	// The following fields are set when the instance is created and can't
	// be changed afterwards, so there is no need to protect them with a
	// separate mutex.
	checkpointsByHeight map[uint64]*params.Checkpoint

	db                  database.DB
	dbInfo              *databaseInfo
	timeSource          MedianTimeSource
	notifications       NotificationCallback
	sigCache            *txscript.SigCache
	indexManager        IndexManager


	// subsidyCache is the cache that provides quick lookup of subsidy
	// values.
	subsidyCache *SubsidyCache

	// chainLock protects concurrent access to the vast majority of the
	// fields in this struct below this point.
	chainLock sync.RWMutex

	// These fields are configuration parameters that can be toggled at
	// runtime.  They are protected by the chain lock.
	noVerify      bool
	noCheckpoints bool

	// These fields are related to the memory block index.  They are
	// protected by the chain lock.
	bestNode *blockNode
	index    *blockIndex

	// This field allows efficient lookup of nodes in the main chain by
	// height.  It is protected by the height lock.
	heightLock        sync.RWMutex
	mainNodesByHeight map[uint64]*blockNode

	// These fields are related to handling of orphan blocks.  They are
	// protected by a combination of the chain lock and the orphan lock.
	orphanLock   sync.RWMutex
	orphans      map[hash.Hash]*orphanBlock
	prevOrphans  map[hash.Hash][]*orphanBlock
	oldestOrphan *orphanBlock

	// The block cache for mainchain blocks, to facilitate faster
	// reorganizations.
	mainchainBlockCacheLock sync.RWMutex
	mainchainBlockCache     map[hash.Hash]*types.SerializedBlock
	mainchainBlockCacheSize int

	// These fields are related to checkpoint handling.  They are protected
	// by the chain lock.
	nextCheckpoint *params.Checkpoint
	checkpointNode *blockNode

	// The state is used as a fairly efficient way to cache information
	// about the current best chain state that is returned to callers when
	// requested.  It operates on the principle of MVCC such that any time a
	// new block becomes the best block, the state pointer is replaced with
	// a new struct and the old state is left untouched.  In this way,
	// multiple callers can be pointing to different best chain states.
	// This is acceptable for most callers because the state is only being
	// queried at a specific point in time.
	//
	// In addition, some of the fields are stored in the database so the
	// chain state can be quickly reconstructed on load.
	stateLock     sync.RWMutex
	stateSnapshot *BestState

	// pruner is the automatic pruner for block nodes and stake nodes,
	// so that the memory may be restored by the garbage collector if
	// it is unlikely to be referenced in the future.
	pruner *chainPruner

}

// Config is a descriptor which specifies the blockchain instance configuration.
type Config struct {
	// DB defines the database which houses the blocks and will be used to
	// store all metadata created by this package such as the utxo set.
	//
	// This field is required.
	DB database.DB

	// Interrupt specifies a channel the caller can close to signal that
	// long running operations, such as catching up indexes or performing
	// database migrations, should be interrupted.
	//
	// This field can be nil if the caller does not desire the behavior.
	Interrupt <-chan struct{}

	// ChainParams identifies which chain parameters the chain is associated
	// with.
	//
	// This field is required.
	ChainParams *params.Params

	// TimeSource defines the median time source to use for things such as
	// block processing and determining whether or not the chain is current.
	//
	// The caller is expected to keep a reference to the time source as well
	// and add time samples from other peers on the network so the local
	// time is adjusted to be in agreement with other peers.
	TimeSource MedianTimeSource

	// Notifications defines a callback to which notifications will be sent
	// when various events take place.  See the documentation for
	// Notification and NotificationType for details on the types and
	// contents of notifications.
	//
	// This field can be nil if the caller is not interested in receiving
	// notifications.
	Notifications NotificationCallback

	// SigCache defines a signature cache to use when when validating
	// signatures.  This is typically most useful when individual
	// transactions are already being validated prior to their inclusion in
	// a block such as what is usually done via a transaction memory pool.
	//
	// This field can be nil if the caller is not interested in using a
	// signature cache.
	SigCache *txscript.SigCache

	// IndexManager defines an index manager to use when initializing the
	// chain and connecting and disconnecting blocks.
	//
	// This field can be nil if the caller does not wish to make use of an
	// index manager.
	IndexManager IndexManager
}

// orphanBlock represents a block that we don't yet have the parent for.  It
// is a normal block plus an expiration time to prevent caching the orphan
// forever.
type orphanBlock struct {
	block      *types.SerializedBlock
	expiration time.Time
}

// BestState houses information about the current best block and other info
// related to the state of the main chain as it exists from the point of view of
// the current best block.
//
// The BestSnapshot method can be used to obtain access to this information
// in a concurrent safe manner and the data will not be changed out from under
// the caller when chain state changes occur as the function name implies.
// However, the returned snapshot must be treated as immutable since it is
// shared by all callers.
type BestState struct {
	Hash         hash.Hash      // The hash of the block.
	Height       uint64          // The height of the block.
	Bits         uint32         // The difficulty bits of the block.
	BlockSize    uint64         // The size of the block.
	NumTxns      uint64         // The number of txns in the block.
	MedianTime   time.Time      // Median time as per CalcPastMedianTime.
	TotalTxns    uint64         // The total number of txns in the chain.
	TotalSubsidy int64          // The total subsidy for the chain.
}

// newBestState returns a new best stats instance for the given parameters.
func newBestState(node *blockNode, blockSize, numTxns uint64, medianTime time.Time, totalTxns uint64, totalSubsidy int64) *BestState {
	return &BestState{
		Hash:         node.hash,
		Height:       node.height,
		Bits:         node.bits,
		BlockSize:    blockSize,
		NumTxns:      numTxns,
		MedianTime:   medianTime,
		TotalTxns:    totalTxns,
		TotalSubsidy: totalSubsidy,
	}
}

// BestSnapshot returns information about the current best chain block and
// related state as of the current point in time.  The returned instance must be
// treated as immutable since it is shared by all callers.
//
// This function is safe for concurrent access.
func (b *BlockChain) BestSnapshot() *BestState {
	b.stateLock.RLock()
	snapshot := b.stateSnapshot
	b.stateLock.RUnlock()
	return snapshot
}

// New returns a BlockChain instance using the provided configuration details.
func New(config *Config) (*BlockChain, error) {
	// Enforce required config fields.
	if config.DB == nil {
		return nil, AssertError("blockchain.New database is nil")
	}
	if config.ChainParams == nil {
		return nil, AssertError("blockchain.New chain parameters nil")
	}

	// Generate a checkpoint by height map from the provided checkpoints.
	par := config.ChainParams
	var checkpointsByHeight map[uint64]*params.Checkpoint
	if len(par.Checkpoints) > 0 {
		checkpointsByHeight = make(map[uint64]*params.Checkpoint)
		for i := range par.Checkpoints {
			checkpoint := &par.Checkpoints[i]
			checkpointsByHeight[checkpoint.Height] = checkpoint
		}
	}

	b := BlockChain{
		checkpointsByHeight:           checkpointsByHeight,
		db:                            config.DB,
		params:                        par,
		timeSource:                    config.TimeSource,
		notifications:                 config.Notifications,
		sigCache:                      config.SigCache,
		indexManager:                  config.IndexManager,
		index:                         newBlockIndex(config.DB,par),
		mainNodesByHeight:             make(map[uint64]*blockNode),
		orphans:                       make(map[hash.Hash]*orphanBlock),
		prevOrphans:                   make(map[hash.Hash][]*orphanBlock),
		mainchainBlockCache:           make(map[hash.Hash]*types.SerializedBlock),
		mainchainBlockCacheSize:       mainchainBlockCacheSize,
	}

	// Initialize the chain state from the passed database.  When the db
	// does not yet contain any chain state, both it and the chain state
	// will be initialized to contain only the genesis block.
	if err := b.initChainState(config.Interrupt); err != nil {
		return nil, err
	}

	// Initialize and catch up all of the currently active optional indexes
	// as needed.
	if config.IndexManager != nil {
		err := config.IndexManager.Init(&b, config.Interrupt)
		if err != nil {
			return nil, err
		}
	}

	b.pruner = newChainPruner(&b)

	b.subsidyCache = NewSubsidyCache(int64(b.bestNode.height), b.params)

	log.Info("Blockchain database version","chain", b.dbInfo.version,"compression", b.dbInfo.compVer,
		"index",b.dbInfo.bidxVer)

	log.Info("Chain state", "height",  b.bestNode.height,
		"hash",b.bestNode.hash,"tx_num", b.stateSnapshot.NumTxns)

	return &b, nil
}


// initChainState attempts to load and initialize the chain state from the
// database.  When the db does not yet contain any chain state, both it and the
// chain state are initialized to the genesis block.
func (b *BlockChain) initChainState(interrupt <-chan struct{}) error {
	// Update database versioning scheme if needed.
	err := b.db.Update(func(dbTx database.Tx) error {
		// No versioning upgrade is needed if the dbinfo bucket does not
		// exist or the legacy key does not exist.
		bucket := dbTx.Metadata().Bucket(dbnamespace.BCDBInfoBucketName)
		if bucket == nil {
			return nil
		}
		legacyBytes := bucket.Get(dbnamespace.BCDBInfoBucketName)
		if legacyBytes == nil {
			return nil
		}

		// No versioning upgrade is needed if the new version key exists.
		if bucket.Get(dbnamespace.BCDBInfoVersionKeyName) != nil {
			return nil
		}

		// Load and deserialize the legacy version information.
		log.Info("Migrating versioning scheme...")
		// TODO legacy support
		/*
		dbi, err := deserializeDatabaseInfoV2(legacyBytes)
		if err != nil {
			return err
		}

		// Store the database version info using the new format.
		if err := dbPutDatabaseInfo(dbTx, dbi); err != nil {
			return err
		}
		*/

		// Remove the legacy version information.
		return bucket.Delete(dbnamespace.BCDBInfoBucketName)
	})
	if err != nil {
		return err
	}

	// Determine the state of the database.
	var isStateInitialized bool
	err = b.db.View(func(dbTx database.Tx) error {
		// Fetch the database versioning information.
		dbInfo, err := dbFetchDatabaseInfo(dbTx)
		if err != nil {
			return err
		}

		// The database bucket for the versioning information is missing.
		if dbInfo == nil && err == nil {
			return nil
		}

		// Don't allow downgrades of the blockchain database.
		if dbInfo.version > currentDatabaseVersion {
			return fmt.Errorf("the current blockchain database is "+
				"no longer compatible with this version of "+
				"the software (%d > %d)", dbInfo.version,
				currentDatabaseVersion)
		}

		// Don't allow downgrades of the database compression version.
		if dbInfo.compVer > currentCompressionVersion {
			return fmt.Errorf("the current database compression "+
				"version is no longer compatible with this "+
				"version of the software (%d > %d)",
				dbInfo.compVer, currentCompressionVersion)
		}

		// Don't allow downgrades of the block index.
		if dbInfo.bidxVer > currentBlockIndexVersion {
			return fmt.Errorf("the current database block index "+
				"version is no longer compatible with this "+
				"version of the software (%d > %d)",
				dbInfo.bidxVer, currentBlockIndexVersion)
		}

		b.dbInfo = dbInfo
		isStateInitialized = true
		return nil
	})
	if err != nil {
		return err
	}

	// Initialize the database if it has not already been done.
	if !isStateInitialized {
		if err := b.createChainState(); err != nil {
			return err
		}
	}

	//  TODO: Upgrade the database as needed.
	/*
	err = upgradeDB(b.db, b.chainParams, b.dbInfo, interrupt)
	if err != nil {
		return err
	}
	*/

	// Attempt to load the chain state from the database.
	err = b.db.View(func(dbTx database.Tx) error {
		// Fetch the stored chain state from the database metadata.
		// When it doesn't exist, it means the database hasn't been
		// initialized for use with chain yet, so break out now to allow
		// that to happen under a writable database transaction.
		meta := dbTx.Metadata()
		serializedData := meta.Get(dbnamespace.ChainStateKeyName)
		if serializedData == nil {
			return nil
		}
		log.Trace("Serialized chain state: ","serializedData", fmt.Sprintf("%x",serializedData))
		state, err := deserializeBestChainState(serializedData)
		if err != nil {
			return err
		}

		log.Info("Loading block index...")
		bidxStart := time.Now()

		// Determine how many blocks will be loaded into the index in order to
		// allocate the right amount as a single alloc versus a whole bunch of
		// littles ones to reduce pressure on the GC.
		blockIndexBucket := meta.Bucket(dbnamespace.BlockIndexBucketName)
		var blockCount int32
		cursor := blockIndexBucket.Cursor()
		for ok := cursor.First(); ok; ok = cursor.Next() {
			blockCount++
		}
		blockNodes := make([]blockNode, blockCount)

		// Load all of the block index entries and construct the block index
		// accordingly.
		//
		// NOTE: No locks are used on the block index here since this is
		// initialization code.
		var i int32
		var lastNode *blockNode
		cursor = blockIndexBucket.Cursor()
		for ok := cursor.First(); ok; ok = cursor.Next() {
			entry, err := deserializeBlockIndexEntry(cursor.Value())
			if err != nil {
				return err
			}
			header := &entry.header

			// Determine the parent block node.  Since the block headers are
			// iterated in order of height, there is a very good chance the
			// previous header processed is the parent.
			var parent *blockNode
			if lastNode == nil {
				blockHash := header.BlockHash()
				if blockHash != *b.params.GenesisHash {
					return AssertError(fmt.Sprintf("initChainState: expected "+
						"first entry in block index to be genesis block, "+
						"found %s", blockHash))
				}
			} else if header.ParentRoot == lastNode.hash {
				parent = lastNode
			} else {
				parent = b.index.lookupNode(&header.ParentRoot)
				if parent == nil {
					return AssertError(fmt.Sprintf("initChainState: could "+
						"not find parent for block %s", header.BlockHash()))
				}
			}

			// Initialize the block node, connect it, and add it to the block
			// index.
			node := &blockNodes[i]
			initBlockNode(node, header, parent)
			b.index.addNode(node)

			lastNode = node
			i++
		}

		// Set the best chain to the stored best state.
		tip := b.index.lookupNode(&state.hash)
		if tip == nil {
			return AssertError(fmt.Sprintf("initChainState: cannot find "+
				"chain tip %s in block index", state.hash))
		}
		b.bestNode = tip

		// Mark all of the nodes from the tip back to the genesis block
		// as part of the main chain and build the by height map.
		for n := tip; n != nil; n = n.parent {
			n.inMainChain = true
			b.mainNodesByHeight[n.height] = n
		}

		log.Debug("Block index loaded","loadTime", time.Since(bidxStart))

		// Load the best and parent blocks and cache them.
		utilBlock, err := dbFetchBlockByHash(dbTx, &tip.hash)
		if err != nil {
			return err
		}
		b.mainchainBlockCache[tip.hash] = utilBlock
		if tip.parent != nil {
			parentBlock, err := dbFetchBlockByHash(dbTx, &tip.parent.hash)
			if err != nil {
				return err
			}
			b.mainchainBlockCache[tip.parent.hash] = parentBlock
		}

		// Initialize the state related to the best block.
		block := utilBlock.Block()
		blockSize := uint64(block.SerializeSize())
		numTxns := uint64(len(block.Transactions))

		b.stateSnapshot = newBestState(tip, blockSize, numTxns,
			 tip.CalcPastMedianTime(),state.totalTxns, state.totalSubsidy)

		return nil
	})
	return err
}

// IsCurrent returns whether or not the chain believes it is current.  Several
// factors are used to guess, but the key factors that allow the chain to
// believe it is current are:
//  - Latest block height is after the latest checkpoint (if enabled)
//  - Latest block has a timestamp newer than 24 hours ago
//
// This function is safe for concurrent access.
func (b *BlockChain) IsCurrent() bool {
	b.chainLock.RLock()
	defer b.chainLock.RUnlock()

	return b.isCurrent()
}

// isCurrent returns whether or not the chain believes it is current.  Several
// factors are used to guess, but the key factors that allow the chain to
// believe it is current are:
//  - Latest block height is after the latest checkpoint (if enabled)
//  - Latest block has a timestamp newer than 24 hours ago
//
// This function MUST be called with the chain state lock held (for reads).
func (b *BlockChain) isCurrent() bool {
	// Not current if the latest main (best) chain height is before the
	// latest known good checkpoint (when checkpoints are enabled).
	checkpoint := b.latestCheckpoint()
	if checkpoint != nil && b.bestNode.height < checkpoint.Height {
		return false
	}

	// Not current if the latest best block has a timestamp before 24 hours
	// ago.
	//
	// The chain appears to be current if none of the checks reported
	// otherwise.
	minus24Hours := b.timeSource.AdjustedTime().Add(-24 * time.Hour).Unix()
	return b.bestNode.timestamp >= minus24Hours
}


// dumpBlockChain dumps a map of the blockchain blocks as serialized bytes.
func (b *BlockChain) DumpBlockChain(dumpFile string, params *params.Params, height uint64) error {
	log.Info("Writing the blockchain to disk as a flat file, " +
		"please wait...")

	progressLogger := progresslog.NewBlockProgressLogger("Written", log)

	file, err := os.Create(dumpFile)
	if err != nil {
		return err
	}
	defer file.Close()

	// Store the network ID in an array for later writing.
	var net [4]byte
	binary.LittleEndian.PutUint32(net[:], uint32(params.Net))

	// Write the blocks sequentially, excluding the genesis block.
	var sz [4]byte
	for i := uint64(1); i <= height; i++ {
		bl, err := b.BlockByHeight(i)
		if err != nil {
			return err
		}

		// Serialize the block for writing.
		blB, err := bl.Bytes()
		if err != nil {
			return err
		}

		// Write the network ID first.
		_, err = file.Write(net[:])
		if err != nil {
			return err
		}

		// Write the size of the block as a little endian uint32,
		// then write the block itself serialized.
		binary.LittleEndian.PutUint32(sz[:], uint32(len(blB)))
		_, err = file.Write(sz[:])
		if err != nil {
			return err
		}

		_, err = file.Write(blB)
		if err != nil {
			return err
		}

		progressLogger.LogBlockHeight(bl)
	}

	log.Info("Successfully dumped the blockchain (%v blocks) to %v.",
		height, dumpFile)

	return nil
}

// BestPrevHash returns the hash of the previous block of the block at HEAD.
//
// This function is safe for concurrent access.
func (b *BlockChain) BestPrevHash() hash.Hash {
	b.chainLock.Lock()
	defer b.chainLock.Unlock()

	var prevHash hash.Hash
	if b.bestNode.parent != nil {
		prevHash = b.bestNode.parent.hash
	}
	return prevHash
}

// BlockByHash returns the block from the main chain with the given hash.
//
// This function is safe for concurrent access.
func (b *BlockChain) BlockByHash(hash *hash.Hash) (*types.SerializedBlock, error) {
	b.chainLock.RLock()
	defer b.chainLock.RUnlock()

	return b.fetchMainChainBlockByHash(hash)
}

// HeaderByHash returns the block header identified by the given hash or an
// error if it doesn't exist.  Note that this will return headers from both the
// main chain and any side chains.
//
// This function is safe for concurrent access.
func (b *BlockChain) HeaderByHash(hash *hash.Hash) (types.BlockHeader, error) {
	node := b.index.LookupNode(hash)
	if node == nil {
		return types.BlockHeader{}, fmt.Errorf("block %s is not known", hash)
	}

	return node.Header(), nil
}

// FetchBlockByHash searches the internal chain block stores and the database
// in an attempt to find the requested block.
//
// This function differs from BlockByHash in that this one also returns blocks
// that are not part of the main chain (if they are known).
//
// This function is safe for concurrent access.
func (b *BlockChain) FetchBlockByHash(hash *hash.Hash) (*types.SerializedBlock, error) {
	return b.fetchBlockByHash(hash)
}

// fetchMainChainBlockByHash returns the block from the main chain with the
// given hash.  It first attempts to use cache and then falls back to loading it
// from the database.
//
// An error is returned if the block is either not found or not in the main
// chain.
//
// This function is safe for concurrent access.
func (b *BlockChain) fetchMainChainBlockByHash(hash *hash.Hash) (*types.SerializedBlock, error) {
	b.mainchainBlockCacheLock.RLock()
	block, ok := b.mainchainBlockCache[*hash]
	b.mainchainBlockCacheLock.RUnlock()
	if ok {
		return block, nil
	}

	// Load the block from the database.
	err := b.db.View(func(dbTx database.Tx) error {
		var err error
		block, err = dbFetchBlockByHash(dbTx, hash)
		return err
	})
	return block, err
}

// removeOrphanBlock removes the passed orphan block from the orphan pool and
// previous orphan index.
func (b *BlockChain) removeOrphanBlock(orphan *orphanBlock) {
	// Protect concurrent access.
	b.orphanLock.Lock()
	defer b.orphanLock.Unlock()

	// Remove the orphan block from the orphan pool.
	orphanHash := orphan.block.Hash()
	delete(b.orphans, *orphanHash)

	// Remove the reference from the previous orphan index too.  An indexing
	// for loop is intentionally used over a range here as range does not
	// reevaluate the slice on each iteration nor does it adjust the index
	// for the modified slice.
	prevHash := &orphan.block.Block().Header.ParentRoot
	orphans := b.prevOrphans[*prevHash]
	for i := 0; i < len(orphans); i++ {
		h := orphans[i].block.Hash()
		if h.IsEqual(orphanHash) {
			copy(orphans[i:], orphans[i+1:])
			orphans[len(orphans)-1] = nil
			orphans = orphans[:len(orphans)-1]
			i--
		}
	}
	b.prevOrphans[*prevHash] = orphans

	// Remove the map entry altogether if there are no longer any orphans
	// which depend on the parent hash.
	if len(b.prevOrphans[*prevHash]) == 0 {
		delete(b.prevOrphans, *prevHash)
	}
}

// addOrphanBlock adds the passed block (which is already determined to be
// an orphan prior calling this function) to the orphan pool.  It lazily cleans
// up any expired blocks so a separate cleanup poller doesn't need to be run.
// It also imposes a maximum limit on the number of outstanding orphan
// blocks and will remove the oldest received orphan block if the limit is
// exceeded.
func (b *BlockChain) addOrphanBlock(block *types.SerializedBlock) {
	// Remove expired orphan blocks.
	for _, oBlock := range b.orphans {
		if time.Now().After(oBlock.expiration) {
			b.removeOrphanBlock(oBlock)
			continue
		}

		// Update the oldest orphan block pointer so it can be discarded
		// in case the orphan pool fills up.
		if b.oldestOrphan == nil ||
			oBlock.expiration.Before(b.oldestOrphan.expiration) {
			b.oldestOrphan = oBlock
		}
	}

	// Limit orphan blocks to prevent memory exhaustion.
	if len(b.orphans)+1 > maxOrphanBlocks {
		// Remove the oldest orphan to make room for the new one.
		b.removeOrphanBlock(b.oldestOrphan)
		b.oldestOrphan = nil
	}

	// Protect concurrent access.  This is intentionally done here instead
	// of near the top since removeOrphanBlock does its own locking and
	// the range iterator is not invalidated by removing map entries.
	b.orphanLock.Lock()
	defer b.orphanLock.Unlock()

	// Insert the block into the orphan map with an expiration time
	// 1 hour from now.
	expiration := time.Now().Add(time.Hour)
	oBlock := &orphanBlock{
		block:      block,
		expiration: expiration,
	}
	b.orphans[*block.Hash()] = oBlock

	// Add to previous hash lookup index for faster dependency lookups.
	prevHash := &block.Block().Header.ParentRoot
	b.prevOrphans[*prevHash] = append(b.prevOrphans[*prevHash], oBlock)
}

// MaximumBlockSize returns the maximum permitted block size for the block
// AFTER the given node.
//
// This function MUST be called with the chain state lock held (for reads).
func (b *BlockChain) maxBlockSize(prevNode *blockNode) (int64, error) {

	maxSize := int64(b.params.MaximumBlockSizes[0])

	// The max block size is not changed in any other cases.
	return maxSize, nil
}

// fetchBlockByHash returns the block with the given hash from all known sources
// such as the internal caches and the database.
//
// This function is safe for concurrent access.
func (b *BlockChain) fetchBlockByHash(hash *hash.Hash) (*types.SerializedBlock, error) {
	// Check orphan cache.
	b.orphanLock.RLock()
	orphan, existsOrphans := b.orphans[*hash]
	b.orphanLock.RUnlock()
	if existsOrphans {
		return orphan.block, nil
	}

	// Check main chain cache.
	b.mainchainBlockCacheLock.RLock()
	block, ok := b.mainchainBlockCache[*hash]
	b.mainchainBlockCacheLock.RUnlock()
	if ok {
		return block, nil
	}

	// Attempt to load the block from the database.
	err := b.db.View(func(dbTx database.Tx) error {
		// NOTE: This does not use the dbFetchBlockByHash function since that
		// function only works with main chain blocks.
		blockBytes, err := dbTx.FetchBlock(hash)
		if err != nil {
			return err
		}

		block, err = types.NewBlockFromBytes(blockBytes)
		return err
	})
	if err == nil && block != nil {
		return block, nil
	}

	return nil, fmt.Errorf("unable to find block %v in cache or db", hash)
}

// TODO, refactor to more general method for panic handling
// panicf is a convenience function that formats according to the given format
// specifier and arguments and then logs the result at the critical level and
// panics with it.
func panicf(format string, args ...interface{}) {
	str := fmt.Sprintf(format, args...)
	log.Crit(str)
	panic(str)
}

// connectBestChain handles connecting the passed block to the chain while
// respecting proper chain selection according to the chain with the most
// proof of work.  In the typical case, the new block simply extends the main
// chain.  However, it may also be extending (or creating) a side chain (fork)
// which may or may not end up becoming the main chain depending on which fork
// cumulatively has the most proof of work.  It returns the resulting fork
// length, that is to say the number of blocks to the fork point from the main
// chain, which will be zero if the block ends up on the main chain (either
// due to extending the main chain or causing a reorganization to become the
// main chain).
//
// The flags modify the behavior of this function as follows:
//  - BFFastAdd: Avoids several expensive transaction validation operations.
//    This is useful when using checkpoints.
//
// This function MUST be called with the chain state lock held (for writes).
func (b *BlockChain) connectBestChain(node *blockNode, block, parent *types.SerializedBlock, flags BehaviorFlags) (int64, error) {
	fastAdd := flags&BFFastAdd == BFFastAdd

	// Ensure the passed parent is actually the parent of the block.
	if *parent.Hash() != node.parent.hash {
		panicf("parent block %v (height %v) does not match expected parent %v "+
			"(height %v)", parent.Hash(), parent.Block().Header.Height,
			node.parent.hash, node.height-1)
	}

	// We are extending the main (best) chain with a new block.  This is the
	// most common case.
	parentHash := &block.Block().Header.ParentRoot
	if *parentHash == b.bestNode.hash {
		// Skip expensive checks if the block has already been fully
		// validated.
		fastAdd = fastAdd || b.index.NodeStatus(node).KnownValid()

		// Perform several checks to verify the block can be connected
		// to the main chain without violating any rules and without
		// actually connecting the block.
		view := NewUtxoViewpoint()
		view.SetBestHash(parentHash)

		var stxos []spentTxOut
		if !fastAdd {
			err := b.checkConnectBlock(node, block, parent, view,
				&stxos)
			if err != nil {
				if _, ok := err.(RuleError); ok {
					b.index.SetStatusFlags(node, statusValidateFailed)
				}
				return 0, err
			}
			b.index.SetStatusFlags(node, statusValid)
		}
		// In the fast add case the code to check the block connection
		// was skipped, so the utxo view needs to load the referenced
		// utxos, spend them, and add the new utxos being created by
		// this block.
		if fastAdd {
			err := view.fetchInputUtxos(b.db, block)
			if err != nil {
				return 0, err
			}
			//TODO, connectTransactions also call fetchInputUtxos & refactor the inner logic
			err = b.connectTransactions(view, block, parent, &stxos)
			if err != nil {
				return 0, err
			}
		}

		// Connect the block to the main chain.
		err := b.connectBlock(node, block, parent, view, stxos)
		if err != nil {
			return 0, err
		}

		validateStr := "validating"

		// TODO, validating previous block
		log.Debug("Block connected to the main chain","hash",node.hash,"height",
			node.height, "operation",fmt.Sprintf( "%v the previous block",validateStr))

		// The fork length is zero since the block is now the tip of the
		// best chain.
		return 0, nil
	}
	if fastAdd {
		log.Warn("fastAdd set in the side chain case? %v\n",
			block.Hash())
	}

	// We're extending (or creating) a side chain which may or may not
	// become the main chain.
	node.inMainChain = false

	// We're extending (or creating) a side chain, but the cumulative
	// work for this new side chain is not enough to make it the new chain.
	if node.workSum.Cmp(b.bestNode.workSum) <= 0 {
		// Find the fork point.
		fork := node
		for ; fork.parent != nil; fork = fork.parent {
			if fork.inMainChain {
				break
			}
		}

		// Log information about how the block is forking the chain.
		if fork.hash == *parentHash {
			log.Info("FORK: Block %v (height %v) forks the chain at height "+
				"%d/block %v, but does not cause a reorganize",
				node.hash, node.height, fork.height, fork.hash)
		} else {
			log.Info("EXTEND FORK: Block %v (height %v) extends a side chain "+
				"which forks the chain at height %d/block %v",
				node.hash, node.height, fork.height, fork.hash)
		}

		forkLen := node.height - fork.height
		return int64(forkLen), nil //TODO, remove type conversion
	}

	// We're extending (or creating) a side chain and the cumulative work
	// for this new side chain is more than the old best chain, so this side
	// chain needs to become the main chain.  In order to accomplish that,
	// find the common ancestor of both sides of the fork, disconnect the
	// blocks that form the (now) old fork from the main chain, and attach
	// the blocks that form the new chain to the main chain starting at the
	// common ancenstor (the point where the chain forked).
	detachNodes, attachNodes := b.getReorganizeNodes(node)

	// Reorganize the chain.
	log.Info("REORGANIZE: Block %v is causing a reorganize.", node.hash)
	err := b.reorganizeChain(detachNodes, attachNodes)
	if err != nil {
		return 0, err
	}

	// The fork length is zero since the block is now the tip of the best
	// chain.
	return 0, nil
}

// connectBlock handles connecting the passed node/block to the end of the main
// (best) chain.
//
// This passed utxo view must have all referenced txos the block spends marked
// as spent and all of the new txos the block creates added to it.  In addition,
// the passed stxos slice must be populated with all of the information for the
// spent txos.  This approach is used because the connection validation that
// must happen prior to calling this function requires the same details, so
// it would be inefficient to repeat it.
//
// This function MUST be called with the chain state lock held (for writes).
func (b *BlockChain) connectBlock(node *blockNode, block, parent *types.SerializedBlock, view *UtxoViewpoint, stxos []spentTxOut) error {
	// Make sure it's extending the end of the best chain.
	prevHash := block.Block().Header.ParentRoot
	if prevHash != b.bestNode.hash {
		panicf("block %v (height %v) connects to block %v instead of "+
			"extending the best chain (hash %v, height %v)", node.hash,
			node.height, prevHash, b.bestNode.hash, b.bestNode.height)
	}


	// Generate a new best state snapshot that will be used to update the
	// database and later memory if all database updates are successful.
	b.stateLock.RLock()
	curTotalTxns := b.stateSnapshot.TotalTxns
	curTotalSubsidy := b.stateSnapshot.TotalSubsidy
	b.stateLock.RUnlock()

	// Calculate the number of transactions that would be added by adding
	// this block.
	numTxns := uint64(len(block.Block().Transactions))

	// Calculate the exact subsidy produced by adding the block.
	subsidy := CalculateAddedSubsidy(block, parent)

	/* TODO, revisit block size in block header
	blockSize := uint64(block.Block().Header.Size)
	*/
	blockSize := uint64(block.Block().SerializeSize())

	state := newBestState(node, uint64(blockSize), uint64(numTxns), node.CalcPastMedianTime(),curTotalTxns+numTxns,
		 curTotalSubsidy+subsidy)


	// Atomically insert info into the database.
	err := b.db.Update(func(dbTx database.Tx) error {
		// Update best block state.
		err := dbPutBestState(dbTx, state, node.workSum)
		if err != nil {
			return err
		}

		// Add the block to the block index.  Ultimately the block index
		// should track modified nodes and persist all of them prior
		// this point as opposed to unconditionally peristing the node
		// again.  However, this is needed for now in lieu of that to
		// ensure the updated status is written to the database.
		err = dbPutBlockNode(dbTx, node)
		if err != nil {
			return err
		}

		// Add the block hash and height to the main chain index.
		err = dbPutMainChainIndex(dbTx, block.Hash(), node.height)
		if err != nil {
			return err
		}

		// Update the utxo set using the state of the utxo view.  This
		// entails removing all of the utxos spent and adding the new
		// ones created by the block.
		err = dbPutUtxoView(dbTx, view)
		if err != nil {
			return err
		}

		// Update the transaction spend journal by adding a record for
		// the block that contains all txos spent by it.
		err = dbPutSpendJournalEntry(dbTx, block.Hash(), stxos)
		if err != nil {
			return err
		}

		// Allow the index manager to call each of the currently active
		// optional indexes with the block being connected so they can
		// update themselves accordingly.
		if b.indexManager != nil {
			err := b.indexManager.ConnectBlock(dbTx, block, parent, view)
			if err != nil {
				return err
			}
		}

		return nil
	})
	if err != nil {
		return err
	}

	// Prune fully spent entries and mark all entries in the view unmodified
	// now that the modifications have been committed to the database.
	view.commit()

	// Mark block as being in the main chain.
	node.inMainChain = true
	b.heightLock.Lock()
	b.mainNodesByHeight[node.height] = node
	b.heightLock.Unlock()

	// This node is now the end of the best chain.
	b.bestNode = node

	// Update the state for the best block.  Notice how this replaces the
	// entire struct instead of updating the existing one.  This effectively
	// allows the old version to act as a snapshot which callers can use
	// freely without needing to hold a lock for the duration.  See the
	// comments on the state variable for more details.
	b.stateLock.Lock()
	b.stateSnapshot = state
	b.stateLock.Unlock()

	// Assemble the current block and the parent into a slice.
	blockAndParent := []*types.SerializedBlock{block, parent}

	// Notify the caller that the block was connected to the main chain.
	// The caller would typically want to react with actions such as
	// updating wallets.
	b.chainLock.Unlock()
	b.sendNotification(BlockConnected, blockAndParent)
	b.chainLock.Lock()


	b.pushMainChainBlockCache(block)

	return nil
}

// FetchSubsidyCache returns the current subsidy cache from the blockchain.
//
// This function is safe for concurrent access.
func (b *BlockChain) FetchSubsidyCache() *SubsidyCache {
	return b.subsidyCache
}

// getReorganizeNodes finds the fork point between the main chain and the passed
// node and returns a list of block nodes that would need to be detached from
// the main chain and a list of block nodes that would need to be attached to
// the fork point (which will be the end of the main chain after detaching the
// returned list of block nodes) in order to reorganize the chain such that the
// passed node is the new end of the main chain.  The lists will be empty if the
// passed node is not on a side chain.
//
// This function MUST be called with the chain state lock held (for reads).
func (b *BlockChain) getReorganizeNodes(node *blockNode) (*list.List, *list.List) {
	// Nothing to detach or attach if there is no node.
	attachNodes := list.New()
	detachNodes := list.New()
	if node == nil {
		return detachNodes, attachNodes
	}

	// Don't allow a reorganize to a descendant of a known invalid block.
	if b.index.NodeStatus(node.parent).KnownInvalid() {
		b.index.SetStatusFlags(node, statusInvalidAncestor)
		return detachNodes, attachNodes
	}

	// Find the fork point (if any) adding each block to the list of nodes
	// to attach to the main tree.  Push them onto the list in reverse order
	// so they are attached in the appropriate order when iterating the list
	// later.
	ancestor := node
	for ; ancestor.parent != nil; ancestor = ancestor.parent {
		if ancestor.inMainChain {
			break
		}
		attachNodes.PushFront(ancestor)
	}

	// Start from the end of the main chain and work backwards until the
	// common ancestor adding each block to the list of nodes to detach from
	// the main chain.
	for n := b.bestNode; n != nil; n = n.parent {
		if n.hash == ancestor.hash {
			break
		}
		detachNodes.PushBack(n)
	}

	return detachNodes, attachNodes
}

// reorganizeChain reorganizes the block chain by disconnecting the nodes in the
// detachNodes list and connecting the nodes in the attach list.  It expects
// that the lists are already in the correct order and are in sync with the
// end of the current best chain.  Specifically, nodes that are being
// disconnected must be in reverse order (think of popping them off the end of
// the chain) and nodes the are being attached must be in forwards order
// (think pushing them onto the end of the chain).
//
// This function MUST be called with the chain state lock held (for writes).
func (b *BlockChain) reorganizeChain(detachNodes, attachNodes *list.List) error {
	// Nothing to do if no reorganize nodes were provided.
	if detachNodes.Len() == 0 && attachNodes.Len() == 0 {
		return nil
	}

	// Ensure the provided nodes match the current best chain.
	if detachNodes.Len() != 0 {
		firstDetachNode := detachNodes.Front().Value.(*blockNode)
		if firstDetachNode.hash != b.bestNode.hash {
			panicf("reorganize nodes to detach are not for the current best "+
				"chain -- first detach node %v, current chain %v",
				&firstDetachNode.hash, &b.bestNode.hash)
		}
	}

	// Ensure the provided nodes are for the same fork point.
	if attachNodes.Len() != 0 && detachNodes.Len() != 0 {
		firstAttachNode := attachNodes.Front().Value.(*blockNode)
		lastDetachNode := detachNodes.Back().Value.(*blockNode)
		if firstAttachNode.parent.hash != lastDetachNode.parent.hash {
			panicf("reorganize nodes do not have the same fork point -- first "+
				"attach parent %v, last detach parent %v",
				&firstAttachNode.parent.hash, &lastDetachNode.parent.hash)
		}
	}

	// Track the old and new best chains heads.
	oldBest := b.bestNode
	newBest := b.bestNode

	// All of the blocks to detach and related spend journal entries needed
	// to unspend transaction outputs in the blocks being disconnected must
	// be loaded from the database during the reorg check phase below and
	// then they are needed again when doing the actual database updates.
	// Rather than doing two loads, cache the loaded data into these slices.
	detachBlocks := make([]*types.SerializedBlock, 0, detachNodes.Len())
	detachSpentTxOuts := make([][]spentTxOut, 0, detachNodes.Len())
	attachBlocks := make([]*types.SerializedBlock, 0, attachNodes.Len())

	// Disconnect all of the blocks back to the point of the fork.  This
	// entails loading the blocks and their associated spent txos from the
	// database and using that information to unspend all of the spent txos
	// and remove the utxos created by the blocks.
	view := NewUtxoViewpoint()
	view.SetBestHash(&oldBest.hash)
	var nextBlockToDetach *types.SerializedBlock
	for e := detachNodes.Front(); e != nil; e = e.Next() {
		// Grab the block to detach based on the node.  Use the fact that the
		// blocks are being detached in reverse order, so the parent of the
		// current block being detached is the next one being detached.
		n := e.Value.(*blockNode)
		block := nextBlockToDetach
		if block == nil {
			var err error
			block, err = b.fetchMainChainBlockByHash(&n.hash)
			if err != nil {
				return err
			}
		}
		if n.hash != *block.Hash() {
			panicf("detach block node hash %v (height %v) does not match "+
				"previous parent block hash %v", &n.hash, n.height,
				block.Hash())
		}

		// Grab the parent of the current block and also save a reference to it
		// as the next block to detach so it doesn't need to be loaded again on
		// the next iteration.
		parent, err := b.fetchMainChainBlockByHash(&n.parent.hash)
		if err != nil {
			return err
		}
		nextBlockToDetach = parent

		// Load all of the spent txos for the block from the spend
		// journal.
		var stxos []spentTxOut
		err = b.db.View(func(dbTx database.Tx) error {
			stxos, err = dbFetchSpendJournalEntry(dbTx, block, parent)
			return err
		})
		if err != nil {
			return err
		}

		// Quick sanity test.
		// TODO, revisit the stxos count
		if len(stxos) != countSpentOutputs(block, parent) {
			panicf("retrieved %v stxos when trying to disconnect block %v "+
				"(height %v), yet counted %v many spent utxos", len(stxos),
				block.Hash(), block.Height(), countSpentOutputs(block, parent))
		}

		// Store the loaded block and spend journal entry for later.
		detachBlocks = append(detachBlocks, block)
		detachSpentTxOuts = append(detachSpentTxOuts, stxos)

		err = b.disconnectTransactions(view, block, parent, stxos)
		if err != nil {
			return err
		}

		newBest = n.parent
	}

	// Set the fork point and grab the fork block when there are nodes to be
	// attached.  The fork block is used as the parent to the first node to be
	// attached below.
	var forkNode *blockNode
	var forkBlock *types.SerializedBlock
	if attachNodes.Len() > 0 {
		forkNode = newBest

		var err error
		forkBlock, err = b.fetchMainChainBlockByHash(&forkNode.hash)
		if err != nil {
			return err
		}
	}

	// Perform several checks to verify each block that needs to be attached
	// to the main chain can be connected without violating any rules and
	// without actually connecting the block.
	//
	// NOTE: These checks could be done directly when connecting a block,
	// however the downside to that approach is that if any of these checks
	// fail after disconnecting some blocks or attaching others, all of the
	// operations have to be rolled back to get the chain back into the
	// state it was before the rule violation (or other failure).  There are
	// at least a couple of ways accomplish that rollback, but both involve
	// tweaking the chain and/or database.  This approach catches these
	// issues before ever modifying the chain.
	for i, e := 0, attachNodes.Front(); e != nil; i, e = i+1, e.Next() {
		// Grab the block to attach based on the node.  Use the fact that the
		// parent of the block is either the fork point for the first node being
		// attached or the previous one that was attached for subsequent blocks
		// to optimize.
		n := e.Value.(*blockNode)
		block, err := b.fetchBlockByHash(&n.hash)
		if err != nil {
			return err
		}
		parent := forkBlock
		if i > 0 {
			parent = attachBlocks[i-1]
		}
		if n.parent.hash != *parent.Hash() {
			panicf("attach block node hash %v (height %v) parent hash %v does "+
				"not match previous parent block hash %v", &n.hash, n.height,
				&n.parent.hash, parent.Hash())
		}

		// Store the loaded block for later.
		attachBlocks = append(attachBlocks, block)

		// Notice the spent txout details are not requested here and
		// thus will not be generated.  This is done because the state
		// is not being immediately written to the database, so it is
		// not needed.
		err = b.checkConnectBlock(n, block, parent, view, nil)
		if err != nil {
			return err
		}

		newBest = n
	}
	log.Debug("New best chain validation completed successfully, " +
		"commencing with the reorganization.")

	// Send a notification that a blockchain reorganization is in progress.
	reorgData := &ReorganizationNotifyData{
		oldBest.hash,
		oldBest.height,
		newBest.hash,
		newBest.height,
	}
	b.chainLock.Unlock()
	b.sendNotification(Reorganization, reorgData)
	b.chainLock.Lock()

	// Reset the view for the actual connection code below.  This is
	// required because the view was previously modified when checking if
	// the reorg would be successful and the connection code requires the
	// view to be valid from the viewpoint of each block being connected or
	// disconnected.
	view = NewUtxoViewpoint()
	view.SetBestHash(&oldBest.hash)

	// Disconnect blocks from the main chain.
	for i, e := 0, detachNodes.Front(); e != nil; i, e = i+1, e.Next() {
		// Since the blocks are being detached in reverse order, the parent of
		// current block being detached is the next one being detached up to
		// the final one at which point it's the block that is already saved
		// from the next block to detach above.
		n := e.Value.(*blockNode)
		block := detachBlocks[i]
		parent := nextBlockToDetach
		if i < len(detachBlocks)-1 {
			parent = detachBlocks[i+1]
		}
		if n.parent.hash != *parent.Hash() {
			panicf("detach block node hash %v (height %v) parent hash %v does "+
				"not match previous parent block hash %v", &n.hash, n.height,
				&n.parent.hash, parent.Hash())
		}

		// Load all of the utxos referenced by the block that aren't
		// already in the view.
		err := view.fetchInputUtxos(b.db, block)
		if err != nil {
			return err
		}

		// Update the view to unspend all of the spent txos and remove
		// the utxos created by the block.
		err = b.disconnectTransactions(view, block, parent,
			detachSpentTxOuts[i])
		if err != nil {
			return err
		}

		// Update the database and chain state.
		err = b.disconnectBlock(n, block, parent, view)
		if err != nil {
			return err
		}
	}

	// Connect the new best chain blocks.
	for i, e := 0, attachNodes.Front(); e != nil; i, e = i+1, e.Next() {
		// Grab the block to attach based on the node.  Use the fact that the
		// parent of the block is either the fork point for the first node being
		// attached or the previous one that was attached for subsequent blocks
		// to optimize.
		n := e.Value.(*blockNode)
		block := attachBlocks[i]
		parent := forkBlock
		if i > 0 {
			parent = attachBlocks[i-1]
		}
		if n.parent.hash != *parent.Hash() {
			panicf("attach block node hash %v (height %v) parent hash %v does "+
				"not match previous parent block hash %v", &n.hash, n.height,
				&n.parent.hash, parent.Hash())
		}

		// Update the view to mark all utxos referenced by the block
		// as spent and add all transactions being created by this block
		// to it.  Also, provide an stxo slice so the spent txout
		// details are generated.
		// TODO, revisit the design of stxos count
		stxos := make([]spentTxOut, 0, countSpentOutputs(block, parent))
		err := b.connectTransactions(view, block, parent, &stxos)
		if err != nil {
			return err
		}

		// Update the database and chain state.
		err = b.connectBlock(n, block, parent, view, stxos)
		if err != nil {
			return err
		}
	}

	// Log the point where the chain forked and old and new best chain
	// heads.
	if forkNode != nil {
		log.Info("REORGANIZE: Chain forks at %v (height %v)",
			forkNode.hash, forkNode.height)
	}
	log.Info("REORGANIZE: Old best chain head was %v (height %v)",
		&oldBest.hash, oldBest.height)
	log.Info("REORGANIZE: New best chain head is %v (height %v)",
		newBest.hash, newBest.height)

	return nil
}

// countSpentOutputs returns the number of utxos the passed block spends.
// TODO, revisit the design of stxos count
func countSpentOutputs(block,parent *types.SerializedBlock) int {
	// Exclude the coinbase transaction since it can't spend anything.
	var numSpent int
	for _, tx := range block.Transactions()[1:] {
		numSpent += len(tx.Transaction().TxIn)
	}
	return numSpent
}

// disconnectBlock handles disconnecting the passed node/block from the end of
// the main (best) chain.
//
// This function MUST be called with the chain state lock held (for writes).
func (b *BlockChain) disconnectBlock(node *blockNode, block, parent *types.SerializedBlock, view *UtxoViewpoint) error {
	// Make sure the node being disconnected is the end of the best chain.
	if node.hash != b.bestNode.hash {
		panicf("block %v (height %v) is not the end of the best chain "+
			"(hash %v, height %v)", node.hash, node.height, b.bestNode.hash,
			b.bestNode.height)
	}

	// Generate a new best state snapshot that will be used to update the
	// database and later memory if all database updates are successful.
	b.stateLock.RLock()
	curTotalTxns := b.stateSnapshot.TotalTxns
	curTotalSubsidy := b.stateSnapshot.TotalSubsidy
	b.stateLock.RUnlock()
	// revisit the size in block headers
	/*
	parentBlockSize := uint64(parent.Block().Header.Size)
	*/
	parentBlockSize := uint64(parent.Block().SerializeSize())

	// Calculate the number of transactions that would be added by adding
	// this block.

	// TODO revisit the tx count logic
	numTxns := uint64(len(parent.Block().Transactions))
	/*
	numTxns := countNumberOfTransactions(block, parent)
	*/
	newTotalTxns := curTotalTxns - numTxns

	// Calculate the exact subsidy produced by adding the block.
	subsidy := CalculateAddedSubsidy(block, parent)
	newTotalSubsidy := curTotalSubsidy - subsidy

	prevNode := node.parent
	state := newBestState(prevNode, parentBlockSize, numTxns,
		prevNode.CalcPastMedianTime(),  newTotalTxns, newTotalSubsidy)

	err := b.db.Update(func(dbTx database.Tx) error {
		// Update best block state.
		err := dbPutBestState(dbTx, state, node.workSum)
		if err != nil {
			return err
		}

		// Remove the block hash and height from the main chain index.
		err = dbRemoveMainChainIndex(dbTx, block.Hash(), int64(node.height))  //TODO, remove type conversion
		if err != nil {
			return err
		}

		// Update the utxo set using the state of the utxo view.  This
		// entails restoring all of the utxos spent and removing the new
		// ones created by the block.
		err = dbPutUtxoView(dbTx, view)
		if err != nil {
			return err
		}

		// Update the transaction spend journal by removing the record
		// that contains all txos spent by the block .
		err = dbRemoveSpendJournalEntry(dbTx, block.Hash())
		if err != nil {
			return err
		}

		// Allow the index manager to call each of the currently active
		// optional indexes with the block being disconnected so they
		// can update themselves accordingly.
		if b.indexManager != nil {
			err := b.indexManager.DisconnectBlock(dbTx, block, parent, view)
			if err != nil {
				return err
			}
		}

		return nil
	})
	if err != nil {
		return err
	}

	// Prune fully spent entries and mark all entries in the view unmodified
	// now that the modifications have been committed to the database.
	view.commit()

	// Mark block as being in a side chain.
	node.inMainChain = false
	b.heightLock.Lock()
	delete(b.mainNodesByHeight, node.height)
	b.heightLock.Unlock()

	// This node's parent is now the end of the best chain.
	b.bestNode = node.parent

	// Update the state for the best block.  Notice how this replaces the
	// entire struct instead of updating the existing one.  This effectively
	// allows the old version to act as a snapshot which callers can use
	// freely without needing to hold a lock for the duration.  See the
	// comments on the state variable for more details.
	b.stateLock.Lock()
	b.stateSnapshot = state
	b.stateLock.Unlock()

	// Assemble the current block and the parent into a slice.
	blockAndParent := []*types.SerializedBlock{block, parent}

	// Notify the caller that the block was disconnected from the main
	// chain.  The caller would typically want to react with actions such as
	// updating wallets.
	b.chainLock.Unlock()
	b.sendNotification(BlockDisconnected, blockAndParent)
	b.chainLock.Lock()

	b.dropMainChainBlockCache(block)

	return nil
}

// pushMainChainBlockCache pushes a block onto the main chain block cache,
// and removes any old blocks from the cache that might be present.
// TODO, refactor the mainchainBlockCache
func (b *BlockChain) pushMainChainBlockCache(block *types.SerializedBlock) {
	curHeight := block.Height()
	curHash := block.Hash()
	b.mainchainBlockCacheLock.Lock()
	b.mainchainBlockCache[*curHash] = block
	for hash, bl := range b.mainchainBlockCache {
		if bl.Height() <= curHeight-uint64(b.mainchainBlockCacheSize) {  //TODO, remove type conversion
			delete(b.mainchainBlockCache, hash)
		}
	}
	b.mainchainBlockCacheLock.Unlock()
}
// dropMainChainBlockCache drops a block from the main chain block cache.
// TODO, refactor the mainchainBlockCache
func (b *BlockChain) dropMainChainBlockCache(block *types.SerializedBlock) {
	curHash := block.Hash()
	b.mainchainBlockCacheLock.Lock()
	delete(b.mainchainBlockCache, *curHash)
	b.mainchainBlockCacheLock.Unlock()
}
