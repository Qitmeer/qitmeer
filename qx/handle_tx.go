package qx

import (
	"bytes"
	"encoding/hex"
	"errors"
	"fmt"
	"github.com/Qitmeer/qitmeer/common/hash"
	"github.com/Qitmeer/qitmeer/common/marshal"
	"github.com/Qitmeer/qitmeer/core/address"
	"github.com/Qitmeer/qitmeer/core/json"
	"github.com/Qitmeer/qitmeer/core/types"
	"github.com/Qitmeer/qitmeer/crypto/ecc"
	"github.com/Qitmeer/qitmeer/engine/txscript"
	"github.com/Qitmeer/qitmeer/params"
	"sort"
	"time"
)

const (
	STANDARD_TX = 0
	LOCK_TX     = 1
	OTHER       = 2
)

type SignInputData struct {
	PrivateKey string
	Params     interface{}
	Type       uint
}

type Amount struct {
	TargetLockTime int64
	Value          int64
	Id             types.CoinID
}

type Input struct {
	TxID     string
	OutIndex uint32
}

func TxEncode(version uint32, lockTime uint32, timestamp *time.Time, inputs []Input, outputs map[string]Amount) (string, error) {
	mtx := types.NewTransaction()
	mtx.Version = uint32(version)
	if lockTime != 0 {
		mtx.LockTime = uint32(lockTime)
	}
	if timestamp != nil {
		mtx.Timestamp = *timestamp
	}

	for _, vout := range inputs {
		txHash, err := hash.NewHashFromStr(vout.TxID)
		if err != nil {
			return "", err
		}
		prevOut := types.NewOutPoint(txHash, vout.OutIndex)
		txIn := types.NewTxInput(prevOut, []byte{})
		if lockTime != 0 {
			txIn.Sequence = types.MaxTxInSequenceNum - 1
		}
		mtx.AddTxIn(txIn)
	}

	outputsSlice := []string{}
	for k := range outputs {
		outputsSlice = append(outputsSlice, k)
	}
	sort.Strings(outputsSlice)

	for _, encodedAddr := range outputsSlice {
		amount := outputs[encodedAddr]
		if amount.Value <= 0 || amount.Value > types.MaxAmount {
			return "", fmt.Errorf("invalid amount: 0 >= %v "+
				"> %v", amount, types.MaxAmount)
		}

		addr, err := address.DecodeAddress(encodedAddr)
		if err != nil {
			return "", fmt.Errorf("could not decode "+
				"address: %v", err)
		}

		switch addr.(type) {
		case *address.PubKeyHashAddress:
		case *address.ScriptHashAddress:
		default:
			return "", fmt.Errorf("invalid type: %T", addr)
		}

		pkScript, err := txscript.PayToAddrScript(addr)
		if err != nil {
			return "", err
		}
		if amount.TargetLockTime > 0 {
			pkScript, err = txscript.PayToCLTVPubKeyHashScript(addr.Script(), amount.TargetLockTime)
			if err != nil {
				return "", err
			}
		}
		txOut := types.NewTxOutput(types.Amount{Value: amount.Value, Id: amount.Id}, pkScript)
		mtx.AddTxOut(txOut)
	}
	mtxHex, err := mtx.Serialize()
	if err != nil {
		return "", err
	}
	return hex.EncodeToString(mtxHex), nil
}

func DecodePkString(pk string) (string, error) {
	b, err := txscript.PkStringToScript(pk)
	if err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}

