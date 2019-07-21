// Copyright 2017-2018 The qitmeer developers
// Use of this source code is governed by an ISC
// license that can be found in the LICENSE file.
package qx

import (
	"bytes"
	"encoding/hex"
	"fmt"
	"github.com/HalalChain/qitmeer-lib/common/hash"
	"github.com/HalalChain/qitmeer-lib/common/marshal"
	"github.com/HalalChain/qitmeer-lib/core/address"
	"github.com/HalalChain/qitmeer-lib/core/json"
	"github.com/HalalChain/qitmeer-lib/core/message"
	"github.com/HalalChain/qitmeer-lib/core/types"
	"github.com/HalalChain/qitmeer-lib/crypto/ecc"
	"github.com/HalalChain/qitmeer-lib/engine/txscript"
	"github.com/HalalChain/qitmeer-lib/params"
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
		ErrExit(fmt.Errorf("invaild raw transaction : %s",rawTxStr))
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
		{Key:"txid", Val:tx.TxHash().String()},
		{Key:"txhash", Val:tx.TxHashFull().String()},
		{Key:"version",  Val:int32(tx.Version)},
		{Key:"locktime", Val:tx.LockTime},
		{Key:"expire",Val:tx.Expire},
		{Key:"vin",      Val:marshal.MarshJsonVin(&tx)},
		{Key:"vout",     Val:marshal.MarshJsonVout(&tx, nil,param)},
	}
	marshaledTx, err := jsonTx.MarshalJSON()
	if err != nil {
		ErrExit(err)
	}

	fmt.Printf("%s",marshaledTx)
}

func TxEncode(version TxVersionFlag, lockTime TxLockTimeFlag, txIn TxInputsFlag,txOut TxOutputsFlag){

	mtx := types.NewTransaction()

	mtx.Version = uint32(version)

	if lockTime!=0 {
		mtx.LockTime = uint32(lockTime)
	}

	for _, input := range txIn.inputs {
		txHash,err := hash.NewHashFromStr(hex.EncodeToString(input.txhash))
		if err!=nil{
			ErrExit(err)
		}
		prevOut := types.NewOutPoint(txHash, input.index)
		txIn := types.NewTxInput(prevOut, types.NullValueIn, []byte{})
		txIn.Sequence = input.sequence
		if lockTime != 0 {
			txIn.Sequence = types.MaxTxInSequenceNum - 1
		}
		mtx.AddTxIn(txIn)
	}

	for _, output:= range txOut.outputs{

		// Decode the provided address.
		addr, err := address.DecodeAddress(output.target)
		if err != nil {
			ErrExit(errors.Wrapf(err,"fail to decode address %s",output.target))
		}

		// Ensure the address is one of the supported types and that
		// the network encoded with the address matches the network the
		// server is currently on.
		switch addr.(type) {
		case *address.PubKeyHashAddress:
		case *address.ScriptHashAddress:
		default:
			ErrExit(errors.Wrapf(err,"invalid type: %T", addr))
		}
		// Create a new script which pays to the provided address.
		pkScript, err := txscript.PayToAddrScript(addr)
		if err != nil {
			ErrExit(errors.Wrapf(err,"fail to create pk script for addr %s",addr))
		}

		atomic, err := types.NewAmount(output.amount)
		if err != nil {
			ErrExit(errors.Wrapf(err,"fail to create the currency amount from a " +
				"floating point value %f",output.amount))
		}
		//TODO fix type conversion
		txOut := types.NewTxOutput(uint64(atomic), pkScript)
		mtx.AddTxOut(txOut)
	}
	mtxHex, err := mtx.Serialize(types.TxSerializeNoWitness)
	if err != nil {
		ErrExit(err)
	}
	fmt.Printf("%x\n",mtxHex)
}

func TxSign(privkeyStr string, rawTxStr string,network string) {
	privkeyByte, err := hex.DecodeString(privkeyStr)
	if err!=nil {
		ErrExit(err)
	}
	if len(privkeyByte) != 32 {
		ErrExit(fmt.Errorf("invaid ec private key bytes: %d",len(privkeyByte)))
	}
	privateKey, pubKey := ecc.Secp256k1.PrivKeyFromBytes(privkeyByte)
	h160 := hash.Hash160(pubKey.SerializeCompressed())

	var param *params.Params
	switch network {
	case "mainnet":
		param = &params.MainNetParams
	case "testnet":
		param = &params.TestNetParams
	case "privnet":
		param = &params.PrivNetParams
	}
	addr,err := address.NewPubKeyHashAddress(h160,param,ecc.ECDSA_Secp256k1)
	if err!=nil {
		ErrExit(err)
	}
	// Create a new script which pays to the provided address.
	pkScript, err := txscript.PayToAddrScript(addr)
	if err!=nil {
		ErrExit(err)
	}

	if len(rawTxStr)%2 != 0 {
		ErrExit(fmt.Errorf("invaild raw transaction : %s",rawTxStr))
	}
	serializedTx, err := hex.DecodeString(rawTxStr)
	if err != nil {
		ErrExit(err)
	}

	var redeemTx types.Transaction
	err = redeemTx.Deserialize(bytes.NewReader(serializedTx))
	if err != nil {
		ErrExit(err)
	}
	var kdb txscript.KeyClosure= func(types.Address) (ecc.PrivateKey, bool, error){
		return privateKey,true,nil // compressed is true
	}
	var sigScripts [][]byte
	for i:= range redeemTx.TxIn {
		sigScript,err := txscript.SignTxOutput(param,&redeemTx,i,pkScript,txscript.SigHashAll,kdb,nil,nil,ecc.ECDSA_Secp256k1)
		if err != nil {
			ErrExit(err)
		}
		sigScripts= append(sigScripts,sigScript)
	}

	for i2:=range sigScripts {
		redeemTx.TxIn[i2].SignScript = sigScripts[i2]
	}

	mtxHex, err := marshal.MessageToHex(&message.MsgTx{Tx:&redeemTx})
	if err != nil {
		ErrExit(err)
	}
	fmt.Printf("%s\n",mtxHex)
}

