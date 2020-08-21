package notify

import (
	"github.com/Qitmeer/qitmeer/core/message"
	"github.com/Qitmeer/qitmeer/services/mempool"
)

// Notify interface manage message announce & relay & notification between mempool, websocket, gbt long pull
// and rpc server.
type Notify interface {
	AnnounceNewTransactions(newTxs []*mempool.TxDesc)
	RelayInventory(invVect *message.InvVect, data interface{})
	BroadcastMessage(msg message.Message)
}
