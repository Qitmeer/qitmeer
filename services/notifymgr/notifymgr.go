package notifymgr

import (
	"fmt"
	"github.com/Qitmeer/qitmeer/core/types"
	"github.com/Qitmeer/qitmeer/log"
	"github.com/Qitmeer/qitmeer/p2p"
	"github.com/Qitmeer/qitmeer/rpc"
	"github.com/libp2p/go-libp2p-core/peer"
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
func (ntmgr *NotifyMgr) AnnounceNewTransactions(newTxs []*types.TxDesc, filters []peer.ID) {
	if len(newTxs) <= 0 {
		return
	}
	for _, tx := range newTxs {
		log.Trace(fmt.Sprintf("Announce new transaction :hash=%s height=%d add=%s", tx.Tx.Hash().String(), tx.Height, tx.Added.String()))
	}
	// reply to p2p
	for _, tx := range newTxs {
		ntmgr.RelayInventory(tx, filters)
	}

	if ntmgr.RpcServer != nil {
		ntmgr.RpcServer.NotifyNewTransactions(newTxs)
	}
}

// RelayInventory relays the passed inventory vector to all connected peers
// that are not already known to have it.
func (ntmgr *NotifyMgr) RelayInventory(data interface{}, filters []peer.ID) {
	ntmgr.Server.RelayInventory(data, filters)
}

func (ntmgr *NotifyMgr) BroadcastMessage(data interface{}) {
	ntmgr.Server.BroadcastMessage(data)
}

func (ntmgr *NotifyMgr) AddRebroadcastInventory(newTxs []*types.TxDesc) {
	for _, tx := range newTxs {
		ntmgr.Server.Rebroadcast().AddInventory(tx.Tx.Hash(), tx)
	}
}

// Transaction has one confirmation on the main chain. Now we can mark it as no
// longer needing rebroadcasting.
func (ntmgr *NotifyMgr) TransactionConfirmed(tx *types.Tx) {
	ntmgr.Server.Rebroadcast().RemoveInventory(tx.Hash())
}
