package rpc

import (
	"fmt"
	"github.com/Qitmeer/qitmeer/common/hash"
	"github.com/Qitmeer/qitmeer/common/marshal"
	"github.com/Qitmeer/qitmeer/core/blockchain"
	"github.com/Qitmeer/qitmeer/core/json"
	"github.com/Qitmeer/qitmeer/core/types"
	"github.com/Qitmeer/qitmeer/engine/txscript"
	"github.com/Qitmeer/qitmeer/params"
	"github.com/Qitmeer/qitmeer/rpc/client/cmds"
	"github.com/Qitmeer/qitmeer/rpc/websocket"
	"sync"
	"time"
)

// Notification types
type notificationBlockConnected types.SerializedBlock
type notificationBlockDisconnected types.SerializedBlock
type notificationBlockAccepted blockchain.BlockAcceptedNotifyData

type notificationReorganization struct {
	OldBlocks []*hash.Hash
	NewBlock  *hash.Hash
	NewOrder  uint64
}

type notificationTxAcceptedByMempool struct {
	isNew bool
	tx    *types.Tx
}

type notificationTxByBlock struct {
	blk *types.SerializedBlock
	tx  *types.Tx
}

// Notification control requests
type notificationRegisterClient wsClient
type notificationUnregisterClient wsClient
type notificationRegisterBlocks wsClient
type notificationRegisterTxConfirms wsClient
type notificationUnregisterBlocks wsClient
type notificationRegisterNewMempoolTxs wsClient
type notificationUnregisterNewMempoolTxs wsClient
type notificationScanComplete wsClient

type wsNotificationManager struct {
	server            *RpcServer
	queueNotification chan interface{}
	notificationMsgs  chan interface{}
	numClients        chan int
	wg                sync.WaitGroup
	quit              chan struct{}
}

func (m *wsNotificationManager) Start() {
	m.wg.Add(2)
	go m.queueHandler()
	go m.notificationHandler()
}

func (m *wsNotificationManager) Stop() {
	close(m.quit)
	m.wg.Wait()
}

func (m *wsNotificationManager) queueHandler() {
	queueHandler(m.queueNotification, m.notificationMsgs, m.quit)
	m.wg.Done()
}

func (m *wsNotificationManager) notificationHandler() {
	// clients is a map of all currently connected websocket clients.
	clients := make(map[chan struct{}]*wsClient)
	blockNotifications := make(map[chan struct{}]*wsClient)
	txNotifications := make(map[chan struct{}]*wsClient)
	txConfirms := make(map[chan struct{}]*wsClient)

out:
	for {
		select {
		case n, ok := <-m.notificationMsgs:
			if !ok {
				// queueHandler quit.
				break out
			}
			switch n := n.(type) {
			case *notificationBlockConnected:
				block := (*types.SerializedBlock)(n)
				if len(blockNotifications) != 0 {
					m.notifyBlockConnected(blockNotifications,
						block)
				}

			case *notificationBlockDisconnected:
				block := (*types.SerializedBlock)(n)
				if len(blockNotifications) != 0 {
					m.notifyBlockDisconnected(blockNotifications,
						block)
				}

			case *notificationBlockAccepted:
				band := (*blockchain.BlockAcceptedNotifyData)(n)
				block := band.Block
				if len(blockNotifications) != 0 {
					m.notifyBlockAccepted(blockNotifications,
						block)
				}
				if band.IsMainChainTipChange {
					// do something
					if len(txConfirms) != 0 {
						for k, wsc := range txConfirms {
							if wsc.Disconnected() {
								delete(txConfirms, k)
								continue
							}
							wsc.TxConfirmsLock.Lock()
							wsc.TxConfirms.Handle(wsc, uint64(block.Height()))
							wsc.TxConfirmsLock.Unlock()
						}
					}
				}

			case *notificationReorganization:
				if len(blockNotifications) != 0 {
					m.notifyReorganization(blockNotifications, n)
				}

			case *notificationTxAcceptedByMempool:

				if n.isNew && len(txNotifications) != 0 {
					m.notifyForNewTx(txNotifications, n.tx)
				}

			case *notificationRegisterBlocks:
				wsc := (*wsClient)(n)
				blockNotifications[wsc.quit] = wsc

			case *notificationRegisterTxConfirms:
				wsc := (*wsClient)(n)
				txConfirms[wsc.quit] = wsc

			case *notificationUnregisterBlocks:
				wsc := (*wsClient)(n)
				delete(blockNotifications, wsc.quit)

			case *notificationRegisterClient:
				wsc := (*wsClient)(n)
				clients[wsc.quit] = wsc

			case *notificationUnregisterClient:
				wsc := (*wsClient)(n)
				// Remove any requests made by the client as well as
				// the client itself.
				delete(blockNotifications, wsc.quit)

				delete(clients, wsc.quit)

			case *notificationRegisterNewMempoolTxs:
				wsc := (*wsClient)(n)
				log.Info(fmt.Sprintf("listen tx %s", wsc.addr))
				txNotifications[wsc.quit] = wsc

			case *notificationUnregisterNewMempoolTxs:
				wsc := (*wsClient)(n)
				delete(txNotifications, wsc.quit)

			default:
				log.Warn("Unhandled notification type")
			}

		case m.numClients <- len(clients):

		case <-m.quit:
			// RPC server shutting down.
			break out
		}
	}

	m.notifyExit(clients)
	for _, c := range clients {
		c.Disconnect()
	}
	m.wg.Done()
}

