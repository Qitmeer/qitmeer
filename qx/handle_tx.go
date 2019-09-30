package qx

import (
	"bytes"
	"encoding/hex"
	"fmt"
	"github.com/Qitmeer/qitmeer/common/hash"
	"github.com/Qitmeer/qitmeer/common/marshal"
	"github.com/Qitmeer/qitmeer/core/address"
	"github.com/Qitmeer/qitmeer/core/message"
	"github.com/Qitmeer/qitmeer/core/types"
	"github.com/Qitmeer/qitmeer/crypto/ecc"
	"github.com/Qitmeer/qitmeer/engine/txscript"
	"github.com/Qitmeer/qitmeer/params"
)

func TxEncode(version uint32, lockTime uint32, inputs map[string]uint32, outputs map[string]uint64) (string, error) {
	mtx := types.NewTransaction()
	mtx.Version = uint32(version)
	if lockTime != 0 {
		mtx.LockTime = uint32(lockTime)
	}

	for txId, vout := range inputs {
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

	for encodedAddr, amount := range outputs {
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
