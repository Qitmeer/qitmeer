// Copyright (c) 2021 The qitmeer developers
// Use of this source code is governed by an ISC
// license that can be found in the LICENSE file.

package types

import (
	"fmt"
	"github.com/Qitmeer/qitmeer/common/math"
)

// TxType indicates the type of transactions
// such as regular or other tx type (coinbase, stake or token).
type TxType int

const (
	TxTypeRegular     TxType = iota
	TxTypeCoinbase    TxType = 0x1
	TxTypeGenesisLock TxType = 0x2 // the tx try to lock the genesis output to the stake pool

	TxTypeStakebase     TxType = 0x10 // the special tx which vote for stake_purchase and reward stake holder from the stake_reserve
	TyTypeStakeReserve  TxType = 0x11 // the tx reserve consensus-based value to a special stake_reserve address
	TxTypeStakePurchase TxType = 0x12 // the tx of stake holder who lock value into stake pool
	TxTypeStakeDispose  TxType = 0x13 // the opposite tx of stake_purchase

	TxTypeTokenRegulation TxType = 0x80 // token-regulation is reserved, not for using.
	TxTypeTokenNew        TxType = 0x81 // new coin-id, owners, up-limit etc. the token is disabled after token-new.
	TxTypeTokenRenew      TxType = 0x82 // update owners, up-limits etc. can't change coin-id. renew works only when the token is disabled.
	TxTypeTokenValidate   TxType = 0x83 // enable the token.
	TxTypeTokenInvalidate TxType = 0x84 // disable the token.
	TxTypeTokenRevoke     TxType = 0x8f // revoke the coin-id. token-revoke is reserved, not used at current stage.

	TxTypeTokenbase   TxType = 0x90 // token-base is reserved, not used at current stage.
	TxTypeTokenMint   TxType = 0x91 // token owner mint token amount by locking MEER. (must validated token)
	TxTypeTokenUnmint TxType = 0x92 // token owner unmint token amount by releasing MEER. (must validated token)
)

func (tt TxType) String() string {
	switch tt {
	case TxTypeRegular:
		return "TxTypeRegular"
	case TxTypeCoinbase:
		return "TxTypeCoinbase"
	case TxTypeGenesisLock:
		return "TxTypeGenesisLock"
	case TxTypeStakebase:
		return "TxTypeStakebase"
	case TyTypeStakeReserve:
		return "TyTypeStakeReserve"
	case TxTypeStakePurchase:
		return "TxTypeStakePurchase"
	case TxTypeStakeDispose:
		return "TxTypeStakeDispose"
	case TxTypeTokenRegulation:
		return "TxTypeTokenRegulation"
	case TxTypeTokenNew:
		return "TxTypeTokenNew"
	case TxTypeTokenRenew:
		return "TxTypeTokenRenew"
	case TxTypeTokenValidate:
		return "TxTypeTokenValidate"
	case TxTypeTokenInvalidate:
		return "TxTypeTokenInvalidate"
	case TxTypeTokenRevoke:
		return "TxTypeTokenRevoke"
	case TxTypeTokenbase:
		return "TxTypeTokenbase"
	case TxTypeTokenMint:
		return "TxTypeTokenMint"
	case TxTypeTokenUnmint:
		return "TxTypeTokenUnmint"
	}
	return "Unknow"
}

