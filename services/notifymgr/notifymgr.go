package notifymgr

import (
	"github.com/noxproject/nox/core/message"
	"github.com/noxproject/nox/core/types"
	"github.com/noxproject/nox/p2p/peerserver"
	"github.com/noxproject/nox/rpc"
)

// NotifyMgr manage message announce & relay & notification between mempool, websocket, gbt long pull
// and rpc server.
type NotifyMgr struct {
	Server *peerserver.PeerServer
	RpcServer *rpc.RpcServer
}

// AnnounceNewTransactions generates and relays inventory vectors and notifies
// both websocket and getblocktemplate long poll clients of the passed
// transactions.  This function should be called whenever new transactions
// are added to the mempool.
func (ntmgr *NotifyMgr) AnnounceNewTransactions(newTxs []*types.Tx) {
	// Generate and relay inventory vectors for all newly accepted
	// transactions into the memory pool due to the original being
	// accepted.
	for _, tx := range newTxs {
		// Generate the inventory vector and relay it.
		iv := message.NewInvVect(message.InvTypeTx, tx.Hash())
		ntmgr.RelayInventory(iv, tx)
		//TODO p2p layer
		if ntmgr.RpcServer != nil {
			//// Notify websocket clients about mempool transactions.
			//nox.node.rpcServer.ntfnMgr.NotifyMempoolTx(tx, true)
			//
			//// Potentially notify any getblocktemplate long poll clients
			//// about stale block templates due to the new transaction.
			//nox.node.rpcServer.gbtWorkState.NotifyMempoolTx(
			//	nox.txMemPool.LastUpdated())
		}
	}
}

// RelayInventory relays the passed inventory vector to all connected peers
// that are not already known to have it.
func (ntmgr *NotifyMgr) RelayInventory(invVect *message.InvVect, data interface{}) {
	ntmgr.Server.RelayInventory(invVect,data)
}
