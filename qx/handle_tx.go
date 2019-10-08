package qx

import (
	"bytes"
	"encoding/hex"
	"fmt"
	"github.com/Qitmeer/qitmeer/common/hash"
	"github.com/Qitmeer/qitmeer/common/marshal"
	"github.com/Qitmeer/qitmeer/core/address"
	"github.com/Qitmeer/qitmeer/core/json"
	"github.com/Qitmeer/qitmeer/core/message"
	"github.com/Qitmeer/qitmeer/core/types"
	"github.com/Qitmeer/qitmeer/crypto/ecc"
	"github.com/Qitmeer/qitmeer/engine/txscript"
	"github.com/Qitmeer/qitmeer/params"
	"github.com/pkg/errors"
	"sort"
)

func TxEncode(version uint32, lockTime uint32, inputs map[string]uint32, outputs map[string]uint64) (string, error) {
	mtx := types.NewTransaction()
	mtx.Version = uint32(version)
	if lockTime != 0 {
		mtx.LockTime = uint32(lockTime)
	}

	inputsSlice := []string{}
	for k := range inputs {
		inputsSlice = append(inputsSlice, k)
	}
	sort.Strings(inputsSlice)

	for _,txId := range inputsSlice {
		vout:=inputs[txId]
		txHash, err := hash.NewHashFromStr(txId)
		if err != nil {
			return "", err
		}
		prevOut := types.NewOutPoint(txHash, vout)
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

	for _,encodedAddr:= range outputsSlice {
		amount:=outputs[encodedAddr]
		if amount <= 0 || amount > types.MaxAmount {
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
		txOut := types.NewTxOutput(amount, pkScript)
		mtx.AddTxOut(txOut)
	}
	mtxHex, err := mtx.Serialize()
	if err != nil {
		return "", err
	}
	return hex.EncodeToString(mtxHex), nil
}

func TxSign(privkeyStr string, rawTxStr string, network string) (string, error) {
	privkeyByte, err := hex.DecodeString(privkeyStr)
	if err != nil {
		return "", err
	}
	if len(privkeyByte) != 32 {
		return "", fmt.Errorf("invaid ec private key bytes: %d", len(privkeyByte))
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
	addr, err := address.NewPubKeyHashAddress(h160, param, ecc.ECDSA_Secp256k1)
	if err != nil {
		return "", err
	}
	// Create a new script which pays to the provided address.
	pkScript, err := txscript.PayToAddrScript(addr)
	if err != nil {
		return "", err
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
	var kdb txscript.KeyClosure = func(types.Address) (ecc.PrivateKey, bool, error) {
		return privateKey, true, nil // compressed is true
	}
	var sigScripts [][]byte
	for i := range redeemTx.TxIn {
		sigScript, err := txscript.SignTxOutput(param, &redeemTx, i, pkScript, txscript.SigHashAll, kdb, nil, nil, ecc.ECDSA_Secp256k1)
		if err != nil {
			return "", err
		}
		sigScripts = append(sigScripts, sigScript)
	}

	for i2 := range sigScripts {
		redeemTx.TxIn[i2].SignScript = sigScripts[i2]
	}

	mtxHex, err := marshal.MessageToHex(&message.MsgTx{Tx: &redeemTx})
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
	mtxHex, err := TxEncode(uint32(version), uint32(lockTime), txInputs, txOutputs)
	if err != nil {
		ErrExit(err)
	}
	fmt.Printf("%s\n", mtxHex)
}

func TxSignSTDO(privkeyStr string, rawTxStr string, network string) {
	mtxHex, err := TxSign(privkeyStr, rawTxStr, network)
	if err != nil {
		ErrExit(err)
	}
	fmt.Printf("%s\n", mtxHex)
}
