// Copyright (c) 2017-2018 The qitmeer developers
// Copyright (c) 2013-2016 The btcsuite developers
// Copyright (c) 2015-2016 The Decred developers
// Use of this source code is governed by an ISC
// license that can be found in the LICENSE file.

package blockchain

import (
	"fmt"
	"github.com/Qitmeer/qng-core/common/hash"
	"github.com/Qitmeer/qng-core/meerdag"
	"github.com/Qitmeer/qng-core/core/types"
	"github.com/Qitmeer/qng-core/engine/txscript"
	"github.com/Qitmeer/qng-core/params"
	"time"
)

// CheckpointConfirmations is the number of blocks before the end of the current
// best block chain that a good checkpoint candidate must be.
const CheckpointConfirmations = 4096

// DisableCheckpoints provides a mechanism to disable validation against
// checkpoints which you DO NOT want to do in production.  It is provided only
// for debug purposes.
//
// This function is safe for concurrent access.
func (b *BlockChain) DisableCheckpoints(disable bool) {
	b.ChainLock()
	b.noCheckpoints = disable
	b.ChainUnlock()
}

// Checkpoints returns a slice of checkpoints (regardless of whether they are
// already known).  When checkpoints are disabled or there are no checkpoints
// for the active network, it will return nil.
//
// This function is safe for concurrent access.
func (b *BlockChain) Checkpoints() []params.Checkpoint {
	if !b.HasCheckpoints() {
		return nil
	}
	return b.params.Checkpoints
}

func (b *BlockChain) HasCheckpoints() bool {
	if b.noCheckpoints {
		return false
	}
	return len(b.params.Checkpoints) > 0
}

// LatestCheckpoint returns the most recent checkpoint (regardless of whether it
// is already known).  When checkpoints are disabled or there are no checkpoints
// for the active network, it will return nil.
//
// This function is safe for concurrent access.
func (b *BlockChain) LatestCheckpoint() *params.Checkpoint {
	if !b.HasCheckpoints() {
		return nil
	}
	checkpoints := b.params.Checkpoints
	return &checkpoints[len(checkpoints)-1]
}

// verifyCheckpoint returns whether the passed block layer and hash combination
// match the hard-coded checkpoint data.  It also returns true if there is no
// checkpoint data for the passed block height.
func (b *BlockChain) verifyCheckpoint(layer uint64, hash *hash.Hash) bool {
	if !b.HasCheckpoints() {
		return true
	}

	// Nothing to check if there is no checkpoint data for the block height.
	checkpoint, exists := b.checkpointsByLayer[layer]
	if !exists {
		return true
	}

	if !checkpoint.Hash.IsEqual(hash) {
		return false
	}

	log.Info(fmt.Sprintf("Verified checkpoint at layer %d/block %s", checkpoint.Layer,
		checkpoint.Hash))
	return true
}

// findPreviousCheckpoint finds the most recent checkpoint that is already
// available in the downloaded portion of the block chain and returns the
// associated block node.  It returns nil if a checkpoint can't be found (this
// should really only happen for blocks before the first checkpoint).
//
// This function MUST be called with the chain lock held (for reads).
func (b *BlockChain) findPreviousCheckpoint() (meerdag.IBlock, error) {
	if !b.HasCheckpoints() {
		return nil, nil
	}

	// Perform the initial search to find and cache the latest known
	// checkpoint if the best chain is not known yet or we haven't already
	// previously searched.
	checkpoints := b.params.Checkpoints
	numCheckpoints := len(checkpoints)
	if b.checkpointNode == nil && b.nextCheckpoint == nil {
		// Loop backwards through the available checkpoints to find one
		// that is already available.
		for i := numCheckpoints - 1; i >= 0; i-- {
			node := b.bd.GetBlock(checkpoints[i].Hash)
			if node == nil {
				continue
			}

			// Checkpoint found.  Cache it for future lookups and
			// set the next expected checkpoint accordingly.
			b.checkpointNode = node
			if i < numCheckpoints-1 {
				b.nextCheckpoint = &checkpoints[i+1]
			}
			return b.checkpointNode, nil
		}

		// No known latest checkpoint.  This will only happen on blocks
		// before the first known checkpoint.  So, set the next expected
		// checkpoint to the first checkpoint and return the fact there
		// is no latest known checkpoint block.
		b.nextCheckpoint = &checkpoints[0]
		return nil, nil
	}

	// At this point we've already searched for the latest known checkpoint,
	// so when there is no next checkpoint, the current checkpoint lockin
	// will always be the latest known checkpoint.
	if b.nextCheckpoint == nil {
		return b.checkpointNode, nil
	}

	// When there is a next checkpoint and the layer of the current best
	// chain does not exceed it, the current checkpoint lockin is still
	// the latest known checkpoint.
	if uint64(b.bd.GetMainChainTip().GetLayer()) < b.nextCheckpoint.Layer {
		return b.checkpointNode, nil
	}

	// We've reached or exceeded the next checkpoint height.  Note that
	// once a checkpoint lockin has been reached, forks are prevented from
	// any blocks before the checkpoint, so we don't have to worry about the
	// checkpoint going away out from under us due to a chain reorganize.

	// Cache the latest known checkpoint for future lookups.  Note that if
	// this lookup fails something is very wrong since the chain has already
	// passed the checkpoint which was verified as accurate before inserting
	// it.
	checkpointNode := b.bd.GetBlock(b.nextCheckpoint.Hash)
	if checkpointNode == nil {
		return nil, AssertError(fmt.Sprintf("findPreviousCheckpoint "+
			"failed lookup of known good block node %s",
			b.nextCheckpoint.Hash))
	}
	b.checkpointNode = checkpointNode

	// Set the next expected checkpoint.
	checkpointIndex := -1
	for i := numCheckpoints - 1; i >= 0; i-- {
		if checkpoints[i].Hash.IsEqual(b.nextCheckpoint.Hash) {
			checkpointIndex = i
			break
		}
	}
	b.nextCheckpoint = nil
	if checkpointIndex != -1 && checkpointIndex < numCheckpoints-1 {
		b.nextCheckpoint = &checkpoints[checkpointIndex+1]
	}

	return b.checkpointNode, nil
}

