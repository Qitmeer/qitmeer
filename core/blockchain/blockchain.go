// Copyright (c) 2017-2018 The qitmeer developers

package blockchain

import (
	"container/list"
	"encoding/binary"
	"fmt"
	"github.com/Qitmeer/qitmeer/common/hash"
	"github.com/Qitmeer/qitmeer/common/roughtime"
	"github.com/Qitmeer/qitmeer/common/util"
	"github.com/Qitmeer/qitmeer/core/blockchain/token"
	"github.com/Qitmeer/qitmeer/core/blockdag"
	"github.com/Qitmeer/qitmeer/core/dbnamespace"
	"github.com/Qitmeer/qitmeer/core/event"
	"github.com/Qitmeer/qitmeer/core/merkle"
	"github.com/Qitmeer/qitmeer/core/serialization"
	"github.com/Qitmeer/qitmeer/core/types"
	"github.com/Qitmeer/qitmeer/core/types/pow"
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

	db           database.DB
	dbInfo       *databaseInfo
	timeSource   MedianTimeSource
	events       *event.Feed
	sigCache     *txscript.SigCache
	indexManager IndexManager

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

	// These fields are related to handling of orphan blocks.  They are
	// protected by a combination of the chain lock and the orphan lock.
	orphanLock   sync.RWMutex
	orphans      map[hash.Hash]*orphanBlock
	oldestOrphan *orphanBlock

	// These fields are related to checkpoint handling.  They are protected
	// by the chain lock.
	nextCheckpoint *params.Checkpoint
	checkpointNode blockdag.IBlock

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

	// Cache Invalid tx
	CacheInvalidTx bool

	// cache notification
	CacheNotifications []*Notification
	notificationsLock sync.RWMutex
	notifications     []NotificationCallback

	// The ID of token state tip for the chain.
	TokenTipID uint32

	warningCaches      []thresholdStateCache
	deploymentCaches   []thresholdStateCache
	unknownRulesWarned bool
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

	// Events defines a event manager to which notifications will be sent
	// when various events take place.  See the documentation for
	// Notification and NotificationType for details on the types and
	// contents of notifications.
	//
	// This field can be nil if the caller is not interested in receiving
	// notifications.
	Events *event.Feed

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

	// Cache Invalid tx
	CacheInvalidTx bool
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
	TokenTipHash *hash.Hash           // The Hash of token state tip for the chain.
	GraphState   *blockdag.GraphState // The graph state of dag
}

