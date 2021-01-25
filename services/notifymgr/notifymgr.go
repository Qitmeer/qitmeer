package notifymgr

import (
	"github.com/Qitmeer/qitmeer/core/types"
	"github.com/Qitmeer/qitmeer/p2p"
	"github.com/Qitmeer/qitmeer/rpc"
)

// NotifyMgr manage message announce & relay & notification between mempool, websocket, gbt long pull
// and rpc server.
type NotifyMgr struct {
	Server    *p2p.Service
	RpcServer *rpc.RpcServer
}

// AnnounceNewTransactions generates and relays inventory vectors and notifies
// both websocket and getblocktemplate long poll clients of the passed
// transactions.  This function should be called whenever new transactions
// are added to the mempool.
func (ntmgr *NotifyMgr) AnnounceNewTransactions(newTxs []*types.TxDesc) {
	// reply to p2p
	ntmgr.RelayInventory(newTxs)

	if ntmgr.RpcServer != nil {
		ntmgr.RpcServer.NotifyNewTransactions(newTxs)
	}
}

// RelayInventory relays the passed inventory vector to all connected peers
// that are not already known to have it.
func (ntmgr *NotifyMgr) RelayInventory(data interface{}) {
	ntmgr.Server.RelayInventory(data)
}

func (ntmgr *NotifyMgr) BroadcastMessage(data interface{}) {
	ntmgr.Server.BroadcastMessage(data)
}
