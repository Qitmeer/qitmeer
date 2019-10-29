// Copyright (c) 2017-2018 The qitmeer developers
// Copyright (c) 2013-2016 The btcsuite developers
// Copyright (c) 2015-2018 The Decred developers
// Use of this source code is governed by an ISC
// license that can be found in the LICENSE file.

package blockchain

import (
	"fmt"
	"github.com/Qitmeer/qitmeer/common/hash"
	`github.com/Qitmeer/qitmeer/core/types/pow`
	"math/big"
	"time"
)

// bigZero is 0 represented as a big.Int.  It is defined here to avoid
// the overhead of creating it multiple times.
var bigZero = big.NewInt(0)

// maxShift is the maximum shift for a difficulty that resets (e.g.
// testnet difficulty).
const maxShift = uint(256)

// calcEasiestDifficulty calculates the easiest possible difficulty that a block
// can have given starting difficulty bits and a duration.  It is mainly used to
// verify that claimed proof of work by a block is sane as compared to a
// known good checkpoint.
func (b *BlockChain) calcEasiestDifficulty(bits uint32, duration time.Duration,powInstance pow.IPow) uint32 {
	// Convert types used in the calculations below.
	durationVal := int64(duration)
	adjustmentFactor := big.NewInt(b.params.RetargetAdjustmentFactor)
	maxRetargetTimespan := int64(b.params.TargetTimespan) *
		b.params.RetargetAdjustmentFactor
	target := powInstance.GetSafeDiff(0)
	// The test network rules allow minimum difficulty blocks once too much
	// time has elapsed without mining a block.
	if b.params.ReduceMinDifficulty {
		if durationVal > int64(b.params.MinDiffReductionTime) {
			return pow.BigToCompact(target)
		}
	}

	// Since easier difficulty equates to higher numbers, the easiest
	// difficulty for a given duration is the largest value possible given
	// the number of retargets for the duration and starting difficulty
	// multiplied by the max adjustment factor.
	newTarget := pow.CompactToBig(bits)

	for durationVal > 0 && powInstance.CompareDiff(newTarget,target) {
		newTarget.Mul(newTarget, adjustmentFactor)
		newTarget = powInstance.GetNextDiffBig(adjustmentFactor,newTarget,big.NewInt(0))
		durationVal -= maxRetargetTimespan
	}

	// Limit new value to the proof of work limit.
	if !powInstance.CompareDiff(newTarget,target) {
		newTarget.Set(target)
	}

	return pow.BigToCompact(newTarget)
}

// findPrevTestNetDifficulty returns the difficulty of the previous block which
// did not have the special testnet minimum difficulty rule applied.
//
// This function MUST be called with the chain state lock held (for writes).
func (b *BlockChain) findPrevTestNetDifficulty(startNode *blockNode,powInstance pow.IPow) uint32 {
	// Search backwards through the chain for the last block without
	// the special rule applied.
	blocksPerRetarget := uint64(b.params.WorkDiffWindowSize *
		b.params.WorkDiffWindows)
	iterBlock:= b.bd.GetBlock(startNode.GetHash())
	var iterNode *blockNode
	target := powInstance.GetSafeDiff(0)
	for {
		if iterBlock == nil ||
			uint64(iterBlock.GetHeight())%blocksPerRetarget == 0{
			break
		}
		iterNode=b.index.lookupNode(iterBlock.GetHash())
		if iterNode.bits != pow.BigToCompact(target) {
			break
		}
	}
	// Return the found difficulty or the minimum difficulty if no
	// appropriate block was found.
	lastBits := pow.BigToCompact(target)
	if iterNode != nil {
		lastBits = iterNode.bits
	}
	return lastBits
}

