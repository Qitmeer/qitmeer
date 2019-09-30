package types

import (
	"bytes"
	"encoding/hex"
	"testing"
)

func createTx() (*Transaction,error) {
	txStr:="010000000184a780b8f5ca533e54a2a4f1e1516abbf149105674b6f701b4810209a9c45f80ffffffffffffffff0280461c86000000001976a914c1777151516afe2b9f59bbd1479231aa2f250d2888ac80b2e60e000000001976a914868b9b6bc7e4a9c804ad3d3d7a2a6be27476941e88ac000000000000000001145108cd9c4f7dcaaef2c8092f7169746d6565722f"
	if len(txStr)%2 != 0 {
		txStr = "0" + txStr
	}
	serializedTx, err := hex.DecodeString(txStr)
	if err != nil {
		return nil,err
	}
	var tx Transaction
	err = tx.Deserialize(bytes.NewReader(serializedTx))
	if err != nil {
		return nil,err
	}
	return &tx,nil
}

func Test_TxHash(t *testing.T) {
	tx,err:=createTx()
	if err != nil {
		t.FailNow()
	}
	wantTxidStr:="8baf9c3b985d5faa1abcfb29faec644f20ff82d5aa6c65c1a01976f093a8fa84"

	if tx.TxHash().String() != wantTxidStr {
		t.FailNow()
	}
}

func Test_TxHashFull(t *testing.T) {
	tx,err:=createTx()
	if err != nil {
		t.FailNow()
	}
	wantTxHashStr:="d3d71396152ea75d0e568528e702c4de8336e8d5f06fdad82194dbb342a981b9"

	if tx.TxHashFull().String() != wantTxHashStr {
		t.FailNow()
	}
}