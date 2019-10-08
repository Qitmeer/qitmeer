// Copyright 2017-2018 The qitmeer developers
// Use of this source code is governed by an ISC
// license that can be found in the LICENSE file.
package core

import (
	"bytes"
	"encoding/hex"
	"fmt"
	"github.com/Qitmeer/qitmeer-lib/common/marshal"
	"github.com/Qitmeer/qitmeer-lib/core/json"
	"github.com/Qitmeer/qitmeer-lib/core/types"
	"github.com/Qitmeer/qitmeer-lib/params"
	"github.com/Qitmeer/qitmeer-lib/qx"
	"github.com/pkg/errors"
)

func TxDecode(network string, rawTxStr string) {
	var param *params.Params
	switch network {
	case "mainnet":
		param = &params.MainNetParams
	case "testnet":
		param = &params.TestNetParams
	case "privnet":
		param = &params.PrivNetParams
	}
	if len(rawTxStr)%2 != 0 {
		ErrExit(fmt.Errorf("invaild raw transaction : %s", rawTxStr))
	}
	serializedTx, err := hex.DecodeString(rawTxStr)
	if err != nil {
		ErrExit(err)
	}
	var tx types.Transaction
	err = tx.Deserialize(bytes.NewReader(serializedTx))
	if err != nil {
		ErrExit(err)
	}

	jsonTx := &json.OrderedResult{
		{Key: "txid", Val: tx.TxHash().String()},
		{Key: "txhash", Val: tx.TxHashFull().String()},
		{Key: "version", Val: int32(tx.Version)},
		{Key: "locktime", Val: tx.LockTime},
		{Key: "expire", Val: tx.Expire},
		{Key: "vin", Val: marshal.MarshJsonVin(&tx)},
		{Key: "vout", Val: marshal.MarshJsonVout(&tx, nil, param)},
	}
	marshaledTx, err := jsonTx.MarshalJSON()
	if err != nil {
		ErrExit(err)
	}

	fmt.Printf("%s", marshaledTx)
}

func TxEncode(version TxVersionFlag, lockTime TxLockTimeFlag, txIn TxInputsFlag, txOut TxOutputsFlag) {
	txInputs := make(map[string]uint32)
	txOutputs := make(map[string]uint64)
	for _, input := range txIn.inputs {
		txInputs[hex.EncodeToString(input.txhash)] = input.index
	}
	for _, output := range txOut.outputs {
		atomic, err := types.NewAmount(output.amount)
		if err != nil {
			ErrExit(errors.Wrapf(err, "fail to create the currency amount from a "+
				"floating point value %f", output.amount))
		}
		txOutputs[output.target] = uint64(atomic)
	}
	mtxHex, err := qx.TxEncode(uint32(version), uint32(lockTime), txInputs, txOutputs)
	if err != nil {
		ErrExit(err)
	}
	fmt.Printf("%s\n", mtxHex)
}

func TxSign(privkeyStr string, rawTxStr string, network string) {
	mtxHex, err := qx.TxSign(privkeyStr, rawTxStr, network)
	if err != nil {
		ErrExit(err)
	}
	fmt.Printf("%s\n", mtxHex)
}
