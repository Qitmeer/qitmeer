// Copyright (c) 2017-2018 The nox developers

package blockchain

import (
	"container/list"
	"encoding/binary"
	"fmt"
	"qitmeer/common/hash"
	"qitmeer/core/dbnamespace"
	"qitmeer/core/types"
	"qitmeer/database"
	"qitmeer/engine/txscript"
	"qitmeer/params"
	"qitmeer/services/common/progresslog"
	"os"
	"sync"
	"time"
	"qitmeer/core/blockdag"
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

	// These fields are related to the memory block index.  They both have
	// their own locks, however they are often also protected by the chain
	// lock to help prevent logic races when blocks are being processed.
	//
	// index houses the entire block index in memory.  The block index is
	// a tree-shaped structure.
	index     *blockIndex

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

	//block dag
	bd *blockdag.BlockDAG
	//badTx hash->block hash
	badTx map[hash.Hash]*blockdag.HashSet
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

	// Setting different dag types will use different consensus
	DAGType string
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
	Order        uint64          // The order of the block.
	Bits         uint32         // The difficulty bits of the block.
	BlockSize    uint64         // The size of the block.
	NumTxns      uint64         // The number of txns in the block.
	MedianTime   time.Time      // Median time as per CalcPastMedianTime.
	TotalTxns    uint64         // The total number of txns in the chain.
	TotalSubsidy int64          // The total subsidy for the chain.
	GraphState   *blockdag.GraphState //The graph state of dag
}

