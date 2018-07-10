// Copyright (c) 2017-2018 The nox developers

package blkmgr

import (
	"sync"
	"time"
	"github.com/noxproject/nox/common/hash"
)

func (b *BlockManager) GetChainState() *ChainState{
	return &b.chainState
}

func (c *ChainState) GetPastMedianTime() time.Time{
	c.Lock()
	medianTime := c.pastMedianTime
	c.Unlock()
	return medianTime
}

func (c *ChainState) GetNextHeightWithState() (prevHash *hash.Hash, nextBlockHeight uint64, poolSize uint32,
	finalState [6]byte){
	c.Lock()
	prevHash = c.newestHash
	nextBlockHeight = c.newestHeight + 1
	poolSize = c.nextPoolSize
	finalState = c.nextFinalState
	c.Unlock()
	return
}

// chainState tracks the state of the best chain as blocks are inserted.  This
// is done because blockchain is currently not safe for concurrent access and the
// block manager is typically quite busy processing block and inventory.
// Therefore, requesting this information from chain through the block manager
// would not be anywhere near as efficient as simply updating it as each block
// is inserted and protecting it with a mutex.
type ChainState struct {
	sync.Mutex
	newestHash          *hash.Hash
	newestHeight        uint64
	nextFinalState      [6]byte
	nextPoolSize        uint32
	nextStakeDifficulty int64
	winningTickets      []hash.Hash
	missedTickets       []hash.Hash
	curPrevHash         hash.Hash
	pastMedianTime      time.Time
}



// headerNode is used as a node in a list of headers that are linked together
// between checkpoints.
type headerNode struct {
	height uint64
	hash   *hash.Hash
}
