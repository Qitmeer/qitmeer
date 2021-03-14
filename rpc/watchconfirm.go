package rpc

import (
	"fmt"
	"github.com/Qitmeer/qitmeer/rpc/client/cmds"
)

type TxConfirm struct {
	Order    uint64
	Confirms uint64
	TxHash   string
}

func (w *WatchTxConfirmServer) AddTxConfirms(confirm TxConfirm) {
	if _, ok := (*w)[confirm.Order]; !ok {
		(*w)[confirm.Order] = map[string]TxConfirm{}
	}
	(*w)[confirm.Order][confirm.TxHash] = confirm
}

type WatchTxConfirmServer map[uint64]map[string]TxConfirm

func (w *WatchTxConfirmServer) Handle(wsc *wsClient) {
	if len(*w) <= 0 {
		return
	}
	bc := wsc.server.BC
	for order, txc := range *w {
		h := bc.BlockDAG().GetBlockByOrder(uint(order))
		if h == nil {
			log.Error("order not exist", "order", order)
			delete(*w, order)
			continue
		}
		ib := bc.BlockDAG().GetBlock(h)
		if ib == nil {
			log.Error("block hash not exist", "hash", h)
			delete(*w, order)
			continue
		}
		node := bc.BlockIndex().LookupNode(h)
		if node == nil {
			log.Error("no node")
			continue
		}
		confirmations := int64(bc.BlockDAG().GetConfirmations(node.GetID()))
		isBlue := bc.BlockDAG().IsBlue(ib.GetID())
		InValid := bc.BlockIndex().NodeStatus(node).KnownInvalid()
		for _, tx := range txc {
			if !isBlue || InValid || confirmations >= int64(tx.Confirms) {
				ntfn := &cmds.NotificationTxConfirmNtfn{
					ConfirmResult: cmds.TxConfirmResult{
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
				err = wsc.QueueNotification(marshalledJSON)
				if err != nil {
					log.Error("notify failed", "err", err)
					continue
				}
				delete((*w)[order], tx.TxHash)
			}
		}
	}
}