// newBestState returns a new best stats instance for the given parameters.
func newBestState(node *blockNode, blockSize, numTxns uint64, medianTime time.Time, totalTxns uint64, totalSubsidy int64,gs *blockdag.GraphState) *BestState {
	return &BestState{
		Hash:         node.hash,
		Order:        node.order,
		Bits:         node.bits,
		BlockSize:    blockSize,
		NumTxns:      numTxns,
		MedianTime:   medianTime,
		TotalTxns:    totalTxns,
		TotalSubsidy: totalSubsidy,
		GraphState:   gs,
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
		orphans:                       make(map[hash.Hash]*orphanBlock),
		prevOrphans:                   make(map[hash.Hash][]*orphanBlock),
		mainchainBlockCache:           make(map[hash.Hash]*types.SerializedBlock),
		mainchainBlockCacheSize:       mainchainBlockCacheSize,
	}
	b.bd=&blockdag.BlockDAG{}
	b.bd.Init(config.DAGType)
	b.badTx=make(map[hash.Hash]*blockdag.HashSet)
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
	b.subsidyCache = NewSubsidyCache(int64(b.BestSnapshot().Order), b.params)

	log.Info(fmt.Sprintf("DAG Type:%s",b.bd.GetName()))
	log.Info("Blockchain database version","chain", b.dbInfo.version,"compression", b.dbInfo.compVer,
		"index",b.dbInfo.bidxVer)

	tips:=b.bd.GetTipsList()
	logStr:=fmt.Sprintf("Chain state:totaltx=%d\ntips=%d\n",b.stateSnapshot.TotalTxns,len(tips))

	for _,v:=range tips{
		tnode:=b.index.lookupNode(v.GetHash())
		logStr+=fmt.Sprintf("hash=%v,order=%d,pastSetNum=%d,work=%v\n",tnode.hash,tnode.order,v.GetWeight(),tnode.workSum)
	}
	log.Info(logStr)

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
		if dbInfo == nil {
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
		return b.createChainState()
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
		blocksM:=make(map[hash.Hash]*types.SerializedBlock)
		blockList:=list.New()


		cursor := blockIndexBucket.Cursor()
		for ok := cursor.First(); ok; ok = cursor.Next() {
			entry, err := deserializeBlockIndexEntry(cursor.Value())
			if err != nil {
				return err
			}
			header := &entry.header
			blockHash := header.BlockHash()
			_,exit:=blocksM[blockHash]
			if exit {
				continue
			}
			block, err := dbFetchBlockByHash(dbTx,&blockHash)
			if err != nil {
				return err
			}
			blocksM[blockHash]=block
			blockList.PushBack(block)

		}
		log.Trace(fmt.Sprintf("load %d blocks",blockList.Len()))

		for blockList.Len()>0 {
			var next *list.Element
			for e := blockList.Front(); e != nil; e = next {
				next = e.Next()
				//
				block:=e.Value.(*types.SerializedBlock)
				parents:=[]*blockNode{}
				needSkip:=false
				for _,pb:=range block.Block().Parents{
					parent:= b.index.LookupNode(pb)
					if parent==nil {
						needSkip=true
						break
					}
					parents=append(parents,parent)
				}
				if needSkip {
					continue
				}
				blockList.Remove(e)
				//
				node := &blockNode{}
				initBlockNode(node, &block.Block().Header, parents)
				list:=b.bd.AddBlock(node)
				dblock:=b.bd.GetBlock(node.GetHash())
				node.SetOrder(uint64(dblock.GetOrder()))
				b.index.addNode(node)
				if list==nil||list.Len()==0 {
					log.Error("Irreparable error!")
					return AssertError(fmt.Sprintf("initChainState: Could "+
						"not add %s",node.hash.String()))
				}
			}

		}
		log.Debug("Block index loaded","loadTime", time.Since(bidxStart))
		/*if !b.dag.GetLastBlock().hash.IsEqual(&state.hash) {
			return AssertError(fmt.Sprintf("initChainState:Data damage"))
		}*/
		// Set the best chain view to the stored best state.
		lastBlock:=b.bd.GetLastBlock()
		tip :=b.index.lookupNode(lastBlock.GetHash())
		if tip == nil {
			return AssertError(fmt.Sprintf("initChainState: cannot find "+
				"chain last %s in block index", state.hash))
		}

		// Load the raw block bytes for the best block.
		block, err := dbFetchBlockByHash(dbTx,&state.hash)
		if err != nil {
			return err
		}
		// Initialize the state related to the best block.
		blockSize := uint64(block.Block().SerializeSize())
		numTxns := uint64(len(block.Block().Transactions))
		b.stateSnapshot = newBestState(tip, blockSize,numTxns,
			tip.CalcPastMedianTime(),state.totalTxns,state.totalSubsidy,b.bd.GetGraphState())

		return nil
	})
	return err
}

// HaveBlock returns whether or not the chain instance has the block represented
// by the passed hash.  This includes checking the various places a block can
// be like part of the main chain, on a side chain, or in the orphan pool.
//
// This function is safe for concurrent access.
func (b *BlockChain) HaveBlock(hash *hash.Hash) (bool, error) {
	return b.index.HaveBlock(hash) || b.IsKnownOrphan(hash), nil
}

// IsKnownOrphan returns whether the passed hash is currently a known orphan.
// Keep in mind that only a limited number of orphans are held onto for a
// limited amount of time, so this function must not be used as an absolute
// way to test if a block is an orphan block.  A full block (as opposed to just
// its hash) must be passed to ProcessBlock for that purpose.  However, calling
// ProcessBlock with an orphan that already exists results in an error, so this
// function provides a mechanism for a caller to intelligently detect *recent*
// duplicate orphans and react accordingly.
//
// This function is safe for concurrent access.
func (b *BlockChain) IsKnownOrphan(hash *hash.Hash) bool {
	// Protect concurrent access.  Using a read lock only so multiple
	// readers can query without blocking each other.
	b.orphanLock.RLock()
	_, exists := b.orphans[*hash]
	b.orphanLock.RUnlock()

	return exists
}

