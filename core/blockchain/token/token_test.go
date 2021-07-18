package token_test

import (
	"github.com/Qitmeer/qitmeer/common/hash"
	"github.com/Qitmeer/qitmeer/core/address"
	"github.com/Qitmeer/qitmeer/core/blockchain/token"
	. "github.com/Qitmeer/qitmeer/core/types"
	"github.com/Qitmeer/qitmeer/engine/txscript"
	"github.com/Qitmeer/qitmeer/params"
	"github.com/Qitmeer/qitmeer/testutils"
	"testing"
	"time"
)

const QITID CoinID = 1

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

func hashFramStr(str string) *hash.Hash {
	h, err := hash.NewHashFromStr(str)
	if err != nil {
		panic(err)
	}
	return h
}

func TestCheckTokenMint(t *testing.T) {
	tests := []struct {
		name     string
		createTx func() *Transaction
		expected bool
	}{
		{
			"invalid empty tx [0:0]",
			func() *Transaction {
				tx := &Transaction{}
				return tx
			},
			false,
		},
		{
			name:     "invalid empty tx [2:2]",
			expected: false,
			createTx: func() *Transaction {
				tx := NewTransaction()
				tx.AddTxIn(&TxInput{})
				tx.AddTxIn(&TxInput{})
				tx.AddTxOut(&TxOutput{})
				tx.AddTxOut(&TxOutput{})
				return tx
			},
		},
		{
			name:     "empty token-mint script",
			expected: false,
			createTx: func() *Transaction {
				tx := NewTransaction()
				tx.AddTxIn(&TxInput{
					PreviousOut: *NewOutPoint(&hash.Hash{}, MaxPrevOutIndex),
					Sequence:    MaxTxInSequenceNum,
					SignScript:  []byte{},
				})
				tx.AddTxIn(&TxInput{})
				tx.AddTxOut(&TxOutput{})
				tx.AddTxOut(&TxOutput{})
				return tx
			},
		},
		{
			name:     "incorrect token-mint token-id",
			expected: false,
			createTx: func() *Transaction {
				tx := NewTransaction()
				builder := txscript.NewScriptBuilder()
				builder.AddData(signature)
				builder.AddData(pubkey)
				builder.AddData(MEERID.Bytes())
				builder.AddOp(txscript.OP_TOKEN_MINT)
				mintScript, _ := builder.Script()
				tx.AddTxIn(&TxInput{
					PreviousOut: *NewOutPoint(&hash.Hash{}, MaxPrevOutIndex),
					Sequence:    MaxTxInSequenceNum,
					SignScript:  mintScript,
				})
				tx.AddTxIn(&TxInput{})
				tx.AddTxOut(&TxOutput{})
				tx.AddTxOut(&TxOutput{})
				return tx
			},
		},
		{
			name:     "token-mint",
			expected: true,
			createTx: func() *Transaction {
				tx := NewTransaction()
				builder := txscript.NewScriptBuilder()
				builder.AddData(signature)
				builder.AddData(pubkey)
				builder.AddData(QITID.Bytes())
				builder.AddOp(txscript.OP_TOKEN_MINT)
				mintScript, _ := builder.Script()
				tx.AddTxIn(&TxInput{
					PreviousOut: *NewOutPoint(&hash.Hash{}, MaxPrevOutIndex),
					Sequence:    MaxTxInSequenceNum,
					SignScript:  mintScript,
					AmountIn:    Amount{Value: 200 * 1e8, Id: QITID},
				})
				fee := int64(5400)
				tx.AddTxIn(&TxInput{
					PreviousOut: *NewOutPoint(hashFramStr("377cfb2c535be289f8e40299e8d4c234283c367e20bc5ff67ca18c1ca1337443"),
						0),
					Sequence:   MaxTxInSequenceNum,
					SignScript: []byte{txscript.OP_DATA_1},
					AmountIn:   Amount{Value: 100*1e8 + fee, Id: MEERID},
				})
				// output[0]
				builder = txscript.NewScriptBuilder()
				builder.AddOp(txscript.OP_MEER_LOCK)
				meerlockScript, _ := builder.Script()
				tx.AddTxOut(&TxOutput{Amount: Amount{Value: 100 * 1e8, Id: MEERID}, PkScript: meerlockScript})
				// output[1]
				addr, err := address.DecodeAddress("XmJvqQiDqCxEKEvSz8QaMJafkyyP4YDjL73")
				if err != nil {
					panic(err)
				}
				p2pkhScript, err := txscript.PayToAddrScript(addr)
				if err != nil {
					panic(err)
				}
				tokenReleaseScript := make([]byte, len(p2pkhScript)+1)
				tokenReleaseScript[0] = txscript.OP_TOKEN_RELEASE
				copy(tokenReleaseScript[1:], p2pkhScript)
				tx.AddTxOut(&TxOutput{Amount: Amount{Value: 200 * 1e8, Id: QITID}, PkScript: tokenReleaseScript})
				return tx
			},
		},
		{
			name:     "token-mint-with-change",
			expected: true,
			createTx: func() *Transaction {
				tx := NewTransaction()
				builder := txscript.NewScriptBuilder()
				builder.AddData(signature)
				builder.AddData(pubkey)
				builder.AddData(QITID.Bytes())
				builder.AddOp(txscript.OP_TOKEN_MINT)
				mintScript, _ := builder.Script()
				tx.AddTxIn(&TxInput{
					PreviousOut: *NewOutPoint(&hash.Hash{}, MaxPrevOutIndex),
					Sequence:    MaxTxInSequenceNum,
					SignScript:  mintScript,
					AmountIn:    Amount{Value: 100 * 1e8, Id: QITID},
				})
				fee := int64(5400)
				mint := int64(90 * 1e8)   // 90meer
				change := int64(10 * 1e8) // 10meer
				tx.AddTxIn(&TxInput{
					PreviousOut: *NewOutPoint(hashFramStr("377cfb2c535be289f8e40299e8d4c234283c367e20bc5ff67ca18c1ca1337443"),
						0),
					Sequence:   MaxTxInSequenceNum,
					SignScript: []byte{txscript.OP_DATA_1},
					AmountIn:   Amount{Value: mint + change + fee, Id: MEERID},
				})
				// output[0]
				builder = txscript.NewScriptBuilder()
				builder.AddOp(txscript.OP_MEER_LOCK)
				meerlockScript, _ := builder.Script()
				tx.AddTxOut(&TxOutput{Amount: Amount{Value: mint, Id: MEERID}, PkScript: meerlockScript})
				// output[1]
				addr, err := address.DecodeAddress("XmJvqQiDqCxEKEvSz8QaMJafkyyP4YDjL73")
				if err != nil {
					panic(err)
				}
				p2pkhScript, err := txscript.PayToAddrScript(addr)
				if err != nil {
					panic(err)
				}
				tokenReleaseScript := make([]byte, len(p2pkhScript)+1)
				tokenReleaseScript[0] = txscript.OP_TOKEN_RELEASE
				copy(tokenReleaseScript[1:], p2pkhScript)
				tx.AddTxOut(&TxOutput{Amount: Amount{Value: 100 * 1e8, Id: QITID}, PkScript: tokenReleaseScript})

				// output[2]
				meerChangeScript := make([]byte, len(p2pkhScript)+1)
				meerChangeScript[0] = txscript.OP_MEER_CHANGE
				copy(meerChangeScript[1:], p2pkhScript)
				tx.AddTxOut(&TxOutput{Amount: Amount{Value: change, Id: MEERID}, PkScript: meerChangeScript})
				return tx
			},
		},
	}

	for i, test := range tests {
		if got := token.IsTokenMint(test.createTx()); got != test.expected {
			_, _, _, err := token.CheckTokenMint(test.createTx())
			t.Errorf("failed test[%d]:[%v], expect [%v] but [%v], error:[%v]", i, test.name, test.expected, got, err)
		}
	}

}

