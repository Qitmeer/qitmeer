package rpc

import (
	"github.com/Qitmeer/qitmeer/core/types"
	"github.com/btcsuite/btcd/btcjson"
	"github.com/btcsuite/btcutil"
	"sync"
)

// Notification types
type notificationBlockConnected types.Block
type notificationBlockDisconnected types.Block

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
				block := (*types.Block)(n)
				if len(blockNotifications) != 0 {
					m.notifyBlockConnected(blockNotifications,
						block)
				}

			case *notificationBlockDisconnected:
				block := (*types.Block)(n)
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

func (*wsNotificationManager) notifyBlockConnected(clients map[chan struct{}]*wsClient, block *types.Block) {
}

func (*wsNotificationManager) notifyBlockDisconnected(clients map[chan struct{}]*wsClient, block *types.Block) {
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