// newBestState returns a new best stats instance for the given parameters.
func newBestState(tipHash *hash.Hash, bits uint32, blockSize, numTxns uint64, medianTime time.Time,
	totalTxns uint64, totalsubsidy uint64, gs *blockdag.GraphState, tokenTipHash *hash.Hash) *BestState {
	return &BestState{
		Hash:         *tipHash,
		Bits:         bits,
		BlockSize:    blockSize,
		NumTxns:      numTxns,
		MedianTime:   medianTime,
		TotalTxns:    totalTxns,
		TotalSubsidy: totalsubsidy,
		TokenTipHash: tokenTipHash,
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

// OrderRange returns a range of block hashes for the given start and end
// orders.  It is inclusive of the start order and exclusive of the end
// order.  The end order will be limited to the current main chain order.
//
// This function is safe for concurrent access.
func (b *BlockChain) OrderRange(startOrder, endOrder uint64) ([]hash.Hash, error) {
	// Ensure requested orders are sane.
	if startOrder < 0 {
		return nil, fmt.Errorf("start order of fetch range must not "+
			"be less than zero - got %d", startOrder)
	}
	if endOrder < startOrder {
		return nil, fmt.Errorf("end order of fetch range must not "+
			"be less than the start order - got start %d, end %d",
			startOrder, endOrder)
	}

	// There is nothing to do when the start and end orders are the same,
	// so return now to avoid the chain view lock.
	if startOrder == endOrder {
		return nil, nil
	}

	// Grab a lock on the chain view to prevent it from changing due to a
	// reorg while building the hashes.
	b.chainLock.Lock()
	defer b.chainLock.Unlock()

	// When the requested start order is after the most recent best chain
	// order, there is nothing to do.
	latestOrder := b.BestSnapshot().GraphState.GetMainOrder()
	if startOrder > uint64(latestOrder) {
		return nil, nil
	}

	// Limit the ending order to the latest order of the chain.
	if endOrder > uint64(latestOrder+1) {
		endOrder = uint64(latestOrder + 1)
	}

	// Fetch as many as are available within the specified range.
	hashes := make([]hash.Hash, 0, endOrder-startOrder)
	for i := startOrder; i < endOrder; i++ {
		h, err := b.BlockHashByOrder(i)
		if err != nil {
			log.Error("order not exist", "order", i)
			return nil, err
		}
		hashes = append(hashes, *h)
	}
	return hashes, nil
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

	if len(par.Deployments) > 0 {
		for _, v := range par.Deployments {
			if v.StartTime < CheckerTimeThreshold &&
				v.ExpireTime < CheckerTimeThreshold &&
				(v.PerformTime < CheckerTimeThreshold || v.PerformTime == 0) {
				continue
			}
			if v.StartTime >= CheckerTimeThreshold &&
				v.ExpireTime >= CheckerTimeThreshold &&
				(v.PerformTime >= CheckerTimeThreshold || v.PerformTime == 0) {
				continue
			}
			if v.StartTime < v.ExpireTime &&
				(v.ExpireTime < v.PerformTime || v.PerformTime == 0) {
				continue
			}
			return nil, AssertError("blockchain.New chain parameters Deployments error")
		}
	}

	b := BlockChain{
		checkpointsByLayer: checkpointsByLayer,
		db:                 config.DB,
		params:             par,
		timeSource:         config.TimeSource,
		events:             config.Events,
		sigCache:           config.SigCache,
		indexManager:       config.IndexManager,
		orphans:            make(map[hash.Hash]*orphanBlock),
		CacheInvalidTx:     config.CacheInvalidTx,
		CacheNotifications: []*Notification{},
		warningCaches:      newThresholdCaches(VBNumBits),
		deploymentCaches:   newThresholdCaches(params.DefinedDeployments),
	}
	b.subsidyCache = NewSubsidyCache(0, b.params)

	b.bd = &blockdag.BlockDAG{}
	b.bd.Init(config.DAGType, b.CalcWeight,
		1.0/float64(par.TargetTimePerBlock/time.Second), b.db, b.getBlockData)
	b.bd.SetTipsDisLimit(int64(par.CoinbaseMaturity))
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
	err := b.CheckCacheInvalidTxConfig()
	if err != nil {
		return nil, err
	}
	b.pruner = newChainPruner(&b)

	// Initialize rule change threshold state caches.
	if err := b.initThresholdCaches(); err != nil {
		return nil, err
	}

	log.Info(fmt.Sprintf("DAG Type:%s", b.bd.GetName()))
	log.Info("Blockchain database version", "chain", b.dbInfo.version, "compression", b.dbInfo.compVer,
		"index", b.dbInfo.bidxVer)

	tips := b.bd.GetTipsList()
	log.Info(fmt.Sprintf("Chain state:totaltx=%d tipsNum=%d mainOrder=%d total=%d", b.BestSnapshot().TotalTxns, len(tips), b.bd.GetMainChainTip().GetOrder(), b.bd.GetBlockTotal()))

	for _, v := range tips {
		log.Info(fmt.Sprintf("hash=%s,order=%s,height=%d", v.GetHash(), blockdag.GetOrderLogStr(v.GetOrder()), v.GetHeight()))
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
		if dbInfo.compVer > serialization.CurrentCompressionVersion {
			return fmt.Errorf("the current database compression "+
				"version is no longer compatible with this "+
				"version of the software (%d > %d)",
				dbInfo.compVer, serialization.CurrentCompressionVersion)
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

	//   Upgrade the database as needed.
	err = b.upgradeDB()
	if err != nil {
		return err
	}

	// Attempt to load the chain state from the database.
	err = b.db.Update(func(dbTx database.Tx) error {
		// Fetch the stored chain state from the database metadata.
		// When it doesn't exist, it means the database hasn't been
		// initialized for use with chain yet, so break out now to allow
		// that to happen under a writable database transaction.
		meta := dbTx.Metadata()
		serializedData := meta.Get(dbnamespace.ChainStateKeyName)
		if serializedData == nil {
			return fmt.Errorf("No chain state data")
		}
		log.Trace("Serialized chain state: ", "serializedData", fmt.Sprintf("%x", serializedData))
		state, err := DeserializeBestChainState(serializedData)
		if err != nil {
			return err
		}
		log.Trace(fmt.Sprintf("Load chain state:%s %d %d %s %s", state.hash.String(), state.total, state.totalTxns, state.tokenTipHash.String(), state.workSum.Text(16)))

		log.Info("Loading dag ...")
		bidxStart := roughtime.Now()

		err = b.bd.Load(dbTx, uint(state.total), b.params.GenesisHash)
		if err != nil {
			return fmt.Errorf("The dag data was damaged (%s). you can cleanup your block data base by '--cleanup'.", err)
		}
		if !b.bd.GetMainChainTip().GetHash().IsEqual(&state.hash) {
			return fmt.Errorf("The dag main tip %s is not the same. %s", state.hash.String(), b.bd.GetMainChainTip().GetHash().String())
		}
		log.Info(fmt.Sprintf("Dag loaded:loadTime=%v", roughtime.Since(bidxStart)))

		// Set the best chain view to the stored best state.
		// Load the raw block bytes for the best block.
		mainTip := b.bd.GetMainChainTip()
		mainTipNode := b.GetBlockNode(mainTip)
		if mainTipNode == nil {
			return fmt.Errorf("No main tip\n")
		}
		block, err := dbFetchBlockByHash(dbTx, mainTip.GetHash())
		if err != nil {
			return err
		}

		// Initialize the state related to the best block.
		blockSize := uint64(block.Block().SerializeSize())
		numTxns := uint64(len(block.Block().Transactions))

		b.TokenTipID = uint32(b.bd.GetBlockId(&state.tokenTipHash))
		b.stateSnapshot = newBestState(mainTip.GetHash(), mainTipNode.Difficulty(), blockSize, numTxns,
			b.CalcPastMedianTime(mainTip), state.totalTxns, b.bd.GetMainChainTip().GetWeight(),
			b.bd.GetGraphState(), &state.tokenTipHash)
		return nil
	})
	if err != nil {
		return err
	}
	ts := b.GetTokenState(b.TokenTipID)
	if ts == nil {
		return fmt.Errorf("token state error")
	}
	return ts.Commit()
}

// HaveBlock returns whether or not the chain instance has the block represented
// by the passed hash.  This includes checking the various places a block can
// be like part of the main chain, on a side chain, or in the orphan pool.
//
// This function is safe for concurrent access.
func (b *BlockChain) HaveBlock(hash *hash.Hash) bool {
	return b.bd.HasBlock(hash) || b.IsOrphan(hash)
}

func (b *BlockChain) HasBlockInDB(h *hash.Hash) bool {
	err := b.db.View(func(dbTx database.Tx) error {
		has,er:=dbTx.HasBlock(h)
		if er != nil {
			return er
		}
		if has {
			return nil
		}
		return fmt.Errorf("no")
	})
	return err == nil
}

// IsCurrent returns whether or not the chain believes it is current.  Several
// factors are used to guess, but the key factors that allow the chain to
// believe it is current are:
//  - Latest block height is after the latest checkpoint (if enabled)
//  - Latest block has a timestamp newer than 24 hours ago
//
// This function is safe for concurrent access.
func (b *BlockChain) IsCurrent() bool {
	b.ChainRLock()
	defer b.ChainRUnlock()
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
	lastNode := b.GetBlockNode(lastBlock)
	if lastNode == nil {
		return false
	}
	return lastNode.GetTimestamp() >= minus24Hours
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
	block, err := b.fetchBlockByHash(hash)
	if err != nil || block == nil {
		return types.BlockHeader{}, fmt.Errorf("block %s is not known", hash)
	}

	return block.Block().Header, nil
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

// MaximumBlockSize returns the maximum permitted block size for the block
// AFTER the given node.
//
// This function MUST be called with the chain state lock held (for reads).
func (b *BlockChain) maxBlockSize() (int64, error) {

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
	block := b.GetOrphan(hash)
	if block != nil {
		return block, nil
	}

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
func (b *BlockChain) connectDagChain(ib blockdag.IBlock, block *types.SerializedBlock, newOrders *list.List, oldOrders *list.List) (bool, error) {
	if newOrders.Len() == 0 {
		return true, nil
	}
	//Fast double spent check
	b.fastDoubleSpentCheck(ib, block)

	// We are extending the main (best) chain with a new block.  This is the
	// most common case.
	if newOrders.Len() == 1 {
		if !ib.IsOrdered() {
			return true, nil
		}
		// Perform several checks to verify the block can be connected
		// to the main chain without violating any rules and without
		// actually connecting the block.
		view := NewUtxoViewpoint()
		view.SetViewpoints([]*hash.Hash{ib.GetHash()})

		stxos := []SpentTxOut{}
		err := b.checkConnectBlock(ib, block, view, &stxos)
		if err != nil {
			b.bd.InvalidBlock(ib)
			stxos = []SpentTxOut{}
			view.Clean()
		}
		// In the fast add case the code to check the block connection
		// was skipped, so the utxo view needs to load the referenced
		// utxos, spend them, and add the new utxos being created by
		// this block.

		// Connect the block to the main chain.
		err = b.connectBlock(ib, block, view, stxos)
		if err != nil {
			b.bd.InvalidBlock(ib)
			return true, err
		}
		if !ib.GetStatus().KnownInvalid() {
			b.bd.ValidBlock(ib)
		}

		// TODO, validating previous block
		log.Debug("Block connected to the main chain", "hash", ib.GetHash(), "order", ib.GetOrder())
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
	log.Debug(fmt.Sprintf("Start DAG REORGANIZE: Block %v is causing a reorganize.", ib.GetHash()))
	err := b.reorganizeChain(ib, oldOrders, newOrders, block)
	if err != nil {
		return false, err
	}
	//b.updateBestState(node, block)
	return true, nil
}

// This function is fast check before global sequencing,it can judge who is the bad block quickly.
func (b *BlockChain) fastDoubleSpentCheck(ib blockdag.IBlock, block *types.SerializedBlock) {
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

func (b *BlockChain) updateBestState(ib blockdag.IBlock, block *types.SerializedBlock, attachNodes *list.List) error {
	// No warnings about unknown rules until the chain is current.
	if b.isCurrent() {
		// Warn if any unknown new rules are either about to activate or
		// have already been activated.
		if err := b.warnUnknownRuleActivations(ib); err != nil {
			return err
		}
	}
	// Must be end node of sequence in dag
	// Generate a new best state snapshot that will be used to update the
	// database and later memory if all database updates are successful.
	lastState := b.BestSnapshot()

	for e := attachNodes.Front(); e != nil; e = e.Next() {
		b.bd.UpdateWeight(e.Value.(blockdag.IBlock))
	}

	// Calculate the number of transactions that would be added by adding
	// this block.
	numTxns := uint64(len(block.Block().Transactions))

	blockSize := uint64(block.Block().SerializeSize())

	mainTip := b.bd.GetMainChainTip()
	mainTipNode := b.GetBlockNode(mainTip)
	if mainTipNode == nil {
		return fmt.Errorf("No main tip node\n")
	}
	state := newBestState(mainTip.GetHash(), mainTipNode.Difficulty(), blockSize, numTxns, b.CalcPastMedianTime(mainTip), lastState.TotalTxns+numTxns,
		b.bd.GetMainChainTip().GetWeight(), b.bd.GetGraphState(), b.GetTokenTipHash())

	// Atomically insert info into the database.
	err := b.db.Update(func(dbTx database.Tx) error {
		// Update best block state.
		err := dbPutBestState(dbTx, state, pow.CalcWork(mainTipNode.Difficulty(), mainTipNode.Pow().GetPowType()))
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

	return b.bd.Commit()
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
func (b *BlockChain) connectBlock(node blockdag.IBlock, block *types.SerializedBlock, view *UtxoViewpoint, stxos []SpentTxOut) error {
	// Atomically insert info into the database.
	err := b.db.Update(func(dbTx database.Tx) error {
		// Update the utxo set using the state of the utxo view.  This
		// entails removing all of the utxos spent and adding the new
		// ones created by the block.
		err := dbPutUtxoView(dbTx, view)
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

	err = b.updateTokenState(node, block, false)
	if err != nil {
		return err
	}

	b.ChainUnlock()
	b.sendNotification(BlockConnected, []*types.SerializedBlock{block})
	b.ChainLock()
	return nil
}

// disconnectBlock handles disconnecting the passed node/block from the end of
// the main (best) chain.
//
// This function MUST be called with the chain state lock held (for writes).
func (b *BlockChain) disconnectBlock(block *types.SerializedBlock, view *UtxoViewpoint, stxos []SpentTxOut) error {
	// Calculate the exact subsidy produced by adding the block.
	err := b.db.Update(func(dbTx database.Tx) error {
		// Update the utxo set using the state of the utxo view.  This
		// entails restoring all of the utxos spent and removing the new
		// ones created by the block.
		err := dbPutUtxoView(dbTx, view)
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

	b.ChainUnlock()
	b.sendNotification(BlockDisconnected, block)
	b.ChainLock()
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

func (b *BlockChain) reorganizeChain(ib blockdag.IBlock, detachNodes *list.List, attachNodes *list.List, newBlock *types.SerializedBlock) error {
	oldBlocks := []*hash.Hash{}
	for e := detachNodes.Front(); e != nil; e = e.Next() {
		ob := e.Value.(*blockdag.BlockOrderHelp)
		oldBlocks = append(oldBlocks, ob.Block.GetHash())
	}

	b.ChainUnlock()
	b.sendNotification(Reorganization, &ReorganizationNotifyData{
		OldBlocks: oldBlocks,
		NewBlock:  newBlock.Hash(),
		NewOrder:  uint64(ib.GetOrder()),
	})
	b.ChainLock()
	// Why the old order is the order that was removed by the new block, because the new block
	// must be one of the tip of the dag.This is very important for the following understanding.
	// In the two case, the perspective is the same.In the other words, the future can not
	// affect the past.
	var block *types.SerializedBlock
	var err error

	for e := detachNodes.Back(); e != nil; e = e.Prev() {
		n := e.Value.(*blockdag.BlockOrderHelp)
		if n == nil {
			panic(err.Error())
		}
		b.updateTokenState(n.Block, nil, true)
		//
		block, err = b.fetchBlockByHash(n.Block.GetHash())
		if err != nil {
			panic(err.Error())
		}

		block.SetOrder(uint64(n.OldOrder))
		// Load all of the utxos referenced by the block that aren't
		// already in the view.
		var stxos []SpentTxOut
		view := NewUtxoViewpoint()
		view.SetViewpoints([]*hash.Hash{block.Hash()})
		if !n.Block.GetStatus().KnownInvalid() {
			b.CalculateDAGDuplicateTxs(block)
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
			err = view.disconnectTransactions(block, stxos, b)
			if err != nil {
				b.bd.InvalidBlock(n.Block)
				log.Info(fmt.Sprintf("%s", err))
			}
		}
		b.bd.ValidBlock(n.Block)

		//newn.FlushToDB(b)

		err = b.disconnectBlock(block, view, stxos)
		if err != nil {
			return err
		}
	}

	for e := attachNodes.Front(); e != nil; e = e.Next() {
		nodeBlock := e.Value.(blockdag.IBlock)
		if nodeBlock.GetID() == ib.GetID() {
			block = newBlock
		} else {
			// If any previous nodes in attachNodes failed validation,
			// mark this one as having an invalid ancestor.
			block, err = b.FetchBlockByHash(nodeBlock.GetHash())

			if err != nil {
				return err
			}
			block.SetOrder(uint64(nodeBlock.GetOrder()))
			block.SetHeight(nodeBlock.GetHeight())
		}
		if !nodeBlock.IsOrdered() {
			continue
		}
		view := NewUtxoViewpoint()
		view.SetViewpoints([]*hash.Hash{nodeBlock.GetHash()})
		stxos := []SpentTxOut{}
		err = b.checkConnectBlock(nodeBlock, block, view, &stxos)
		if err != nil {
			b.bd.InvalidBlock(nodeBlock)
			stxos = []SpentTxOut{}
			view.Clean()
			log.Info(fmt.Sprintf("%s", err))
		}
		err = b.connectBlock(nodeBlock, block, view, stxos)
		if err != nil {
			b.bd.InvalidBlock(nodeBlock)
			log.Info(fmt.Sprintf("%s", err))
			continue
		}
		if !nodeBlock.GetStatus().KnownInvalid() {
			b.bd.ValidBlock(nodeBlock)
		}

	}

	// Log the point where the chain forked and old and new best chain
	// heads.
	log.Debug(fmt.Sprintf("End DAG REORGANIZE: Old Len= %d;New Len= %d", attachNodes.Len(), detachNodes.Len()))

	return nil
}

// countSpentOutputs returns the number of utxos the passed block spends.
func (b *BlockChain) countSpentOutputs(block *types.SerializedBlock) int {
	// Exclude the coinbase transaction since it can't spend anything.
	var numSpent int
	for _, tx := range block.Transactions()[1:] {
		if tx.IsDuplicate {
			continue
		}
		if types.IsTokenTx(tx.Tx) {
			if types.IsTokenMintTx(tx.Tx) {
				numSpent--
			} else {
				continue
			}
		}
		numSpent += len(tx.Transaction().TxIn)

	}
	return numSpent
}

// Return the dag instance
func (b *BlockChain) BlockDAG() *blockdag.BlockDAG {
	return b.bd
}

// Return median time source
func (b *BlockChain) TimeSource() MedianTimeSource {
	return b.timeSource
}

// FetchSpendJournal can return the set of outputs spent for the target block.
func (b *BlockChain) FetchSpendJournal(targetBlock *types.SerializedBlock) ([]SpentTxOut, error) {
	b.ChainRLock()
	defer b.ChainRUnlock()

	return b.fetchSpendJournal(targetBlock)
}

func (b *BlockChain) fetchSpendJournal(targetBlock *types.SerializedBlock) ([]SpentTxOut, error) {
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

// expect priority
func (b *BlockChain) GetMiningTips(expectPriority int) []*hash.Hash {
	return b.BlockDAG().GetValidTips(expectPriority)
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

func (b *BlockChain) IsDuplicateTx(txid *hash.Hash, blockHash *hash.Hash) bool {
	err := b.db.Update(func(dbTx database.Tx) error {
		if b.indexManager != nil {
			if b.indexManager.IsDuplicateTx(dbTx, txid, blockHash) {
				return nil
			}
		}
		return fmt.Errorf("null")
	})
	return err == nil
}

func (b *BlockChain) CalculateDAGDuplicateTxs(block *types.SerializedBlock) {
	txs := block.Transactions()
	for _, tx := range txs {
		tx.IsDuplicate = b.IsDuplicateTx(tx.Hash(), block.Hash())
	}
}

func (b *BlockChain) CalculateFees(block *types.SerializedBlock) types.AmountMap {
	transactions := block.Transactions()
	totalAtomOut := types.AmountMap{}
	for i, tx := range transactions {
		if i == 0 || tx.Tx.IsCoinBase() || tx.IsDuplicate {
			continue
		}
		for _, txOut := range tx.Transaction().TxOut {
			totalAtomOut[txOut.Amount.Id] += int64(txOut.Amount.Value)
		}
	}
	spentTxos, err := b.fetchSpendJournal(block)
	if err != nil {
		return nil
	}
	totalAtomIn := types.AmountMap{}
	if spentTxos != nil {
		for _, st := range spentTxos {
			if transactions[st.TxIndex].IsDuplicate {
				continue
			}
			totalAtomIn[st.Amount.Id] += int64(st.Amount.Value + st.Fees.Value)
		}

		totalFees := types.AmountMap{}
		for _, coinId := range types.CoinIDList {
			totalFees[coinId] = totalAtomIn[coinId] - totalAtomOut[coinId]
			if totalFees[coinId] < 0 {
				totalFees[coinId] = 0
			}
		}
		return totalFees
	}
	return nil
}

// GetFees
func (b *BlockChain) GetFees(h *hash.Hash) types.AmountMap {
	ib := b.GetBlock(h)
	if ib == nil {
		return nil
	}
	if ib.GetStatus().KnownInvalid() {
		return nil
	}
	block, err := b.FetchBlockByHash(h)
	if err != nil {
		return nil
	}
	b.CalculateDAGDuplicateTxs(block)

	return b.CalculateFees(block)
}

func (b *BlockChain) GetFeeByCoinID(h *hash.Hash, coinId types.CoinID) int64 {
	fees := b.GetFees(h)
	if fees == nil {
		return 0
	}
	return fees[coinId]
}

func (b *BlockChain) CalcWeight(ib blockdag.IBlock, bi *blockdag.BlueInfo) int64 {
	if ib.GetStatus().KnownInvalid() {
		return 0
	}
	block, err := b.FetchBlockByHash(ib.GetHash())
	if err != nil {
		log.Error(fmt.Sprintf("CalcWeight:%v", err))
		return 0
	}
	if b.IsDuplicateTx(block.Transactions()[0].Hash(), ib.GetHash()) {
		return 0
	}
	return b.subsidyCache.CalcBlockSubsidy(bi)
}

func (b *BlockChain) CheckCacheInvalidTxConfig() error {
	if b.CacheInvalidTx {
		hasConfig := true
		b.db.View(func(dbTx database.Tx) error {
			meta := dbTx.Metadata()
			citData := meta.Get(dbnamespace.CacheInvalidTxName)
			if citData == nil {
				hasConfig = false
			}
			return nil
		})
		if hasConfig {
			return nil
		}
		return fmt.Errorf("You must use --droptxindex before you use --cacheinvalidtx.")
	}
	return nil
}

// Return chain params
func (b *BlockChain) ChainParams() *params.Params {
	return b.params
}

func (b *BlockChain) GetTokenTipHash() *hash.Hash {
	if uint(b.TokenTipID) == blockdag.MaxId {
		return nil
	}
	ib := b.bd.GetBlockById(uint(b.TokenTipID))
	if ib == nil {
		return nil
	}
	return ib.GetHash()
}

func (b *BlockChain) CalculateTokenStateRoot(txs []*types.Tx, parents []*hash.Hash) hash.Hash {
	updates := []token.ITokenUpdate{}
	for _, tx := range txs {
		if types.IsTokenTx(tx.Tx) {
			update, err := token.NewUpdateFromTx(tx.Tx)
			if err != nil {
				log.Error(err.Error())
				continue
			}
			updates = append(updates, update)
		}
	}
	if len(updates) <= 0 {
		if len(parents) <= 0 {
			return hash.ZeroHash
		}
		block, err := b.fetchBlockByHash(parents[0])
		if err != nil {
			return hash.ZeroHash
		}
		return block.Block().Header.StateRoot
	}
	balanceUpdate := []*hash.Hash{}
	for _, u := range updates {
		balanceUpdate = append(balanceUpdate, u.GetHash())
	}
	tsMerkle := merkle.BuildTokenBalanceMerkleTreeStore(balanceUpdate)

	return *tsMerkle[0]
}

func (b *BlockChain) getBlockData(hash *hash.Hash) blockdag.IBlockData {
	block, err := b.fetchBlockByHash(hash)
	if err != nil {
		log.Error(err.Error())
		return nil
	}
	return NewBlockNode(block, block.Block().Parents)
}

// CalcPastMedianTime calculates the median time of the previous few blocks
// prior to, and including, the block node.
//
// This function is safe for concurrent access.
func (b *BlockChain) CalcPastMedianTime(block blockdag.IBlock) time.Time {
	// Create a slice of the previous few block timestamps used to calculate
	// the median per the number defined by the constant medianTimeBlocks.
	timestamps := make([]int64, medianTimeBlocks)
	numNodes := 0
	iterBlock := block
	for i := 0; i < medianTimeBlocks && iterBlock != nil; i++ {
		iterNode := b.GetBlockNode(iterBlock)
		if iterNode == nil {
			break
		}
		timestamps[i] = iterNode.GetTimestamp()
		numNodes++

		iterBlock = b.bd.GetBlockById(iterBlock.GetMainParent())
	}

	// Prune the slice to the actual number of available timestamps which
	// will be fewer than desired near the beginning of the block chain
	// and sort them.
	timestamps = timestamps[:numNodes]
	sort.Sort(util.TimeSorter(timestamps))

	// NOTE: The consensus rules incorrectly calculate the median for even
	// numbers of blocks.  A true median averages the middle two elements
	// for a set with an even number of elements in it.   Since the constant
	// for the previous number of blocks to be used is odd, this is only an
	// issue for a few blocks near the beginning of the chain.  I suspect
	// this is an optimization even though the result is slightly wrong for
	// a few of the first blocks since after the first few blocks, there
	// will always be an odd number of blocks in the set per the constant.
	//
	// This code follows suit to ensure the same rules are used, however, be
	// aware that should the medianTimeBlocks constant ever be changed to an
	// even number, this code will be wrong.
	medianTimestamp := timestamps[numNodes/2]
	return time.Unix(medianTimestamp, 0)
}

func (b *BlockChain) GetSubsidyCache() *SubsidyCache {
	return b.subsidyCache
}
