/*
 * Copyright (c) 2017-2020 The qitmeer developers
 */

package qnode

import (
	"bytes"
	"crypto/ecdsa"
	"encoding/hex"
	"math/big"
	"testing"

	"github.com/Qitmeer/qitmeer/common/encode/rlp"
	"github.com/Qitmeer/qitmeer/crypto"
	"github.com/Qitmeer/qitmeer/p2p/qnr"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var (
	privkey, _ = crypto.HexToECDSA("b71c71a67e1177ad4e901695e1b4b9ee17ae16c6668d313eac2f96dbcda3f291")
	pubkey     = &privkey.PublicKey
)

func TestEmptyNodeID(t *testing.T) {
	var r qnr.Record
	if addr := ValidSchemes.NodeAddr(&r); addr != nil {
		t.Errorf("wrong address on empty record: got %v, want %v", addr, nil)
	}

	require.NoError(t, SignV4(&r, privkey))
	expected := "a448f24c6d18e575453db13171562b71999873db5b286df957af199ec94617f7"
	assert.Equal(t, expected, hex.EncodeToString(ValidSchemes.NodeAddr(&r)))
}

// Checks that failure to sign leaves the record unmodified.
func TestSignError(t *testing.T) {
	invalidKey := &ecdsa.PrivateKey{D: new(big.Int), PublicKey: *pubkey}

	var r qnr.Record
	emptyEnc, _ := rlp.EncodeToBytes(&r)
	if err := SignV4(&r, invalidKey); err == nil {
		t.Fatal("expected error from SignV4")
	}
	newEnc, _ := rlp.EncodeToBytes(&r)
	if !bytes.Equal(newEnc, emptyEnc) {
		t.Fatal("record modified even though signing failed")
	}
}

// TestGetSetSecp256k1 tests encoding/decoding and setting/getting of the Secp256k1 key.
func TestGetSetSecp256k1(t *testing.T) {
	var r qnr.Record
	if err := SignV4(&r, privkey); err != nil {
		t.Fatal(err)
	}

	var pk Secp256k1
	require.NoError(t, r.Load(&pk))
	assert.EqualValues(t, pubkey, &pk)
}
