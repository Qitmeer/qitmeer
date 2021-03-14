package testutils

import (
	"github.com/Qitmeer/qitmeer/core/types"
	"github.com/Qitmeer/qitmeer/params"
	"testing"
	"time"
)

func TestWsNotify(t *testing.T) {
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
	GenerateBlock(t, h, 1)
	AssertBlockOrderAndHeight(t, h, 3, 3, 2)
	GenerateBlock(t, h, 16)
	AssertBlockOrderAndHeight(t, h, 19, 19, 18)
	err = h.Notifier.NotifyNewTransactions(true)
	if err != nil {
		t.Fatal(err)
		return
	}
	err = h.Notifier.NotifyTxsByAddr(false, h.Wallet.Addresses(), nil)
	if err != nil {
		t.Fatal(err)
		return
	}
	spendAmt := types.Amount{Value: 50 * types.AtomsPerCoin, Id: types.MEERID}
	txid := Spend(t, h, spendAmt)
	t.Logf("[%v]: tx %v which spend %v has been sent", h.Node.Id(), txid, spendAmt.String())
	blocks := GenerateBlock(t, h, 4)
	AssertTxMinedUseNotifierAPI(t, h, txid, blocks[0])
	t.Logf("[%v]: tx %v which spend %v has been sent", h.Node.Id(), txid, spendAmt.String())
	blocks = GenerateBlock(t, h, 3)
	err = h.Notifier.Rescan(0, 17, h.Wallet.Addresses(), nil)
	if err != nil {
		t.Fatal(err)
		return
	}
}
