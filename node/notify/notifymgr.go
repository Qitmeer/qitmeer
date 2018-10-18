package notify

import (
	"github.com/noxproject/nox/core/message"
	"github.com/noxproject/nox/core/types"
)

// NotifyMgr manage message announce & relay & notification between mempool, websocket, gbt long pull
// and rpc server.
type NotifyMgr struct {

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
		/*
		if nox.node.rpcServer != nil {
			// Notify websocket clients about mempool transactions.
			nox.node.rpcServer.ntfnMgr.NotifyMempoolTx(tx, true)

			// Potentially notify any getblocktemplate long poll clients
			// about stale block templates due to the new transaction.
			nox.node.rpcServer.gbtWorkState.NotifyMempoolTx(
				nox.txMemPool.LastUpdated())
		}
		*/
	}
}

// RelayInventory relays the passed inventory vector to all connected peers
// that are not already known to have it.
func (ntmgr *NotifyMgr) RelayInventory(invVect *message.InvVect, data interface{}) {
	// TODO p2p layer
	/*
	s.relayInv <- relayMsg{invVect: invVect, data: data}
	*/
}
