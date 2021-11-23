package blockchain

import (
	"bytes"
	"encoding/hex"
	"github.com/Qitmeer/qng-core/common/hash"
	"github.com/Qitmeer/qng-core/core/address"
	"github.com/Qitmeer/qng-core/core/types"
	"github.com/Qitmeer/qng-core/engine/txscript"
	"github.com/Qitmeer/qng-core/params"
	"testing"
)

const QITID types.CoinID = 1

func Test_CheckTransactionSanity(t *testing.T) {
	txStr := "0100000001df44f16415ecc836e1c8e7f8cd3af896dc1c81f0f4e159535e744a0f75c76b7501000000ffffffff0101000093459e0b0000001976a9149a2aff5fcbfc29f9a30262bd6ba976f08f9fcf9188ac00000000000000004c790a60016a4730440220336757a0c9bdcb2510fe627520ad79694fc36d1880bf24ecc7496a65fc65a991022073853aba8fa9f3f29a8f1135f3b570aa3f804887cdbfa260d1544a0ea8c9aa530121036f2b25ef58a673b255d63802d1a9f25ffeb6f2a78c0eab33ee17831e07e5e487"
	if len(txStr)%2 != 0 {
		txStr = "0" + txStr
	}
	serializedTx, err := hex.DecodeString(txStr)
	if err != nil {
		t.Fatal(err)
	}
	var tx types.Transaction
	err = tx.Deserialize(bytes.NewReader(serializedTx))
	if err != nil {
		t.Fatal(err)
	}
	err = checkTransactionSanityForAllNet(&tx)
	if err != nil {
		t.Fatal(err)
	}

	// We create an attack transacton data
	attackerPkScript, err := hex.DecodeString("76a914c0f0b73c320e1fe38eb1166a57b953e509c8f93e88ac")
	if err != nil {
		panic(err)
	}
	tx.AddTxOut(&types.TxOutput{
		Amount:   types.Amount{Value: 99999999999999999, Id: types.MEERID},
		PkScript: attackerPkScript,
	})

	err = checkTransactionSanityForAllNet(&tx)
	if err == nil {
		t.Fatal("Successful attack")
	}
}

func checkTransactionSanityForAllNet(tx *types.Transaction) error {
	err := CheckTransactionSanity(tx, &params.TestNetParams)
	if err != nil {
		return err
	}

	err = CheckTransactionSanity(tx, &params.PrivNetParams)
	if err != nil {
		return err
	}

	err = CheckTransactionSanity(tx, &params.MixNetParams)
	if err != nil {
		return err
	}

	err = CheckTransactionSanity(tx, &params.MainNetParams)
	if err != nil {
		return err
	}
	return nil
}

func TestTokenStateRoot(t *testing.T) {
	bc := BlockChain{}
	expected := "5b7d48b6c505d90b21355081cf4f5a332a925ac87e24ceedd3ddf02e0f387cc3"
	stateRoot := bc.CalculateTokenStateRoot([]*types.Tx{types.NewTx(createTx())}, nil)
	if stateRoot.String() != expected {
		t.Fatalf("token state root is %s, but expected is %s", stateRoot, expected)
	}
}

