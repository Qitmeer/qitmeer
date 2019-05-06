// Copyright (c) 2017-2018 The nox developers

package blkmgr

import (
	"sync"
	"time"
	"qitmeer/common/hash"
	"qitmeer/core/blockchain"
	"qitmeer/config"
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
