package types

import (
	"encoding/hex"
	"github.com/Qitmeer/qitmeer/common/hash"
	"testing"
	"time"
)

func createTx(ctime *time.Time) (*Transaction, error) {
	tx := NewTransaction()
	tx.AddTxIn(&TxInput{
		PreviousOut: *NewOutPoint(&hash.Hash{}, MaxPrevOutIndex),
		Sequence:    MaxTxInSequenceNum,
		SignScript:  []byte{},
	})

	if ctime != nil {
		tx.Timestamp = *ctime
	} else {
		tx.Timestamp, _ = time.Parse("2016-01-02 15:04:05", "2019-13-14 00:00:00")
	}
	ds, err := hex.DecodeString("76a914868b9b6bc7e4a9c804ad3d3d7a2a6be27476941e88ac")
	if err != nil {
		return nil, err
	}
	var amt *Amount
	amt,err = NewMeer(1200000000)
	if err != nil {
		return nil, err
	}
	tx.AddTxOut(&TxOutput{
		Amount: *amt,
		PkScript: ds,
	})

	return tx, nil
}

func Test_TxHash(t *testing.T) {
	tx, err := createTx(nil)
	if err != nil {
		t.FailNow()
	}
	wantTxidStr := "a790ccc7d2039dcd06696c5f1a4d0b0d5ef3c70df03452811d190e5c5963aa7e"

	if tx.TxHash().String() != wantTxidStr {
		t.FailNow()
	}
}

func Test_TxHashFull(t *testing.T) {
	tx, err := createTx(nil)
	if err != nil {
		t.FailNow()
	}
	wantTxHashStr := "2d31999fb44ec7cef84456396d2de95a0dac0ef49087daae2ba2d16e1d63c081"

	if tx.TxHashFull().String() != wantTxHashStr {
		t.FailNow()
	}
}

func Test_TxExtensibility(t *testing.T) {
	tx, err := createTx(nil)
	if err != nil {
		t.FailNow()
	}
	newTX := NewTxDeep(tx)

	// Although the tx content is the same, the tx creation time is different.
	newTX.Tx.Timestamp = newTX.Tx.Timestamp.Add(time.Second)

	txHash := tx.TxHash()
	newTxHash := newTX.Tx.TxHash()
	if !newTxHash.IsEqual(&txHash) {
		t.Fatal()
	}
	txHashFull := tx.TxHashFull()
	newTxHashFull := newTX.Tx.TxHashFull()
	if newTxHashFull.IsEqual(&txHashFull) {
		t.Fatal()
	}
}
