package rpc

import (
	"bytes"
	"encoding/hex"
	"fmt"
	"github.com/Qitmeer/qitmeer/common/hash"
	"github.com/Qitmeer/qitmeer/core/blockchain"
	"github.com/Qitmeer/qitmeer/core/types"
	"github.com/Qitmeer/qitmeer/rpc/client/cmds"
)

type TxConfirm struct {
	Confirms uint64
	TxHash   string
}

func (w *WatchTxConfirmServer) AddTxConfirms(confirm TxConfirm) {
	if w == nil {
		w = &WatchTxConfirmServer{}
	}
	if _, ok := (*w)[confirm.TxHash]; !ok {
		(*w)[confirm.TxHash] = TxConfirm{}
	}
	(*w)[confirm.TxHash] = confirm
}

type WatchTxConfirmServer map[string]TxConfirm

func (w *WatchTxConfirmServer) Handle(wsc *wsClient) {
	if w == nil || len(*w) <= 0 {
		return
	}
	bc := wsc.server.BC
	txIndex := wsc.server.TxIndex
	if txIndex == nil {
		log.Error("specify --txindex in configuration")
	}
	for tx, txconf := range *w {
		txHash := hash.MustHexToDecodedHash(tx)
		blockRegion, err := txIndex.TxBlockRegion(txHash)
		if err != nil {
			log.Error(err.Error(), "txhash", txHash)
			continue
		}
		if blockRegion == nil {
			if bc.CacheInvalidTx {
				blockRegion, err = txIndex.InvalidTxBlockRegion(txHash)
				if err != nil {
					log.Error(err.Error(), "txhash", txHash)
					continue
				}
			} else {
				log.Warn("tx hash not found", "txhash", txHash)
				continue
			}
		}
		txBytes, err := txIndex.GetTxBytes(blockRegion)
		if err != nil {
			log.Error("tx not found")
			continue
		}

		// Deserialize the transaction
		var msgTx types.Transaction
		err = msgTx.Deserialize(bytes.NewReader(txBytes))
		log.Trace("GetRawTx", "hex", hex.EncodeToString(txBytes))
		if err != nil {
			log.Error("Failed to deserialize transaction")
			continue
		}
		mtx := types.NewTx(&msgTx)
		mtx.IsDuplicate = bc.IsDuplicateTx(mtx.Hash(), blockRegion.Hash)
		ib := bc.BlockDAG().GetBlock(blockRegion.Hash)
		if ib == nil {
			log.Error("block hash not exist", "hash", blockRegion.Hash)
			delete(*w, tx)
			continue
		}
		confirmations := bc.BlockDAG().GetConfirmations(ib.GetID())
		isBlue := true
		if mtx.Tx.IsCoinBase() {
			isBlue = bc.BlockDAG().IsBlue(ib.GetID())
		}
		InValid := blockchain.BlockStatus(ib.GetStatus()).KnownInvalid()
		if uint64(confirmations) >= txconf.Confirms || InValid || !isBlue {
			ntfn := &cmds.NotificationTxConfirmNtfn{
				ConfirmResult: cmds.TxConfirmResult{
					Tx:       tx,
					Confirms: uint64(confirmations),
					IsBlue:   isBlue,
					IsValid:  !InValid,
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
			delete(*w, tx)
		}
	}
}
