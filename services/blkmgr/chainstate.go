// Copyright (c) 2017-2018 The qitmeer developers

package blkmgr

import (
	"github.com/HalalChain/qitmeer-lib/common/hash"
	"sync"
	"time"
)

func (b *BlockManager) GetChainState() *ChainState{
	return &b.chainState
}

// updateChainState updates the chain state associated with the block manager.
// This allows fast access to chain information since blockchain is currently not
// safe for concurrent access and the block manager is typically quite busy
// processing block and inventory.
func (c *ChainState) UpdateChainState(newestHash *hash.Hash,
	newestHeight uint64, bestMedianTime time.Time) {

	c.Lock()
	defer c.Unlock()

	c.newestHash = newestHash
	c.newestHeight = newestHeight
	c.pastMedianTime = bestMedianTime
}

// chainState tracks the state of the best chain as blocks are inserted.  This
// is done because blockchain is currently not safe for concurrent access and the
// block manager is typically quite busy processing block and inventory.
// Therefore, requesting this information from chain through the block manager
// would not be anywhere near as efficient as simply updating it as each block
// is inserted and protecting it with a mutex.
type ChainState struct {
	sync.RWMutex
	newestHash          *hash.Hash
	newestHeight        uint64
	nextFinalState      [6]byte
	nextPoolSize        uint32
	nextStakeDifficulty int64
	winningTickets      []hash.Hash
	missedTickets       []hash.Hash
	pastMedianTime      time.Time
}


// headerNode is used as a node in a list of headers that are linked together
// between checkpoints.
type headerNode struct {
	height uint64
	hash   *hash.Hash
}
