package qx

import (
	"bytes"
	"encoding/hex"
	"fmt"
	"github.com/Qitmeer/qng-core/common/hash"
	"github.com/Qitmeer/qng-core/common/marshal"
	"github.com/Qitmeer/qng-core/core/address"
	"github.com/Qitmeer/qng-core/core/json"
	"github.com/Qitmeer/qng-core/core/types"
	"github.com/Qitmeer/qng-core/crypto/ecc"
	"github.com/Qitmeer/qng-core/engine/txscript"
	"github.com/Qitmeer/qng-core/params"
	"sort"
	"time"
)

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

func TxSign(privkeyStrs []string, rawTxStr string, network string, pks []string) (string, error) {
	privateKeyMaps := map[string]ecc.PrivateKey{}
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
	for _, privkeyStr := range privkeyStrs {
		privkeyByte, err := hex.DecodeString(privkeyStr)
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
		privateKeyMaps[addr.String()] = privateKey
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

	if len(redeemTx.TxIn) != len(pks) {
		return "", fmt.Errorf("input pkscript len :%d not equal %d txIn length", len(pks), len(redeemTx.TxIn))
	}
	var sigScripts [][]byte
	for i := range redeemTx.TxIn {
		pkScript, err := hex.DecodeString(pks[i])
		if err != nil {
			return "", fmt.Errorf("pkscript %d error:%s", i, err.Error())
		}
		_, addrs, _, err := txscript.ExtractPkScriptAddrs(pkScript, param)
		privateKey, ok := privateKeyMaps[addrs[0].String()]
		if !ok {
			return "", fmt.Errorf("addrress : %s  privatekey not exist,can not sign", addrs[0].String())
		}
		var kdb txscript.KeyClosure = func(types.Address) (ecc.PrivateKey, bool, error) {
			return privateKey, true, nil // compressed is true
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

func TxSignSTDO(privkeyStr string, rawTxStr string, network string, pks []string) {
	mtxHex, err := TxSign([]string{privkeyStr}, rawTxStr, network, pks)
	if err != nil {
		ErrExit(err)
	}
	fmt.Printf("%s\n", mtxHex)
}
