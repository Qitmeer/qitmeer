package notify

import (
	"github.com/HalalChain/qitmeer/core/message"
	"github.com/HalalChain/qitmeer/core/types"
)

// Notify interface manage message announce & relay & notification between mempool, websocket, gbt long pull
// and rpc server.
type Notify interface {
	AnnounceNewTransactions(newTxs []*types.Tx)
	RelayInventory(invVect *message.InvVect, data interface{})
}

