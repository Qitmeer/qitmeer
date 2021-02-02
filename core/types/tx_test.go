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
	wantTxidStr := "3a73ec7b175f06eb2aa1a1184b76b626416bcef99af8699605b0fc082cd9d032"
	got := tx.TxHash().String()
	if  got != wantTxidStr {
		t.Errorf("want %s, got %s", wantTxidStr, got)
	}
}

func Test_TxHashFull(t *testing.T) {
	tx, err := createTx(nil)
	if err != nil {
		t.FailNow()
	}
	wantTxHashStr := "fe5ffd5eb8f349fe2281b8f9bcba0c8db8778c27b83ae2d6c8624513821b4fc4"
	got := tx.TxHashFull().String()
	if  got != wantTxHashStr {
		t.Errorf("want %s, got %s", wantTxHashStr, got)
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