// isNonstandardTransaction determines whether a transaction contains any
// scripts which are not one of the standard types.
func isNonstandardTransaction(tx *types.Tx) bool {
	// Check all of the output public key scripts for non-standard scripts.
	for _, txOut := range tx.Transaction().TxOut {
		//TODO, hardcoded version dependence
		scriptClass := txscript.GetScriptClass(0, txOut.PkScript)
		if scriptClass == txscript.NonStandardTy {
			return true
		}
	}
	return false
}

// IsCheckpointCandidate returns whether or not the passed block is a good
// checkpoint candidate.
//
// The factors used to determine a good checkpoint are:
//  - The block must be in the main chain
//  - The block must be at least 'CheckpointConfirmations' blocks prior to the
//    current end of the main chain
//  - The timestamps for the blocks before and after the checkpoint must have
//    timestamps which are also before and after the checkpoint, respectively
//    (due to the median time allowance this is not always the case)
//  - The block must not contain any strange transaction such as those with
//    nonstandard scripts
//
// The intent is that candidates are reviewed by a developer to make the final
// decision and then manually added to the list of checkpoints for a network.
//
// This function is safe for concurrent access.
func (b *BlockChain) IsCheckpointCandidate(preBlock, block meerdag.IBlock) (bool, error) {
	b.ChainRLock()
	defer b.ChainRUnlock()

	if preBlock.GetHash().IsEqual(block.GetHash()) {
		return false, nil
	}

	// A checkpoint must be at least CheckpointConfirmations blocks
	// before the end of the main chain.
	mainChainLayer := b.BlockDAG().GetMainChainTip().GetLayer()
	if block.GetLayer() > (mainChainLayer - CheckpointConfirmations) {
		return false, nil
	}

	// A checkpoint must be have at least one block after it.
	//
	// This should always succeed since the check above already made sure it
	// is CheckpointConfirmations back, but be safe in case the constant
	// changes.
	if block == nil {
		return false, nil
	}
	nextBlockH := block.GetMainParent()
	if nextBlockH == meerdag.MaxId {
		return false, nil
	}
	nextBlock := b.BlockDAG().GetBlockById(nextBlockH)
	if nextBlock == nil {
		return false, nil
	}
	nextNode := b.GetBlockNode(nextBlock)
	if nextNode == nil {
		return false, nil
	}

	preNode := b.GetBlockNode(preBlock)
	if preNode == nil {
		return false, nil
	}

	node := b.GetBlockNode(block)
	if node == nil {
		return false, nil
	}

	// A checkpoint must have timestamps for the block and the blocks on
	// either side of it in order (due to the median time allowance this is
	// not always the case).
	prevTime := time.Unix(nextNode.GetTimestamp(), 0)
	curTime := time.Unix(node.GetTimestamp(), 0)
	nextTime := time.Unix(preNode.GetTimestamp(), 0)
	if prevTime.After(curTime) || nextTime.Before(curTime) {
		return false, nil
	}

	// A checkpoint must have transactions that only contain standard
	// scripts.
	serblock, err := b.fetchBlockByHash(block.GetHash())
	if err != nil {
		return false, err
	}
	for _, tx := range serblock.Transactions() {
		if isNonstandardTransaction(tx) {
			return false, nil
		}
	}

	return b.BlockDAG().IsHourglass(block.GetID()), nil
}
