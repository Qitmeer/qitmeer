// Copyright (c) 2017-2018 The qitmeer developers
// Copyright (c) 2017-2018 The Decred developers
// Use of this source code is governed by an ISC
// license that can be found in the LICENSE file.

package blockchain

import (
	"fmt"
	"github.com/Qitmeer/qng-core/common/hash"
	"time"

	"github.com/Qitmeer/qng-core/core/types"
)

// SequenceLock represents the minimum timestamp and minimum block height after
// which a transaction can be included into a block while satisfying the
// relative lock times of all of its input sequence numbers.  It is calculated
// via the CalcSequenceLock function.  Each field may be -1 if none of the input
// sequence numbers require a specific relative lock time for the respective
// type.  Since all valid heights and times are larger than -1, this implies
// that it will not prevent a transaction from being included due to the
// sequence lock, which is the desired behavior.
type SequenceLock struct {
	BlockHeight int64
	Time        int64
}

// calcSequenceLock computes the relative lock times for the passed transaction
// from the point of view of the block node passed in as the first argument.
//
// See the CalcSequenceLock comments for more details.
func (b *BlockChain) calcSequenceLock(tx *types.Tx, view *UtxoViewpoint, isActive bool) (*SequenceLock, error) {
	// A value of -1 for each lock type allows a transaction to be included
	// in a block at any given height or time.
	sequenceLock := &SequenceLock{BlockHeight: -1, Time: -1}

	// Sequence locks do not apply if they are not yet active, the tx
	// version is less than 2, or the tx is a coinbase or stakebase, so
	// return now with a sequence lock that indicates the tx can possibly be
	// included in a block at any given height or time.
	msgTx := tx.Transaction()
	//TODO, revisit the tx version for lock time
	enforce := isActive && msgTx.Version >= 2
	if !enforce || msgTx.IsCoinBase() || tx.IsDuplicate || types.IsTokenTx(tx.Tx) {
		return sequenceLock, nil

	}

	for txInIndex, txIn := range msgTx.TxIn {
		// Nothing to calculate for this input when relative time locks
		// are disabled for it.
		sequenceNum := txIn.Sequence
		// TODO, refactor config item
		if types.IsSequenceLockTimeDisabled(sequenceNum) {
			continue
		}

		utxo := view.LookupEntry(txIn.PreviousOut)
		if utxo == nil {
			str := fmt.Sprintf("output %v referenced from "+
				"transaction %s:%d either does not exist or "+
				"has already been spent", txIn.PreviousOut,
				tx.Hash(), txInIndex)
			return sequenceLock, ruleError(ErrMissingTxOut, str)
		}

		// Mask off the value portion of the sequence number to obtain
		// the time lock delta required before this input can be spent.
		// The relative lock can be time based or block based.
		relativeLock := int64(sequenceNum & types.SequenceLockTimeMask)

		if sequenceNum&types.SequenceLockTimeIsSeconds == types.SequenceLockTimeIsSeconds {
			// This input requires a time based relative lock
			// expressed in seconds before it can be spent and time
			// based locks are calculated relative to the earliest
			// possible time the block that contains the referenced
			// output could have been.  That time is the past
			// median time of the block before it (technically one
			// second after that, but that complexity is ignored for
			// time based locks which already have a granularity
			// associated with them anyways).  Therefore, the block
			// prior to the one in which the referenced output was
			// included is needed to compute its past median time.
			var medianTime time.Time
			if hash.ZeroHash.IsEqual(utxo.BlockHash()) {
				medianTime = b.BestSnapshot().MedianTime
			} else {
				blockNode := b.bd.GetBlock(utxo.BlockHash())
				if blockNode == nil {
					return sequenceLock, nil
				}
				medianTime = b.CalcPastMedianTime(blockNode)
			}
			// Calculate the minimum required timestamp based on the
			// sum of the aforementioned past median time and
			// required relative number of seconds.  Since time
			// based relative locks have a granularity associated
			// with them, shift left accordingly in order to convert
			// to the proper number of relative seconds.  Also,
			// subtract one from the relative lock to maintain the
			// original lock time semantics.
			relativeSecs := relativeLock << types.SequenceLockTimeGranularity
			minTime := medianTime.Unix() + relativeSecs - 1
			if minTime > sequenceLock.Time {
				sequenceLock.Time = minTime
			}
		} else {
			// This input requires a relative lock expressed in
			// blocks before it can be spent.  Therefore, calculate
			// the minimum required height based on the sum of the
			// input height and required relative number of blocks.
			// Also, subtract one from the relative lock in order to
			// maintain the original lock time semantics.
			var inputHeight uint
			if hash.ZeroHash.IsEqual(utxo.BlockHash()) {
				inputHeight = b.BestSnapshot().GraphState.GetMainHeight()
			} else {
				block := b.bd.GetBlock(utxo.BlockHash())
				if block == nil {
					return sequenceLock, nil
				}
				inputHeight = block.GetHeight()
			}

			minLayer := int64(inputHeight) + relativeLock - 1 //TODO,remove type conversion
			if minLayer > sequenceLock.BlockHeight {
				sequenceLock.BlockHeight = minLayer
			}
		}
	}

	return sequenceLock, nil
}

