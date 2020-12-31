package rpc

import (
	"fmt"
	"github.com/Qitmeer/qitmeer/core/types"
	"github.com/Qitmeer/qitmeer/rpc/client/cmds"
	"sync"
)

// Notification types
type notificationBlockConnected types.SerializedBlock
type notificationBlockDisconnected types.SerializedBlock

// Notification control requests
type notificationRegisterClient wsClient
type notificationUnregisterClient wsClient
type notificationRegisterBlocks wsClient
type notificationUnregisterBlocks wsClient

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

			case *notificationRegisterBlocks:
				wsc := (*wsClient)(n)
				blockNotifications[wsc.quit] = wsc

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

			default:
				log.Warn("Unhandled notification type")
			}

		case m.numClients <- len(clients):

		case <-m.quit:
			// RPC server shutting down.
			break out
		}
	}

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
	ntfn := cmds.NewBlockConnectedNtfn(block.Hash().String(), int64(block.Order()), block.Block().Header.Timestamp.Unix(), txs)
	marshalledJSON, err := cmds.MarshalCmd(nil, ntfn)
	if err != nil {
		log.Error(fmt.Sprintf("Failed to marshal block connected notification: "+
			"%v", err))
		return
	}
	for _, wsc := range clients {
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
		int64(block.Order()), block.Block().Header.Timestamp.Unix(), txs)
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

func (m *wsNotificationManager) UnregisterBlockUpdates(wsc *wsClient) {
	m.queueNotification <- (*notificationUnregisterBlocks)(wsc)
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