// GetOrphansParents returns the parents for the provided hash from the
// map of orphan blocks.
func (b *BlockChain) GetOrphansParents() []*hash.Hash{
	b.orphanLock.RLock()
	defer b.orphanLock.RUnlock()
	//
	result:=blockdag.NewHashSet()
	for _,v:=range b.orphans{
		for _,h:=range v.block.Block().Parents{
			exists, err := b.HaveBlock(h)
			if err != nil||exists {
				continue
			}
			result.Add(h)
		}

	}
	return result.List()
}

// GetOrphansParents returns the parents for the provided hash from the
// map of orphan blocks.
func (b *BlockChain) GetOrphanParents(h *hash.Hash) []*hash.Hash{
	b.orphanLock.RLock()
	defer b.orphanLock.RUnlock()
	//
	ob, exists := b.orphans[*h]
	if !exists {
		return nil
	}
	result:=blockdag.NewHashSet()
	for _,h:=range ob.block.Block().Parents{
		exists, err := b.HaveBlock(h)
		if err != nil||exists {
			continue
		}
		result.Add(h)
	}

	return result.List()
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
	lastBlock:=b.bd.GetLastBlock()
	if checkpoint != nil && uint64(lastBlock.GetOrder()) < checkpoint.Height {
		return false
	}

	// Not current if the latest best block has a timestamp before 24 hours
	// ago.
	//
	// The chain appears to be current if none of the checks reported
	// otherwise.
	minus24Hours := b.timeSource.AdjustedTime().Add(-24 * time.Hour).Unix()
	lastNode:=b.index.lookupNode(lastBlock.GetHash())
	return lastNode.timestamp >= minus24Hours
}

// TipGeneration returns the entire generation of blocks stemming from the
// parent of the current tip.
//
// The function is safe for concurrent access.
func (b *BlockChain) TipGeneration() ([]hash.Hash, error) {
	b.chainLock.Lock()
	b.index.RLock()
	tips := b.bd.GetTipsList()
	tiphashs:=[]hash.Hash{}
	for _, block := range tips {
		tiphashs=append(tiphashs,*block.GetHash())
	}
	b.index.RUnlock()
	b.chainLock.Unlock()
	return tiphashs, nil
}

