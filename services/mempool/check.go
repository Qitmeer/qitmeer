// Copyright (c) 2017-2018 The qitmeer developers
// Copyright (c) 2013-2016 The btcsuite developers
// Copyright (c) 2017-2018 The Decred developers
// Use of this source code is governed by an ISC
// license that can be found in the LICENSE file.
package mempool

import (
	"fmt"
	"github.com/Qitmeer/qitmeer/core/blockchain"
	"github.com/Qitmeer/qitmeer/core/message"
	"github.com/Qitmeer/qitmeer/core/types"
	"github.com/Qitmeer/qitmeer/engine/txscript"
	"time"
)

// checkTransactionStandard performs a series of checks on a transaction to
// ensure it is a "standard" transaction.  A standard transaction is one that
// conforms to several additional limiting cases over what is considered a
// "sane" transaction such as having a version in the supported range, being
// finalized, conforming to more stringent size constraints, having scripts
// of recognized forms, and not containing "dust" outputs (those that are
// so small it costs more to process them than they are worth).
func checkTransactionStandard(tx *types.Tx, height uint64,
	medianTime time.Time, minRelayTxFee types.Amount,
	maxTxVersion uint16) error {

	// The transaction must be a currently supported version and serialize
	// type.
	msgTx := tx.Transaction()
	version := uint32(msgTx.Version & 0xffff) //TODO fix type conversion
	serType := types.TxSerializeType(version >> 16)

	if serType != types.TxSerializeFull {
		str := fmt.Sprintf("transaction is not serialized with all "+
			"required data -- type %v", serType)
		return txRuleError(message.RejectNonstandard, str)
	}
	if msgTx.Version > uint32(maxTxVersion) || msgTx.Version < 1 {
		str := fmt.Sprintf("transaction version %d is not in the "+
			"valid range of %d-%d", msgTx.Version, 1, maxTxVersion)
		return txRuleError(message.RejectNonstandard, str)
	}

	// The transaction must be finalized to be standard and therefore
	// considered for inclusion in a block.
	// TODO fix type conversion
	if !blockchain.IsFinalizedTransaction(tx, height, medianTime) {
		return txRuleError(message.RejectNonstandard,
			"transaction is not finalized")
	}

	// Since extremely large transactions with a lot of inputs can cost
	// almost as much to process as the sender fees, limit the maximum
	// size of a transaction.  This also helps mitigate CPU exhaustion
	// attacks.
	serializedLen := msgTx.SerializeSize()
	if serializedLen > maxStandardTxSize {
		str := fmt.Sprintf("transaction size of %v is larger than max "+
			"allowed size of %v", serializedLen, maxStandardTxSize)
		return txRuleError(message.RejectNonstandard, str)
	}

	for i, txIn := range msgTx.TxIn {
		// Each transaction input signature script must not exceed the
		// maximum size allowed for a standard transaction.  See
		// the comment on maxStandardSigScriptSize for more details.
		sigScriptLen := len(txIn.SignScript)
		if sigScriptLen > maxStandardSigScriptSize {
			str := fmt.Sprintf("transaction input %d: signature "+
				"script size of %d bytes is large than max "+
				"allowed size of %d bytes", i, sigScriptLen,
				maxStandardSigScriptSize)
			return txRuleError(message.RejectNonstandard, str)
		}

		// Each transaction input signature script must only contain
		// opcodes which push data onto the stack.
		if !txscript.IsPushOnlyScript(txIn.SignScript) {
			str := fmt.Sprintf("transaction input %d: signature "+
				"script is not push only", i)
			return txRuleError(message.RejectNonstandard, str)
		}

	}

	// None of the output public key scripts can be a non-standard script or
	// be "dust" (except when the script is a null data script).
	numNullDataOutputs := 0
	for i, txOut := range msgTx.TxOut {
		//TODO the tx version
		scriptClass := txscript.GetScriptClass(txscript.DefaultScriptVersion, txOut.PkScript)
		err := checkPkScriptStandard(txOut.PkScript, scriptClass)
		if err != nil {
			// Attempt to extract a reject code from the error so
			// it can be retained.  When not possible, fall back to
			// a non standard error.
			rejectCode, found := extractRejectCode(err)
			if !found {
				rejectCode = message.RejectNonstandard
			}
			str := fmt.Sprintf("transaction output %d: %v", i, err)
			return txRuleError(rejectCode, str)
		}

		// Accumulate the number of outputs which only carry data.  For
		// all other script types, ensure the output value is not
		// "dust".
		// TODO DUST decision (may careful about reject Dust for token base tx)
		if scriptClass == txscript.NullDataTy {
			numNullDataOutputs++
		} else if isDust(txOut, minRelayTxFee) {
			str := fmt.Sprintf("transaction output %d: payment "+
				"of %d is dust", i, txOut.Amount)
			return txRuleError(message.RejectDust, str)
		}
	}

	// A standard transaction must not have more than one output script that
	// only carries data. However, certain types of standard stake transactions
	// are allowed to have multiple OP_RETURN outputs, so only throw an error here
	// if the tx is TxTypeRegular.
	if numNullDataOutputs > maxNullDataOutputs {
		str := "more than one transaction output in a nulldata script for a " +
			"regular type tx"
		return txRuleError(message.RejectNonstandard, str)
	}

	return nil
}

