// Copyright (c) 2017-2018 The qitmeer developers
// Copyright (c) 2013-2016 The btcsuite developers
// Copyright (c) 2017-2018 The Decred developers
// Use of this source code is governed by an ISC
// license that can be found in the LICENSE file.
package mempool

import (
	"github.com/Qitmeer/qitmeer/common/hash"
	"github.com/Qitmeer/qitmeer/core/blockchain"
	"github.com/Qitmeer/qitmeer/core/blockdag"
	"github.com/Qitmeer/qitmeer/core/types"
)

// minInt is a helper function to return the minimum of two ints.  This avoids
// a math import and the need to cast to floats.
func minInt(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// calcInputValueAge is a helper function used to calculate the input age of
// a transaction.  The input age for a txin is the number of confirmations
// since the referenced txout multiplied by its output value.  The total input
// age is the sum of this value for each txin.  Any inputs to the transaction
// which are currently in the mempool and hence not mined into a block yet,
// contribute no additional input age to the transaction.
func calcInputValueAge(tx *types.Transaction, utxoView *blockchain.UtxoViewpoint, nextBlockHeight uint64, bd *blockdag.BlockDAG) float64 {
	var totalInputAge float64
	for _, txIn := range tx.TxIn {
		// Don't attempt to accumulate the total input age if the
		// referenced transaction output doesn't exist.
		txEntry := utxoView.LookupEntry(txIn.PreviousOut)
		if txEntry != nil && !txEntry.IsSpent() {
			// Inputs with dependencies currently in the mempool
			// have their block height set to a special constant.
			// Their input age should be computed as zero since
			// their parent hasn't made it into a block yet.
			var inputAge uint64
			if txEntry.BlockHash().IsEqual(&hash.ZeroHash) {
				inputAge = 0
			} else {
				block := bd.GetBlock(txEntry.BlockHash())
				if block == nil {
					return 0
				}
				inputAge = nextBlockHeight - uint64(block.GetHeight())
			}
			// Sum the input value times age.
			inputValue := uint64(txEntry.Amount().Value)
			totalInputAge += float64(inputValue * inputAge)
		}
	}

	return totalInputAge
}

// CalcPriority returns a transaction priority given a transaction and the sum
// of each of its input values multiplied by their age (# of confirmations).
// Thus, the final formula for the priority is:
// sum(inputValue * inputAge) / adjustedTxSize
func CalcPriority(tx *types.Transaction, utxoView *blockchain.UtxoViewpoint, nextBlockHeight uint64, bd *blockdag.BlockDAG) float64 {
	// In order to encourage spending multiple old unspent transaction
	// outputs thereby reducing the total set, don't count the constant
	// overhead for each input as well as enough bytes of the signature
	// script to cover a pay-to-script-hash redemption with a compressed
	// pubkey.  This makes additional inputs free by boosting the priority
	// of the transaction accordingly.  No more incentive is given to avoid
	// encouraging gaming future transactions through the use of junk
	// outputs.  This is the same logic used in the reference
	// implementation.
	//
	// The constant overhead for a txin is 41 bytes since the previous
	// outpoint is 36 bytes + 4 bytes for the sequence + 1 byte the
	// signature script length.
	//
	// A compressed pubkey pay-to-script-hash redemption with a maximum len
	// signature is of the form:
	// [OP_DATA_73 <73-byte sig> + OP_DATA_35 + {OP_DATA_33
	// <33 byte compresed pubkey> + OP_CHECKSIG}]
	//
	// Thus 1 + 73 + 1 + 1 + 33 + 1 = 110
	overhead := 0
	for _, txIn := range tx.TxIn {
		// Max inputs + size can't possibly overflow here.
		overhead += 41 + minInt(110, len(txIn.SignScript))
	}

	serializedTxSize := tx.SerializeSize()
	if overhead >= serializedTxSize {
		return 0.0
	}

	inputValueAge := calcInputValueAge(tx, utxoView, nextBlockHeight, bd)
	return inputValueAge / float64(serializedTxSize-overhead)
}