func (m *wsNotificationManager) NotifyBlockConnected(block *types.SerializedBlock) {
	select {
	case m.queueNotification <- (*notificationBlockConnected)(block):
	case <-m.quit:
	}
}

func (m *wsNotificationManager) notifyBlockConnected(clients map[chan struct{}]*wsClient, block *types.SerializedBlock) {
	txs, err := GetTxsHexFromBlock(block, true)
	if err != nil {
		log.Error(err.Error())
		return
	}
	ntfn := cmds.NewBlockConnectedNtfn(block.Hash().String(), int64(block.Height()), int64(block.Order()), block.Block().Header.Timestamp.Unix(), txs)
	marshalledJSON, err := cmds.MarshalCmd(nil, ntfn)
	for _, wsc := range clients {
		// Marshal and queue notification.
		wsc.QueueNotification(marshalledJSON)
	}
}

func (m *wsNotificationManager) NotifyBlockDisconnected(block *types.SerializedBlock) {
	select {
	case m.queueNotification <- (*notificationBlockDisconnected)(block):
	case <-m.quit:
	}
}

func (*wsNotificationManager) notifyBlockDisconnected(clients map[chan struct{}]*wsClient, block *types.SerializedBlock) {
	if len(clients) == 0 {
		return
	}
	txs, err := GetTxsHexFromBlock(block, false)
	if err != nil {
		log.Error(err.Error())
		return
	}
	// Notify interested websocket clients about the disconnected block.
	ntfn := cmds.NewBlockDisconnectedNtfn(block.Hash().String(),
		int64(block.Height()), int64(block.Order()), block.Block().Header.Timestamp.Unix(), txs)
	marshalledJSON, err := cmds.MarshalCmd(nil, ntfn)
	if err != nil {
		log.Error(fmt.Sprintf("Failed to marshal block disconnected "+
			"notification: %v", err))
		return
	}
	for _, wsc := range clients {
		wsc.QueueNotification(marshalledJSON)
	}
}

func (m *wsNotificationManager) NotifyBlockAccepted(band *blockchain.BlockAcceptedNotifyData) {
	select {
	case m.queueNotification <- (*notificationBlockAccepted)(band):
	case <-m.quit:
	}
}

func (m *wsNotificationManager) notifyBlockAccepted(clients map[chan struct{}]*wsClient, block *types.SerializedBlock) {
	for _, tx := range block.Transactions() {
		subClients := m.subscribedClients(tx, clients)
		for quitChan, _ := range subClients {
			m.notifyForBlockTx(clients[quitChan], tx, block)
		}
	}
}

