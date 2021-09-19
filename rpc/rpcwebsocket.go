/*
 * Copyright (c) 2017-2020 The qitmeer developers
 */

package rpc

import (
	"fmt"
	"github.com/Qitmeer/qitmeer/common/hash"
	"github.com/Qitmeer/qitmeer/core/blockchain"
	"github.com/Qitmeer/qitmeer/core/event"
	"github.com/Qitmeer/qitmeer/core/types"
	"github.com/Qitmeer/qitmeer/rpc/client/cmds"
	"github.com/Qitmeer/qitmeer/rpc/websocket"
	"time"
)

var timeZeroVal time.Time

func (s *RpcServer) subscribe(events *event.Feed) {
	ch := make(chan *event.Event)
	sub := events.Subscribe(ch)
	go func() {
		defer sub.Unsubscribe()
		for {
			select {
			case ev := <-ch:
				if ev.Data != nil {
					switch value := ev.Data.(type) {
					case *blockchain.Notification:
						s.handleNotifyMsg(value)
					}
				}
				if ev.Ack != nil {
					ev.Ack <- struct{}{}
				}
			case <-s.quit:
				log.Info("Close RpcServer Event Subscribe")
				return
			}
		}
	}()
}

func (s *RpcServer) handleNotifyMsg(notification *blockchain.Notification) {
	switch notification.Type {
	case blockchain.BlockAccepted:
		bnd, ok := notification.Data.(*blockchain.BlockAcceptedNotifyData)
		if !ok {
			log.Warn("Chain accepted notification is not " +
				"BlockAcceptedNotifyData.")
			break
		}
		s.ntfnMgr.NotifyBlockAccepted(bnd)

	case blockchain.BlockConnected:
		blockSlice, ok := notification.Data.([]*types.SerializedBlock)
		if !ok {
			log.Warn("Chain connected notification is not a block slice.")
			break
		}

		if len(blockSlice) != 1 {
			log.Warn("Chain connected notification is wrong size slice.")
			break
		}
		s.ntfnMgr.NotifyBlockConnected(blockSlice[0])

	case blockchain.BlockDisconnected:
		block, ok := notification.Data.(*types.SerializedBlock)
		if !ok {
			log.Warn("Chain disconnected notification is not a block slice.")
			break
		}
		s.ntfnMgr.NotifyBlockDisconnected(block)

	case blockchain.Reorganization:
		rnd, ok := notification.Data.(*blockchain.ReorganizationNotifyData)
		if !ok {
			log.Warn("Chain accepted notification is not " +
				"ReorganizationNotifyData.")
			break
		}
		s.ntfnMgr.NotifyReorganization(rnd)
	}
}

func (s *RpcServer) NotifyNewTransactions(txns []*types.TxDesc) {
	for _, txD := range txns {
		// Notify websocket clients about mempool transactions.
		s.ntfnMgr.NotifyMempoolTx(txD.Tx, true)
	}
}

func (s *RpcServer) WebsocketHandler(conn *websocket.Conn, remoteAddr string, isAdmin bool) {
	// Clear the read deadline that was set before the websocket hijacked
	// the connection.
	conn.SetReadDeadline(timeZeroVal)
	// Limit max number of websocket clients.
	log.Info(fmt.Sprintf("New websocket client %s", remoteAddr))
	if s.ntfnMgr.NumClients()+1 > s.config.RPCMaxWebsockets {
		log.Info(fmt.Sprintf("Max websocket clients exceeded [%d] - disconnecting client %s", s.config.RPCMaxWebsockets,
			remoteAddr))
		conn.Close()
		return
	}

	// Create a new websocket client to handle the new websocket connection
	// and wait for it to shutdown.  Once it has shutdown (and hence
	// disconnected), remove it and any notifications it registered for.
	client, err := newWebsocketClient(s, conn, remoteAddr, isAdmin)
	if err != nil {
		log.Error(fmt.Sprintf("Failed to serve client %s: %v", remoteAddr, err))
		conn.Close()
		return
	}
	s.ntfnMgr.AddClient(client)
	client.Start()
	client.WaitForShutdown()
	s.ntfnMgr.RemoveClient(client)
	log.Info(fmt.Sprintf("Disconnected websocket client %s", remoteAddr))
}

type wsCommandHandler func(*wsClient, interface{}) (interface{}, error)