// dumpBlockChain dumps a map of the blockchain blocks as serialized bytes.
func (b *BlockChain) DumpBlockChain(dumpFile string, params *params.Params, order uint64) error {
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
	for i := uint64(1); i <= order; i++ {
		bl, err := b.BlockByOrder(i)
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
		order, dumpFile)

	return nil
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
		// Check if the block is in the main chain.
		if !dbMainChainHasBlock(dbTx, hash) {
			str := fmt.Sprintf("block %s is not in the main chain", hash)
			return errNotInMainChain(str)
		}

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
	// Load the block from the database.
	dbErr := b.db.View(func(dbTx database.Tx) error {
		var err error
		block, err= dbFetchBlockByHash(dbTx, hash)
		return err
	})
	if dbErr==nil&&block!=nil {
		return block,nil
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
func (b *BlockChain) connectDagChain(node *blockNode, block *types.SerializedBlock,newOrders *list.List) (bool, error) {
	if newOrders.Len()==0 {
		return false,nil
	}
	//Fast double spent check
	b.fastDoubleSpentCheck(node,block)

	// We are extending the main (best) chain with a new block.  This is the
	// most common case.
	if newOrders.Len()==1 {
		// Perform several checks to verify the block can be connected
		// to the main chain without violating any rules and without
		// actually connecting the block.
		view := NewUtxoViewpoint()
		view.SetBestHash(b.bd.GetPrevious(&node.hash))

		stxos:=[]spentTxOut{}
		err := b.checkConnectBlock(node, block, view,&stxos)
		if err != nil {
			b.RemoveBadTx(block.Hash())
			return false, err
		}
		// In the fast add case the code to check the block connection
		// was skipped, so the utxo view needs to load the referenced
		// utxos, spend them, and add the new utxos being created by
		// this block.

		// Connect the block to the main chain.
		err = b.connectBlock(node, block, view, stxos)
		if err != nil {
			b.RemoveBadTx(block.Hash())
			return false, err
		}

		validateStr := "validating"

		// TODO, validating previous block
		log.Debug("Block connected to the main chain","hash",node.hash,"order",
			node.order, "operation",fmt.Sprintf( "%v the previous block",validateStr))

		// The fork length is zero since the block is now the tip of the
		// best chain.
		return true, nil
	}


	// We're extending (or creating) a side chain and the cumulative work
	// for this new side chain is more than the old best chain, so this side
	// chain needs to become the main chain.  In order to accomplish that,
	// find the common ancestor of both sides of the fork, disconnect the
	// blocks that form the (now) old fork from the main chain, and attach
	// the blocks that form the new chain to the main chain starting at the
	// common ancenstor (the point where the chain forked).

	// Reorganize the chain.
	log.Info("DAG REORGANIZE: Block %v is causing a reorganize.", node.hash)
	oldOrder:=list.New()
	for e := newOrders.Front(); e != nil; e = e.Next() {
		log.Info(e.Value.(*hash.Hash).String())
		if e.Value.(*hash.Hash).IsEqual(&node.hash) {
			continue
		}
		oldOrder.PushBack(e.Value)
	}
	err := b.reorganizeChain(oldOrder, newOrders,block)
	if err!=nil {
		return false,err
	}
	return true, nil
}

// This function is fast check before global sequencing,it can judge who is the bad block quickly.
func (b *BlockChain) fastDoubleSpentCheck(node *blockNode,block *types.SerializedBlock) {
	/*transactions:=block.Transactions()
	if len(transactions)>1 {
		for i, tx := range transactions {
			if i==0 {
				continue
			}
			for _, txIn := range tx.Transaction().TxIn {
				entry,err:= b.fetchUtxoEntry(&txIn.PreviousOut.Hash)
				if entry == nil || err!=nil || !entry.IsOutputSpent(txIn.PreviousOut.OutIndex) {
					continue
				}
				preBlockH:=b.dag.GetBlockByOrder(uint(entry.height))
				if preBlockH==nil {
					continue
				}
				preBlock:=b.index.LookupNode(preBlockH)
				if preBlock==nil {
					continue
				}
				ret, err := b.dag.s.Vote(preBlock,node)
				if err!=nil {
					continue
				}
				if ret {
					b.AddBadTx(tx.Hash(),block.Hash())
				}
			}
		}
	}*/
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
func (b *BlockChain) connectBlock(node *blockNode, block *types.SerializedBlock, view *UtxoViewpoint, stxos []spentTxOut) error {
	// Must be end node of sequence in dag
	lastBlock:=b.bd.GetLastBlock()
	lastTip:=b.index.lookupNode(lastBlock.GetHash())
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
	subsidy := CalculateAddedSubsidy(block)

	/* TODO, revisit block size in block header
	blockSize := uint64(block.Block().Header.Size)
	*/
	blockSize := uint64(block.Block().SerializeSize())

	state := newBestState(lastTip, uint64(blockSize), uint64(numTxns),*b.bd.GetLastTime(),curTotalTxns+numTxns,
		 curTotalSubsidy+subsidy,b.bd.GetGraphState())


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
		err = dbPutMainChainIndex(dbTx, block.Hash(), node.order)
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
			err := b.indexManager.ConnectBlock(dbTx, block, view)
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

	// Update the state for the best block.  Notice how this replaces the
	// entire struct instead of updating the existing one.  This effectively
	// allows the old version to act as a snapshot which callers can use
	// freely without needing to hold a lock for the duration.  See the
	// comments on the state variable for more details.
	b.stateLock.Lock()
	b.stateSnapshot = state
	b.stateLock.Unlock()

	// Assemble the current block and the parent into a slice.
	blockAndParent := []*types.SerializedBlock{block}

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

// reorganizeChain reorganizes the block chain by disconnecting the nodes in the
// detachNodes list and connecting the nodes in the attach list.  It expects
// that the lists are already in the correct order and are in sync with the
// end of the current best chain.  Specifically, nodes that are being
// disconnected must be in reverse order (think of popping them off the end of
// the chain) and nodes the are being attached must be in forwards order
// (think pushing them onto the end of the chain).
//
// This function MUST be called with the chain state lock held (for writes).

func (b *BlockChain) reorganizeChain(detachNodes, attachNodes *list.List,newBlock *types.SerializedBlock) error {

	node:=b.index.LookupNode(newBlock.Hash())
	// Why the old order is the order that was removed by the new block, because the new block
	// must be one of the tip of the dag.This is very important for the following understanding.
	// In the two case, the perspective is the same.In the other words, the future can not
	// affect the past.
	var n *blockNode
	var block *types.SerializedBlock
	var err error

	for e := detachNodes.Back(); e != nil; e = e.Prev() {
		n=b.index.LookupNode(e.Value.(*hash.Hash))

		block, err = b.fetchMainChainBlockByHash(&n.hash)

		if err != nil {
			return err
		}
		if n==nil {
			return fmt.Errorf("no node")
		}
		block.SetOrder(n.order-1)
		// Load all of the utxos referenced by the block that aren't
		// already in the view.
		view := NewUtxoViewpoint()
		view.SetBestHash(block.Hash())
		err = view.fetchInputUtxos(b.db, block,b)
		if err != nil {
			return err
		}

		// Load all of the spent txos for the block from the spend
		// journal.
		var stxos map[string]spentTxOut
		err = b.db.View(func(dbTx database.Tx) error {
			stxos, err = dbFetchSpendJournalEntry(dbTx, block)
			return err
		})
		if err != nil {
			return err
		}
		// Store the loaded block and spend journal entry for later.

		prevNode:=e.Prev()
		var prevH *hash.Hash
		if prevNode!=nil {
			prevH=e.Value.(*hash.Hash)
		}else{
			prevH=b.bd.GetPrevious(block.Hash())
			if prevH.IsEqual(&node.hash) {
				prevH=b.bd.GetPrevious(prevH)
			}
		}
		err=b.disconnectTransactions(view,block,stxos,prevH)
		if err != nil {
			return err
		}
		err = b.disconnectBlock(n, block, view,prevH)
		if err != nil {
			return err
		}

	}

	for e := attachNodes.Front(); e != nil; e = e.Next() {
		nodeHash:=e.Value.(*hash.Hash)
		if nodeHash.IsEqual(newBlock.Hash()) {
			n=node
			block=newBlock
		} else {
			n=b.index.LookupNode(nodeHash)
			// If any previous nodes in attachNodes failed validation,
			// mark this one as having an invalid ancestor.
			block, err = b.FetchBlockByHash(&n.hash)

			if err != nil {
				return err
			}
		}


		view := NewUtxoViewpoint()
		view.SetBestHash(b.bd.GetPrevious(&n.hash))
		stxos:=[]spentTxOut{}
		err= b.checkConnectBlock(n, block, view, &stxos)
		if err != nil {
			return err
		}
		err = b.connectBlock(n, block, view, stxos)
		if err != nil {
			return err
		}
	}

	// Log the point where the chain forked and old and new best chain
	// heads.
	firstAttachNode := attachNodes.Front().Value.(*hash.Hash)
	lastAttachNode := attachNodes.Back().Value.(*hash.Hash)
	log.Info(fmt.Sprintf("DAG REORGANIZE: Start at %s", firstAttachNode.String()))
	log.Info(fmt.Sprintf("DAG REORGANIZE: End at %s", lastAttachNode.String()))
	log.Info(fmt.Sprintf("DAG REORGANIZE: New Len= %d;Old Len= %d",attachNodes.Len(),detachNodes.Len()))

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
func (b *BlockChain) disconnectBlock(node *blockNode, block *types.SerializedBlock, view *UtxoViewpoint,prev *hash.Hash) error {

	prevNode := b.index.LookupNode(prev)
	if prevNode==nil {
		return fmt.Errorf("no node")
	}
	prevBlock, err := b.fetchMainChainBlockByHash(prev)

	if err != nil {
		return err
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
	parentBlockSize := uint64(prevBlock.Block().SerializeSize())

	// Calculate the number of transactions that would be added by adding
	// this block.

	// TODO revisit the tx count logic
	numTxns := uint64(len(prevBlock.Block().Transactions))
	/*
	numTxns := countNumberOfTransactions(block, parent)
	*/
	newTotalTxns := curTotalTxns - numTxns

	// Calculate the exact subsidy produced by adding the block.
	subsidy := CalculateAddedSubsidy(block)
	newTotalSubsidy := curTotalSubsidy - subsidy

	state := newBestState(prevNode, parentBlockSize, numTxns,
		prevNode.CalcPastMedianTime(),  newTotalTxns, newTotalSubsidy,b.bd.GetGraphState())

	err = b.db.Update(func(dbTx database.Tx) error {
		// Update best block state.
		err := dbPutBestState(dbTx, state, node.workSum)
		if err != nil {
			return err
		}

		// Remove the block hash and height from the main chain index.
		err = dbRemoveMainChainIndex(dbTx, block.Hash(), int64(node.order))  //TODO, remove type conversion
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
			err := b.indexManager.DisconnectBlock(dbTx, block, view)
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

	b.RemoveBadTx(&node.hash)

	// Update the state for the best block.  Notice how this replaces the
	// entire struct instead of updating the existing one.  This effectively
	// allows the old version to act as a snapshot which callers can use
	// freely without needing to hold a lock for the duration.  See the
	// comments on the state variable for more details.
	b.stateLock.Lock()
	b.stateSnapshot = state
	b.stateLock.Unlock()

	// Assemble the current block and the parent into a slice.
	blockAndParent := []*types.SerializedBlock{block}

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
	curHeight := block.Order()
	curHash := block.Hash()
	b.mainchainBlockCacheLock.Lock()
	b.mainchainBlockCache[*curHash] = block
	for hash, bl := range b.mainchainBlockCache {
		if bl.Order() <= curHeight-uint64(b.mainchainBlockCacheSize) {  //TODO, remove type conversion
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

func (b *BlockChain) IsBadTx(txh *hash.Hash) bool{
	_, ok := b.badTx[*txh]
	return ok
}

func (b *BlockChain) GetBadTxFromBlock(bh *hash.Hash) []*hash.Hash{
	result:=[]*hash.Hash{}
	for k,v:=range b.badTx{
		if v.Has(bh) {
			txHash:=k
			result=append(result,&txHash)
		}
	}
	return result
}

func (b *BlockChain) AddBadTx(txh *hash.Hash,bh *hash.Hash){
	if b.IsBadTx(txh) {
		b.badTx[*txh].Add(bh)
	}else{
		set:=blockdag.NewHashSet()
		set.Add(bh)
		b.badTx[*txh]=set
	}
}

func (b *BlockChain) AddBadTxArray(txha []*hash.Hash,bh *hash.Hash){
	if len(txha)==0 {
		return
	}
	for _,v:=range txha{
		b.AddBadTx(v,bh)
	}
}

func (b *BlockChain) RemoveBadTx(bh *hash.Hash){
	for k,v:=range b.badTx{
		if v.Has(bh) {
			v.Remove(bh)
			if v.IsEmpty() {
				delete(b.badTx,k)
			}
		}
	}
}

// Return the dag instance
func (b *BlockChain) BlockDAG() *blockdag.BlockDAG{
	return b.bd
}

// Return the blockindex instance
func (b *BlockChain) BlockIndex() *blockIndex{
	return b.index
}