func (m *wsNotificationManager) NotifyReorganization(rnd *blockchain.ReorganizationNotifyData) {
	nr := &notificationReorganization{
		OldBlocks: rnd.OldBlocks,
		NewBlock:  rnd.NewBlock,
		NewOrder:  rnd.NewOrder,
	}
	select {
	case m.queueNotification <- nr:
	case <-m.quit:
	}
}

func (m *wsNotificationManager) notifyReorganization(clients map[chan struct{}]*wsClient, nr *notificationReorganization) {

	olds := []string{}
	for _, old := range nr.OldBlocks {
		olds = append(olds, old.String())
	}
	ntfn := cmds.NewReorganizationNtfn(nr.NewBlock.String(), int64(nr.NewOrder), olds)
	marshalledJSON, err := cmds.MarshalCmd(nil, ntfn)
	if err != nil {
		log.Error(fmt.Sprintf("Failed to marshal block accepted notification: "+
			"%v", err))
		return
	}
	for _, wsc := range clients {
		wsc.QueueNotification(marshalledJSON)
	}
}

func (m *wsNotificationManager) NumClients() (n int) {
	select {
	case n = <-m.numClients:
	case <-m.quit: // Use default n (0) if server has shut down.
	}
	return
}

func (m *wsNotificationManager) AddClient(wsc *wsClient) {
	m.queueNotification <- (*notificationRegisterClient)(wsc)
}

func (m *wsNotificationManager) RemoveClient(wsc *wsClient) {
	select {
	case m.queueNotification <- (*notificationUnregisterClient)(wsc):
	case <-m.quit:
	}
}

func (m *wsNotificationManager) RegisterBlockUpdates(wsc *wsClient) {
	m.queueNotification <- (*notificationRegisterBlocks)(wsc)
}

func (m *wsNotificationManager) RegisterTxConfirm(wsc *wsClient) {
	m.queueNotification <- (*notificationRegisterTxConfirms)(wsc)
}

func (m *wsNotificationManager) UnregisterBlockUpdates(wsc *wsClient) {
	m.queueNotification <- (*notificationUnregisterBlocks)(wsc)
}

func (m *wsNotificationManager) RegisterNewMempoolTxsUpdates(wsc *wsClient) {
	m.queueNotification <- (*notificationRegisterNewMempoolTxs)(wsc)
}

func (m *wsNotificationManager) RegisterNotifyTxsByAddr(wsc *wsClient) {
	m.queueNotification <- (*notificationRegisterNewMempoolTxs)(wsc)
}

func (m *wsNotificationManager) UnregisterNewMempoolTxsUpdates(wsc *wsClient) {
	m.queueNotification <- (*notificationUnregisterNewMempoolTxs)(wsc)
}

func (m *wsNotificationManager) RegisterNotifyComplete(wsc *wsClient) {
	m.queueNotification <- (*notificationScanComplete)(wsc)
}

func (m *wsNotificationManager) NotifyMempoolTx(tx *types.Tx, isNew bool) {
	n := &notificationTxAcceptedByMempool{
		isNew: isNew,
		tx:    tx,
	}

	select {
	case m.queueNotification <- n:
	case <-m.quit:
	}
}

func (m *wsNotificationManager) NotifyBlockTx(wsc *wsClient, tx *types.Tx, blk *types.SerializedBlock) {
	m.notifyForBlockTx(wsc, tx, blk)
}