var wsHandlers map[string]wsCommandHandler
var wsHandlersBeforeInit = map[string]wsCommandHandler{
	"notifyBlocks":              handleNotifyBlocks,
	"stopNotifyBlocks":          handleStopNotifyBlocks,
	"session":                   handleSession,
	"notifynewtransactions":     handleNotifyNewTransactions,
	"stopnotifynewtransactions": handleStopNotifyNewTransactions,
	"notifyTxsByAddr":           handleNotifyTxsByAddr,
	"stopnotifyTxsByAddr":       handleStopNotifyTxsByAddr,
	"rescan":                    handleRescan,
	"notifyTxsConfirmed":        handleNotifyTxsConfirmed,
	"removeTxsConfirmed":        handleRemoveTxsConfirmed,
}

func handleNotifyBlocks(wsc *wsClient, icmd interface{}) (interface{}, error) {
	wsc.server.ntfnMgr.RegisterBlockUpdates(wsc)
	return nil, nil
}

func handleStopNotifyBlocks(wsc *wsClient, icmd interface{}) (interface{}, error) {
	wsc.server.ntfnMgr.UnregisterBlockUpdates(wsc)
	return nil, nil
}

func handleSession(wsc *wsClient, icmd interface{}) (interface{}, error) {
	return &SessionResult{SessionID: wsc.sessionID}, nil
}

func handleNotifyNewTransactions(wsc *wsClient, icmd interface{}) (interface{}, error) {
	cmd, ok := icmd.(*cmds.NotifyNewTransactionsCmd)
	if !ok {
		return nil, cmds.ErrRPCInternal
	}
	wsc.verboseTxUpdates = cmd.Verbose
	wsc.server.ntfnMgr.RegisterNewMempoolTxsUpdates(wsc)
	return nil, nil
}

func handleNotifyTxsByAddr(wsc *wsClient, icmd interface{}) (interface{}, error) {
	cmd, ok := icmd.(*cmds.NotifyTxsByAddrCmd)
	if !ok {
		return nil, cmds.ErrRPCInternal
	}
	outPoints := make([]types.TxOutPoint, len(cmd.OutPoints))
	for i := range cmd.OutPoints {
		h, err := hash.NewHashFromStr(cmd.OutPoints[i].Hash)
		if err != nil {
			return nil, cmds.ErrRPCInvalidParams
		}
		outPoints[i] = types.TxOutPoint{
			Hash:     *h,
			OutIndex: cmd.OutPoints[i].Index,
		}
	}

	wsc.Lock()
	if cmd.Reload || wsc.filterData == nil {
		wsc.filterData = newWSClientFilter(cmd.Addresses, outPoints)
		wsc.Unlock()
	} else {
		wsc.Unlock()

		wsc.filterData.mu.Lock()
		for _, a := range cmd.Addresses {
			wsc.filterData.addAddressStr(a)
		}
		wsc.filterData.mu.Unlock()
	}
	return nil, nil
}

func handleStopNotifyTxsByAddr(wsc *wsClient, icmd interface{}) (interface{}, error) {
	cmd, ok := icmd.(*cmds.UnNotifyTxsByAddrCmd)
	if !ok {
		return nil, cmds.ErrRPCInternal
	}
	for _, addr := range cmd.Addresses {
		wsc.filterData.removeAddressStr(addr)
	}
	return nil, nil
}

func handleStopNotifyNewTransactions(wsc *wsClient, icmd interface{}) (interface{}, error) {
	wsc.server.ntfnMgr.UnregisterNewMempoolTxsUpdates(wsc)
	return nil, nil
}

func init() {
	wsHandlers = wsHandlersBeforeInit
}

// rpcDecodeHexError is a convenience function for returning a nicely formatted
// RPC error which indicates the provided hex string failed to decode.
func rpcDecodeHexError(gotHex string) *cmds.RPCError {
	log.Error(gotHex)
	return cmds.ErrRPCDecodeHexString
}

