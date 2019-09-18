// Copyright (c) 2017-2018 The qitmeer developers
// Copyright (c) 2013-2016 The btcsuite developers
// Copyright (c) 2015-2018 The Decred developers
// Use of this source code is governed by an ISC
// license that can be found in the LICENSE file.

package blockchain

import (
	"fmt"
	"math/big"
	"time"
	"github.com/Qitmeer/qitmeer-lib/common/hash"
)

var (
	// bigZero is 0 represented as a big.Int.  It is defined here to avoid
	// the overhead of creating it multiple times.
	bigZero = big.NewInt(0)

	// bigOne is 1 represented as a big.Int.  It is defined here to avoid
	// the overhead of creating it multiple times.
	bigOne = big.NewInt(1)

	// oneLsh256 is 1 shifted left 256 bits.  It is defined here to avoid
	// the overhead of creating it multiple times.
	oneLsh256 = new(big.Int).Lsh(bigOne, 256)
)

// maxShift is the maximum shift for a difficulty that resets (e.g.
// testnet difficulty).
const maxShift = uint(256)

// HashToBig converts a hash.Hash into a big.Int that can be used to
// perform math comparisons.
func HashToBig(hash *hash.Hash) *big.Int {
	// A Hash is in little-endian, but the big package wants the bytes in
	// big-endian, so reverse them.
	buf := *hash
	blen := len(buf)
	for i := 0; i < blen/2; i++ {
		buf[i], buf[blen-1-i] = buf[blen-1-i], buf[i]
	}

	return new(big.Int).SetBytes(buf[:])
}

// CompactToBig converts a compact representation of a whole number N to an
// unsigned 32-bit number.  The representation is similar to IEEE754 floating
// point numbers.
//
// Like IEEE754 floating point, there are three basic components: the sign,
// the exponent, and the mantissa.  They are broken out as follows:
//
//	* the most significant 8 bits represent the unsigned base 256 exponent
// 	* bit 23 (the 24th bit) represents the sign bit
//	* the least significant 23 bits represent the mantissa
//
//	-------------------------------------------------
//	|   Exponent     |    Sign    |    Mantissa     |
//	-------------------------------------------------
//	| 8 bits [31-24] | 1 bit [23] | 23 bits [22-00] |
//	-------------------------------------------------
//
// The formula to calculate N is:
// 	N = (-1^sign) * mantissa * 256^(exponent-3)
//
// This compact form is only used to encode unsigned 256-bit numbers which
// represent difficulty targets, thus there really is not a need for a sign
// bit, but it is implemented here to stay consistent with bitcoind.
// TODO, revisit the compact difficulty form design
func CompactToBig(compact uint32) *big.Int {
	// Extract the mantissa, sign bit, and exponent.
	mantissa := compact & 0x007fffff
	isNegative := compact&0x00800000 != 0
	exponent := uint(compact >> 24)

	// Since the base for the exponent is 256, the exponent can be treated
	// as the number of bytes to represent the full 256-bit number.  So,
	// treat the exponent as the number of bytes and shift the mantissa
	// right or left accordingly.  This is equivalent to:
	// N = mantissa * 256^(exponent-3)
	var bn *big.Int
	if exponent <= 3 {
		mantissa >>= 8 * (3 - exponent)
		bn = big.NewInt(int64(mantissa))
	} else {
		bn = big.NewInt(int64(mantissa))
		bn.Lsh(bn, 8*(exponent-3))
	}

	// Make it negative if the sign bit is set.
	if isNegative {
		bn = bn.Neg(bn)
	}

	return bn
}

// BigToCompact converts a whole number N to a compact representation using
// an unsigned 32-bit number.  The compact representation only provides 23 bits
// of precision, so values larger than (2^23 - 1) only encode the most
// significant digits of the number.  See CompactToBig for details.
func BigToCompact(n *big.Int) uint32 {
	// No need to do any work if it's zero.
	if n.Sign() == 0 {
		return 0
	}

	// Since the base for the exponent is 256, the exponent can be treated
	// as the number of bytes.  So, shift the number right or left
	// accordingly.  This is equivalent to:
	// mantissa = mantissa / 256^(exponent-3)
	var mantissa uint32
	exponent := uint(len(n.Bytes()))
	if exponent <= 3 {
		mantissa = uint32(n.Bits()[0])
		mantissa <<= 8 * (3 - exponent)
	} else {
		// Use a copy to avoid modifying the caller's original number.
		tn := new(big.Int).Set(n)
		mantissa = uint32(tn.Rsh(tn, 8*(exponent-3)).Bits()[0])
	}

	// When the mantissa already has the sign bit set, the number is too
	// large to fit into the available 23-bits, so divide the number by 256
	// and increment the exponent accordingly.
	if mantissa&0x00800000 != 0 {
		mantissa >>= 8
		exponent++
	}

	// Pack the exponent, sign bit, and mantissa into an unsigned 32-bit
	// int and return it.
	compact := uint32(exponent<<24) | mantissa
	if n.Sign() < 0 {
		compact |= 0x00800000
	}
	return compact
}

