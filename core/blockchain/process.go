// Copyright (c) 2013-2016 The btcsuite developers
// Copyright (c) 2015-2018 The Decred developers
// Use of this source code is governed by an ISC
// license that can be found in the LICENSE file.

package blockchain

import (
	"fmt"
	"time"
	"github.com/noxproject/nox/core/types"
	"github.com/noxproject/nox/common/hash"
)

// BehaviorFlags is a bitmask defining tweaks to the normal behavior when
// performing chain processing and consensus rules checks.
type BehaviorFlags uint32

const (
	// BFFastAdd may be set to indicate that several checks can be avoided
	// for the block since it is already known to fit into the chain due to
	// already proving it correct links into the chain up to a known
	// checkpoint.  This is primarily used for headers-first mode.
	BFFastAdd BehaviorFlags = 1 << iota

	// BFNoPoWCheck may be set to indicate the proof of work check which
	// ensures a block hashes to a value less than the required target will
	// not be performed.
	BFNoPoWCheck

	// BFNone is a convenience value to specifically indicate no flags.
	BFNone BehaviorFlags = 0
)

// processOrphans determines if there are any orphans which depend on the passed
// block hash (they are no longer orphans if true) and potentially accepts them.
// It repeats the process for the newly accepted blocks (to detect further
// orphans which may no longer be orphans) until there are no more.
//
// The flags do not modify the behavior of this function directly, however they
// are needed to pass along to maybeAcceptBlock.
//
// This function MUST be called with the chain state lock held (for writes).
func (b *BlockChain) processOrphans(h *hash.Hash, flags BehaviorFlags) error {
	for  {
		needLoop:=false
		for _,v:=range b.orphans{
			allExists:=true
			for _,h:=range v.block.Block().Parents{
				exists:= b.index.HaveBlock(h)
				if !exists {
					allExists=false
					break
				}
			}
			if allExists {

				//
				exists := b.index.HaveBlock(v.block.Hash())
				if exists {
					b.removeOrphanBlock(v)
					continue
				}
				// Potentially accept the block into the block chain.
				_, err := b.maybeAcceptBlock(v.block, flags)
				if err != nil {
					return err
				}
				b.removeOrphanBlock(v)
				needLoop=true
				break
			}
		}
		if !needLoop {
			return nil
		}
	}

	return nil
}

// ProcessBlock is the main workhorse for handling insertion of new blocks into
// the block chain.  It includes functionality such as rejecting duplicate
// blocks, ensuring blocks follow all rules, orphan handling, and insertion into
// the block chain along with best chain selection and reorganization.
//
// When no errors occurred during processing, the first return value indicates
// the length of the fork the block extended.  In the case it either exteneded
// the best chain or is now the tip of the best chain due to causing a
// reorganize, the fork length will be 0.  The second return value indicates
// whether or not the block is an orphan, in which case the fork length will
// also be zero as expected, because it, by definition, does not connect ot the
// best chain.
//
// This function is safe for concurrent access.
func (b *BlockChain) ProcessBlock(block *types.SerializedBlock, flags BehaviorFlags) (bool, bool, error) {
	b.chainLock.Lock()
	defer b.chainLock.Unlock()

	fastAdd := flags&BFFastAdd == BFFastAdd

	blockHash := block.Hash()
	log.Trace("Processing block ","hash", blockHash)
	currentTime := time.Now()
	defer func() {
		elapsedTime := time.Since(currentTime)
		log.Debug(fmt.Sprintf("Finished block processing"),
			"hash",blockHash,"height",block.Height(),"elapsed time", elapsedTime)
	}()

	// The block must not already exist in the main chain or side chains.
	if b.index.HaveBlock(blockHash) {
		str := fmt.Sprintf("already have block %v", blockHash)
		return false, false, ruleError(ErrDuplicateBlock, str)
	}

	// The block must not already exist as an orphan.
	if _, exists := b.orphans[*blockHash]; exists {
		str := fmt.Sprintf("already have block (orphan) %v", blockHash)
		return false, false, ruleError(ErrDuplicateBlock, str)
	}

	// Perform preliminary sanity checks on the block and its transactions.
	err := checkBlockSanity(block, b.timeSource, flags, b.params)
	if err != nil {
		return false, false, err
	}

	// Find the previous checkpoint and perform some additional checks based
	// on the checkpoint.  This provides a few nice properties such as
	// preventing old side chain blocks before the last checkpoint,
	// rejecting easy to mine, but otherwise bogus, blocks that could be
	// used to eat memory, and ensuring expected (versus claimed) proof of
	// work requirements since the previous checkpoint are met.
	blockHeader := &block.Block().Header
	checkpointNode, err := b.findPreviousCheckpoint()
	if err != nil {
		return false, false, err
	}
	if checkpointNode != nil {
		// Ensure the block timestamp is after the checkpoint timestamp.
		checkpointTime := time.Unix(checkpointNode.timestamp, 0)
		if blockHeader.Timestamp.Before(checkpointTime) {
			str := fmt.Sprintf("block %v has timestamp %v before "+
				"last checkpoint timestamp %v", blockHash,
				blockHeader.Timestamp, checkpointTime)
			return false, false, ruleError(ErrCheckpointTimeTooOld, str)
		}

		if !fastAdd {
			// Even though the checks prior to now have already ensured the
			// proof of work exceeds the claimed amount, the claimed amount
			// is a field in the block header which could be forged.  This
			// check ensures the proof of work is at least the minimum
			// expected based on elapsed time since the last checkpoint and
			// maximum adjustment allowed by the retarget rules.
			duration := blockHeader.Timestamp.Sub(checkpointTime)
			requiredTarget := CompactToBig(b.calcEasiestDifficulty(
				checkpointNode.bits, duration))
			currentTarget := CompactToBig(blockHeader.Difficulty)
			if currentTarget.Cmp(requiredTarget) > 0 {
				str := fmt.Sprintf("block target difficulty of %064x "+
					"is too low when compared to the previous "+
					"checkpoint", currentTarget)
				return false, false, ruleError(ErrDifficultyTooLow, str)
			}
		}
	}

	// Handle orphan blocks.
	for _,pb:=range block.Block().Parents{
		if !b.index.HaveBlock(pb) {
			log.Info("Adding orphan block %v with parent %v", blockHash,pb)
			b.addOrphanBlock(block)

			// The fork length of orphans is unknown since they, by definition, do
			// not connect to the best chain.
			return false, true, nil
		}
	}

	// The block has passed all context independent checks and appears sane
	// enough to potentially accept it into the block chain.
	_, err = b.maybeAcceptBlock(block, flags)
	if err != nil {
		return false, false, err
	}

	// Accept any orphan blocks that depend on this block (they are no
	// longer orphans) and repeat for those accepted blocks until there are
	// no more.
	err = b.processOrphans(blockHash, flags)
	if err != nil {
		return false, false, err
	}

	log.Debug("Accepted block", "hash", blockHash)

	return false, false, nil
}
