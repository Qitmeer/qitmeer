package txtype

import (
	"github.com/Qitmeer/qitmeer/core/types"
	"github.com/Qitmeer/qitmeer/engine/txscript"
)

// DetermineTxType determines the type of transaction
func DetermineTxType(tx *types.Transaction) types.TxType {
	if types.IsCoinBaseTx(tx) {
		return types.TxTypeCoinbase
	}
	if types.IsGenesisLockTx(tx) {
		return types.TxTypeGenesisLock
	}
	if IsTokenNewTx(tx) {
		return types.TxTypeTokenNew
	}

	//TODO more txType
	return types.TxTypeRegular
}

func IsTokenNewTx(tx *types.Transaction) bool {
	if len(tx.TxOut) != 1 {
		return false
	}
	scriptClass := txscript.GetScriptClass(txscript.DefaultScriptVersion, tx.TxOut[0].PkScript)
	if scriptClass != txscript.TokenPubKeyHashTy {
		return false
	}
	return true
}