// checkPkScriptStandard performs a series of checks on a transaction output
// script (public key script) to ensure it is a "standard" public key script.
// A standard public key script is one that is a recognized form, and for
// multi-signature scripts, only contains from 1 to maxStandardMultiSigKeys
// public keys.
func checkPkScriptStandard(pkScript []byte,
	scriptClass txscript.ScriptClass) error {

	// TODO the DefaultPkScriptVersion check
	// Only default Bitcoin-style script is standard except for
	// null data outputs.
	/*
		if version != message.DefaultPkScriptVersion {
			str := fmt.Sprintf("versions other than default pkscript version " +
				"are currently non-standard except for provably unspendable " +
				"outputs")
			return txRuleError(message.RejectNonstandard, str)
		}
	*/

	switch scriptClass {
	case txscript.MultiSigTy:
		numPubKeys, numSigs, err := txscript.CalcMultiSigStats(pkScript)
		if err != nil {
			str := fmt.Sprintf("multi-signature script parse "+
				"failure: %v", err)
			return txRuleError(message.RejectNonstandard, str)
		}

		// A standard multi-signature public key script must contain
		// from 1 to maxStandardMultiSigKeys public keys.
		if numPubKeys < 1 {
			str := "multi-signature script with no pubkeys"
			return txRuleError(message.RejectNonstandard, str)
		}
		if numPubKeys > maxStandardMultiSigKeys {
			str := fmt.Sprintf("multi-signature script with %d "+
				"public keys which is more than the allowed "+
				"max of %d", numPubKeys, maxStandardMultiSigKeys)
			return txRuleError(message.RejectNonstandard, str)
		}

		// A standard multi-signature public key script must have at
		// least 1 signature and no more signatures than available
		// public keys.
		if numSigs < 1 {
			return txRuleError(message.RejectNonstandard,
				"multi-signature script with no signatures")
		}
		if numSigs > numPubKeys {
			str := fmt.Sprintf("multi-signature script with %d "+
				"signatures which is more than the available "+
				"%d public keys", numSigs, numPubKeys)
			return txRuleError(message.RejectNonstandard, str)
		}

	case txscript.NonStandardTy:
		return txRuleError(message.RejectNonstandard,
			"non-standard script form")
	}

	return nil
}