var (
	// private key
	privateKey []byte = []byte{
		0x9a, 0xf3, 0xb7, 0xc0, 0xb4, 0xf1, 0x96, 0x35,
		0xf9, 0x0a, 0x5f, 0xc7, 0x22, 0xde, 0xfb, 0x96,
		0x1a, 0xc4, 0x35, 0x08, 0xc6, 0x6f, 0xfe, 0x5d,
		0xf9, 0x92, 0xe9, 0x31, 0x4f, 0x2a, 0x29, 0x48,
	}
	// compressed public key
	pubkey []byte = []byte{
		0x02, 0xab, 0xb1, 0x3c, 0xd5, 0x26, 0x0d, 0x3e,
		0x9f, 0x8b, 0xc3, 0xdb, 0x86, 0x87, 0x14, 0x7a,
		0xce, 0x7b, 0x6e, 0x5b, 0x63, 0xb0, 0x61, 0xaf,
		0xe3, 0x7d, 0x09, 0xa8, 0xe4, 0x55, 0x0c, 0xd1,
		0x74,
	}
	// schnorr signature for hash.HashB([]byte("qitmeer"))
	signature []byte = []byte{
		0xb2, 0xcb, 0x95, 0xbb, 0x27, 0x32, 0xac, 0xb9,
		0xcc, 0x14, 0x5f, 0xe8, 0x78, 0xc8, 0x99, 0xc8,
		0xd0, 0xf6, 0x19, 0x0a, 0x3b, 0x97, 0xcd, 0x44,
		0xf1, 0x20, 0xaa, 0x78, 0x17, 0xc8, 0x08, 0x6d,
		0x43, 0xc1, 0x6d, 0x61, 0x1d, 0xa6, 0x40, 0x1d,
		0xd1, 0x72, 0x3b, 0x4d, 0x9f, 0x6e, 0xc1, 0x76,
		0xd8, 0x4b, 0x23, 0xaa, 0x82, 0xc2, 0xca, 0x44,
		0xf9, 0x4a, 0x9a, 0x24, 0xd2, 0x7e, 0x80, 0x7b,
	}
)

func createTx() *types.Transaction {
	tx := types.NewTransaction()
	builder := txscript.NewScriptBuilder()
	builder.AddData(signature)
	builder.AddData(pubkey)
	builder.AddData(QITID.Bytes())
	builder.AddOp(txscript.OP_TOKEN_UNMINT)
	unmintScript, _ := builder.Script()
	tx.AddTxIn(&types.TxInput{
		PreviousOut: *types.NewOutPoint(&hash.Hash{}, types.TokenPrevOutIndex),
		Sequence:    uint32(types.TxTypeTokenUnmint),
		SignScript:  unmintScript,
		AmountIn:    types.Amount{Value: 10 * 1e8, Id: types.MEERID},
	})

	txid := hash.MustHexToDecodedHash("377cfb2c535be289f8e40299e8d4c234283c367e20bc5ff67ca18c1ca1337443")
	tx.AddTxIn(&types.TxInput{
		PreviousOut: *types.NewOutPoint(&txid,
			0),
		Sequence:   types.MaxTxInSequenceNum,
		SignScript: []byte{txscript.OP_DATA_1},
		AmountIn:   types.Amount{Value: 100 * 1e8, Id: QITID},
	})
	// output[0]
	builder = txscript.NewScriptBuilder()
	builder.AddOp(txscript.OP_TOKEN_DESTORY)
	tokenDestoryScript, _ := builder.Script()
	tx.AddTxOut(&types.TxOutput{Amount: types.Amount{Value: 99 * 1e8, Id: QITID}, PkScript: tokenDestoryScript})
	// output[1]
	addr, err := address.DecodeAddress("XmJvqQiDqCxEKEvSz8QaMJafkyyP4YDjL73")
	if err != nil {
		panic(err)
	}
	p2pkhScript, err := txscript.PayToAddrScript(addr)
	if err != nil {
		panic(err)
	}
	meerReleaseScript := make([]byte, len(p2pkhScript)+1)
	meerReleaseScript[0] = txscript.OP_MEER_RELEASE
	copy(meerReleaseScript[1:], p2pkhScript)
	fee := int64(5400)
	tx.AddTxOut(&types.TxOutput{Amount: types.Amount{Value: 10*1e8 - fee, Id: types.MEERID}, PkScript: meerReleaseScript})
	// output[2] token-change
	tokenChangeScript := make([]byte, len(p2pkhScript)+1)
	tokenChangeScript[0] = txscript.OP_TOKEN_CHANGE
	copy(tokenChangeScript[1:], p2pkhScript)
	tx.AddTxOut(&types.TxOutput{Amount: types.Amount{Value: 1 * 1e8, Id: QITID}, PkScript: tokenChangeScript})
	return tx
}
