// Copyright (c) 2020 The qitmeer developers
// Use of this source code is governed by an ISC
// license that can be found in the LICENSE file.

package blockchain

import (
	"bytes"
	"encoding/hex"
	"github.com/Qitmeer/qitmeer/common/hash"
	"github.com/Qitmeer/qitmeer/core/dbnamespace"
	"github.com/Qitmeer/qitmeer/core/types"
	"github.com/Qitmeer/qitmeer/database"
	_ "github.com/Qitmeer/qitmeer/database/ffldb"
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

func TestTokeStateSerialization(t *testing.T) {
	tests := []struct {
		name  string
		state *tokenState
		bytes []byte
		ok    bool
	}{
		{name: "test1",
			state: &tokenState{
				balances: tokenBalances{
					types.QITID: tokenBalance{200 * 1e8, 100 * 1e8}},
				updates: []balanceUpdate{
					{typ: tokenMint,
						meerAmount:  100 * 1e8,
						tokenAmount: types.Amount{200 * 1e8, types.QITID}},
				},
			},
			bytes: bytesFromStr("0101c9bfde8f00a49faec7000101a49faec70001c9bfde8f00"),
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
			types.QITID: tokenBalance{200 * 1e8, 100 * 1e8}},
		updates: []balanceUpdate{
			{typ: tokenMint,
				meerAmount:  100 * 1e8,
				tokenAmount: types.Amount{200 * 1e8, types.QITID}},
		}}
	// create a fake block hash for testing
	b := make([]byte, 32)
	b[0] = 0xa
	hash := hash.HashH(b)

	err = tokendb.Update(func(dbTx database.Tx) error {
		return dbPutTokenState(dbTx, &hash, ts)
	})
	if err != nil {
		t.Fatalf("%v", err)
	}

	// fetch record from tokenstate db
	var tsfromdb *tokenState
	err = tokendb.View(func(dbTx database.Tx) error {
		tsfromdb, err = dbFetchTokenState(dbTx, hash)
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