// subscribedClients returns the set of all websocket client quit channels that
// are registered to receive notifications regarding tx, either due to tx
// spending a watched output or outputting to a watched address.  Matching
// client's filters are updated based on this transaction's outputs and output
// addresses that may be relevant for a client.
func (m *wsNotificationManager) subscribedClients(tx *types.Tx,
	clients map[chan struct{}]*wsClient) map[chan struct{}]struct{} {

	// Use a map of client quit channels as keys to prevent duplicates when
	// multiple inputs and/or outputs are relevant to the client.
	subscribed := make(map[chan struct{}]struct{})

	msgTx := tx.Tx
	for _, input := range msgTx.TxIn {
		for quitChan, wsc := range clients {
			wsc.Lock()
			filter := wsc.filterData
			wsc.Unlock()
			if filter == nil {
				continue
			}
			pkScript, err := txscript.ComputePkScript(
				input.SignScript,
			)
			if err != nil {
				continue
			}
			addr, err := pkScript.Address(wsc.server.ChainParams)
			if err != nil {
				continue
			}
			filter.mu.Lock()
			if filter.existsAddress(addr) {
				subscribed[quitChan] = struct{}{}
			}
			filter.mu.Unlock()
		}
	}

	for i, output := range msgTx.TxOut {
		_, addrs, _, err := txscript.ExtractPkScriptAddrs(
			output.PkScript, m.server.ChainParams)
		if err != nil {
			// Clients are not able to subscribe to
			// nonstandard or non-address outputs.
			continue
		}
		for quitChan, wsc := range clients {
			wsc.Lock()
			filter := wsc.filterData
			wsc.Unlock()
			if filter == nil {
				continue
			}
			filter.mu.Lock()
			for _, a := range addrs {
				if filter.existsAddress(a) {
					subscribed[quitChan] = struct{}{}
					op := types.TxOutPoint{
						Hash:     *tx.Hash(),
						OutIndex: uint32(i),
					}
					filter.addUnspentOutPoint(&op)
				}
			}
			filter.mu.Unlock()
		}
	}

	return subscribed
}

func (m *wsNotificationManager) notifyForNewTx(clients map[chan struct{}]*wsClient, tx *types.Tx) {
	if len(clients) <= 0 {
		return
	}
	needTx := false
	needVerboseTx := false
	for _, wsc := range clients {
		if wsc.verboseTxUpdates {
			needVerboseTx = true
		} else {
			needTx = true
		}

		if needVerboseTx && needTx {
			break
		}
	}

	var marshalledJSON []byte
	var err error
	if needTx {
		txHashStr := tx.Hash().String()
		amountsM := map[string]*types.Amount{}
		for _, txOut := range tx.Tx.TxOut {
			amount, ok := amountsM[txOut.Amount.Id.Name()]
			if ok {
				_, err := amount.Add(amount, &txOut.Amount)
				if err != nil {
					log.Error("notifyForNewTx fail")
					return
				}
			} else {
				amountsM[txOut.Amount.Id.Name()] = &types.Amount{Value: txOut.Amount.Value, Id: txOut.Amount.Id}
			}
		}

		var amounts types.AmountGroup
		for _, amount := range amountsM {
			amounts = append(amounts, *amount)
		}

		ntfn := cmds.NewTxAcceptedNtfn(txHashStr, amounts)
		marshalledJSON, err = cmds.MarshalCmd(nil, ntfn)
		if err != nil {
			log.Error(fmt.Sprintf("Failed to marshal tx notification: %s", err.Error()))
			return
		}
	}

	var marshalledJSONVerbose []byte
	if needVerboseTx {
		rawTx := &json.DecodeRawTransactionResult{
			Txid:       tx.Hash().String(),
			Duplicate:  tx.IsDuplicate,
			IsCoinbase: tx.Tx.IsCoinBase(),
			Hash:       tx.Tx.TxHashFull().String(),
			Version:    tx.Tx.Version,
			LockTime:   tx.Tx.LockTime,
			Time:       tx.Tx.Timestamp.Format(time.RFC3339),
			Vin:        marshal.MarshJsonVin(tx.Tx),
			Vout:       marshal.MarshJsonVout(tx.Tx, nil, params.ActiveNetParams.Params),
		}
		verboseNtfn := cmds.NewTxAcceptedVerboseNtfn(*rawTx)
		marshalledJSONVerbose, err = cmds.MarshalCmd(nil, verboseNtfn)
		if err != nil {
			log.Error(fmt.Sprintf("Failed to marshal verbose tx "+
				"notification: %s", err.Error()))
			return
		}
	}
	clientsToNotify := m.subscribedClients(tx, clients)

	for quitChan := range clientsToNotify {
		wsc := clients[quitChan]
		if wsc.verboseTxUpdates {
			wsc.QueueNotification(marshalledJSONVerbose)
		} else {
			wsc.QueueNotification(marshalledJSON)
		}
	}
}

