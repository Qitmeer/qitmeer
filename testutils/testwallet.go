// Copyright (c) 2020 The qitmeer developers
// Use of this source code is governed by an ISC
// license that can be found in the LICENSE file.

package testutils

import (
	"encoding/binary"
	"github.com/Qitmeer/qitmeer/common/hash"
	"github.com/Qitmeer/qitmeer/core/address"
	"github.com/Qitmeer/qitmeer/core/types"
	"github.com/Qitmeer/qitmeer/crypto/bip32"
	"github.com/Qitmeer/qitmeer/crypto/ecc/secp256k1"
	"github.com/Qitmeer/qitmeer/params"
	"testing"
)

var (
	// the default seed used in the testWallet
	defaultSeed = [hash.HashSize]byte{
		0x7e, 0x44, 0x5a, 0xa5, 0xff, 0xd8, 0x34, 0xcb,
		0x2d, 0x3b, 0x2d, 0xb5, 0x0f, 0x89, 0x97, 0xdd,
		0x21, 0xaf, 0x29, 0xbe, 0xc3, 0xd2, 0x96, 0xaa,
		0xa0, 0x66, 0xd9, 0x02, 0xb9, 0x3f, 0x48, 0x4b,
	}
)

// testWallet is a simple in-memory wallet works for a test harness instance's
// node. the purpose of testWallet is to provide basic wallet functionality for
// the integrated-test, such as send tx & verify balance etc.
// testWallet works as a HD (BIP-32) wallet
type testWallet struct {
	// the bip32 master extended private key from a seed
	hdMaster *bip32.Key
	// the next hd child number from the master
	hdChildNumer uint32
	// addrs are all addresses which belong to the master private key.
	// the key of address map is their hd child number.
	addrs map[uint32]types.Address

	netParams *params.Params
	t         *testing.T
}

func NewTestWallet(t *testing.T, params *params.Params, nodeId uint32) (*testWallet, error) {
	return NewTestWalletWithSeed(t, params, &defaultSeed, nodeId)
}

func NewTestWalletWithSeed(t *testing.T, params *params.Params, seed *[hash.HashSize]byte, nodeId uint32) (*testWallet, error) {
	// The final seed is seed || nodeId, the purpose to make sure that each harness
	// node use a deterministic private key based on the its node id.
	var finalSeed [hash.HashSize + 4]byte
	// t.Logf("seed is %v",hexutil.Encode(seed[:]))
	copy(finalSeed[:], seed[:])
	// t.Logf("finalseed is %v",hexutil.Encode(finalSeed[:]))
	binary.LittleEndian.PutUint32(finalSeed[hash.HashSize:], nodeId)
	version := bip32.Bip32Version{
		PrivKeyVersion: params.HDPrivateKeyID[:],
		PubKeyVersion:  params.HDPublicKeyID[:],
	}
	// t.Logf("finalseed is %v",hexutil.Encode(finalSeed[:]))
	hdMaster, err := bip32.NewMasterKey2(finalSeed[:], version)
	if err != nil {
		return nil, err
	}
	child0, err := hdMaster.NewChildKey(0)
	if err != nil {
		return nil, err
	}
	key0 := child0.Key
	addr0, err := privKeyToAddr(key0, params)
	if err != nil {
		return nil, err
	}
	addrs := make(map[uint32]types.Address)
	addrs[0] = addr0
	return &testWallet{
		hdMaster:     hdMaster,
		hdChildNumer: 1,
		addrs:        addrs,
		netParams:    params,
	}, nil
}

// newAddress create a new address from the wallet's key chain.
func (w *testWallet) newAddress() (types.Address, error) {
	num := w.hdChildNumer
	childx, err := w.hdMaster.NewChildKey(num)
	if err != nil {
		return nil, err
	}
	addrx, err := privKeyToAddr(childx.Key, w.netParams)
	if err != nil {
		return nil, err
	}
	w.addrs[num] = addrx
	w.hdChildNumer++
	return addrx, nil
}

func (w *testWallet) coinBaseAddr() types.Address {
	return w.addrs[0]
}

// convert the serialized private key into the p2pkh address
func privKeyToAddr(privKey []byte, params *params.Params) (types.Address, error) {
	_, pubKey := secp256k1.PrivKeyFromBytes(privKey)
	serializedKey := pubKey.SerializeCompressed()
	addr, err := address.NewSecpPubKeyAddress(serializedKey, params)
	if err != nil {
		return nil, err
	}
	return addr.PKHAddress(), nil
}
