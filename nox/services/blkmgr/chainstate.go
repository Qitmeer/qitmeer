// Copyright (c) 2017-2018 The nox developers

package blkmgr

import (
	"sync"
	"time"
	"github.com/noxproject/nox/common/hash"
	"github.com/noxproject/nox/core/blockchain"
	"github.com/noxproject/nox/config"
)

func (b *BlockManager) GetChainState() *ChainState{
	return &b.chainState
}

// updateChainState updates the chain state associated with the block manager.
// This allows fast access to chain information since blockchain is currently not
// safe for concurrent access and the block manager is typically quite busy
// processing block and inventory.
func (c *ChainState) UpdateChainState(newestHash *hash.Hash,
	newestHeight uint64, bestMedianTime time.Time, curPrevHash hash.Hash) {

	c.Lock()
	defer c.Unlock()

	c.newestHash = newestHash
	c.newestHeight = newestHeight
	c.pastMedianTime = bestMedianTime
	c.curPrevHash = curPrevHash
}

//TODO revisit concurrent lock/unclock
func (c *ChainState) GetPastMedianTime() time.Time{
	c.RLock()
	defer c.RUnlock()
	medianTime := c.pastMedianTime
	return medianTime
}

// MedianAdjustedTime returns the current time adjusted to ensure it is at least
// one second after the median timestamp of the last several blocks per the
// chain consensus rules.
func (c *ChainState) MedianAdjustedTime(timeSource blockchain.MedianTimeSource, cfg *config.Config) (time.Time, error) {
	c.RLock()
	defer c.RUnlock()

	// The timestamp for the block must not be before the median timestamp
	// of the last several blocks.  Thus, choose the maximum between the
	// current time and one second after the past median time.  The current
	// timestamp is truncated to a second boundary before comparison since a
	// block timestamp does not supported a precision greater than one
	// second.
	newTimestamp := timeSource.AdjustedTime()
	pastMedianTime := c.pastMedianTime
	minTimestamp := pastMedianTime.Add(time.Second)
	if newTimestamp.Before(minTimestamp) {
		newTimestamp = minTimestamp
	}
	//TODO, refactor the config dependence
	// Adjust by the amount requested from the command line argument.
	newTimestamp = newTimestamp.Add(
		time.Duration(-cfg.MiningTimeOffset) * time.Second)

	return newTimestamp, nil
}

//TODO revisit concurrent lock/unclock
func (c *ChainState) GetNextHeightWithState() (prevHash *hash.Hash, nextBlockHeight uint64, poolSize uint32,
	finalState [6]byte){
	c.RLock()
	defer c.RUnlock()
	prevHash = c.newestHash
	nextBlockHeight = c.newestHeight + 1
	poolSize = c.nextPoolSize
	finalState = c.nextFinalState
	return
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
	curPrevHash         hash.Hash
	pastMedianTime      time.Time
}

// Best returns the block hash and height known for the tip of the best known
// chain.
//
// This function is safe for concurrent access.
func (c *ChainState) Best() (*hash.Hash, uint64) {
	c.RLock()
	defer c.RUnlock()

	return c.newestHash, c.newestHeight
}



// headerNode is used as a node in a list of headers that are linked together
// between checkpoints.
type headerNode struct {
	height uint64
	hash   *hash.Hash
}
