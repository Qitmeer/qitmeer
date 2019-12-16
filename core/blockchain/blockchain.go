// Copyright (c) 2017-2018 The qitmeer developers

package blockchain

import (
	"container/list"
	"encoding/binary"
	"fmt"
	"github.com/Qitmeer/qitmeer/common/hash"
	"github.com/Qitmeer/qitmeer/core/blockdag"
	"github.com/Qitmeer/qitmeer/core/dbnamespace"
	"github.com/Qitmeer/qitmeer/core/types"
	"github.com/Qitmeer/qitmeer/database"
	"github.com/Qitmeer/qitmeer/engine/txscript"
	"github.com/Qitmeer/qitmeer/params"
	"github.com/Qitmeer/qitmeer/services/common/progresslog"
	"os"
	"sort"
	"sync"
	"time"
)

const (

	// maxOrphanBlocks is the maximum number of orphan blocks that can be
	// queued.
	MaxOrphanBlocks = 500
	// minMemoryNodes is the minimum number of consecutive nodes needed
	// in memory in order to perform all necessary validation.  It is used
	// to determine when it's safe to prune nodes from memory without
	// causing constant dynamic reloading.  This value should be larger than
	// that for minMemoryStakeNodes.
	minMemoryNodes = 2880
)

// BlockChain provides functions such as rejecting duplicate blocks, ensuring
// blocks follow all rules, orphan handling, checkpoint handling, and best chain
// selection with reorganization.
type BlockChain struct {
	params *params.Params

	// The following fields are set when the instance is created and can't
	// be changed afterwards, so there is no need to protect them with a
	// separate mutex.
	checkpointsByLayer map[uint64]*params.Checkpoint

	db            database.DB
	dbInfo        *databaseInfo
	timeSource    MedianTimeSource
	notifications NotificationCallback
	sigCache      *txscript.SigCache
	indexManager  IndexManager

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
	index *blockIndex

	// These fields are related to handling of orphan blocks.  They are
	// protected by a combination of the chain lock and the orphan lock.
	orphanLock   sync.RWMutex
	orphans      map[hash.Hash]*orphanBlock
	prevOrphans  map[hash.Hash][]*orphanBlock
	oldestOrphan *orphanBlock

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

	//tx manager
	txManager TxManager

	// block version
	BlockVersion uint32
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

	// block version
	BlockVersion uint32
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
	Hash         hash.Hash            // The hash of the main chain tip.
	Bits         uint32               // The difficulty bits of the main chain tip.
	BlockSize    uint64               // The size of the main chain tip.
	NumTxns      uint64               // The number of txns in the main chain tip.
	MedianTime   time.Time            // Median time as per CalcPastMedianTime.
	TotalTxns    uint64               // The total number of txns in the chain.
	TotalSubsidy uint64               // The total subsidy for the chain.
	GraphState   *blockdag.GraphState // The graph state of dag
}

