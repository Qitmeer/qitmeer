package testutils

import (
	"github.com/Qitmeer/qitmeer/core/types"
	"github.com/Qitmeer/qitmeer/params"
	"github.com/Qitmeer/qitmeer/rpc/client/cmds"
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
	AssertBlockOrderAndHeight(t, h, 2, 2, 1)

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
	lockT := int64(1)
	txid, addr := Spend(t, h, spendAmt, nil, &lockT)
	t.Logf("[%v]: tx %v which spend %v has been sent", h.Node.Id(), txid, spendAmt.String())
	t.Log(h.Wallet.Addresses())
	AssertMempoolTxNotify(t, h, txid.String(), addr.String(), 5)
	blocks := GenerateBlock(t, h, 4)
	AssertTxMinedUseNotifierAPI(t, h, txid, blocks[0])
	AssertBlockOrderAndHeight(t, h, 6, 6, 5)
	h.Wallet.OnRescanComplete = func() {
		AssertScan(t, h, 5, 6)
	}
	err = h.Notifier.Rescan(0, 6, h.Wallet.Addresses(), nil)
	if err != nil {
		t.Fatal(err)
		return
	}
	err = h.Notifier.NotifyTxsConfirmed([]cmds.TxConfirm{
		{
			Txid:          txid.String(),
			Confirmations: 5,
		},
	})
	if err != nil {
		t.Fatal(err)
		return
	}
	GenerateBlock(t, h, 2)
	AssertBlockOrderAndHeight(t, h, 8, 8, 7)
	AssertTxConfirm(t, h, txid.String(), 5)
	spendAmt = types.Amount{Value: 50 * types.AtomsPerCoin, Id: types.MEERID}
	lockT = int64(1)
	txid, addr = Spend(t, h, spendAmt, nil, &lockT)
	t.Logf("[%v]: tx %v which spend %v has been sent", h.Node.Id(), txid, spendAmt.String())
	GenerateBlock(t, h, 4)
	err = h.Notifier.NotifyTxsConfirmed([]cmds.TxConfirm{
		{
			Txid:          txid.String(),
			Confirmations: 5,
		},
	})
	if err != nil {
		t.Fatal(err)
		return
	}
	err = h.Notifier.RemoveTxsConfirmed([]cmds.TxConfirm{
		{
			Txid: txid.String(),
		},
	})
	if err != nil {
		t.Fatal(err)
		return
	}
	GenerateBlock(t, h, 2)
	AssertTxNotConfirm(t, h, txid.String())
}