// CalcSequenceLock computes the minimum block order and time after which the
// passed transaction can be included into a block while satisfying the relative
// lock times of all of its input sequence numbers.  The passed view is used to
// obtain the past median time and block heights of the blocks in which the
// referenced outputs of the inputs to the transaction were included.  The
// generated sequence lock can be used in conjunction with a block height and
// median time to determine if all inputs to the transaction have reached the
// required maturity allowing it to be included in a block.
//
// NOTE: This will calculate the sequence locks regardless of the state of the
// agenda which conditionally activates it.  This is acceptable for standard
// transactions, however, callers which are intending to perform any type of
// consensus checking must check the status of the agenda first.
//
// This function is safe for concurrent access.
func (b *BlockChain) CalcSequenceLock(tx *types.Tx, view *UtxoViewpoint) (*SequenceLock, error) {
	b.ChainRLock()
	seqLock, err := b.calcSequenceLock(tx, view, true)
	b.ChainRUnlock()
	return seqLock, err
}

// LockTimeToSequence converts the passed relative lock time to a sequence
// number in accordance with DCP0003.
//
// A sequence number is defined as follows:
//
//   - bit 31 is the disable bit
//   - the next 8 bits are reserved
//   - bit 22 is the relative lock type (unset = block height, set = seconds)
//   - the next 6 bites are reserved
//   - the least significant 16 bits represent the value
//     - value has a granularity of 512 when interpreted as seconds (bit 22 set)
//
//   ---------------------------------------------------
//   | Disable | Reserved |  Type | Reserved |  Value  |
//   ---------------------------------------------------
//   |  1 bit  |  8 bits  | 1 bit |  6 bits  | 16 bits |
//   ---------------------------------------------------
//   |   [31]  |  [30-23] |  [22] |  [21-16] | [15-0]  |
//   ---------------------------------------------------
//
// The above implies that the maximum relative block height that can be encoded
// is 65535 and the maximum relative number of seconds that can be encoded is
// 65535*512 = 33,553,920 seconds (~1.06 years).  It also means that seconds are
// truncated to the nearest granularity towards 0 (e.g. 536 seconds will end up
// round tripping as 512 seconds and 1500 seconds will end up round tripping as
// 1024 seconds).
//
// An error will be returned for values that are larger than can be represented.
func LockTimeToSequence(isSeconds bool, lockTime uint32) (uint32, error) {
	// The corresponding sequence number is simply the desired input age
	// when expressing the relative lock time in blocks.
	if !isSeconds {
		if lockTime > types.SequenceLockTimeMask {
			return 0, fmt.Errorf("max relative block height a "+
				"sequence number can represent is %d",
				types.SequenceLockTimeMask)
		}
		return lockTime, nil
	}

	maxSeconds := uint32(types.SequenceLockTimeMask <<
		types.SequenceLockTimeGranularity)
	if lockTime > maxSeconds {
		return 0, fmt.Errorf("max relative seconds a sequence number "+
			"can represent is %d", maxSeconds)
	}

	// Set the 22nd bit which indicates the lock time is in seconds, then
	// shift the lock time over by 9 since the time granularity is in
	// 512-second intervals (2^9). This results in a max lock time of
	// 33,553,920 seconds (~1.06 years).
	return types.SequenceLockTimeIsSeconds |
		lockTime>>types.SequenceLockTimeGranularity, nil
}
