package main

import (
	"fmt"
	"github.com/Qitmeer/qitmeer/common/hash"
	"github.com/Qitmeer/qitmeer/core/types"
	"github.com/Qitmeer/qitmeer/params"
	"github.com/Qitmeer/qitmeer/testutils"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

func TestLockedLedger(t *testing.T) {
	args := []string{"--modules=miner", "--modules=qitmeer"}
	h, err := testutils.NewHarness(t, params.PrivNetParam.Params, args...)
	defer h.Teardown()
	if err != nil {
		t.Errorf("new harness failed: %v", err)
		h.Teardown()
	}
	err = h.Setup()
	if err != nil {
		t.Fatalf("setup harness failed:%v", err)
	}
	time.Sleep(500 * time.Millisecond)

	if info, err := h.Client.NodeInfo(); err != nil {
		t.Errorf("test failed : %v", err)
	} else {
		expect := "privnet"
		if info.Network != expect {
			t.Errorf("test failed, expect %v , but got %v", expect, info.Network)
		}
	}
	addr, err := h.Wallet.NewAddress()
	if err != nil {
		t.Fatalf("failed to generate new address for test wallet: %v", err)
	}

	testutils.AssertBlockOrderAndHeight(t, h, 1, 1, 0)
	testutils.GenerateBlock(t, h, 2)
	testutils.AssertBlockOrderAndHeight(t, h, 3, 3, 2)

	spendAmt := types.Amount{Value: 50 * types.AtomsPerCoin, Id: types.MEERID}
	txid, addr := testutils.Spend(t, h, spendAmt)
	t.Logf("[%v]: tx %v which spend %v has been sent", h.Node.Id(), txid, spendAmt.String())
	t.Log(h.Wallet.Addresses(), addr)
}

func TestGenesisShuffle(t *testing.T) {
	seed := "RmBKxMWg4C4EMzYowisDEGSBwmnR6tPgjLs:234,RmBKxMWg4C4EMzYowisDEGSBwmnR6tPgjLs:67,RmHFARk5xmoMNUVJ6UCHFiWQML1vxwUhw1b:234,"
	seedHash := hash.HashB([]byte(seed))
	arr := []int{6, 43, 23, 1, 5, 7, 1, 5, 7, 1, 5, 7, 1, 5, 7, 1, 5, 7, 1, 5, 7, 1, 5, 7, 1, 5, 7, 1, 5, 7, 1, 5, 7, 1, 5, 7, 1, 5, 7, 1, 5, 7, 1, 5, 7}
	arr1 := []int{6, 43, 23, 1, 5, 7, 1, 5, 7, 1, 5, 7, 1, 5, 7, 1, 5, 7, 1, 5, 7, 1, 5, 7, 1, 5, 7, 1, 5, 7, 1, 5, 7, 1, 5, 7, 1, 5, 7, 1, 5, 7, 1, 5, 7}
	arr2 := []int{6, 43, 23, 1, 5, 7, 1, 5, 7, 1, 5, 7, 1, 5, 7, 1, 5, 7, 1, 5, 7, 1, 5, 7, 1, 5, 7, 1, 5, 7, 1, 5, 7, 1, 5, 7, 1, 5, 7, 1, 5, 7, 1, 5, 7}
	arr3 := []int{6, 43, 23, 1, 5, 7, 1, 5, 7, 1, 5, 7, 1, 5, 7, 1, 5, 7, 1, 5, 7, 1, 5, 7, 1, 5, 7, 1, 5, 7, 1, 5, 7, 1, 5, 7, 1, 5, 7, 1, 5, 7, 1, 5, 7}
	fmt.Println(GenesisShuffle(arr, seedHash))
	fmt.Println(GenesisShuffle(arr1, seedHash))
	assert.Equal(t, arr, arr1)
	seed = "RmHFARk5xmoMNUVJ6UCHFiWQML1vxwUhw1b:234,RmBKxMWg4C4EMzYowisDEGSBwmnR6tPgjLs:353,RmHFARk5xmoMNUVJ6UCHFiWQML1vxwUhw1b:235,"
	seedHash = hash.HashB([]byte(seed))
	fmt.Println(GenesisShuffle(arr2, seedHash))
	fmt.Println(GenesisShuffle(arr3, seedHash))
	assert.Equal(t, arr2, arr3)
}
