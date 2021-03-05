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
	bc         *blockchain.BlockChain
	TxConfirms map[uint64]map[string]TxConfirm
	lock       sync.Mutex
	m          *wsNotificationManager
	wg         *sync.WaitGroup
	quit       chan struct{}
}

func newWatchTxConfirmServer(server *RpcServer, m *wsNotificationManager) *WatchTxConfirmServer {
	return &WatchTxConfirmServer{
		bc:         server.BC,
		m:          m,
		TxConfirms: map[uint64]map[string]TxConfirm{},
		quit:       make(chan struct{}),
		wg:         &sync.WaitGroup{},
	}
}

func (w *WatchTxConfirmServer) AddTxConfirms(confirm TxConfirm) {
	w.lock.Lock()
	defer w.lock.Unlock()
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
					h := w.bc.BlockDAG().GetBlockByOrder(uint(order))
					ib := w.bc.BlockDAG().GetBlock(h)
					node := w.bc.BlockIndex().LookupNode(h)
					if node == nil {
						log.Error("no node")
						continue
					}
					confirmations := int64(w.bc.BlockDAG().GetConfirmations(node.GetID()))
					isBlue := w.bc.BlockDAG().IsBlue(ib.GetID())
					IsValid := w.bc.BlockIndex().NodeStatus(node).KnownInvalid()
					for _, tx := range txc {
						if !isBlue || !IsValid || confirmations > int64(tx.Confirms) {
							ntfn := &notificationTxConfirm{
								Tx:       tx.TxHash,
								Confirms: uint64(confirmations),
								IsBlue:   isBlue,
								IsValid:  IsValid,
								Order:    order,
							}
							marshalledJSON, err := cmds.MarshalCmd(nil, ntfn)
							if err != nil {
								log.Error(fmt.Sprintf("Failed to marshal tx confirm notification: "+
									"%v", err))
								continue
							}
							tx.wsc.QueueNotification(marshalledJSON)
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
