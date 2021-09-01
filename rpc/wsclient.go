package rpc

import (
	"container/list"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/Qitmeer/qitmeer/core/serialization"
	"github.com/Qitmeer/qitmeer/rpc/client/cmds"
	"github.com/Qitmeer/qitmeer/rpc/websocket"
	"io"
	"sync"
)

const (
	// websocketSendBufferSize is the number of elements the send channel
	// can queue before blocking.  Note that this only applies to requests
	// handled directly in the websocket client input handler or the async
	// handler since notifications have their own queuing mechanism
	// independent of the send channel buffer.
	websocketSendBufferSize = 50
)

var ErrClientQuit = errors.New("client quit")

type wsClient struct {
	sync.Mutex

	// server is the RPC server that is servicing the client.
	server *RpcServer

	// conn is the underlying websocket connection.
	conn *websocket.Conn

	// disconnected indicated whether or not the websocket client is
	// disconnected.
	disconnected bool

	// addr is the remote address of the client.
	addr string

	// authenticated specifies whether a client has been authenticated
	// and therefore is allowed to communicated over the websocket.
	authenticated bool

	// isAdmin specifies whether a client may change the state of the server;
	// false means its access is only to the limited set of RPC calls.
	isAdmin bool

	// sessionID is a random ID generated for each client when connected.
	// These IDs may be queried by a client using the session RPC.  A change
	// to the session ID indicates that the client reconnected.
	sessionID uint64

	// verboseTxUpdates specifies whether a client has requested verbose
	// information about all new transactions.
	verboseTxUpdates bool

	// filterData is the new generation transaction filter backported from
	// github.com/decred/dcrd for the new backported `loadtxfilter` and
	// `rescanblocks` methods.
	filterData *wsClientFilter

	// Networking infrastructure.
	serviceRequestSem semaphore
	ntfnChan          chan []byte
	sendChan          chan wsResponse
	quit              chan struct{}
	wg                sync.WaitGroup
	TxConfirms        *WatchTxConfirmServer
	TxConfirmsLock    sync.Mutex
}

func (c *wsClient) Start() {
	log.Trace(fmt.Sprintf("Starting websocket client %s", c.addr))

	// Start processing input and output.
	c.wg.Add(3)
	go c.inHandler()
	go c.notificationQueueHandler()
	go c.outHandler()
}

func (c *wsClient) Disconnect() {
	c.Lock()
	defer c.Unlock()

	// Nothing to do if already disconnected.
	if c.disconnected {
		return
	}

	log.Trace(fmt.Sprintf("Disconnecting websocket client %s", c.addr))
	close(c.quit)
	c.conn.Close()
	c.disconnected = true
}

func (c *wsClient) Disconnected() bool {
	c.Lock()
	isDisconnected := c.disconnected
	c.Unlock()

	return isDisconnected
}

func (c *wsClient) WaitForShutdown() {
	c.wg.Wait()
}

func (c *wsClient) inHandler() {
out:
	for {
		// Break out of the loop once the quit channel has been closed.
		// Use a non-blocking select here so we fall through otherwise.
		select {
		case <-c.quit:
			break out
		default:
		}
		_, msg, err := c.conn.ReadMessage()
		if err != nil {
			// Log the error if it's not due to disconnecting.
			if err != io.EOF && !c.Disconnected() {
				log.Error(fmt.Sprintf("Websocket receive error from "+
					"%s: %v", c.addr, err))
			}
			break out
		}
		success, exit := c.wsServiceRequest(msg)
		if exit {
			break out
		}
		if success {
			continue
		}
		codec := NewWSCodec(msg, c)
		c.serviceRequestSem.acquire()
		go func() {
			defer codec.Close()
			ctx := context.Background()
			c.server.ServeSingleRequest(ctx, codec, OptionMethodInvocation)

			c.serviceRequestSem.release()
		}()
	}

	// Ensure the connection is closed.
	c.Disconnect()
	c.wg.Done()
	log.Trace(fmt.Sprintf("Websocket client input handler done for %s", c.addr))
}

func (c *wsClient) wsServiceRequest(msg []byte) (bool, bool) {
	var request cmds.Request
	err := json.Unmarshal(msg, &request)
	if err != nil {
		if !c.isAdmin {
			return false, true
		}
		return false, false
	}
	cmd := parseCmd(&request)
	if cmd.err != nil {
		if !c.isAdmin {
			return false, true
		}
		return false, false
	}
	log.Debug(fmt.Sprintf("Received command <%s> from %s", cmd.method, c.addr))

	var result interface{}

	// Lookup the websocket extension for the command and if it doesn't
	// exist fallback to handling the command as a standard command.
	wsHandler, ok := wsHandlers[cmd.method]
	if ok {
		result, err = wsHandler(c, cmd.cmd)
	} else {
		return false, false
	}
	reply, err := createMarshalledReply(cmd.id, result, err)
	if err != nil {
		return false, false
	}

	c.serviceRequestSem.acquire()
	go func() {
		c.SendMessage(reply, nil)
		c.serviceRequestSem.release()
	}()

	return true, false
}

