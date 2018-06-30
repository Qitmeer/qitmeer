// Copyright (c) 2017-2018 The nox developers
package blockchain

import (
	"github.com/noxproject/nox/core/types"
	"math"
	"github.com/noxproject/nox/common/hash"
)

var (
	// zeroHash is the zero value for a hash.Hash and is defined as a
	// package level variable to avoid the need to create a new instance
	// every time a check is needed.
	zeroHash = &hash.Hash{}
)

// IsCoinBaseTx determines whether or not a transaction is a coinbase.  A
// coinbase is a special transaction created by miners that has no inputs.
// This is represented in the block chain by a transaction with a single input
// that has a previous output transaction index set to the maximum value along
// with a zero hash.
//
// This function only differs from IsCoinBase in that it works with a raw wire
// transaction as opposed to a higher level util transaction.
func IsCoinBaseTx(tx *types.Transaction) bool {
	// A coin base must only have one transaction input.
	if len(tx.TxIn) != 1 {
		return false
	}
	// The previous output of a coin base must have a max value index and a
	// zero hash.
	prevOut := &tx.TxIn[0].PreviousOut
	if prevOut.OutIndex != math.MaxUint32 || !prevOut.Hash.IsEqual(zeroHash) {
		return false
	}
	return true
}