// DetermineTxType determines the type of transaction
func DetermineTxType(tx *Transaction) TxType {
	if IsCoinBaseTx(tx) {
		return TxTypeCoinbase
	}
	if IsGenesisLockTx(tx) {
		return TxTypeGenesisLock
	}
	if IsTokenNewTx(tx) {
		return TxTypeTokenNew
	}
	if IsTokenRenewTx(tx) {
		return TxTypeTokenRenew
	}
	if IsTokenValidateTx(tx) {
		return TxTypeTokenValidate
	}
	if IsTokenInvalidateTx(tx) {
		return TxTypeTokenInvalidate
	}
	if IsTokenMintTx(tx) {
		return TxTypeTokenMint
	}
	if IsTokenUnmintTx(tx) {
		return TxTypeTokenUnmint
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

// --------------------------------------------------------------------------------
// The Genesis_XXX transactions
//
//    the transactions related to the genesis UTXOs
//
//  - genesis_lock   how lock the genesis output to the stake pool
// --------------------------------------------------------------------------------

// CheckGenesisLock returns an error if a transaction is not a genesis_lock transaction.
// It makes sure the number of inputs/outputs, and related scripts are valid.
//
// genesis_lock transactions are specified as below:
//
// 1.) Inputs:
// genesis output 1 [index 0]
// genesis output 2 [index 1]
// ...
//
// 2.) Outputs:
// OP_STAKE_LOCK tagged output [index 0]
// OP_RETURN push of input 1's stake reward receiving address [index 1]
// OP_STAKE_LOCK tagged output [index 2]
// OP_RETURN push of input 2's stake reward receiving address [index 3]
// ...
// Note: The output OP_RETURN pushes should be of size 20 bytes (standard address).
//
// 3.) NO changes are allowed for the tx && NO fee are allowed for this tx.
// inputs amount == output amount
//
func CheckGenesisLock(tx *Transaction) error {
	return fmt.Errorf("CheckGenesisLock is not supported yet")
}

// IsGenesisLockTx returns whether or not a transaction is a genesis_lock transaction.
func IsGenesisLockTx(tx *Transaction) bool {
	return CheckGenesisLock(tx) == nil
}

// --------------------------------------------------------------------------------
// Stake_XXX Transaction
// --------------------------------------------------------------------------------

// --------------------------------------------------------------------------------
// Token_XXX Transaction
//
//
// The workflow of an new token
// roles
//   - token operator : the controller of a token who provide & organize the financial services using the token and
//     take the responsibility as the legislate entity.
//   - token regulator : the main supervising body who assure all token operator follow regulatory guidelines, such as AML policy etc.
//     who in charge of new token approve and supervision the running token service.
// workflow
//   1. token operator request to token regulator, fulfil the requirement of the new token application. (off-chain)
//   2. if 1. is accepted, token regulator issue token_new (on chain).
//   3. consensus-based vote for token_validate (on chain).
//   4. if 3. is ok, token can be operated by token operator officially.
//   5. token operator do token_mint, the consensus-based token amount assessable. (on chain)
// --------------------------------------------------------------------------------

func IsTokenNewTx(tx *Transaction) bool {
	if len(tx.TxOut) != 1 || len(tx.TxIn) != 1 {
		return false
	}
	if tx.TxIn[0].PreviousOut.OutIndex != TokenPrevOutIndex {
		return false
	}
	return TxType(tx.TxIn[0].Sequence) == TxTypeTokenNew
}

func IsTokenRenewTx(tx *Transaction) bool {
	if len(tx.TxOut) != 1 || len(tx.TxIn) != 1 {
		return false
	}
	if tx.TxIn[0].PreviousOut.OutIndex != TokenPrevOutIndex {
		return false
	}
	return TxType(tx.TxIn[0].Sequence) == TxTypeTokenRenew
}

func IsTokenValidateTx(tx *Transaction) bool {
	if len(tx.TxOut) != 1 || len(tx.TxIn) != 1 {
		return false
	}
	if tx.TxIn[0].PreviousOut.OutIndex != TokenPrevOutIndex {
		return false
	}
	return TxType(tx.TxIn[0].Sequence) == TxTypeTokenValidate
}

func IsTokenInvalidateTx(tx *Transaction) bool {
	if len(tx.TxOut) != 1 || len(tx.TxIn) != 1 {
		return false
	}
	if tx.TxIn[0].PreviousOut.OutIndex != TokenPrevOutIndex {
		return false
	}
	return TxType(tx.TxIn[0].Sequence) == TxTypeTokenInvalidate
}

func IsTokenMintTx(tx *Transaction) bool {
	if len(tx.TxOut) < 1 || len(tx.TxIn) <= 1 {
		return false
	}
	if tx.TxIn[0].PreviousOut.OutIndex != TokenPrevOutIndex {
		return false
	}
	return TxType(tx.TxIn[0].Sequence) == TxTypeTokenMint
}

func IsTokenUnmintTx(tx *Transaction) bool {
	if len(tx.TxOut) < 1 || len(tx.TxIn) <= 1 {
		return false
	}
	if tx.TxIn[0].PreviousOut.OutIndex != TokenPrevOutIndex {
		return false
	}
	return TxType(tx.TxIn[0].Sequence) == TxTypeTokenUnmint
}

func IsTokenTx(tx *Transaction) bool {
	return IsTokenNewTx(tx) ||
		IsTokenRenewTx(tx) ||
		IsTokenValidateTx(tx) ||
		IsTokenInvalidateTx(tx) ||
		IsTokenMintTx(tx) ||
		IsTokenUnmintTx(tx)
}

// Standard transaction type
var StdTxs = []TxType{TxTypeRegular, TxTypeCoinbase}
var NonStdTxs = []TxType{
	TxTypeTokenNew,
	TxTypeTokenRenew,
	TxTypeTokenInvalidate,
	TxTypeTokenValidate,
	TxTypeTokenMint,
}
