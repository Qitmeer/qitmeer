/*
 * Copyright (c) 2017-2020 The qitmeer developers
 */

package blockchain

import (
	"github.com/Qitmeer/qitmeer/core/address"
	"github.com/Qitmeer/qitmeer/engine/txscript"
	"strings"
)

func PayToCltvAddrScriptWithMainHeight(addrStr string, mainHeight int64) ([]byte, error) {
	addr, err := address.DecodeAddress(addrStr)
	if err != nil {
		return nil, err
	}
	return txscript.NewScriptBuilder().AddInt64(mainHeight).AddOp(txscript.OP_CHECKLOCKTIMEVERIFY).AddOp(txscript.OP_DROP).AddOp(txscript.OP_DUP).AddOp(txscript.OP_HASH160).
		AddData(addr.Script()).AddOp(txscript.OP_EQUALVERIFY).AddOp(txscript.OP_CHECKSIG).
		Script()
}

func IsCltvPublicKeyHashTy(pkscript []byte) bool {
	ds, err := txscript.DisasmString(pkscript)
	if err != nil {
		log.Error(err.Error())
		return false
	}
	if len(ds) <= 0 {
		return false
	}
	return strings.Contains(ds, "OP_CHECKLOCKTIMEVERIFY")
}