// newBestState returns a new best stats instance for the given parameters.
func newBestState(tipHash *hash.Hash, bits uint32, blockSize, numTxns uint64, medianTime time.Time, totalTxns uint64, totalsubsidy uint64, gs *blockdag.GraphState) *BestState {
	return &BestState{
		Hash:         *tipHash,
		Bits:         bits,
		BlockSize:    blockSize,
		NumTxns:      numTxns,
		MedianTime:   medianTime,
		TotalTxns:    totalTxns,
		TotalSubsidy: totalsubsidy,
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
	var checkpointsByLayer map[uint64]*params.Checkpoint
	var prevCheckpointLayer uint64
	if len(par.Checkpoints) > 0 {
		checkpointsByLayer = make(map[uint64]*params.Checkpoint)
		for i := range par.Checkpoints {
			checkpoint := &par.Checkpoints[i]
			if checkpoint.Layer <= prevCheckpointLayer {
				return nil, AssertError("blockchain.New " +
					"checkpoints are not sorted by height")
			}
			checkpointsByLayer[checkpoint.Layer] = checkpoint
			prevCheckpointLayer = checkpoint.Layer
		}
	}

	b := BlockChain{
		checkpointsByLayer: checkpointsByLayer,
		db:                 config.DB,
		params:             par,
		timeSource:         config.TimeSource,
		notifications:      config.Notifications,
		sigCache:           config.SigCache,
		indexManager:       config.IndexManager,
		index:              newBlockIndex(config.DB, par),
		orphans:            make(map[hash.Hash]*orphanBlock),
		prevOrphans:        make(map[hash.Hash][]*orphanBlock),
		BlockVersion:       config.BlockVersion,
	}
	b.subsidyCache = NewSubsidyCache(0, b.params)

	b.bd = &blockdag.BlockDAG{}
	b.bd.Init(config.DAGType, b.subsidyCache.CalcBlockSubsidy,
		1.0/float64(par.TargetTimePerBlock/time.Second))
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

	log.Info(fmt.Sprintf("DAG Type:%s", b.bd.GetName()))
	log.Info("Blockchain database version", "chain", b.dbInfo.version, "compression", b.dbInfo.compVer,
		"index", b.dbInfo.bidxVer)

	tips := b.bd.GetTipsList()
	log.Info(fmt.Sprintf("Chain state:totaltx=%d tipsNum=%d mainOrder=%d total=%d", b.BestSnapshot().TotalTxns, len(tips), b.bd.GetMainChainTip().GetOrder(), b.bd.GetBlockTotal()))

	for _, v := range tips {
		tnode := b.index.LookupNode(v.GetHash())
		log.Info(fmt.Sprintf("hash=%v,order=%s,work=%v", tnode.hash, blockdag.GetOrderLogStr(uint(tnode.GetOrder())), tnode.workSum))
	}

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
		log.Trace("Serialized chain state: ", "serializedData", fmt.Sprintf("%x", serializedData))
		state, err := deserializeBestChainState(serializedData)
		if err != nil {
			return err
		}
		log.Trace(fmt.Sprintf("Load chain state:%s %d %d %d %s", state.hash.String(), state.total, state.totalTxns, state.totalsubsidy, state.workSum.Text(16)))

		log.Info("Loading dag ...")
		bidxStart := time.Now()

		err = b.bd.Load(dbTx, uint(state.total), b.params.GenesisHash)
		if err != nil {
			return fmt.Errorf("The dag data was damaged (%s). you can cleanup your block data base by '--cleanup'.", err)
		}
		if !b.bd.GetMainChainTip().GetHash().IsEqual(&state.hash) {
			return fmt.Errorf("The dag main tip %s is not the same. %s", state.hash.String(), b.bd.GetMainChainTip().GetHash().String())
		}
		log.Info(fmt.Sprintf("Dag loaded:loadTime=%v", time.Since(bidxStart)))

		// Determine how many blocks will be loaded into the index in order to
		// allocate the right amount as a single alloc versus a whole bunch of
		// littles ones to reduce pressure on the GC.
		var block *types.SerializedBlock
		for i := uint(0); i < uint(state.total); i++ {
			blockHash := b.bd.GetBlockHash(i)
			block, err = dbFetchBlockByHash(dbTx, blockHash)
			if err != nil {
				return err
			}
			if i != 0 && block.Block().Header.Version != b.BlockVersion {
				return fmt.Errorf("The dag block is not match current genesis block. you can cleanup your block data base by '--cleanup'.")
			}
			parents := []*blockNode{}
			for _, pb := range block.Block().Parents {
				parent := b.index.LookupNode(pb)
				if parent == nil {
					return fmt.Errorf("Can't find parent %s", pb.String())
				}
				parents = append(parents, parent)
			}
			refblock := b.bd.GetBlock(blockHash)
			//
			node := &blockNode{}
			initBlockNode(node, &block.Block().Header, parents)
			b.index.addNode(node)
			node.status = blockStatus(refblock.GetStatus())
			node.SetOrder(uint64(refblock.GetOrder()))
			node.SetHeight(refblock.GetHeight())
			if i != 0 {
				node.CalcWorkSum(node.GetMainParent(b))
			}
		}

		// Set the best chain view to the stored best state.
		// Load the raw block bytes for the best block.
		mainTip := b.index.LookupNode(b.bd.GetMainChainTip().GetHash())
		// Initialize the state related to the best block.
		blockSize := uint64(block.Block().SerializeSize())
		numTxns := uint64(len(block.Block().Transactions))
		b.stateSnapshot = newBestState(mainTip.GetHash(), mainTip.bits, blockSize, numTxns,
			mainTip.CalcPastMedianTime(b), state.totalTxns, b.bd.GetMainChainTip().GetWeight(), b.bd.GetGraphState())

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
	return b.index.HaveBlock(hash) || b.IsOrphan(hash), nil
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
func (b *BlockChain) IsOrphan(hash *hash.Hash) bool {
	// Protect concurrent access.  Using a read lock only so multiple
	// readers can query without blocking each other.
	b.orphanLock.RLock()
	_, exists := b.orphans[*hash]
	b.orphanLock.RUnlock()

	return exists
}

// Whether it is connected by all parents
func (b *BlockChain) IsUnconnectedOrphan(hash *hash.Hash) bool {
	op := b.GetRecentOrphanParents(hash)
	return len(op) > 0
}

// GetOrphansParents returns the parents for the provided hash from the
// map of orphan blocks.
func (b *BlockChain) GetOrphansParents() []*hash.Hash {
	b.orphanLock.RLock()
	defer b.orphanLock.RUnlock()
	//
	result := blockdag.NewHashSet()
	for _, v := range b.orphans {
		for _, h := range v.block.Block().Parents {
			exists, err := b.HaveBlock(h)
			if err != nil || exists {
				continue
			}
			result.Add(h)
		}

	}
	return result.List()
}

// GetOrphansParents returns the parents for the provided hash from the
// map of orphan blocks.
func (b *BlockChain) GetRecentOrphanParents(h *hash.Hash) []*hash.Hash {
	b.orphanLock.RLock()
	defer b.orphanLock.RUnlock()
	//
	ob, exists := b.orphans[*h]
	if !exists {
		log.Debug(fmt.Sprintf("GetRecentOrphanParents not found hash %v in the orphans list",h))
		return nil
	}
	if !b.IsOrphanNear(ob) {
		return nil
	}
	result := blockdag.NewHashSet()
	for _, h := range ob.block.Block().Parents {
		exists, err := b.HaveBlock(h)
		if err != nil || exists {
			continue
		}
		result.Add(h)
	}
	log.Trace(fmt.Sprintf("GetRecentOrphanParents result list %v",result))
	return result.List()
}

// Get the total of all orphans
func (b *BlockChain) GetOrphansTotal() int {
	b.orphanLock.RLock()
	ol := len(b.orphans)
	b.orphanLock.RUnlock()
	return ol
}

func (b *BlockChain) GetRecentOrphansParents() []*hash.Hash {
	b.orphanLock.RLock()
	defer b.orphanLock.RUnlock()

	result := blockdag.NewHashSet()
	for _, v := range b.orphans {
		if !b.IsOrphanNear(v) {
			continue
		}
		for _, h := range v.block.Block().Parents {
			exists, err := b.HaveBlock(h)
			if err != nil || exists {
				continue
			}
			result.Add(h)
		}

	}
	return result.List()
}

func (b *BlockChain) IsOrphanNear(ob *orphanBlock) bool {
	serializedHeight, err := ExtractCoinbaseHeight(ob.block.Block().Transactions[0])
	if err != nil {
		return false
	}
	dist := int64(serializedHeight) - int64(blockdag.StableConfirmations)
	if dist <= 0 {
		return true
	}
	if uint(dist) > b.BestSnapshot().GraphState.GetMainHeight() {
		return false
	}
	return true
}

// IsCurrent returns whether or not the chain believes it is current.  Several
// factors are used to guess, but the key factors that allow the chain to
// believe it is current are:
//  - Latest block height is after the latest checkpoint (if enabled)
//  - Latest block has a timestamp newer than 24 hours ago
//
// This function is safe for concurrent access.
func (b *BlockChain) IsCurrent() bool {
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
	checkpoint := b.LatestCheckpoint()
	lastBlock := b.bd.GetMainChainTip()
	if checkpoint != nil && uint64(lastBlock.GetLayer()) < checkpoint.Layer {
		return false
	}

	// Not current if the latest best block has a timestamp before 24 hours
	// ago.
	//
	// The chain appears to be current if none of the checks reported
	// otherwise.
	minus24Hours := b.timeSource.AdjustedTime().Add(-24 * time.Hour).Unix()
	lastNode := b.index.LookupNode(lastBlock.GetHash())
	return lastNode.timestamp >= minus24Hours
}

// TipGeneration returns the entire generation of blocks stemming from the
// parent of the current tip.
//
// The function is safe for concurrent access.
func (b *BlockChain) TipGeneration() ([]hash.Hash, error) {
	tips := b.bd.GetTipsList()
	tiphashs := []hash.Hash{}
	for _, block := range tips {
		tiphashs = append(tiphashs, *block.GetHash())
	}
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
	b.ChainRLock()
	defer b.ChainRUnlock()

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
	if !b.MainChainHasBlock(hash) {
		return nil, fmt.Errorf("No block in main chain")
	}
	block, err := b.fetchBlockByHash(hash)
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
	if len(b.orphans)+1 > MaxOrphanBlocks*2 {
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
	var block *types.SerializedBlock
	// Load the block from the database.
	dbErr := b.db.View(func(dbTx database.Tx) error {
		var err error
		block, err = dbFetchBlockByHash(dbTx, hash)
		return err
	})
	if dbErr == nil && block != nil {
		return block, nil
	}
	return nil, fmt.Errorf("unable to find block %v db", hash)
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
func (b *BlockChain) connectDagChain(node *blockNode, block *types.SerializedBlock, newOrders *list.List, oldOrders BlockNodeList) (bool, error) {
	if newOrders.Len() == 0 {
		return true, nil
	}
	//Fast double spent check
	b.fastDoubleSpentCheck(node, block)

	// We are extending the main (best) chain with a new block.  This is the
	// most common case.
	if newOrders.Len() == 1 {
		if !node.IsOrdered() {
			return true, nil
		}
		// Perform several checks to verify the block can be connected
		// to the main chain without violating any rules and without
		// actually connecting the block.
		view := NewUtxoViewpoint()
		view.SetViewpoints([]*hash.Hash{&node.hash})

		stxos := make([]SpentTxOut, 0, countSpentOutputs(block))
		err := b.checkConnectBlock(node, block, view, &stxos)
		if err != nil {
			node.Invalid(b)
			stxos = []SpentTxOut{}
			view.Clean()
		}
		// In the fast add case the code to check the block connection
		// was skipped, so the utxo view needs to load the referenced
		// utxos, spend them, and add the new utxos being created by
		// this block.

		// Connect the block to the main chain.
		err = b.connectBlock(node, block, view, stxos)
		if err != nil {
			node.Invalid(b)
			return true, err
		}
		if !node.GetStatus().KnownInvalid() {
			node.Valid(b)
		}
		// TODO, validating previous block
		log.Debug("Block connected to the main chain", "hash", node.hash, "order", node.order)
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
	log.Debug(fmt.Sprintf("Start DAG REORGANIZE: Block %v is causing a reorganize.", node.hash))
	err := b.reorganizeChain(oldOrders, newOrders, block)
	if err != nil {
		return false, err
	}
	//b.updateBestState(node, block)
	return true, nil
}

// This function is fast check before global sequencing,it can judge who is the bad block quickly.
func (b *BlockChain) fastDoubleSpentCheck(node *blockNode, block *types.SerializedBlock) {
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
					b.AddInvalidTx(tx.Hash(),block.Hash())
				}
			}
		}
	}*/
}

func (b *BlockChain) updateBestState(node *blockNode, block *types.SerializedBlock) error {
	// Must be end node of sequence in dag
	// Generate a new best state snapshot that will be used to update the
	// database and later memory if all database updates are successful.
	b.stateLock.RLock()
	curTotalTxns := b.stateSnapshot.TotalTxns
	b.stateLock.RUnlock()

	// Calculate the number of transactions that would be added by adding
	// this block.
	numTxns := uint64(len(block.Block().Transactions))

	blockSize := uint64(block.Block().SerializeSize())

	mainTip := b.index.LookupNode(b.bd.GetMainChainTip().GetHash())

	state := newBestState(mainTip.GetHash(), mainTip.bits, blockSize, numTxns, mainTip.CalcPastMedianTime(b), curTotalTxns+numTxns,
		b.bd.GetMainChainTip().GetWeight(), b.bd.GetGraphState())

	// Atomically insert info into the database.
	err := b.db.Update(func(dbTx database.Tx) error {
		// Update best block state.
		err := dbPutBestState(dbTx, state, mainTip.workSum)
		if err != nil {
			return err
		}
		return nil
	})

	if err != nil {
		return err
	}
	// Update the state for the best block.  Notice how this replaces the
	// entire struct instead of updating the existing one.  This effectively
	// allows the old version to act as a snapshot which callers can use
	// freely without needing to hold a lock for the duration.  See the
	// comments on the state variable for more details.
	b.stateLock.Lock()
	b.stateSnapshot = state
	b.stateLock.Unlock()

	return nil
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
func (b *BlockChain) connectBlock(node *blockNode, block *types.SerializedBlock, view *UtxoViewpoint, stxos []SpentTxOut) error {
	// Atomically insert info into the database.
	err := b.db.Update(func(dbTx database.Tx) error {
		// Add the block hash and height to the block index.
		err := dbPutBlockIndex(dbTx, block.Hash(), node.order)
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
			err := b.indexManager.ConnectBlock(dbTx, block, stxos)
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

	b.sendNotification(BlockConnected, []*types.SerializedBlock{block})
	return nil
}

// disconnectBlock handles disconnecting the passed node/block from the end of
// the main (best) chain.
//
// This function MUST be called with the chain state lock held (for writes).
func (b *BlockChain) disconnectBlock(node *blockNode, block *types.SerializedBlock, view *UtxoViewpoint, stxos []SpentTxOut) error {
	// Calculate the exact subsidy produced by adding the block.
	err := b.db.Update(func(dbTx database.Tx) error {
		// Remove the block hash and order from the block index.
		err := dbRemoveBlockIndex(dbTx, block.Hash(), int64(node.order)) //TODO, remove type conversion
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
			err := b.indexManager.DisconnectBlock(dbTx, block, stxos)
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

	b.sendNotification(BlockDisconnected, block)

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

func (b *BlockChain) reorganizeChain(detachNodes BlockNodeList, attachNodes *list.List, newBlock *types.SerializedBlock) error {

	node := b.index.LookupNode(newBlock.Hash())
	// Why the old order is the order that was removed by the new block, because the new block
	// must be one of the tip of the dag.This is very important for the following understanding.
	// In the two case, the perspective is the same.In the other words, the future can not
	// affect the past.
	var n *blockNode
	var block *types.SerializedBlock
	var err error

	dl := len(detachNodes)
	for i := dl - 1; i >= 0; i-- {
		n = detachNodes[i]
		newn := b.index.LookupNode(n.GetHash())
		block, err = b.fetchBlockByHash(&n.hash)
		if err != nil || n == nil {
			panic(err.Error())
		}
		if !n.IsOrdered() {
			panic("no ordered")
		}
		block.SetOrder(n.order)
		// Load all of the utxos referenced by the block that aren't
		// already in the view.
		var stxos []SpentTxOut
		view := NewUtxoViewpoint()
		view.SetViewpoints([]*hash.Hash{block.Hash()})
		if !b.index.NodeStatus(n).KnownInvalid() {
			err = view.fetchInputUtxos(b.db, block, b)
			if err != nil {
				return err
			}

			// Load all of the spent txos for the block from the spend
			// journal.

			err = b.db.View(func(dbTx database.Tx) error {
				stxos, err = dbFetchSpendJournalEntry(dbTx, block)
				return err
			})
			if err != nil {
				return err
			}
			// Store the loaded block and spend journal entry for later.
			err = view.disconnectTransactions(block, stxos)
			if err != nil {
				n.Invalid(b)
				newn.Invalid(b)
				log.Info(fmt.Sprintf("%s", err))
			}
		}

		n.UnsetStatusFlags(statusValid)
		newn.UnsetStatusFlags(statusValid)
		n.UnsetStatusFlags(statusInvalid)
		newn.UnsetStatusFlags(statusInvalid)
		newn.FlushToDB(b)

		err = b.disconnectBlock(n, block, view, stxos)
		if err != nil {
			return err
		}
	}

	for e := attachNodes.Front(); e != nil; e = e.Next() {
		nodeHash := e.Value.(*hash.Hash)
		if nodeHash.IsEqual(newBlock.Hash()) {
			n = node
			block = newBlock
		} else {
			n = b.index.LookupNode(nodeHash)
			// If any previous nodes in attachNodes failed validation,
			// mark this one as having an invalid ancestor.
			block, err = b.FetchBlockByHash(&n.hash)

			if err != nil {
				return err
			}
			block.SetOrder(n.GetOrder())
		}
		if !n.IsOrdered() {
			continue
		}
		view := NewUtxoViewpoint()
		view.SetViewpoints([]*hash.Hash{n.GetHash()})
		stxos := []SpentTxOut{}
		err = b.checkConnectBlock(n, block, view, &stxos)
		if err != nil {
			n.Invalid(b)
			stxos = []SpentTxOut{}
			view.Clean()
			log.Info(fmt.Sprintf("%s", err))
		}
		err = b.connectBlock(n, block, view, stxos)
		if err != nil {
			n.Invalid(b)
			log.Info(fmt.Sprintf("%s", err))
			continue
		}
		if !n.GetStatus().KnownInvalid() {
			n.Valid(b)
		}
	}

	// Log the point where the chain forked and old and new best chain
	// heads.
	log.Debug(fmt.Sprintf("End DAG REORGANIZE: Old Len= %d;New Len= %d", attachNodes.Len(), detachNodes.Len()))

	return nil
}

// countSpentOutputs returns the number of utxos the passed block spends.
// TODO, revisit the design of stxos count
func countSpentOutputs(block *types.SerializedBlock) int {
	// Exclude the coinbase transaction since it can't spend anything.
	var numSpent int
	for _, tx := range block.Transactions()[1:] {
		numSpent += len(tx.Transaction().TxIn)
	}
	return numSpent
}

// Return the dag instance
func (b *BlockChain) BlockDAG() *blockdag.BlockDAG {
	return b.bd
}

// Return the blockindex instance
func (b *BlockChain) BlockIndex() *blockIndex {
	return b.index
}

// Return median time source
func (b *BlockChain) TimeSource() MedianTimeSource {
	return b.timeSource
}

// Return the reorganization information
func (b *BlockChain) getReorganizeNodes(newNode *blockNode, block *types.SerializedBlock, newOrders *list.List, oldOrders *BlockNodeList) {
	var refnode *blockNode
	var oldOrdersTemp BlockNodeList

	for e := newOrders.Front(); e != nil; e = e.Next() {
		refHash := e.Value.(*hash.Hash)
		refblock := b.bd.GetBlock(refHash)
		if refHash.IsEqual(&newNode.hash) {
			newNode.SetLayer(refblock.GetLayer())
			refnode = newNode
			block.SetOrder(uint64(refblock.GetOrder()))
			if refblock.GetHeight() != newNode.GetHeight() {
				log.Warn(fmt.Sprintf("The consensus main height is not match (%s) %d-%d", newNode.GetHash(), newNode.GetHeight(), refblock.GetHeight()))
				newNode.SetHeight(refblock.GetHeight())
				block.SetHeight(refblock.GetHeight())
			}
		} else {
			refnode = b.index.LookupNode(refHash)
			if refnode.IsOrdered() {
				oldOrdersTemp = append(oldOrdersTemp, refnode.Clone())
			}
		}
		refnode.SetOrder(uint64(refblock.GetOrder()))
	}
	if newOrders.Len() <= 1 || len(oldOrdersTemp) == 0 {
		return
	}

	if len(oldOrdersTemp) > 1 {
		sort.Sort(oldOrdersTemp)
	}
	oldOrdersList := list.New()
	for i := 0; i < len(oldOrdersTemp); i++ {
		oldOrdersList.PushBack(oldOrdersTemp[i])
	}

	// optimization
	ne := newOrders.Front()
	oe := oldOrdersList.Front()
	for {
		if ne == nil || oe == nil {
			break
		}
		neNext := ne.Next()
		oeNext := oe.Next()

		neHash := ne.Value.(*hash.Hash)
		oeNode := oe.Value.(*blockNode)
		if neHash.IsEqual(oeNode.GetHash()) {
			newOrders.Remove(ne)
			oldOrdersList.Remove(oe)
		} else {
			break
		}

		ne = neNext
		oe = oeNext
	}
	//
	for e := oldOrdersList.Front(); e != nil; e = e.Next() {
		node := e.Value.(*blockNode)
		*oldOrders = append(*oldOrders, node)
	}
}

// FetchSpendJournal can return the set of outputs spent for the target block.
func (b *BlockChain) FetchSpendJournal(targetBlock *types.SerializedBlock) ([]SpentTxOut, error) {
	b.ChainRLock()
	defer b.ChainRUnlock()

	var spendEntries []SpentTxOut
	err := b.db.View(func(dbTx database.Tx) error {
		var err error

		spendEntries, err = dbFetchSpendJournalEntry(dbTx, targetBlock)
		return err
	})
	if err != nil {
		return nil, err
	}

	return spendEntries, nil
}

func (b *BlockChain) SetTxManager(txManager TxManager) {
	b.txManager = txManager
}

func (b *BlockChain) GetTxManager() TxManager {
	return b.txManager
}

func (b *BlockChain) GetMiningTips() []*hash.Hash {
	return b.BlockDAG().GetValidTips()
}

func (b *BlockChain) ChainLock() {
	b.chainLock.Lock()

}

func (b *BlockChain) ChainUnlock() {
	b.chainLock.Unlock()
}

func (b *BlockChain) ChainRLock() {
	b.chainLock.RLock()
}

func (b *BlockChain) ChainRUnlock() {
	b.chainLock.RUnlock()
}