func TxSign(signInputData []SignInputData, rawTxStr string, network string) (string, error) {
	var param *params.Params
	switch network {
	case "mainnet":
		param = &params.MainNetParams
	case "testnet":
		param = &params.TestNetParams
	case "privnet":
		param = &params.PrivNetParams
	case "mixnet":
		param = &params.MixNetParams
	}
	if len(rawTxStr)%2 != 0 {
		return "", fmt.Errorf("invaild raw transaction : %s", rawTxStr)
	}
	serializedTx, err := hex.DecodeString(rawTxStr)
	if err != nil {
		return "", err
	}

	var redeemTx types.Transaction
	err = redeemTx.Deserialize(bytes.NewReader(serializedTx))
	if err != nil {
		return "", err
	}

	if len(redeemTx.TxIn) > len(signInputData) && signInputData[0].Type == STANDARD_TX {
		for i := 0; i < len(redeemTx.TxIn)-len(signInputData); i++ {
			signInputData = append(signInputData, signInputData[0])
		}
	}
	if len(redeemTx.TxIn) != len(signInputData) {
		return "", fmt.Errorf("input pkscript len :%d not equal %d txIn length", len(signInputData), len(redeemTx.TxIn))
	}
	var sigScripts [][]byte
	for i := range redeemTx.TxIn {
		var pkScript []byte

		privkeyByte, err := hex.DecodeString(signInputData[i].PrivateKey)
		if err != nil {
			return "", err
		}
		if len(privkeyByte) != 32 {
			return "", fmt.Errorf("invaid ec private key bytes: %d", len(privkeyByte))
		}
		privateKey, pubKey := ecc.Secp256k1.PrivKeyFromBytes(privkeyByte)
		h160 := hash.Hash160(pubKey.SerializeCompressed())
		addr, err := address.NewPubKeyHashAddress(h160, param, ecc.ECDSA_Secp256k1)
		if err != nil {
			return "", err
		}
		var kdb txscript.KeyClosure = func(types.Address) (ecc.PrivateKey, bool, error) {
			return privateKey, true, nil // compressed is true
		}
		switch signInputData[i].Type {
		case STANDARD_TX:
			pkScript, err = txscript.PayToAddrScript(addr)
			if err != nil {
				return "", err
			}
		case LOCK_TX:
			height, ok := signInputData[i].Params.(int64)
			if !ok {
				return "", fmt.Errorf("invaid lock tx input params: %v", signInputData[i].Params)
			}
			pkScript, err = txscript.PayToCLTVPubKeyHashScript(addr.Script(), height)
			if err != nil {
				return "", err
			}
		case OTHER:
			pkHex, ok := signInputData[i].Params.(string)
			if !ok {
				return "", fmt.Errorf("invaid pkscript hex string: %v", signInputData[i].Params)
			}
			pkScript, err = hex.DecodeString(pkHex)
			if err != nil {
				return "", err
			}
		}
		sigScript, err := txscript.SignTxOutput(param, &redeemTx, i, pkScript, txscript.SigHashAll, kdb, nil, nil, ecc.ECDSA_Secp256k1)
		if err != nil {
			return "", err
		}
		sigScripts = append(sigScripts, sigScript)
	}

	for i2 := range sigScripts {
		redeemTx.TxIn[i2].SignScript = sigScripts[i2]
	}

	mtxHex, err := marshal.MessageToHex(&redeemTx)
	if err != nil {
		return "", err
	}
	return mtxHex, nil
}

func TxDecode(network string, rawTxStr string) {
	var param *params.Params
	switch network {
	case "mainnet":
		param = &params.MainNetParams
	case "testnet":
		param = &params.TestNetParams
	case "privnet":
		param = &params.PrivNetParams
	case "mixnet":
		param = &params.MixNetParams
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

func TxEncodeSTDO(version TxVersionFlag, lockTime TxLockTimeFlag, txIn TxInputsFlag, txOut TxOutputsFlag) {
	txInputs := []Input{}
	txOutputs := make(map[string]Amount)
	for _, input := range txIn.inputs {
		txInputs = append(txInputs, Input{
			TxID:     hex.EncodeToString(input.txhash),
			OutIndex: input.index,
		})
	}
	for _, output := range txOut.outputs {
		atomic, err := types.NewAmount(output.amount)
		if err != nil {
			ErrExit(fmt.Errorf("fail to create the currency amount from a "+
				"floating point value %f : %w", output.amount, err))
		}
		txOutputs[output.target] = Amount{
			TargetLockTime: 0,
			Id:             types.MEERID,
			Value:          atomic.Value,
		}
	}
	mtxHex, err := TxEncode(uint32(version), uint32(lockTime), nil, txInputs, txOutputs)
	if err != nil {
		ErrExit(err)
	}
	fmt.Printf("%s\n", mtxHex)
}

func TxSignSTDO(privkeyStr string, params []SignInputData, rawTxStr string, network string) {
	if privkeyStr != "" && params != nil && len(params) > 0 {
		ErrExit(errors.New("-k and -p can't be used with the same time"))
	}
	var inputData []SignInputData

	if privkeyStr != "" {
		inputData = []SignInputData{
			{PrivateKey: privkeyStr, Type: STANDARD_TX},
		}
	} else {
		inputData = params
	}
	mtxHex, err := TxSign(inputData, rawTxStr, network)
	if err != nil {
		ErrExit(err)
	}
	fmt.Printf("%s\n", mtxHex)
}
