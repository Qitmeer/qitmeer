package notifymgr

import (
	"qitmeer/core/message"
	"qitmeer/core/types"
	"qitmeer/p2p/peerserver"
	"qitmeer/rpc"
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
		// reply to p2p
		ntmgr.RelayInventory(iv, tx)
		// reply to rpc
		if ntmgr.RpcServer != nil {
			//TODO reply to rpc layer (if websockect long connection or gbt long poll)
			// Notify websocket clients about mempool transactions.
			//nox.node.rpcServer.ntfnMgr.NotifyMempoolTx(tx, true)
			//
			// Potentially notify any getblocktemplate long poll clients
			// about stale block templates due to the new transaction.
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
