// Copyright 2017-2018 The nox developers
// Use of this source code is governed by an ISC
// license that can be found in the LICENSE file.
package main

import (
	"bytes"
	"encoding/hex"
	"fmt"
	"github.com/noxproject/nox/common/hash"
	"github.com/noxproject/nox/core/address"
	"github.com/noxproject/nox/core/json"
	"github.com/noxproject/nox/core/message"
	"github.com/noxproject/nox/core/types"
	"github.com/noxproject/nox/engine/txscript"
	"github.com/noxproject/nox/params"
	"github.com/noxproject/nox/services/common/marshal"
	"github.com/pkg/errors"
)

func txDecode(network string, rawTxStr string) {
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
		errExit(fmt.Errorf("invaild raw transaction : %s",rawTxStr))
	}
	serializedTx, err := hex.DecodeString(rawTxStr)
	if err != nil {
		errExit(err)
	}
	var tx types.Transaction
	err = tx.Deserialize(bytes.NewReader(serializedTx))
	if err != nil {
		errExit(err)
	}

	jsonTx := &json.OrderedResult{
		{"txid", tx.TxHash().String()},
		{"txhash", tx.TxHashFull().String()},
		{"version",  int32(tx.Version)},
		{"locktime", tx.LockTime},
		{"vin",      marshal.MarshJsonVin(&tx)},
		{"vout",     marshal.MarshJsonVout(&tx, nil,param)},
	}
	marshaledTx, err := jsonTx.MarshalJSON()
	if err != nil {
		errExit(err)
	}

	fmt.Printf("%s",marshaledTx)
}

func txEncode(version txVersionFlag, lockTime txLockTimeFlag, txIn txInputsFlag,txOut txOutputsFlag){

	mtx := types.NewTransaction()

	mtx.Version = uint32(version)

	if lockTime!=0 {
		mtx.LockTime = uint32(lockTime)
	}

	for _, input := range txIn.inputs {
		txHash,err := hash.NewHashFromStr(hex.EncodeToString(input.txhash))
		if err!=nil{
			errExit(err)
		}
		prevOut := types.NewOutPoint(txHash, input.index)
		txIn := types.NewTxInput(prevOut, types.NullValueIn, []byte{})
		txIn.Sequence = input.sequence
		if lockTime != 0 {
			txIn.Sequence = types.MaxTxInSequenceNum - 1
		}
		mtx.AddTxIn(txIn)
	}

	for _, output:= range txOutputs.outputs{

		// Decode the provided address.
		addr, err := address.DecodeAddress(output.target)
		if err != nil {
			errExit(errors.Wrapf(err,"fail to decode address %s",output.target))
		}

		// Ensure the address is one of the supported types and that
		// the network encoded with the address matches the network the
		// server is currently on.
		switch addr.(type) {
		case *address.PubKeyHashAddress:
		case *address.ScriptHashAddress:
		default:
			errExit(errors.Wrapf(err,"invalid type: %T", addr))
		}
		// Create a new script which pays to the provided address.
		pkScript, err := txscript.PayToAddrScript(addr)
		if err != nil {
			errExit(errors.Wrapf(err,"fail to create pk script for addr %s",addr))
		}

		atomic, err := types.NewAmount(output.amount)
		if err != nil {
			errExit(errors.Wrapf(err,"fail to create the currency amount from a " +
				"floating point value %f",output.amount))
		}
		//TODO fix type conversion
		txOut := types.NewTxOutput(uint64(atomic), pkScript)
		mtx.AddTxOut(txOut)
	}
	mtxHex, err := marshal.MessageToHex(&message.MsgTx{mtx})
	if err != nil {
		errExit(err)
	}
	fmt.Printf("%s\n",mtxHex)
}



