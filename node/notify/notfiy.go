package notify

import (
	"github.com/noxproject/nox/core/message"
	"github.com/noxproject/nox/core/types"
)

// Notify interface manage message announce & relay & notification between mempool, websocket, gbt long pull
// and rpc server.
type Notify interface {
	AnnounceNewTransactions(newTxs []*types.Tx)
	RelayInventory(invVect *message.InvVect, data interface{})
}

