/*
 * Copyright (c) 2017-2020 The qitmeer developers
 */

package token

import (
	"bytes"
	"encoding/hex"
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

const QITID types.CoinID = 1

func bytesFromStr(s string) []byte {
	b, _ := hex.DecodeString(s)
	return b
}

func TestTokenStateSerialization(t *testing.T) {
	tests := []struct {
		name  string
		state *TokenState
		bytes []byte
		ok    bool
	}{
		{name: "test1",
			state: &TokenState{
				Balances: TokenBalancesMap{
					QITID: TokenBalance{Balance: 200 * 1e8, LockedMeer: 100 * 1e8}},
				Updates: []ITokenUpdate{&BalanceUpdate{
					TokenUpdate: &TokenUpdate{Typ: types.TxTypeTokenMint},
					MeerAmount:  100 * 1e8,
					TokenAmount: types.Amount{Value: 200 * 1e8, Id: QITID},
				}},
			},
			bytes: bytesFromStr("000101c9bfde8f00a49faec700018011a49faec70001c9bfde8f0000"),
			ok:    true,
		},
	}
	for i, test := range tests {
		serialized, err := test.state.Serialize()
		if !bytes.Equal(serialized, test.bytes) {
			t.Errorf("test[%d][%s] failed: want %x but got %x : %w", i, test.name, test.bytes, serialized, err)
		}
		deserialized, err := test.state.Deserialize(test.bytes)
		if !reflect.DeepEqual(deserialized, len(test.bytes)) {
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
	ts := &TokenState{
		Balances: TokenBalancesMap{
			QITID: TokenBalance{Balance: 200 * 1e8, LockedMeer: 100 * 1e8}},
		Updates: []ITokenUpdate{&BalanceUpdate{
			TokenUpdate: &TokenUpdate{Typ: types.TxTypeTokenMint},
			MeerAmount:  100 * 1e8,
			TokenAmount: types.Amount{Value: 200 * 1e8, Id: QITID},
		}},
	}

	// create a fake block id for testing
	bid := uint32(0xa)

	err = tokendb.Update(func(dbTx database.Tx) error {
		return DBPutTokenState(dbTx, bid, ts)
	})
	if err != nil {
		t.Fatalf("%v", err)
	}

	// fetch record from tokenstate db
	var tsfromdb *TokenState
	err = tokendb.View(func(dbTx database.Tx) error {
		tsfromdb, err = DBFetchTokenState(dbTx, bid)
		return err
	})
	if err != nil {
		t.Fatalf("%v", err)
	}

	// compare result
	if !reflect.DeepEqual(ts.Balances, tsfromdb.Balances) {
		t.Fatalf("token state put db is %v ,but from db is %v", ts, *tsfromdb)
	}
}

func Test_FeeConfigs(t *testing.T) {

	tests := []struct {
		txFees types.AmountMap
		expect bool
	}{
		{
			txFees: types.AmountMap{
				types.MEERID: types.AtomsPerCoin,
				QITID:        types.AtomsPerCoin,
			},
			expect: true,
		},
		{
			txFees: types.AmountMap{
				types.MEERID: types.AtomsPerCoin * 2,
				QITID:        types.AtomsPerCoin * 3,
			},
			expect: true,
		},
		{
			txFees: types.AmountMap{
				types.MEERID: types.AtomsPerCoin,
				QITID:        types.AtomsPerCoin,
			},
			expect: true,
		},
		{
			txFees: types.AmountMap{
				types.MEERID: types.AtomsPerCoin,
				QITID:        types.AtomsPerCoin,
			},
			expect: true,
		},
	}

	ts := TokenState{Types: TokenTypesMap{}}
	ts.Types[types.MEERID] = TokenType{FeeCfg: *NewTokenFeeConfig(0)}
	ts.Types[QITID] = TokenType{FeeCfg: *NewTokenFeeConfig(0)}

	for _, test := range tests {
		err := ts.CheckFees(test.txFees)
		if (err == nil && test.expect) ||
			(err != nil && !test.expect) {
			continue
		}
		t.Fatalf("txFees:%v Expect:%v", test.txFees, test.expect)
	}
}
