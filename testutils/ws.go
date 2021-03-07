package testutils

import (
	"fmt"
	"github.com/Qitmeer/qitmeer/common/hash"
	j "github.com/Qitmeer/qitmeer/core/json"
	"github.com/Qitmeer/qitmeer/core/types"
	"github.com/Qitmeer/qitmeer/rpc/client"
	"github.com/Qitmeer/qitmeer/rpc/client/cmds"
	"time"
)

// Register Addrs Filter
func (h *Harness) NotifyTxsByAddr(reload bool, addr []string, outpoint []cmds.OutPoint) error {
	return h.Notifier.NotifyTxsByAddr(reload, addr, outpoint)
}

// Register NotifyNewTransactions
func (h *Harness) NotifyNewTransactions(verbose bool) error {
	return h.Notifier.NotifyNewTransactions(verbose)
}

// Register Rescan by address
func (h *Harness) Rescan(beginBlock, endBlock uint64, addrs []string, op []cmds.OutPoint) error {
	return h.Notifier.Rescan(beginBlock, endBlock, addrs, op)
}

// Register NotifyTxsConfirmed
func (h *Harness) NotifyTxsConfirmed(txs []cmds.TxConfirm) error {
	return h.Notifier.NotifyTxsConfirmed(txs)
}

func (w *testWallet) blockConnected(hash *hash.Hash, order int64, t time.Time, txs []*types.Transaction) {
	w.t.Logf("node [%v] OnBlockConnected hash=%v,order=%v", w.nodeId, hash, order)
	for _, tx := range txs {
		w.t.Logf("node [%v] OnBlockConnected tx=%v", w.nodeId, tx.TxHash())
	}
	// Append the new update to the end of the queue of block dag updates.
	w.updateMtx.Lock()
	w.updates = append(w.updates, &update{order, hash, txs})
	w.updateMtx.Unlock()

	// signal the update watcher that a new update is arrived . use a goroutine
	// in order to avoid blocking this callback itself from the websocket client.
	go func() {
		w.updateArrived <- struct{}{}
	}()
}

func (w *testWallet) blockDisconnected(hash *hash.Hash, order int64, t time.Time, txs []*types.Transaction) {
	w.t.Logf("node [%v] OnBlockDisconnected hash=%v,order=%v", w.nodeId, hash, order)
	w.Lock()
	defer w.Unlock()

	undo, ok := w.undoes[hash]
	if !ok {
		w.t.Fatalf("the disconnected a unknown block, hash=%v, order=%v", hash, order)
	}

	for _, utxo := range undo.utxosCreated {
		delete(w.utxos, utxo)
	}

	for outPoint, utxo := range undo.utxosDestroyed {
		w.utxos[outPoint] = utxo
	}

	delete(w.undoes, hash)
}

func (w *testWallet) OnTxConfirm(txConfirm *cmds.TxConfirmResult) {
	fmt.Println("OnTxConfirm", txConfirm.Tx, txConfirm.Confirms, txConfirm.Order)
}
func (w *testWallet) OnTxAcceptedVerbose(c *client.Client, tx *j.DecodeRawTransactionResult) {
	fmt.Println("OnTxAcceptedVerbose", tx.Hash, tx.Confirms, tx.Order, tx.Txvalid, tx.IsBlue, tx.Duplicate)
}
func (w *testWallet) OnRescanProgress(rescanPro *cmds.RescanProgressNtfn) {
	fmt.Println("OnRescanProgress", rescanPro.Order, rescanPro.Hash)
}
func (w *testWallet) OnRescanFinish(rescanFinish *cmds.RescanFinishedNtfn) {
	fmt.Println("OnRescanFinish", rescanFinish.Order, rescanFinish.Hash)
}