// calcNextRequiredDifficulty calculates the required difficulty for the block
// after the passed previous block node based on the difficulty retarget rules.
// This function differs from the exported CalcNextRequiredDifficulty in that
// the exported version uses the current best chain as the previous block node
// while this function accepts any block node.
func (b *BlockChain) calcNextRequiredDifficulty(curNode *blockNode, newBlockTime time.Time,powInstance pow.IPow) (uint32, error) {
	baseTarget := powInstance.GetSafeDiff(0)
	originCurrentNode := curNode
	// Genesis block.
	if curNode == nil {
		return pow.BigToCompact(baseTarget), nil
	}

	curNode = b.getPowTypeNode(curNode,powInstance.GetPowType())
	curBlock:=b.bd.GetBlock(curNode.GetHash())
	if curBlock == nil{
		return pow.BigToCompact(baseTarget),nil
	}

	// Get the old difficulty; if we aren't at a block height where it changes,
	// just return this.
	oldDiff := curNode.bits
	oldDiffBig := pow.CompactToBig(curNode.bits)
	windowsSizeBig := big.NewInt(b.params.WorkDiffWindowSize)
	// percent is *100 * 2^32
	windowsSizeBig.Mul(windowsSizeBig,powInstance.PowPercent())
	windowsSizeBig.Div(windowsSizeBig,big.NewInt(100))
	windowsSizeBig.Rsh(windowsSizeBig,32)
	needAjustCount := int64(windowsSizeBig.Uint64())
	// We're not at a retarget point, return the oldDiff.
	if !b.needAjustPowDifficulty(curNode,powInstance.GetPowType(),needAjustCount) {
		// For networks that support it, allow special reduction of the
		// required difficulty once too much time has elapsed without
		// mining a block.
		if b.params.ReduceMinDifficulty {
			// Return minimum difficulty when more than the desired
			// amount of time has elapsed without mining a block.
			reductionTime := int64(b.params.MinDiffReductionTime /
				time.Second)
			allowMinTime := curNode.timestamp + reductionTime

			// For every extra target timespan that passes, we halve the
			// difficulty.
			if newBlockTime.Unix() > allowMinTime {
				timePassed := newBlockTime.Unix() - curNode.timestamp
				timePassed -= reductionTime
				shifts := uint((timePassed / int64(b.params.TargetTimePerBlock/
					time.Second)) + 1)

				// Scale the difficulty with time passed.
				oldTarget := pow.CompactToBig(curNode.bits)
				newTarget := new(big.Int)
				if shifts < maxShift {
					newTarget.Lsh(oldTarget, shifts)
				} else {
					newTarget.Set(pow.OneLsh256)
				}

				// Limit new value to the proof of work limit.
				if powInstance.CompareDiff(newTarget,baseTarget) {
					newTarget.Set(baseTarget)
				}

				return pow.BigToCompact(newTarget), nil
			}

			// The block was mined within the desired timeframe, so
			// return the difficulty for the last block which did
			// not have the special minimum difficulty rule applied.
			return b.findPrevTestNetDifficulty(curNode,powInstance), nil
		}

		return oldDiff, nil
	}
	// Declare some useful variables.
	RAFBig := big.NewInt(b.params.RetargetAdjustmentFactor)
	nextDiffBigMin := pow.CompactToBig(curNode.bits)
	nextDiffBigMin.Div(nextDiffBigMin, RAFBig)
	nextDiffBigMax := pow.CompactToBig(curNode.bits)
	nextDiffBigMax.Mul(nextDiffBigMax, RAFBig)

	alpha := b.params.WorkDiffAlpha

	// Number of nodes to traverse while calculating difficulty.
	nodesToTraverse := needAjustCount * b.params.WorkDiffWindows

	//calc pow block count in last nodesToTraverse blocks
	currentPowBlockCount := b.calcCurrentPowCount(originCurrentNode,nodesToTraverse,powInstance.GetPowType())

	// Initialize bigInt slice for the percentage changes for each window period
	// above or below the target.
	windowChanges := make([]*big.Int, b.params.WorkDiffWindows)

	// Regress through all of the previous blocks and store the percent changes
	// per window period; use bigInts to emulate 64.32 bit fixed point.
	var olderTime, windowPeriod int64
	var weights uint64
	oldNode := curNode
	recentTime := curNode.timestamp
	for i := uint64(0); ; i++ {
		// Store and reset after reaching the end of every window period.
		if i%uint64(needAjustCount) == 0 && i != 0 {
			olderTime = oldNode.timestamp
			timeDifference := recentTime - olderTime
			// Just assume we're at the target (no change) if we've
			// gone all the way back to the genesis block.
			if oldNode.order == 0 {
				timeDifference = int64(b.params.TargetTimespan /
					time.Second)
			}
			timeDifBig := big.NewInt(timeDifference)
			timeDifBig.Lsh(timeDifBig, 32) // Add padding
			targetTemp := big.NewInt(int64(b.params.TargetTimespan /
				time.Second))
			windowAdjusted := targetTemp.Div(timeDifBig, targetTemp)

			// Weight it exponentially. Be aware that this could at some point
			// overflow if alpha or the number of blocks used is really large.
			windowAdjusted = windowAdjusted.Lsh(windowAdjusted,
				uint((b.params.WorkDiffWindows-windowPeriod)*alpha))

			// Sum up all the different weights incrementally.
			weights += 1 << uint64((b.params.WorkDiffWindows-windowPeriod)*
				alpha)

			// Store it in the slice.
			windowChanges[windowPeriod] = windowAdjusted

			windowPeriod++

			recentTime = olderTime
		}

		if i == uint64(nodesToTraverse) {
			break // Exit for loop when we hit the end.
		}
		// Get the previous node while staying at the genesis block as
		// needed.
		if oldNode.parents != nil {
			oldBlock:=b.bd.GetBlock(oldNode.GetHash())
			oldMainParent:=b.bd.GetBlock(oldBlock.GetMainParent())
			if oldMainParent != nil {
				oldNode=b.index.lookupNode(oldMainParent.GetHash())
				oldNode = b.getPowTypeNode(oldNode,powInstance.GetPowType())
			}
		}
	}
	// Sum up the weighted window periods.
	weightedSum := big.NewInt(0)
	for i := int64(0); i < b.params.WorkDiffWindows; i++ {
		weightedSum.Add(weightedSum, windowChanges[i])
	}

	// Divide by the sum of all weights.
	weightsBig := big.NewInt(int64(weights))
	weightedSumDiv := weightedSum.Div(weightedSum, weightsBig)
	// if current pow count is zero , set 1 min 1
	if currentPowBlockCount <= 0{
		currentPowBlockCount = 1
	}
	//percent calculate
	currentPowPercent := big.NewInt(int64(currentPowBlockCount))
	currentPowPercent.Lsh(currentPowPercent,32)
	nodesToTraverseBig := big.NewInt(nodesToTraverse)
	currentPowPercent = currentPowPercent.Div(currentPowPercent,nodesToTraverseBig)
	// Multiply by the old diff.
	nextDiffBig := powInstance.GetNextDiffBig(weightedSumDiv,oldDiffBig,currentPowPercent)
	// Right shift to restore the original padding (restore non-fixed point).
	// Check to see if we're over the limits for the maximum allowable retarget;
	// if we are, return the maximum or minimum except in the case that oldDiff
	// is zero.
	if oldDiffBig.Cmp(bigZero) == 0 { // This should never really happen,
		nextDiffBig.Set(nextDiffBig) // but in case it does...
	} else if nextDiffBig.Cmp(bigZero) == 0 {
		nextDiffBig.Set(baseTarget)
	} else if nextDiffBig.Cmp(nextDiffBigMax) == 1 {
		nextDiffBig.Set(nextDiffBigMax)
	} else if nextDiffBig.Cmp(nextDiffBigMin) == -1 {
		nextDiffBig.Set(nextDiffBigMin)
	}

	// Limit new value to the proof of work limit.
	if !powInstance.CompareDiff(nextDiffBig,baseTarget) {
		nextDiffBig.Set(baseTarget)
	}
	// Log new target difficulty and return it.  The new target logging is
	// intentionally converting the bits back to a number instead of using
	// newTarget since conversion to the compact representation loses
	// precision.
	nextDiffBits := pow.BigToCompact(nextDiffBig)
	log.Debug("Difficulty retarget", "block main height", curBlock.GetHeight()+1)
	log.Debug("Old target", "bits",fmt.Sprintf("%08x", curNode.bits),
		"diff",fmt.Sprintf( "(%064x)",oldDiffBig))
	log.Debug("New target", "bits",fmt.Sprintf("%08x", nextDiffBits),
		"diff",fmt.Sprintf( "(%064x)",nextDiffBig))

	return nextDiffBits, nil
}