func (c *wsClient) SendMessage(marshalledJSON []byte, doneChan chan bool) {
	// Don't send the message if disconnected.
	if c.Disconnected() {
		if doneChan != nil {
			doneChan <- false
		}
		return
	}

	c.sendChan <- wsResponse{msg: marshalledJSON, doneChan: doneChan}
}

func (c *wsClient) notificationQueueHandler() {
	ntfnSentChan := make(chan bool, 1) // nonblocking sync

	// pendingNtfns is used as a queue for notifications that are ready to
	// be sent once there are no outstanding notifications currently being
	// sent.  The waiting flag is used over simply checking for items in the
	// pending list to ensure cleanup knows what has and hasn't been sent
	// to the outHandler.  Currently no special cleanup is needed, however
	// if something like a done channel is added to notifications in the
	// future, not knowing what has and hasn't been sent to the outHandler
	// (and thus who should respond to the done channel) would be
	// problematic without using this approach.
	pendingNtfns := list.New()
	waiting := false
out:
	for {
		select {
		// This channel is notified when a message is being queued to
		// be sent across the network socket.  It will either send the
		// message immediately if a send is not already in progress, or
		// queue the message to be sent once the other pending messages
		// are sent.
		case msg := <-c.ntfnChan:
			if !waiting {
				c.SendMessage(msg, ntfnSentChan)
			} else {
				pendingNtfns.PushBack(msg)
			}
			waiting = true

		// This channel is notified when a notification has been sent
		// across the network socket.
		case <-ntfnSentChan:
			// No longer waiting if there are no more messages in
			// the pending messages queue.
			next := pendingNtfns.Front()
			if next == nil {
				waiting = false
				continue
			}

			// Notify the outHandler about the next item to
			// asynchronously send.
			msg := pendingNtfns.Remove(next).([]byte)
			c.SendMessage(msg, ntfnSentChan)

		case <-c.quit:
			break out
		}
	}

	// Drain any wait channels before exiting so nothing is left waiting
	// around to send.
cleanup:
	for {
		select {
		case <-c.ntfnChan:
		case <-ntfnSentChan:
		default:
			break cleanup
		}
	}
	c.wg.Done()
	log.Trace(fmt.Sprintf("Websocket client notification queue handler done "+
		"for %s", c.addr))
}

func (c *wsClient) outHandler() {
out:
	for {
		// Send any messages ready for send until the quit channel is
		// closed.
		select {
		case r := <-c.sendChan:
			err := c.conn.WriteMessage(websocket.TextMessage, r.msg)
			if err != nil {
				c.Disconnect()
				break out
			}
			if r.doneChan != nil {
				r.doneChan <- true
			}

		case <-c.quit:
			break out
		}
	}

	// Drain any wait channels before exiting so nothing is left waiting
	// around to send.
cleanup:
	for {
		select {
		case r := <-c.sendChan:
			if r.doneChan != nil {
				r.doneChan <- false
			}
		default:
			break cleanup
		}
	}
	c.wg.Done()
	log.Trace(fmt.Sprintf("Websocket client output handler done for %s", c.addr))
}

func (c *wsClient) QueueNotification(marshalledJSON []byte) error {
	// Don't queue the message if disconnected.
	if c.Disconnected() {
		return ErrClientQuit
	}
	c.ntfnChan <- marshalledJSON
	return nil
}

func newWebsocketClient(server *RpcServer, conn *websocket.Conn,
	remoteAddr string, isAdmin bool) (*wsClient, error) {

	sessionID, err := serialization.RandomUint64()
	if err != nil {
		return nil, err
	}

	client := &wsClient{
		conn:              conn,
		addr:              remoteAddr,
		isAdmin:           isAdmin,
		sessionID:         sessionID,
		server:            server,
		serviceRequestSem: makeSemaphore(server.config.RPCMaxConcurrentReqs),
		ntfnChan:          make(chan []byte, 1), // nonblocking sync
		sendChan:          make(chan wsResponse, websocketSendBufferSize),
		quit:              make(chan struct{}),
		TxConfirms:        &WatchTxConfirmServer{},
	}
	return client, nil
}

type wsResponse struct {
	msg      []byte
	doneChan chan bool
}