func TestCheckTokenUnMint(t *testing.T) {
	tests := []struct {
		name     string
		expected bool
		createTx func() *Transaction
	}{
		{
			name:     "can not unmint meer",
			expected: false,
			createTx: func() *Transaction {
				tx := NewTransaction()
				builder := txscript.NewScriptBuilder()
				builder.AddData(signature)
				builder.AddData(pubkey)
				builder.AddData(MEERID.Bytes())
				builder.AddOp(txscript.OP_TOKEN_UNMINT)
				unmintScript, _ := builder.Script()
				tx.AddTxIn(&TxInput{
					PreviousOut: *NewOutPoint(&hash.Hash{}, MaxPrevOutIndex),
					Sequence:    MaxTxInSequenceNum,
					SignScript:  unmintScript,
				})
				tx.AddTxIn(&TxInput{})
				tx.AddTxOut(&TxOutput{})
				tx.AddTxOut(&TxOutput{})
				return tx
			},
		},
		{
			name:     "invalid input from meer",
			expected: false,
			createTx: func() *Transaction {
				tx := NewTransaction()
				builder := txscript.NewScriptBuilder()
				builder.AddData(signature)
				builder.AddData(pubkey)
				builder.AddData(QITID.Bytes())
				builder.AddOp(txscript.OP_TOKEN_UNMINT)
				unmintScript, _ := builder.Script()
				tx.AddTxIn(&TxInput{
					PreviousOut: *NewOutPoint(&hash.Hash{}, MaxPrevOutIndex),
					Sequence:    MaxTxInSequenceNum,
					SignScript:  unmintScript,
				})
				tx.AddTxIn(&TxInput{
					PreviousOut: *NewOutPoint(hashFramStr("377cfb2c535be289f8e40299e8d4c234283c367e20bc5ff67ca18c1ca1337443"),
						0),
					Sequence:   MaxTxInSequenceNum,
					SignScript: []byte{txscript.OP_DATA_1},
				})
				tx.AddTxOut(&TxOutput{})
				return tx
			},
		},
		{
			name:     "token unmint",
			expected: true,
			createTx: func() *Transaction {
				tx := NewTransaction()
				builder := txscript.NewScriptBuilder()
				builder.AddData(signature)
				builder.AddData(pubkey)
				builder.AddData(QITID.Bytes())
				builder.AddOp(txscript.OP_TOKEN_UNMINT)
				unmintScript, _ := builder.Script()
				tx.AddTxIn(&TxInput{
					PreviousOut: *NewOutPoint(&hash.Hash{}, MaxPrevOutIndex),
					Sequence:    MaxTxInSequenceNum,
					SignScript:  unmintScript,
					AmountIn:    Amount{Value: 10 * 1e8, Id: MEERID},
				})
				tx.AddTxIn(&TxInput{
					PreviousOut: *NewOutPoint(hashFramStr("377cfb2c535be289f8e40299e8d4c234283c367e20bc5ff67ca18c1ca1337443"),
						0),
					Sequence:   MaxTxInSequenceNum,
					SignScript: []byte{txscript.OP_DATA_1},
					AmountIn:   Amount{Value: 100 * 1e8, Id: QITID},
				})
				// output[0]
				builder = txscript.NewScriptBuilder()
				builder.AddOp(txscript.OP_TOKEN_DESTORY)
				tokenDestoryScript, _ := builder.Script()
				tx.AddTxOut(&TxOutput{Amount: Amount{Value: 100 * 1e8, Id: QITID}, PkScript: tokenDestoryScript})
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
				tx.AddTxOut(&TxOutput{Amount: Amount{Value: 10*1e8 - fee, Id: MEERID}, PkScript: meerReleaseScript})
				return tx
			},
		},
		{
			name:     "token unmint with change",
			expected: true,
			createTx: func() *Transaction {
				tx := NewTransaction()
				builder := txscript.NewScriptBuilder()
				builder.AddData(signature)
				builder.AddData(pubkey)
				builder.AddData(QITID.Bytes())
				builder.AddOp(txscript.OP_TOKEN_UNMINT)
				unmintScript, _ := builder.Script()
				tx.AddTxIn(&TxInput{
					PreviousOut: *NewOutPoint(&hash.Hash{}, MaxPrevOutIndex),
					Sequence:    MaxTxInSequenceNum,
					SignScript:  unmintScript,
					AmountIn:    Amount{Value: 10 * 1e8, Id: MEERID},
				})
				tx.AddTxIn(&TxInput{
					PreviousOut: *NewOutPoint(hashFramStr("377cfb2c535be289f8e40299e8d4c234283c367e20bc5ff67ca18c1ca1337443"),
						0),
					Sequence:   MaxTxInSequenceNum,
					SignScript: []byte{txscript.OP_DATA_1},
					AmountIn:   Amount{Value: 100 * 1e8, Id: QITID},
				})
				// output[0]
				builder = txscript.NewScriptBuilder()
				builder.AddOp(txscript.OP_TOKEN_DESTORY)
				tokenDestoryScript, _ := builder.Script()
				tx.AddTxOut(&TxOutput{Amount: Amount{Value: 99 * 1e8, Id: QITID}, PkScript: tokenDestoryScript})
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
				tx.AddTxOut(&TxOutput{Amount: Amount{Value: 10*1e8 - fee, Id: MEERID}, PkScript: meerReleaseScript})
				// output[2] token-change
				tokenChangeScript := make([]byte, len(p2pkhScript)+1)
				tokenChangeScript[0] = txscript.OP_TOKEN_CHANGE
				copy(tokenChangeScript[1:], p2pkhScript)
				tx.AddTxOut(&TxOutput{Amount: Amount{Value: 1 * 1e8, Id: QITID}, PkScript: tokenChangeScript})
				return tx
			},
		},
	}
	for i, test := range tests {
		if got := token.IsTokenUnMint(test.createTx()); got != test.expected {
			_, _, _, err := token.CheckTokenUnMint(test.createTx())
			t.Errorf("failed test[%d]:[%v], expect [%v] but [%v], error:[%v]", i, test.name, test.expected, got, err)
		}
	}
}

