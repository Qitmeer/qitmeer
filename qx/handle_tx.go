package qx

import (
	"bytes"
	"encoding/hex"
	"fmt"
	"github.com/Qitmeer/qitmeer-lib/common/hash"
	"github.com/Qitmeer/qitmeer-lib/common/marshal"
	"github.com/Qitmeer/qitmeer-lib/core/address"
	"github.com/Qitmeer/qitmeer-lib/core/message"
	"github.com/Qitmeer/qitmeer-lib/core/types"
	"github.com/Qitmeer/qitmeer-lib/crypto/ecc"
	"github.com/Qitmeer/qitmeer-lib/engine/txscript"
	"github.com/Qitmeer/qitmeer-lib/params"
)

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
