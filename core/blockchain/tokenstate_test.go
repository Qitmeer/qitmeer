// Copyright (c) 2020 The qitmeer developers
// Use of this source code is governed by an ISC
// license that can be found in the LICENSE file.

package blockchain

import (
	"bytes"
	"encoding/hex"
	"github.com/Qitmeer/qitmeer/common/hash"
	"github.com/Qitmeer/qitmeer/core/address"
	"github.com/Qitmeer/qitmeer/core/dbnamespace"
	"github.com/Qitmeer/qitmeer/core/types"
	"github.com/Qitmeer/qitmeer/database"
	_ "github.com/Qitmeer/qitmeer/database/ffldb"
	"github.com/Qitmeer/qitmeer/engine/txscript"
	"github.com/Qitmeer/qitmeer/params"
	"io/ioutil"
	"os"
	"reflect"
	"testing"
)

func bytesFromStr(s string) []byte {
	b, _ := hex.DecodeString(s)
	return b
}

func TestTokenStateSerialization(t *testing.T) {
	tests := []struct {
		name  string
		state *tokenState
		bytes []byte
		ok    bool
	}{
		{name: "test1",
			state: &tokenState{
				balances: tokenBalances{
					types.QITID: tokenBalance{balance: 200 * 1e8, lockedMeer: 100 * 1e8}},
				updates: []balanceUpdate{
					{typ: tokenMint,
						meerAmount:  100 * 1e8,
						tokenAmount: types.Amount{Value: 200 * 1e8, Id: types.QITID}},
				},
			},
			bytes: bytesFromStr("000101c9bfde8f00a49faec7000101a49faec70001c9bfde8f00"),
			ok:    true,
		},
	}
	for i, test := range tests {
		serialized, err := serializeTokeState(*test.state)
		if !bytes.Equal(serialized, test.bytes) {
			t.Errorf("test[%d][%s] failed: want %x but got %x : %w", i, test.name, test.bytes, serialized, err)
		}
		deserialized, err := deserializeTokenState(test.bytes)
		if !reflect.DeepEqual(deserialized, test.state) {
			t.Errorf("test[%d][%s] failed: want %v but got %v", i, test.name, test.state, deserialized)
		}
	}

}

func TestTokenStateDB(t *testing.T) {
	//create a test token state database
	dbPath, err := ioutil.TempDir("", "test_tokenstate_db")
	if err != nil {
		t.Fatalf("failed to create token state db : %v", err)
	}
	// clean up db file when the test is finished
	defer os.RemoveAll(dbPath)

	tokendb, err := database.Create("ffldb", dbPath, params.PrivNetParam.Net)
	if err != nil {
		t.Fatalf("failed to create token state db : %v", err)
	}
	defer tokendb.Close()

	// prepare token db.
	err = tokendb.Update(func(dbTx database.Tx) error {
		_, err := dbTx.Metadata().CreateBucketIfNotExists(dbnamespace.TokenBucketName)
		return err
	})
	if err != nil {
		t.Fatal(err)
	}

	// put a test token state record into tokenstate db
	ts := tokenState{
		balances: tokenBalances{
			types.QITID: tokenBalance{balance: 200 * 1e8, lockedMeer: 100 * 1e8}},
		updates: []balanceUpdate{
			{typ: tokenMint,
				meerAmount:  100 * 1e8,
				tokenAmount: types.Amount{Value: 200 * 1e8, Id: types.QITID}},
		}}
	// create a fake block id for testing
	bid := uint32(0xa)

	err = tokendb.Update(func(dbTx database.Tx) error {
		return dbPutTokenState(dbTx, bid, ts)
	})
	if err != nil {
		t.Fatalf("%v", err)
	}

	// fetch record from tokenstate db
	var tsfromdb *tokenState
	err = tokendb.View(func(dbTx database.Tx) error {
		tsfromdb, err = dbFetchTokenState(dbTx, bid)
		return err
	})
	if err != nil {
		t.Fatalf("%v", err)
	}

	// compare result
	if !reflect.DeepEqual(ts, *tsfromdb) {
		t.Fatalf("token state put db is %v ,but from db is %v", ts, *tsfromdb)
	}
}

func TestTokenStateRoot(t *testing.T) {
	bc := BlockChain{}
	expected := "033138b7ff15e4a8a1ad8f269a8d488e1c32a89466db4a84f2d67e56472c5f41"
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
	builder.AddData(types.QITID.Bytes())
	builder.AddOp(txscript.OP_TOKEN_UNMINT)
	unmintScript, _ := builder.Script()
	tx.AddTxIn(&types.TxInput{
		PreviousOut: *types.NewOutPoint(&hash.Hash{}, types.MaxPrevOutIndex),
		Sequence:    types.MaxTxInSequenceNum,
		SignScript:  unmintScript,
		AmountIn:    types.Amount{Value: 10 * 1e8, Id: types.MEERID},
	})

	txid := hash.MustHexToDecodedHash("377cfb2c535be289f8e40299e8d4c234283c367e20bc5ff67ca18c1ca1337443")
	tx.AddTxIn(&types.TxInput{
		PreviousOut: *types.NewOutPoint(&txid,
			0),
		Sequence:   types.MaxTxInSequenceNum,
		SignScript: []byte{txscript.OP_DATA_1},
		AmountIn:   types.Amount{Value: 100 * 1e8, Id: types.QITID},
	})
	// output[0]
	builder = txscript.NewScriptBuilder()
	builder.AddOp(txscript.OP_TOKEN_DESTORY)
	tokenDestoryScript, _ := builder.Script()
	tx.AddTxOut(&types.TxOutput{Amount: types.Amount{Value: 99 * 1e8, Id: types.QITID}, PkScript: tokenDestoryScript})
	// output[1]
	addr, err := address.DecodeAddress("XmiGSPpX7v8hC4Mb59pufnhwYcUe1GvZVEx")
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
	tx.AddTxOut(&types.TxOutput{Amount: types.Amount{Value: 1 * 1e8, Id: types.QITID}, PkScript: tokenChangeScript})
	return tx
}