func TestTokenIssue(t *testing.T) {
	args := []string{"--modules=miner", "--modules=qitmeer", "--miningaddr=RmFa5hnPd3uQRpzr3xWTfr8EFZdX7dS1qzV"}
	netParams := params.PrivNetParam.Params
	h, err := testutils.NewHarness(t, netParams, args...)
	if err != nil {
		t.Errorf("failed to create test harness")
	}
	defer func() {
		err := h.Teardown()
		if err != nil {
			t.Errorf("failed to teardown test harness")
		}
	}()
	if err = h.Setup(); err != nil {
		t.Fatalf("failed to setup test harness: %v", err)
	}
	time.Sleep(500 * time.Millisecond)
	testutils.AssertBlockOrderAndHeight(t, h, 1, 1, 0)
	startHeight, err := h.Client.MainHeight()
	if err != nil {
		t.Errorf("failed to get main height: %v", err)
	}
	matureHeight := uint64(netParams.CoinbaseMaturity)
	if matureHeight > startHeight {
		testutils.GenerateBlock(t, h, matureHeight-startHeight)
	}
	testutils.AssertBlockOrderAndHeight(t, h, 17, 17, 16)
}

//func generateKeys() {
//	printKey := func(key []byte){
//		for i, v := range key {
//			if i != 0 && i%8 == 0 {
//				fmt.Printf("\n")
//			}
//			fmt.Printf("0x%02x,", v)
//		}
//		fmt.Println()
//	}
//	// private key
//	priKeyStr := "9af3b7c0b4f19635f90a5fc722defb961ac43508c66ffe5df992e9314f2a2948"
//	priKey,_ :=hex.DecodeString(priKeyStr)
//	key,_ := secp256k1.PrivKeyFromBytes(priKey)
//	fmt.Printf("Private key : %x\n",string(privateKey))
//	printKey(priKey)
//
//	// public key
//	pubKey := key.PubKey()
//	fmt.Printf("Public key len(0x%x): %x\n", len(pubKey.SerializeCompressed()),
//		pubKey.SerializeCompressed())
//	printKey(pubKey.SerializeCompressed())
//
//	// sigature
//	message := "qitmeer"
//	messageHash := hash.HashB([]byte(message))
//	r, s, err := schnorr.Sign(key, messageHash)
//	signature := schnorr.Signature{r,s}
//	if err != nil {
//		panic(err)
//	}
//	sig := signature.Serialize()
//	fmt.Printf("Signature len(0x%x): %x\n", len(sig), sig)
//	printKey(sig)
//
//}
//
//func init() {
//	generateKeys()
//}
