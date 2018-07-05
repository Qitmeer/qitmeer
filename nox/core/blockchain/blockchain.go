// Copyright (c) 2017-2018 The nox developers

package blockchain

import (
	"sync"
	"time"
	"github.com/noxproject/nox/common/hash"
	"github.com/noxproject/nox/core/types"
	"github.com/noxproject/nox/database"
	"github.com/noxproject/nox/params"
	"github.com/noxproject/nox/engine/txscript"
	"github.com/noxproject/nox/core/dbnamespace"
	"fmt"
	"encoding/binary"
	"github.com/noxproject/nox/services/common/progresslog"
	"os"
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
	block      *types.Block
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
}

// newBestState returns a new best stats instance for the given parameters.
func newBestState(node *blockNode, blockSize, numTxns uint64, medianTime time.Time) *BestState {
	return &BestState{
		Hash:         node.hash,
		Height:       node.height,
		Bits:         node.bits,
		BlockSize:    blockSize,
		NumTxns:      numTxns,
		MedianTime:   medianTime,
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

	//TODO chain Pruner
	// b.pruner = newChainPruner(&b)

	log.Info("Blockchain database version info: chain: %d, compression: "+
		"%d, block index: %d", b.dbInfo.version, b.dbInfo.compVer,
		b.dbInfo.bidxVer)

	log.Info("Chain state: height %d, hash %v, total transactions %d, "+
		"work %v, stake version %v", b.bestNode.height, b.bestNode.hash,
		b.stateSnapshot.NumTxns, b.bestNode.workSum,
		0)

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

	//  TODO: Uphrade the database as needed.
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
			 tip.CalcPastMedianTime(),
			)

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

// latestCheckpoint returns the most recent checkpoint (regardless of whether it
// is already known).  When checkpoints are disabled or there are no checkpoints
// for the active network, it will return nil.
//
// This function MUST be called with the chain state lock held (for reads).
func (b *BlockChain) latestCheckpoint() *params.Checkpoint {
	if b.noCheckpoints || len(b.params.Checkpoints) == 0 {
		return nil
	}

	checkpoints := b.params.Checkpoints
	return &checkpoints[len(checkpoints)-1]
}

// DisableCheckpoints provides a mechanism to disable validation against
// checkpoints which you DO NOT want to do in production.  It is provided only
// for debug purposes.
//
// This function is safe for concurrent access.
func (b *BlockChain) DisableCheckpoints(disable bool) {
	b.chainLock.Lock()
	b.noCheckpoints = disable
	b.chainLock.Unlock()
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

