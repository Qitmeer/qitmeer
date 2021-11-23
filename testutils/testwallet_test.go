// Copyright (c) 2020 The qitmeer developers
// Use of this source code is governed by an ISC
// license that can be found in the LICENSE file.
package testutils

import (
	"github.com/Qitmeer/qng-core/common/util/hexutil"
	"github.com/Qitmeer/qng-core/params"
	"testing"
)

var (
	expect = struct {
		ver       string
		key       string
		chaincode string
		priv0     string
		priv1     string
		addr0     string
		addr1     string
	}{
		"0x040bee6e",
		"0x38015593945529cc0bd761108ad2fbd98a3f5f8e030c5acd3747ce3e54d95c16",
		"0x4eb4e56ada09795313734db329c362923c5b6fac75b924780e68b9c9b18a24b3",
		"0xe0b26a52b1a9676a365d6452fb04a1c05b58e959683862d73105e58d4416baba",
		"0xfff2cefe258ca60ae5f5abec99b5d63e2a561c40d784ee50b04eddf8efc84b0d",
		"RmBKxMWg4C4EMzYowisDEGSBwmnR6tPgjLs",
		"RmHFARk5xmoMNUVJ6UCHFiWQML1vxwUhw1b",
	}
)

func Test_newTestWallet(t *testing.T) {
	wallet, err := newTestWallet(t, params.PrivNetParam.Params, 0)
	if err != nil {
		t.Errorf("create the test wallet failed: %v", err)
	}

	if hexutil.Encode(wallet.hdMaster.Key) != expect.key {
		t.Errorf("hd master key not matched, expect %v but got %v", wallet.hdMaster.Key, expect.key)
	}
	if hexutil.Encode(wallet.hdMaster.Version) != expect.ver {
		t.Errorf("hd master version not matched, expect %v but got %v", wallet.hdMaster.Version, expect.ver)
	}
	if hexutil.Encode(wallet.hdMaster.ChainCode) != expect.chaincode {
		t.Errorf("hd master chain code not matched, expect %v but got %v", wallet.hdMaster.ChainCode, expect.chaincode)
	}
	if wallet.addrs[0].Encode() != expect.addr0 {
		t.Errorf("hd key0 addr not matched, expect %v but got %v", wallet.addrs[0].Encode(), expect.addr0)
	}
	addr1, err := wallet.NewAddress()
	if err != nil {
		t.Errorf("failed get new address : %v", err)
	}
	if addr1.Encode() != expect.addr1 {
		t.Errorf("hd key1 addr not matched, expect %v but got %v", wallet.addrs[1].Encode(), expect.addr1)
	}
	if hexutil.Encode(wallet.privkeys[0]) != expect.priv0 {
		t.Errorf("hd key0 priv key not matched, expect %x but got %v", wallet.privkeys[0], expect.priv0)
	}
	if hexutil.Encode(wallet.privkeys[1]) != expect.priv1 {
		t.Errorf("hd key0 priv key not matched, expect %x but got %v", wallet.privkeys[1], expect.priv1)
	}
	if wallet.coinBaseAddr().Encode() != expect.addr0 {
		t.Errorf("hd coinbase addr not matched, expect %v but got %v", wallet.coinBaseAddr(), expect.addr0)
	}
	if hexutil.Encode(wallet.coinBasePrivKey()) != expect.priv0 {
		t.Errorf("hd coinbase priv key not matched, expect %x but got %v", wallet.coinBasePrivKey(), expect.priv0)
	}
}

func TestHarnessWallet(t *testing.T) {
	h, err := NewHarness(t, params.PrivNetParam.Params)
	h2, err := NewHarness(t, params.PrivNetParam.Params)
	defer func() {
		h.Teardown()
		h2.Teardown()
	}()
	if err != nil {
		t.Errorf("new harness failed: %v", err)
		h.Teardown()
		h2.Teardown()
	}
	if h.Wallet.coinBaseAddr().Encode() != expect.addr0 {
		t.Errorf("hd coinbase addr not matched, expect %v but got %v", h.Wallet.coinBaseAddr(), expect.addr0)
	}
	if hexutil.Encode(h.Wallet.coinBasePrivKey()) != expect.priv0 {
		t.Errorf("hd coinbase priv key not matched, expect %x but got %v", h.Wallet.coinBasePrivKey(), expect.priv0)
	}
	h2expect := struct {
		coinbaseAddr   string
		coinbasePivkey string
	}{
		"RmQsTTCZCWEjgnzRm8XWRQRitPoWvtzs1rJ",
		"0xe7487ce542af7360e4e1560baa60c405c041065c3ccbb715fbd81da154910829",
	}

	if h2.Wallet.coinBaseAddr().Encode() != h2expect.coinbaseAddr {
		t.Errorf("h2 coinbase addr not matched, expect %v but got %v", h2.Wallet.coinBaseAddr(), h2expect.coinbaseAddr)
	}
	if hexutil.Encode(h2.Wallet.coinBasePrivKey()) != h2expect.coinbasePivkey {
		t.Errorf("h2 coinbase priv key not matched, expect %x but got %v", h2.Wallet.coinBasePrivKey(), h2expect.coinbasePivkey)
	}
}
