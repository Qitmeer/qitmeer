package rpc

import (
	"fmt"
	"github.com/Qitmeer/qitmeer/core/blockchain"
	"github.com/Qitmeer/qitmeer/rpc/client/cmds"
	"sync"
)

type TxConfirm struct {
	Order    uint64
	Confirms uint64
	TxHash   string
	wsc      *wsClient
}

type WatchTxConfirmServer struct {
	BC         *blockchain.BlockChain
	TxConfirms map[uint64]map[string]TxConfirm
	lock       sync.Mutex
	m          *wsNotificationManager
	wg         *sync.WaitGroup
	quit       chan struct{}
}

func newWatchTxConfirmServer(server *RpcServer, m *wsNotificationManager) *WatchTxConfirmServer {
	return &WatchTxConfirmServer{
		BC:         server.BC,
		m:          m,
		TxConfirms: map[uint64]map[string]TxConfirm{},
		quit:       make(chan struct{}),
		wg:         &sync.WaitGroup{},
	}
}

func (w *WatchTxConfirmServer) AddTxConfirms(confirm TxConfirm) {
	w.lock.Lock()
	defer w.lock.Unlock()
	if _, ok := w.TxConfirms[confirm.Order]; !ok {
		w.TxConfirms[confirm.Order] = map[string]TxConfirm{}
	}
	w.TxConfirms[confirm.Order][confirm.TxHash] = confirm
}

func (w *WatchTxConfirmServer) Start() {
	w.wg.Add(1)
	go w.HandleTxConfirm()
}

func (w *WatchTxConfirmServer) HandleTxConfirm() {
	defer w.wg.Done()
	for {
		select {
		case <-w.quit:
			log.Info("exit watch tx confirm")
			return
		case <-w.m.NewBlockMsg:
			if len(w.TxConfirms) > 0 {
				w.lock.Lock()
				for order, txc := range w.TxConfirms {
					h := w.BC.BlockDAG().GetBlockByOrder(uint(order))
					if h == nil {
						log.Error("order not exist", "order", order)
						delete(w.TxConfirms, order)
						continue
					}
					ib := w.BC.BlockDAG().GetBlock(h)
					if ib == nil {
						if h == nil {
							log.Error("block hash not exist", "hash", h)
							delete(w.TxConfirms, order)
							continue
						}
					}
					node := w.BC.BlockIndex().LookupNode(h)
					if node == nil {
						log.Error("no node")
						continue
					}
					confirmations := int64(w.BC.BlockDAG().GetConfirmations(node.GetID()))
					isBlue := w.BC.BlockDAG().IsBlue(ib.GetID())
					InValid := w.BC.BlockIndex().NodeStatus(node).KnownInvalid()
					for _, tx := range txc {
						if !isBlue || InValid || confirmations >= int64(tx.Confirms) {
							ntfn := &cmds.NotificationTxConfirmNtfn{
								cmds.TxConfirmResult{
									Tx:       tx.TxHash,
									Confirms: uint64(confirmations),
									IsBlue:   isBlue,
									IsValid:  !InValid,
									Order:    order,
								},
							}
							marshalledJSON, err := cmds.MarshalCmd(nil, ntfn)
							if err != nil {
								log.Error(fmt.Sprintf("Failed to marshal tx confirm notification: "+
									"%v", err))
								continue
							}
							err = tx.wsc.QueueNotification(marshalledJSON)
							if err != nil {
								log.Error("notify failed", "err", err)
								continue
							}
							delete(w.TxConfirms[order], tx.TxHash)
						}
					}
				}
				w.lock.Unlock()
			}
		}
	}
}

func (w *WatchTxConfirmServer) Stop() {
	close(w.quit)
	w.wg.Wait()
}