// stats current pow count in nodesToTraverse
func (b *BlockChain) calcCurrentPowCount(curNode *blockNode, nodesToTraverse int64,powType pow.PowType) int64 {
	// Genesis block.
	if curNode == nil {
		return 0
	}
	currentPowBlockCount := nodesToTraverse
	// Regress through all of the previous blocks and store the percent changes
	// per window period; use bigInts to emulate 64.32 bit fixed point.
	oldNode := curNode
	for i := int64(0) ;i < nodesToTraverse ; i++ {
		// Get the previous node while staying at the genesis block as
		// needed.
		if oldNode.order == 0 {
			currentPowBlockCount--
		}
		if oldNode.parents != nil {
			oldBlock:=b.bd.GetBlock(oldNode.GetHash())
			oldMainParent:=b.bd.GetBlock(oldBlock.GetMainParent())
			if oldMainParent != nil {
				oldNode=b.index.lookupNode(oldMainParent.GetHash())
				if oldNode.order != 0 && oldNode.GetPowType() != powType{
					currentPowBlockCount--
				}
			}
		}
	}

	return currentPowBlockCount
}

// whether need ajust Pow Difficulty
// recent b.params.WorkDiffWindowSize blocks
// if current count arrived target block count . need ajustment difficulty
func (b *BlockChain) needAjustPowDifficulty(curNode *blockNode, powType pow.PowType,needAjustCount int64) bool {
	countFromLastAdjustment := b.getDistanceFromLastAdjustment(curNode,powType,needAjustCount)
	// countFromLastAdjustment stats b.params.WorkDiffWindows Multiple count
	countFromLastAdjustment /= b.params.WorkDiffWindows
	return countFromLastAdjustment > 0 && countFromLastAdjustment % needAjustCount == 0
}

