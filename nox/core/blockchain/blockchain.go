// Copyright (c) 2017-2018 The nox developers
package blockchain

import (
	"sync"
	"github.com/noxproject/nox/common/hash"
	"github.com/noxproject/nox/core/types"
	"github.com/noxproject/nox/database"
	"github.com/noxproject/nox/params"
	"time"
)

// BlockChain provides functions such as rejecting duplicate blocks, ensuring
// blocks follow all rules, orphan handling, checkpoint handling, and best chain
// selection with reorganization.
type BlockChain struct {

	params         		*params.Params

	// The following fields are set when the instance is created and can't
	// be changed afterwards, so there is no need to protect them with a
	// separate mutex.
	checkpointsByHeight map[int64]*params.Checkpoint

	db                  database.DB

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
	mainNodesByHeight map[int64]*blockNode

	// These fields are related to handling of orphan blocks.  They are
	// protected by a combination of the chain lock and the orphan lock.
	orphanLock   sync.RWMutex
	orphans      map[hash.Hash]*orphanBlock
	prevOrphans  map[hash.Hash][]*orphanBlock
	oldestOrphan *orphanBlock

	// The block cache for mainchain blocks, to facilitate faster
	// reorganizations.
	mainchainBlockCacheLock sync.RWMutex
	mainchainBlockCache     map[hash.Hash]*types.Block
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
	Height       int64          // The height of the block.
	Bits         uint32         // The difficulty bits of the block.
	BlockSize    uint64         // The size of the block.
	NumTxns      uint64         // The number of txns in the block.
	TotalTxns    uint64         // The total number of txns in the chain.
	MedianTime   time.Time      // Median time as per CalcPastMedianTime.
	TotalSubsidy int64          // The total subsidy for the chain.
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