// CalcWork calculates a work value from difficulty bits. it increases the difficulty
// for generating a block by decreasing the value which the generated hash must be
// less than.
//
// This difficulty target is stored in each block header using a compact
// representation as described in the documentation for CompactToBig.
//
// The main chain is selected by choosing the chain that has the most proof of
// work (highest difficulty).
//
// Since a lower target difficulty value equates to higher actual difficulty, the
// work value which will be accumulated must be the inverse of the difficulty.
// Also, in order to avoid potential division by zero and really small floating
// point numbers, the result adds 1 to the denominator and multiplies the numerator
// by 2^256.
func CalcWork(bits uint32) *big.Int {
	// Return a work value of zero if the passed difficulty bits represent
	// a negative number. Note this should not happen in practice with valid
	// blocks, but an invalid block could trigger it.
	difficultyNum := CompactToBig(bits)
	if difficultyNum.Sign() <= 0 {
		return big.NewInt(0)
	}

	// (1 << 256) / (difficultyNum + 1)
	denominator := new(big.Int).Add(difficultyNum, bigOne)
	return new(big.Int).Div(oneLsh256, denominator)
}

// calcEasiestDifficulty calculates the easiest possible difficulty that a block
// can have given starting difficulty bits and a duration.  It is mainly used to
// verify that claimed proof of work by a block is sane as compared to a
// known good checkpoint.
func (b *BlockChain) calcEasiestDifficulty(bits uint32, duration time.Duration) uint32 {
	// Convert types used in the calculations below.
	durationVal := int64(duration)
	adjustmentFactor := big.NewInt(b.params.RetargetAdjustmentFactor)
	maxRetargetTimespan := int64(b.params.TargetTimespan) *
		b.params.RetargetAdjustmentFactor

	// The test network rules allow minimum difficulty blocks once too much
	// time has elapsed without mining a block.
	if b.params.ReduceMinDifficulty {
		if durationVal > int64(b.params.MinDiffReductionTime) {
			return b.params.PowLimitBits
		}
	}

	// Since easier difficulty equates to higher numbers, the easiest
	// difficulty for a given duration is the largest value possible given
	// the number of retargets for the duration and starting difficulty
	// multiplied by the max adjustment factor.
	newTarget := CompactToBig(bits)
	for durationVal > 0 && newTarget.Cmp(b.params.PowLimit) < 0 {
		newTarget.Mul(newTarget, adjustmentFactor)
		durationVal -= maxRetargetTimespan
	}

	// Limit new value to the proof of work limit.
	if newTarget.Cmp(b.params.PowLimit) > 0 {
		newTarget.Set(b.params.PowLimit)
	}

	return BigToCompact(newTarget)
}

