package notify

import (
	"github.com/Qitmeer/qitmeer/core/types"
)

// Notify interface manage message announce & relay & notification between mempool, websocket, gbt long pull
// and rpc server.
type Notify interface {
	AnnounceNewTransactions(newTxs []*types.TxDesc)
	RelayInventory(data interface{})
	BroadcastMessage(data interface{})
	TransactionConfirmed(tx *types.Tx)
	AddRebroadcastInventory(newTxs []*types.TxDesc)
}