// Distance block count from last adjustment
func (b *BlockChain) getDistanceFromLastAdjustment(curNode *blockNode, powType pow.PowType,needAjustCount int64) int64 {
	if curNode == nil{
		return 0
	}
	//calculate
	oldBits := curNode.bits
	count := int64(0)
	currentTime := curNode.timestamp
	for{
		if curNode.pow.GetPowType() == powType{
			if oldBits != curNode.bits{
				return count
			}
			count++
		}
		if curNode.order == 0{
			//geniess block
			return count
		}
		// if TargetTimespan have only one pow block need ajustment difficulty
		// or count >= needAjustCount
		if ( count > 1 && currentTime - curNode.timestamp > (count-1)*int64(b.params.TargetTimespan /time.Second) ) ||
			count >= needAjustCount{
			return needAjustCount * b.params.WorkDiffWindows
		}
		if curNode.parents == nil{
			return count
		}

		oldBlock:=b.bd.GetBlock(curNode.GetHash())
		oldMainParent :=b.bd.GetBlock(oldBlock.GetMainParent())
		if oldMainParent != nil {
			curNode = b.index.lookupNode(oldMainParent.GetHash())
		} else{
			return count
		}
	}
}

// CalcNextRequiredDiffFromNode calculates the required difficulty for the block
// given with the passed hash along with the given timestamp.
//
// This function is NOT safe for concurrent access.
func (b *BlockChain) CalcNextRequiredDiffFromNode(hash *hash.Hash, timestamp time.Time,powType pow.PowType) (uint32, error) {
	node := b.index.LookupNode(hash)
	if node == nil {
		return 0, fmt.Errorf("block %s is not known", hash)
	}
	instance := pow.GetInstance(powType,0,[]byte{})
	instance.SetParams(b.params.PowConfig)
	instance.SetHeight(int64(node.height+1))
	return b.calcNextRequiredDifficulty(node, timestamp,instance)
}

// CalcNextRequiredDifficulty calculates the required difficulty for the block
// after the end of the current best chain based on the difficulty retarget
// rules.
//
// This function is safe for concurrent access.
func (b *BlockChain) CalcNextRequiredDifficulty(timestamp time.Time,powType pow.PowType) (uint32, error) {
	b.chainLock.Lock()
	block:=b.bd.GetMainChainTip()
	node:=b.index.lookupNode(block.GetHash())
	instance := pow.GetInstance(powType,0,[]byte{})
	instance.SetParams(b.params.PowConfig)
	instance.SetHeight(int64(node.height+1))
	difficulty, err := b.calcNextRequiredDifficulty(node, timestamp,instance)
	b.chainLock.Unlock()
	return difficulty, err
}

// find block node by pow type
func (b *BlockChain) getPowTypeNode(curNode *blockNode, powType pow.PowType) *blockNode {
	for{
		if curNode.pow.GetPowType() == powType{
			return curNode
		}

		if curNode.parents == nil{
			return &blockNode{
				pow:pow.GetInstance(powType,0,[]byte{}),
			}
		}

		oldBlock:=b.bd.GetBlock(curNode.GetHash())
		oldMainParent :=b.bd.GetBlock(oldBlock.GetMainParent())
		if oldMainParent != nil {
			curNode = b.index.lookupNode(oldMainParent.GetHash())
		} else{
			return &blockNode{
				pow:pow.GetInstance(powType,0,[]byte{}),
			}
		}
	}
}