// handleRescan implements the rescan command extension for websocket
// connections.
//
// NOTE: This does not smartly handle the re-org and BlockDAG case should be covered carefully and fixing requires database
// changes (for safe, concurrent access to full block ranges, and support
// for other chains than the best chain).  It will, however, detect whether
// a reorg removed a block that was previously processed, and result in the
// handler erroring.  Clients must handle this by finding a block still in
// the chain (perhaps from a rescanprogress notification) to resume their
// rescan.
func handleRescan(wsc *wsClient, icmd interface{}) (interface{}, error) {
	cmd, ok := icmd.(*cmds.RescanCmd)
	if !ok {
		return nil, cmds.ErrRPCInternal
	}

	outpoints := make([]*types.TxOutPoint, 0, len(cmd.OutPoints))
	for i := range cmd.OutPoints {
		cmdOutpoint := &cmd.OutPoints[i]
		blockHash, err := hash.NewHashFromStr(cmdOutpoint.Hash)
		if err != nil {
			return nil, rpcDecodeHexError(cmdOutpoint.Hash)
		}
		outpoint := types.NewOutPoint(blockHash, cmdOutpoint.Index)
		outpoints = append(outpoints, outpoint)
	}

	numAddrs := len(cmd.Addresses)
	if numAddrs == 1 {
		log.Info("Beginning rescan for 1 address")
	} else {
		log.Info(fmt.Sprintf("Beginning rescan for %d addresses", numAddrs))
	}

	// Build lookup maps.
	lookups := rescanKeys{
		addrs:   map[string]struct{}{},
		unspent: map[types.TxOutPoint]struct{}{},
	}
	for _, addrStr := range cmd.Addresses {
		lookups.addrs[addrStr] = struct{}{}
	}
	for _, outpoint := range outpoints {
		lookups.unspent[*outpoint] = struct{}{}
	}

	chain := wsc.server.BC

	var (
		err           error
		lastBlock     *types.SerializedBlock
		lastTxHash    *hash.Hash
		lastBlockHash *hash.Hash
	)
	if len(lookups.addrs) != 0 || len(lookups.unspent) != 0 {
		// With all the arguments parsed, we'll execute our chunked rescan
		// which will notify the clients of any address deposits or output
		// spends.
		lastBlock, lastBlockHash, lastTxHash, err = scanBlockChunks(
			wsc, cmd, &lookups, cmd.BeginBlock, cmd.EndBlock, chain,
		)
		if err != nil {
			return nil, err
		}

		// If the last block is nil, then this means that the client
		// disconnected mid-rescan. As a result, we don't need to send
		// anything back to them.
		if lastBlock == nil {
			return nil, nil
		}
	} else {
		log.Info("Skipping rescan as client has no addrs/utxos")

		// If we didn't actually do a rescan, then we'll give the
		// client our best known block within the final rescan finished
		// notification.
		chainTip := chain.BestSnapshot()
		lastBlockHash = &chainTip.Hash
		lastBlock, err = chain.FetchBlockByHash(lastBlockHash)
		if err != nil {
			return nil, cmds.ErrRPCBlockNotFound
		}
	}
	lastTx := ""
	if lastTxHash != nil {
		lastTx = lastTxHash.String()
	}
	// Notify websocket client of the finished rescan.  Due to how btcd
	// asynchronously queues notifications to not block calling code,
	// there is no guarantee that any of the notifications created during
	// rescan (such as rescanprogress, recvtx and redeemingtx) will be
	// received before the rescan RPC returns.  Therefore, another method
	// is needed to safely inform clients that all rescan notifications have
	// been sent.
	n := cmds.NewRescanFinishedNtfn(
		lastBlockHash.String(),
		lastTx,
		lastBlock.Order(),
		lastBlock.Block().Header.Timestamp.Unix(),
	)
	if mn, err := cmds.MarshalCmd(nil, n); err != nil {
		log.Error(fmt.Sprintf("Failed to marshal rescan finished "+
			"notification: %v", err))
	} else {
		// The rescan is finished, so we don't care whether the client
		// has disconnected at this point, so discard error.
		_ = wsc.QueueNotification(mn)
	}

	log.Info("Finished rescan")
	return nil, nil
}

func handleNotifyTxsConfirmed(wsc *wsClient, icmd interface{}) (interface{}, error) {
	cmd, ok := icmd.(*cmds.NotifyTxsConfirmedCmd)
	if !ok {
		return nil, cmds.ErrRPCInternal
	}
	wsc.TxConfirmsLock.Lock()
	defer wsc.TxConfirmsLock.Unlock()
	for _, tx := range cmd.Txs {
		wsc.TxConfirms.AddTxConfirms(TxConfirm{
			Confirms:  uint64(tx.Confirmations),
			TxHash:    tx.Txid,
			EndHeight: tx.EndHeight,
		})
	}
	wsc.server.ntfnMgr.RegisterTxConfirm(wsc)
	return nil, nil
}

func handleRemoveTxsConfirmed(wsc *wsClient, icmd interface{}) (interface{}, error) {
	cmd, ok := icmd.(*cmds.RemoveTxsConfirmedCmd)
	if !ok {
		return nil, cmds.ErrRPCInternal
	}
	wsc.TxConfirmsLock.Lock()
	defer wsc.TxConfirmsLock.Unlock()
	for _, tx := range cmd.Txs {
		wsc.TxConfirms.RemoveTxConfirms(TxConfirm{
			TxHash: tx.Txid,
		})
	}
	return nil, nil
}
