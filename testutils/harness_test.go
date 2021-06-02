// Copyright (c) 2020 The qitmeer developers
// Use of this source code is governed by an ISC
// license that can be found in the LICENSE file.

package testutils_test

import (
	"github.com/Qitmeer/qitmeer/core/types"
	"github.com/Qitmeer/qitmeer/params"
	. "github.com/Qitmeer/qitmeer/testutils"
	"sync"
	"testing"
	"time"
)

func TestHarness(t *testing.T) {
	h, err := NewHarness(t, params.PrivNetParam.Params)
	if err != nil {
		t.Errorf("create new test harness instance failed %v", err)
	}
	if err := h.Setup(); err != nil {
		t.Errorf("setup test harness instance failed %v", err)
	}

	h2, err := NewHarness(t, params.PrivNetParam.Params)
	defer func() {

		if err := h.Teardown(); err != nil {
			t.Errorf("tear down test harness instance failed %v", err)
		}
		numOfHarnessInstances := len(AllHarnesses())
		if numOfHarnessInstances != 10 {
			t.Errorf("harness num is wrong, expect %d , but got %d", 10, numOfHarnessInstances)
			for _, h := range AllHarnesses() {
				t.Errorf("%v\n", h.Id())
			}
		}

		if err := TearDownAll(); err != nil {
			t.Errorf("tear down all error %v", err)
		}
		numOfHarnessInstances = len(AllHarnesses())
		if numOfHarnessInstances != 0 {
			t.Errorf("harness num is wrong, expect %d , but got %d", 0, numOfHarnessInstances)
			for _, h := range AllHarnesses() {
				t.Errorf("%v\n", h.Id())
			}
		}

	}()
	numOfHarnessInstances := len(AllHarnesses())
	if numOfHarnessInstances != 2 {
		t.Errorf("harness num is wrong, expect %d , but got %d", 2, numOfHarnessInstances)
	}
	if err := h2.Teardown(); err != nil {
		t.Errorf("teardown h2 error:%v", err)
	}

	numOfHarnessInstances = len(AllHarnesses())
	if numOfHarnessInstances != 1 {
		t.Errorf("harness num is wrong, expect %d , but got %d", 1, numOfHarnessInstances)
	}
	var wg sync.WaitGroup
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func() {
			NewHarness(t, params.PrivNetParam.Params)
			wg.Done()
		}()
	}
	wg.Wait()
}

func TestHarnessNodePorts(t *testing.T) {
	var setup, teardown sync.WaitGroup
	ha := make(map[int]*Harness, 10)
	for i := 0; i < 10; i++ {
		setup.Add(1)
		teardown.Add(1)
		// new and setup
		go func(index int) {
			defer setup.Done()
			h, err := NewHarness(t, params.PrivNetParam.Params)
			if err != nil {
				t.Errorf("new harness failed: %v", err)
			}
			ha[index] = h
			if err := ha[index].Setup(); err != nil {
				t.Errorf("setup harness failed: %v", err)
			}
			time.Sleep(500 * time.Millisecond)
		}(i)
		go func(index int) {
			defer teardown.Done()
			setup.Wait() //wait for all setup finished
			if err := ha[index].Teardown(); err != nil {
				t.Errorf("tear down harness failed: %v", err)
			}
		}(i)
	}
	teardown.Wait()
}

func TestHarness_RpcAPI(t *testing.T) {
	args := []string{"--modules=miner", "--modules=qitmeer"}
	h, err := NewHarness(t, params.PrivNetParam.Params, args...)
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
	AssertBlockOrderAndHeight(t, h, 1, 1, 0)
	GenerateBlock(t, h, 2)
	AssertBlockOrderAndHeight(t, h, 3, 3, 2)
	GenerateBlock(t, h, 16)
	AssertBlockOrderAndHeight(t, h, 19, 19, 18)
	spendAmt := types.Amount{Value: 50 * types.AtomsPerCoin, Id: types.MEERID}
	lockTime := int64(18)
	txid, _ := Spend(t, h, spendAmt, nil, &lockTime)
	t.Logf("[%v]: tx %v which spend %v has been sent", h.Node.Id(), txid, spendAmt.String())
	blocks := GenerateBlock(t, h, 1)
	AssertTxMinedUseNotifierAPI(t, h, txid, blocks[0])
	lockTime = int64(19)
	spendAmt = types.Amount{Value: 5000 * types.AtomsPerCoin, Id: types.MEERID}

	txid, _ = Spend(t, h, spendAmt, nil, &lockTime)
	t.Logf("[%v]: tx %v which spend %v has been sent", h.Node.Id(), txid, spendAmt.String())
	blocks = GenerateBlock(t, h, 1)
	AssertTxMinedUseNotifierAPI(t, h, txid, blocks[0])
}

func TestHarness_SpentGenesis(t *testing.T) {
	args := []string{"--modules=miner", "--modules=qitmeer"}
	h, err := NewHarness(t, params.PrivNetParam.Params, args...)
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
	// max can spent 13000 meer in genesis
	spendAmt := types.Amount{Value: 14000 * types.AtomsPerCoin, Id: types.MEERID}
	_, _ = CanNotSpend(t, h, spendAmt, nil, nil)

	GenerateBlock(t, h, 20)
	AssertBlockOrderAndHeight(t, h, 21, 21, 20)

	lockTime := int64(20)
	_, _ = Spend(t, h, spendAmt, nil, &lockTime)
	GenerateBlock(t, h, 5)
	AssertBlockOrderAndHeight(t, h, 26, 26, 25)

	spendAmt = types.Amount{Value: 50 * types.AtomsPerCoin, Id: types.MEERID}
	txid, _ := Spend(t, h, spendAmt, nil, nil)
	t.Logf("[%v]: tx %v which spend %v has been sent", h.Node.Id(), txid, spendAmt.String())
	blocks := GenerateBlock(t, h, 1)
	AssertTxMinedUseNotifierAPI(t, h, txid, blocks[0])

}