// findPrevTestNetDifficulty returns the difficulty of the previous block which
// did not have the special testnet minimum difficulty rule applied.
//
// This function MUST be called with the chain state lock held (for writes).
func (b *BlockChain) findPrevTestNetDifficulty(startNode *blockNode) uint32 {
	// Search backwards through the chain for the last block without
	// the special rule applied.
	blocksPerRetarget := uint64(b.params.WorkDiffWindowSize *
		b.params.WorkDiffWindows)
	iterBlock:= b.bd.GetBlock(startNode.GetHash())
	var iterNode *blockNode
	for {
		if iterBlock == nil ||
			uint64(iterBlock.GetHeight())%blocksPerRetarget == 0{
			break
		}
		iterNode=b.index.lookupNode(iterBlock.GetHash())
		if iterNode.bits != b.params.PowLimitBits {
			break
		}
	}
	// Return the found difficulty or the minimum difficulty if no
	// appropriate block was found.
	lastBits := b.params.PowLimitBits
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
func (b *BlockChain) calcNextRequiredDifficulty(curNode *blockNode, newBlockTime time.Time) (uint32, error) {
	// Genesis block.
	if curNode == nil {
		return b.params.PowLimitBits, nil
	}
	// Get the old difficulty; if we aren't at a block height where it changes,
	// just return this.
	oldDiff := curNode.bits
	oldDiffBig := CompactToBig(curNode.bits)
	curBlock:=b.bd.GetBlock(curNode.GetHash())
	// We're not at a retarget point, return the oldDiff.
	if int64(curBlock.GetHeight()+1)%b.params.WorkDiffWindowSize != 0 {
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
				oldTarget := CompactToBig(curNode.bits)
				newTarget := new(big.Int)
				if shifts < maxShift {
					newTarget.Lsh(oldTarget, shifts)
				} else {
					newTarget.Set(oneLsh256)
				}

				// Limit new value to the proof of work limit.
				if newTarget.Cmp(b.params.PowLimit) > 0 {
					newTarget.Set(b.params.PowLimit)
				}

				return BigToCompact(newTarget), nil
			}

			// The block was mined within the desired timeframe, so
			// return the difficulty for the last block which did
			// not have the special minimum difficulty rule applied.
			return b.findPrevTestNetDifficulty(curNode), nil
		}

		return oldDiff, nil
	}

	// Declare some useful variables.
	RAFBig := big.NewInt(b.params.RetargetAdjustmentFactor)
	nextDiffBigMin := CompactToBig(curNode.bits)
	nextDiffBigMin.Div(nextDiffBigMin, RAFBig)
	nextDiffBigMax := CompactToBig(curNode.bits)
	nextDiffBigMax.Mul(nextDiffBigMax, RAFBig)

	alpha := b.params.WorkDiffAlpha

	// Number of nodes to traverse while calculating difficulty.
	nodesToTraverse := (b.params.WorkDiffWindowSize *
		b.params.WorkDiffWindows)

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
		if i%uint64(b.params.WorkDiffWindowSize) == 0 && i != 0 {
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

	// Multiply by the old diff.
	nextDiffBig := weightedSumDiv.Mul(weightedSumDiv, oldDiffBig)

	// Right shift to restore the original padding (restore non-fixed point).
	nextDiffBig = nextDiffBig.Rsh(nextDiffBig, 32)

	// Check to see if we're over the limits for the maximum allowable retarget;
	// if we are, return the maximum or minimum except in the case that oldDiff
	// is zero.
	if oldDiffBig.Cmp(bigZero) == 0 { // This should never really happen,
		nextDiffBig.Set(nextDiffBig) // but in case it does...
	} else if nextDiffBig.Cmp(bigZero) == 0 {
		nextDiffBig.Set(b.params.PowLimit)
	} else if nextDiffBig.Cmp(nextDiffBigMax) == 1 {
		nextDiffBig.Set(nextDiffBigMax)
	} else if nextDiffBig.Cmp(nextDiffBigMin) == -1 {
		nextDiffBig.Set(nextDiffBigMin)
	}

	// Limit new value to the proof of work limit.
	if nextDiffBig.Cmp(b.params.PowLimit) > 0 {
		nextDiffBig.Set(b.params.PowLimit)
	}

	// Log new target difficulty and return it.  The new target logging is
	// intentionally converting the bits back to a number instead of using
	// newTarget since conversion to the compact representation loses
	// precision.
	nextDiffBits := BigToCompact(nextDiffBig)
	log.Debug("Difficulty retarget", "block main height", curBlock.GetHeight()+1)
	log.Debug("Old target", "bits",fmt.Sprintf("%08x", curNode.bits),
		"diff",fmt.Sprintf( "(%064x)",oldDiffBig))
	log.Debug("New target", "bits",fmt.Sprintf("%08x", nextDiffBits),
		"diff",fmt.Sprintf( "(%064x)",CompactToBig(nextDiffBits)))

	return nextDiffBits, nil
}

// CalcNextRequiredDiffFromNode calculates the required difficulty for the block
// given with the passed hash along with the given timestamp.
//
// This function is NOT safe for concurrent access.
func (b *BlockChain) CalcNextRequiredDiffFromNode(hash *hash.Hash, timestamp time.Time) (uint32, error) {
	node := b.index.LookupNode(hash)
	if node == nil {
		return 0, fmt.Errorf("block %s is not known", hash)
	}

	return b.calcNextRequiredDifficulty(node, timestamp)
}

// CalcNextRequiredDifficulty calculates the required difficulty for the block
// after the end of the current best chain based on the difficulty retarget
// rules.
//
// This function is safe for concurrent access.
func (b *BlockChain) CalcNextRequiredDifficulty(timestamp time.Time) (uint32, error) {
	b.chainLock.Lock()
	block:=b.bd.GetMainChainTip()
	node:=b.index.lookupNode(block.GetHash())
	difficulty, err := b.calcNextRequiredDifficulty(node, timestamp)
	b.chainLock.Unlock()
	return difficulty, err
}

// mergeDifficulty takes an original stake difficulty and two new, scaled
// stake difficulties, merges the new difficulties, and outputs a new
// merged stake difficulty.
func mergeDifficulty(oldDiff int64, newDiff1 int64, newDiff2 int64) int64 {
	newDiff1Big := big.NewInt(newDiff1)
	newDiff2Big := big.NewInt(newDiff2)
	newDiff2Big.Lsh(newDiff2Big, 32)

	oldDiffBig := big.NewInt(oldDiff)
	oldDiffBigLSH := big.NewInt(oldDiff)
	oldDiffBigLSH.Lsh(oldDiffBig, 32)

	newDiff1Big.Div(oldDiffBigLSH, newDiff1Big)
	newDiff2Big.Div(newDiff2Big, oldDiffBig)

	// Combine the two changes in difficulty.
	summedChange := big.NewInt(0)
	summedChange.Set(newDiff2Big)
	summedChange.Lsh(summedChange, 32)
	summedChange.Div(summedChange, newDiff1Big)
	summedChange.Mul(summedChange, oldDiffBig)
	summedChange.Rsh(summedChange, 32)

	return summedChange.Int64()
}

