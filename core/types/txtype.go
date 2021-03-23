// Copyright (c) 2021 The qitmeer developers
// Use of this source code is governed by an ISC
// license that can be found in the LICENSE file.

package types

import (
	"github.com/Qitmeer/qitmeer/common/math"
)

// TxType indicates the type of transactions
// such as regular or other tx type (coinbase, stake or token).
type TxType int

const (
	TxTypeRegular       TxType = iota
	TxTypeCoinbase      TxType = 0x1
	TxTypeStakebase     TxType = 0x10
	TyTypeStakeReserve  TxType = 0x11
	TxTypeStakePurchase TxType = 0x12
	TxTypeStakeDispose  TxType = 0x13
	TxTypeTokenbase     TxType = 0x80  // token-base is reserved, not used at current stage
	TxTypeTokenMint     TxType = 0x81
	TxTypeTokenUnmint   TxType = 0x82
)

// DetermineTxType determines the type of transaction
func DetermineTxType(tx *Transaction) TxType {
	if IsCoinBaseTx(tx) {
		return TxTypeCoinbase
	}
	//TODO more txType
	return TxTypeRegular
}

// IsCoinBaseTx determines whether or not a transaction is a coinbase.  A
// coinbase is a special transaction created by miners that has no inputs.
// This is represented in the block chain by a transaction with a single input
// that has a previous output transaction index set to the maximum value along
// with a zero hash.
//
// This function only differs from IsCoinBase in that it works with a raw wire
// transaction as opposed to a higher level util transaction.
func IsCoinBaseTx(tx *Transaction) bool {
	// A coin base must only have one transaction input.
	if len(tx.TxIn) != 1 {
		return false
	}
	// The previous output of a coin base must have a max value index and a
	// zero hash.
	prevOut := &tx.TxIn[0].PreviousOut
	/*if prevOut.OutIndex != math.MaxUint32 || !prevOut.Hash.IsEqual(&hash.ZeroHash) {
		return false
	}*/
	return prevOut.OutIndex == math.MaxUint32
}