func (m *wsNotificationManager) notifyExit(clients map[chan struct{}]*wsClient) {
	if len(clients) <= 0 {
		return
	}
	for _, wsc := range clients {
		exitNtfn := &cmds.NodeExitNtfn{}
		marshalledJSON, err := cmds.MarshalCmd(nil, exitNtfn)
		if err != nil {
			log.Error(fmt.Sprintf("Failed to marshal exitNtfn "+
				"notification: %s", err.Error()))
			return
		}
		_ = wsc.conn.WriteMessage(websocket.TextMessage, marshalledJSON)
	}
}

func (m *wsNotificationManager) notifyForBlockTx(wsc *wsClient, tx *types.Tx,
	blk *types.SerializedBlock) {
	ib := m.server.BC.BlockDAG().GetBlock(blk.Hash())
	node := m.server.BC.BlockDAG().GetBlock(blk.Hash())
	if node == nil {
		log.Error("no node")
		return
	}
	confirmations := int64(m.server.BC.BlockDAG().GetConfirmations(node.GetID()))
	isBlue := m.server.BC.BlockDAG().IsBlue(ib.GetID())
	InValid := node.GetStatus().KnownInvalid()

	var err error

	var marshalledJSONVerbose []byte

	rawTx := &json.DecodeRawTransactionResult{
		IsCoinbase: tx.Tx.IsCoinBase(),
		Txvalid:    !InValid,
		BlockHash:  blk.Hash().String(),
		Duplicate:  tx.IsDuplicate,
		Order:      uint64(node.GetOrder()),
		Confirms:   uint64(confirmations),
		IsBlue:     isBlue,
		Txid:       tx.Hash().String(),
		Hash:       tx.Tx.TxHashFull().String(),
		Version:    tx.Tx.Version,
		LockTime:   tx.Tx.LockTime,
		Time:       tx.Tx.Timestamp.Format(time.RFC3339),
		Vin:        marshal.MarshJsonVin(tx.Tx),
		Vout:       marshal.MarshJsonVout(tx.Tx, nil, params.ActiveNetParams.Params),
	}
	verboseNtfn := cmds.NewTxAcceptedVerboseNtfn(*rawTx)
	marshalledJSONVerbose, err = cmds.MarshalCmd(nil, verboseNtfn)
	if err != nil {
		log.Error(fmt.Sprintf("Failed to marshal verbose tx "+
			"notification: %s", err.Error()))
		return
	}
	wsc.QueueNotification(marshalledJSONVerbose)
}

func newWsNotificationManager(server *RpcServer) *wsNotificationManager {
	return &wsNotificationManager{
		server:            server,
		queueNotification: make(chan interface{}),
		notificationMsgs:  make(chan interface{}),
		numClients:        make(chan int),
		quit:              make(chan struct{}),
	}
}

func queueHandler(in <-chan interface{}, out chan<- interface{}, quit <-chan struct{}) {
	var q []interface{}
	var dequeue chan<- interface{}
	skipQueue := out
	var next interface{}
out:
	for {
		select {
		case n, ok := <-in:
			if !ok {
				// Sender closed input channel.
				break out
			}

			// Either send to out immediately if skipQueue is
			// non-nil (queue is empty) and reader is ready,
			// or append to the queue and send later.
			select {
			case skipQueue <- n:
			default:
				q = append(q, n)
				dequeue = out
				skipQueue = nil
				next = q[0]
			}

		case dequeue <- next:
			copy(q, q[1:])
			q[len(q)-1] = nil // avoid leak
			q = q[:len(q)-1]
			if len(q) == 0 {
				dequeue = nil
				skipQueue = out
			} else {
				next = q[0]
			}

		case <-quit:
			break out
		}
	}
	close(out)
}