// isDust returns whether or not the passed transaction output amount is
// considered dust or not based on the passed minimum transaction relay fee.
// Dust is defined in terms of the minimum transaction relay fee.  In
// particular, if the cost to the network to spend coins is more than 1/3 of the
// minimum transaction relay fee, it is considered dust.
func isDust(txOut *types.TxOutput, minRelayTxFee types.Amount) bool {
	// Unspendable outputs are considered dust.
	if txscript.IsUnspendable(txOut.PkScript) {
		return true
	}

	// Only MeerCoin need to compare with RelayTxFee
	if txOut.Amount.Id != types.MEERID {
		// TODO the Dust rule for coin other than meer
		return false
	}

	// The total serialized size consists of the output and the associated
	// input script to redeem it.  Since there is no input script
	// to redeem it yet, use the minimum size of a typical input script.
	//
	// Pay-to-pubkey-hash bytes breakdown:
	//
	//  Output to hash (36 bytes):
	//   8 value, 2 script version, 1 script len, 25 script [1 OP_DUP,
	//   1 OP_HASH_160, 1 OP_DATA_20, 20 hash, 1 OP_EQUALVERIFY,
	//   1 OP_CHECKSIG]
	//
	//  Input with compressed pubkey (165 bytes):
	//   37 prev outpoint, 4 sequence, 16 fraud proof, 1 script len,
	//   107 script [1 OP_DATA_72, 72 sig, 1 OP_DATA_33, 33 compressed
	//   pubkey]
	//
	//  Input with uncompressed pubkey (197 bytes):
	//   37 prev outpoint, 4 sequence, 16 fraud proof, 1 script len,
	//   139 script [1 OP_DATA_72, 72 sig, 1 OP_DATA_65, 65 uncompressed
	//   pubkey]
	//
	// Pay-to-pubkey bytes breakdown:
	//
	//  Output to compressed pubkey (46 bytes):
	//   8 value, 2 script version, 1 script len, 35 script [1 OP_DATA_33,
	//   33 compressed pubkey, 1 OP_CHECKSIG]
	//
	//  Output to uncompressed pubkey (78 bytes):
	//   8 value, 2 script version, 1 script len, 67 script [1 OP_DATA_65,
	//   65 uncompressed pubkey, 1 OP_CHECKSIG]
	//
	//  Input (131 bytes):
	//   37 prev outpoint, 4 sequence, 16 fraud proof, 1 script len,
	//   73 script [1 OP_DATA_72, 72 sig]
	//
	// Theoretically this could examine the script type of the output script
	// and use a different size for the typical input script size for
	// pay-to-pubkey vs pay-to-pubkey-hash inputs per the above breakdowns,
	// but the only combinination which is less than the value chosen is
	// a pay-to-pubkey script with a compressed pubkey, which is not very
	// common.
	//
	// The most common scripts are pay-to-pubkey-hash, and as per the above
	// breakdown, the minimum size of a p2pkh input script is 165 bytes.  So
	// that figure is used.
	totalSize := txOut.SerializeSize() + 165

	// The output is considered dust if the cost to the network to spend the
	// coins is more than 1/3 of the minimum free transaction relay fee.
	// minFreeTxRelayFee is in Atom/KB, so multiply by 1000 to convert to
	// bytes.
	//
	// Using the typical values for a pay-to-pubkey-hash transaction from
	// the breakdown above and the default minimum free transaction relay
	// fee of 10000, this equates to values less than 6030 atoms being
	// considered dust.
	//
	// The following is equivalent to (value/totalSize) * (1/3) * 1000
	// without needing to do floating point math.
	// TODO fix type conversion
	// TODO consider the case of token output
	return int64(txOut.Amount.Value)*1000/(3*int64(totalSize)) < int64(minRelayTxFee.Value)
}

// checkPoolDoubleSpend checks whether or not the passed transaction is
// attempting to spend coins already spent by other transactions in the pool.
// Note it does not check for double spends against transactions already in the
// main chain.
//
// This function MUST be called with the mempool lock held (for reads).
func (mp *TxPool) checkPoolDoubleSpend(tx *types.Tx) error {
	for _, txIn := range tx.Transaction().TxIn {
		if txR, exists := mp.outpoints[txIn.PreviousOut]; exists {
			str := fmt.Sprintf("transaction %v in the pool "+
				"already spends the same coins", txR.Hash())
			return txRuleError(message.RejectDuplicate, str)
		}
	}
	return nil
}

// checkInputsStandard performs a series of checks on a transaction's inputs
// to ensure they are "standard".  A standard transaction input within the
// context of this function is one whose referenced public key script is of a
// standard form and, for pay-to-script-hash, does not have more than
// maxStandardP2SHSigOps signature operations.  However, it should also be noted
// that standard inputs also are those which have a clean stack after execution
// and only contain pushed data in their signature scripts.  This function does
// not perform those checks because the script engine already does this more
// accurately and concisely via the txscript.ScriptVerifyCleanStack and
// txscript.ScriptVerifySigPushOnly flags.
func checkInputsStandard(tx *types.Tx, utxoView *blockchain.UtxoViewpoint) error {

	// NOTE: The reference implementation also does a coinbase check here,
	// but coinbases have already been rejected prior to calling this
	// function so no need to recheck.

	for i, txIn := range tx.Transaction().TxIn {

		// It is safe to elide existence and index checks here since
		// they have already been checked prior to calling this
		// function.
		prevOut := txIn.PreviousOut
		entry := utxoView.LookupEntry(prevOut)
		originPkScript := entry.PkScript()
		switch txscript.GetScriptClass(txscript.DefaultScriptVersion, originPkScript) {
		case txscript.ScriptHashTy:
			numSigOps := txscript.GetPreciseSigOpCount(
				txIn.SignScript, originPkScript, true)
			if numSigOps > maxStandardP2SHSigOps {
				str := fmt.Sprintf("transaction input #%d has "+
					"%d signature operations which is more "+
					"than the allowed max amount of %d",
					i, numSigOps, maxStandardP2SHSigOps)
				return txRuleError(message.RejectNonstandard, str)
			}
		case txscript.NonStandardTy:
			str := fmt.Sprintf("transaction input #%d has a "+
				"non-standard script form", i)
			return txRuleError(message.RejectNonstandard, str)
		}
	}

	return nil
}
