/*
 * Copyright (c) 2017-2020 The qitmeer developers
 */

package client

import (
	"bytes"
	"container/list"
	"encoding/json"
	"fmt"
	"github.com/Qitmeer/qitmeer/params"
	"github.com/Qitmeer/qitmeer/rpc/client/cmds"
	"github.com/Qitmeer/qitmeer/rpc/websocket"
	"io"
	"io/ioutil"
	"math"
	"net"
	"net/http"
	"sync"
	"sync/atomic"
	"time"
)

type Client struct {
	id uint64 // atomic, so must stay 64-bit aligned

	// config holds the connection configuration assoiated with this client.
	config *ConnConfig

	// chainParams holds the params for the chain that this client is using,
	// and is used for many wallet methods.
	chainParams *params.Params

	// wsConn is the underlying websocket connection when not in HTTP POST
	// mode.
	wsConn *websocket.Conn

	// httpClient is the underlying HTTP client to use when running in HTTP
	// POST mode.
	httpClient *http.Client

	// mtx is a mutex to protect access to connection related fields.
	mtx sync.Mutex

	// disconnected indicated whether or not the server is disconnected.
	disconnected bool

	// retryCount holds the number of times the client has tried to
	// reconnect to the RPC server.
	retryCount int64

	// Track command and their response channels by ID.
	requestLock sync.Mutex
	requestMap  map[uint64]*list.Element
	requestList *list.List

	// Notifications.
	ntfnHandlers  *NotificationHandlers
	ntfnStateLock sync.Mutex
	ntfnState     *notificationState

	// Networking infrastructure.
	sendChan        chan []byte
	sendPostChan    chan *sendPostDetails
	connEstablished chan struct{}
	disconnect      chan struct{}
	shutdown        chan struct{}
	wg              sync.WaitGroup
}

func (c *Client) start() {
	log.Trace(fmt.Sprintf("Starting RPC client %s", c.config.Host))

	// Start the I/O processing handlers depending on whether the client is
	// in HTTP POST mode or the default websocket mode.
	if c.config.HTTPPostMode {
		c.wg.Add(1)
		go c.sendPostHandler()
	} else {
		c.wg.Add(3)
		go func() {
			if c.ntfnHandlers != nil {
				if c.ntfnHandlers.OnClientConnected != nil {
					c.ntfnHandlers.OnClientConnected()
				}
			}
			c.wg.Done()
		}()
		go c.wsInHandler()
		go c.wsOutHandler()
	}
}

func (c *Client) sendPostHandler() {
out:
	for {
		// Send any messages ready for send until the shutdown channel
		// is closed.
		select {
		case details := <-c.sendPostChan:
			c.handleSendPostMessage(details)

		case <-c.shutdown:
			break out
		}
	}

	// Drain any wait channels before exiting so nothing is left waiting
	// around to send.
cleanup:
	for {
		select {
		case details := <-c.sendPostChan:
			details.jsonRequest.responseChan <- &response{
				result: nil,
				err:    ErrClientShutdown,
			}

		default:
			break cleanup
		}
	}
	c.wg.Done()
	log.Trace(fmt.Sprintf("RPC client send handler done for %s", c.config.Host))

}

func (c *Client) handleSendPostMessage(details *sendPostDetails) {
	jReq := details.jsonRequest
	log.Trace(fmt.Sprintf("Sending command [%s] with id %d", jReq.method, jReq.id))
	httpResponse, err := c.httpClient.Do(details.httpRequest)
	if err != nil {
		jReq.responseChan <- &response{err: err}
		return
	}

	// Read the raw bytes and close the response.
	respBytes, err := ioutil.ReadAll(httpResponse.Body)
	httpResponse.Body.Close()
	if err != nil {
		err = fmt.Errorf("error reading json reply: %v", err)
		jReq.responseChan <- &response{err: err}
		return
	}

	// Try to unmarshal the response as a regular JSON-RPC response.
	var resp rawResponse
	err = json.Unmarshal(respBytes, &resp)
	if err != nil {
		// When the response itself isn't a valid JSON-RPC response
		// return an error which includes the HTTP status code and raw
		// response bytes.
		err = fmt.Errorf("status code: %d, response: %q",
			httpResponse.StatusCode, string(respBytes))
		jReq.responseChan <- &response{err: err}
		return
	}

	res, err := resp.result()
	jReq.responseChan <- &response{result: res, err: err}
}

func (c *Client) wsInHandler() {
out:
	for {
		// Break out of the loop once the shutdown channel has been
		// closed.  Use a non-blocking select here so we fall through
		// otherwise.
		select {
		case <-c.shutdown:
			break out
		default:
		}

		_, msg, err := c.wsConn.ReadMessage()
		if err != nil {
			// Log the error if it's not due to disconnecting.
			if c.shouldLogReadError(err) {
				log.Error(fmt.Sprintf("Websocket receive error from "+
					"%s: %v", c.config.Host, err))
			}
			break out
		}
		c.handleMessage(msg)
	}

	// Ensure the connection is closed.
	c.Disconnect()
	c.wg.Done()
	log.Trace(fmt.Sprintf("RPC client input handler done for %s", c.config.Host))
}

func (c *Client) handleMessage(msg []byte) {
	// Attempt to unmarshal the message as either a notification or
	// response.
	var in inMessage
	in.rawResponse = new(rawResponse)
	in.rawNotification = new(rawNotification)
	err := json.Unmarshal(msg, &in)
	if err != nil {
		log.Warn("Remote server sent invalid message: %v", err)
		return
	}

	// JSON-RPC 1.0 notifications are requests with a null id.
	if in.ID == nil {
		ntfn := in.rawNotification
		if ntfn == nil {
			log.Warn("Malformed notification: missing " +
				"method and parameters")
			return
		}
		if ntfn.Method == "" {
			log.Warn("Malformed notification: missing method")
			return
		}
		// params are not optional: nil isn't valid (but len == 0 is)
		if ntfn.Params == nil {
			log.Warn("Malformed notification: missing params")
			return
		}
		// Deliver the notification.
		log.Trace(fmt.Sprintf("Received notification [%s]", in.Method))
		c.handleNotification(in.rawNotification)
		return
	}

	// ensure that in.ID can be converted to an integer without loss of precision
	if *in.ID < 0 || *in.ID != math.Trunc(*in.ID) {
		log.Warn("Malformed response: invalid identifier")
		return
	}

	if in.rawResponse == nil {
		log.Warn("Malformed response: missing result and error")
		return
	}

	id := uint64(*in.ID)
	log.Trace(fmt.Sprintf("Received response for id %d (result %s)", id, in.Result))
	request := c.removeRequest(id)

	// Nothing more to do if there is no request associated with this reply.
	if request == nil || request.responseChan == nil {
		log.Warn(fmt.Sprintf("Received unexpected reply: %s (id %d)", in.Result,
			id))
		return
	}

	// Since the command was successful, examine it to see if it's a
	// notification, and if is, add it to the notification state so it
	// can automatically be re-established on reconnect.
	c.trackRegisteredNtfns(request.cmd)

	// Deliver the response.
	result, err := in.rawResponse.result()
	request.responseChan <- &response{result: result, err: err}
}

func (c *Client) handleNotification(ntfn *rawNotification) {
	// Ignore the notification if the client is not interested in any
	// notifications.
	if c.ntfnHandlers == nil {
		return
	}
	switch ntfn.Method {
	// OnBlockConnected
	case cmds.BlockConnectedNtfnMethod:
		// Ignore the notification if the client is not interested in
		// it.
		if c.ntfnHandlers.OnBlockConnected == nil {
			return
		}
		blockHash, height, blockOrder, blockTime, txs, err := parseChainNtfnParams(ntfn.Params)
		if err != nil {
			log.Warn(fmt.Sprintf("Received invalid block connected "+
				"notification: %v", err))
			return
		}
		c.ntfnHandlers.OnBlockConnected(blockHash, height, blockOrder, blockTime, txs)

	// OnBlockDisconnected
	case cmds.BlockDisconnectedNtfnMethod:
		// Ignore the notification if the client is not interested in
		// it.
		if c.ntfnHandlers.OnBlockDisconnected == nil {
			return
		}

		blockHash, height, blockOrder, blockTime, txs, err := parseChainNtfnParams(ntfn.Params)
		if err != nil {
			log.Warn(fmt.Sprintf("Received invalid block connected "+
				"notification: %v", err))
			return
		}

		c.ntfnHandlers.OnBlockDisconnected(blockHash, height, blockOrder, blockTime, txs)

	case cmds.BlockAcceptedNtfnMethod:
		// Ignore the notification if the client is not interested in
		// it.
		if c.ntfnHandlers.OnBlockAccepted == nil {
			return
		}

		blockHash, height, blockOrder, blockTime, txs, err := parseChainNtfnParams(ntfn.Params)
		if err != nil {
			log.Warn(fmt.Sprintf("Received invalid block accepted "+
				"notification: %v", err))
			return
		}

		c.ntfnHandlers.OnBlockAccepted(blockHash, height, blockOrder, blockTime, txs)

	case cmds.ReorganizationNtfnMethod:
		// Ignore the notification if the client is not interested in
		// it.
		if c.ntfnHandlers.OnReorganization == nil {
			return
		}

		blockHash, blockOrder, olds, err := parseReorganizationNtfnParams(ntfn.Params)
		if err != nil {
			log.Warn(fmt.Sprintf("Received invalid block reorganization "+
				"notification: %v", err))
			return
		}

		c.ntfnHandlers.OnReorganization(blockHash, blockOrder, olds)

		// OnTxAccepted
	case cmds.TxAcceptedNtfnMethod:
		// Ignore the notification if the client is not interested in
		// it.
		if c.ntfnHandlers.OnTxAccepted == nil {
			return
		}

		hash, amt, err := parseTxAcceptedNtfnParams(ntfn.Params)
		if err != nil {
			log.Warn(fmt.Sprintf("Received invalid tx accepted "+
				"notification: %v", err))
			return
		}

		c.ntfnHandlers.OnTxAccepted(hash, amt)

	// OnTxAcceptedVerbose
	case cmds.TxAcceptedVerboseNtfnMethod:
		// Ignore the notification if the client is not interested in
		// it.
		if c.ntfnHandlers.OnTxAcceptedVerbose == nil {
			return
		}

		rawTx, err := parseTxAcceptedVerboseNtfnParams(ntfn.Params)
		if err != nil {
			log.Warn(fmt.Sprintf("Received invalid tx accepted verbose "+
				"notification: %v", err))
			return
		}

		c.ntfnHandlers.OnTxAcceptedVerbose(c, rawTx)

	// OnTxConfirm
	case cmds.TxConfirmNtfnMethod:
		// Ignore the notification if the client is not interested in
		// it.
		if c.ntfnHandlers.OnTxConfirm == nil {
			return
		}
		rawTx, err := parseTxConfirm(ntfn.Params)
		if err != nil {
			log.Warn(fmt.Sprintf("Received invalid tx confirm struct "+
				"notification: %v", err))
			return
		}
		c.ntfnHandlers.OnTxConfirm(rawTx)

	// OnRescanProgress
	case cmds.RescanProgressNtfnMethod:
		// Ignore the notification if the client is not interested in
		// it.
		if c.ntfnHandlers.OnRescanProgress == nil {
			return
		}

		rawTx, err := parseRescanProgress(ntfn.Params)
		if err != nil {
			log.Warn(fmt.Sprintf("Received invalid tx accepted verbose "+
				"notification: %v", err))
			return
		}

		c.ntfnHandlers.OnRescanProgress(rawTx)

	// OnRescanFinish
	case cmds.RescanCompleteNtfnMethod:
		// Ignore the notification if the client is not interested in
		// it.
		if c.ntfnHandlers.OnRescanFinish == nil {
			return
		}

		rawTx, err := parseRescanFinish(ntfn.Params)
		if err != nil {
			log.Warn(fmt.Sprintf("Received invalid tx accepted verbose "+
				"notification: %v", err))
			return
		}

		c.ntfnHandlers.OnRescanFinish(rawTx)

	// OnNodeExit
	case cmds.NodeExitMethod:
		// Ignore the notification if the client is not interested in
		// it.
		if c.ntfnHandlers.OnNodeExit == nil {
			return
		}

		c.ntfnHandlers.OnNodeExit(&cmds.NodeExitNtfn{})

	// OnUnknownNotification
	default:
		if c.ntfnHandlers.OnUnknownNotification == nil {
			return
		}

		c.ntfnHandlers.OnUnknownNotification(ntfn.Method, ntfn.Params)
	}
}

func (c *Client) shouldLogReadError(err error) bool {
	// No logging when the connetion is being forcibly disconnected.
	select {
	case <-c.shutdown:
		return false
	default:
	}

	// No logging when the connection has been disconnected.
	if err == io.EOF {
		return false
	}
	if opErr, ok := err.(*net.OpError); ok && !opErr.Temporary() {
		return false
	}

	return true
}

func (c *Client) removeRequest(id uint64) *jsonRequest {
	c.requestLock.Lock()
	defer c.requestLock.Unlock()

	element := c.requestMap[id]
	if element != nil {
		delete(c.requestMap, id)
		request := c.requestList.Remove(element).(*jsonRequest)
		return request
	}

	return nil
}

func (c *Client) trackRegisteredNtfns(cmd interface{}) {
	// Nothing to do if the caller is not interested in notifications.
	if c.ntfnHandlers == nil {
		return
	}

	c.ntfnStateLock.Lock()
	defer c.ntfnStateLock.Unlock()

	switch bcmd := cmd.(type) {
	case *cmds.NotifyBlocksCmd:
		c.ntfnState.notifyBlocks = true

	case *cmds.NotifyReceivedCmd:
		for _, addr := range bcmd.Addresses {
			c.ntfnState.notifyReceived[addr] = struct{}{}
		}
	}
}

func (c *Client) sendCmd(cmd interface{}) chan *response {
	// Get the method associated with the command.
	method, err := cmds.CmdMethod(cmd)
	if err != nil {
		return newFutureError(err)
	}

	// Marshal the command.
	id := c.NextID()
	marshalledJSON, err := cmds.MarshalCmd(id, cmd)
	if err != nil {
		return newFutureError(err)
	}

	// Generate the request and send it along with a channel to respond on.
	responseChan := make(chan *response, 1)
	jReq := &jsonRequest{
		id:             id,
		method:         method,
		cmd:            cmd,
		marshalledJSON: marshalledJSON,
		responseChan:   responseChan,
	}
	c.sendRequest(jReq)

	return responseChan
}

func (c *Client) NextID() uint64 {
	return atomic.AddUint64(&c.id, 1)
}

func (c *Client) sendRequest(jReq *jsonRequest) {
	// Choose which marshal and send function to use depending on whether
	// the client running in HTTP POST mode or not.  When running in HTTP
	// POST mode, the command is issued via an HTTP client.  Otherwise,
	// the command is issued via the asynchronous websocket channels.
	if c.config.HTTPPostMode {
		c.sendPost(jReq)
		return
	}

	// Check whether the websocket connection has never been established,
	// in which case the handler goroutines are not running.
	select {
	case <-c.connEstablished:
	default:
		jReq.responseChan <- &response{err: ErrClientNotConnected}
		return
	}

	// Add the request to the internal tracking map so the response from the
	// remote server can be properly detected and routed to the response
	// channel.  Then send the marshalled request via the websocket
	// connection.
	if err := c.addRequest(jReq); err != nil {
		jReq.responseChan <- &response{err: err}
		return
	}
	log.Trace("Sending command [%s] with id %d", jReq.method, jReq.id)
	c.sendMessage(jReq.marshalledJSON)
}

func (c *Client) sendPost(jReq *jsonRequest) {
	// Generate a request to the configured RPC server.
	protocol := "http"
	if !c.config.DisableTLS {
		protocol = "https"
	}
	url := protocol + "://" + c.config.Host
	bodyReader := bytes.NewReader(jReq.marshalledJSON)
	httpReq, err := http.NewRequest("POST", url, bodyReader)
	if err != nil {
		jReq.responseChan <- &response{result: nil, err: err}
		return
	}
	httpReq.Close = true
	httpReq.Header.Set("Content-Type", "application/json")
	for key, value := range c.config.ExtraHeaders {
		httpReq.Header.Set(key, value)
	}

	// Configure basic access authorization.
	user, pass, err := c.config.getAuth()
	if err != nil {
		jReq.responseChan <- &response{result: nil, err: err}
		return
	}
	httpReq.SetBasicAuth(user, pass)

	log.Trace(fmt.Sprintf("Sending command [%s] with id %d", jReq.method, jReq.id))
	c.sendPostRequest(httpReq, jReq)
}

func (c *Client) sendPostRequest(httpReq *http.Request, jReq *jsonRequest) {
	// Don't send the message if shutting down.
	select {
	case <-c.shutdown:
		jReq.responseChan <- &response{result: nil, err: ErrClientShutdown}
	default:
	}

	c.sendPostChan <- &sendPostDetails{
		jsonRequest: jReq,
		httpRequest: httpReq,
	}
}

func (c *Client) addRequest(jReq *jsonRequest) error {
	c.requestLock.Lock()
	defer c.requestLock.Unlock()

	// A non-blocking read of the shutdown channel with the request lock
	// held avoids adding the request to the client's internal data
	// structures if the client is in the process of shutting down (and
	// has not yet grabbed the request lock), or has finished shutdown
	// already (responding to each outstanding request with
	// ErrClientShutdown).
	select {
	case <-c.shutdown:
		return ErrClientShutdown
	default:
	}

	element := c.requestList.PushBack(jReq)
	c.requestMap[jReq.id] = element
	return nil
}

func (c *Client) sendMessage(marshalledJSON []byte) {
	// Don't send the message if disconnected.
	select {
	case c.sendChan <- marshalledJSON:
	case <-c.disconnectChan():
		return
	}
}

func (c *Client) disconnectChan() <-chan struct{} {
	c.mtx.Lock()
	ch := c.disconnect
	c.mtx.Unlock()
	return ch
}

func (c *Client) Disconnected() bool {
	c.mtx.Lock()
	defer c.mtx.Unlock()

	select {
	case <-c.connEstablished:
		return c.disconnected
	default:
		return false
	}
}

func (c *Client) Disconnect() {
	// Nothing to do if already disconnected or running in HTTP POST mode.
	if !c.doDisconnect() {
		return
	}

	c.requestLock.Lock()
	defer c.requestLock.Unlock()

	// When operating without auto reconnect, send errors to any pending
	// requests and shutdown the client.
	if c.config.DisableAutoReconnect {
		for e := c.requestList.Front(); e != nil; e = e.Next() {
			req := e.Value.(*jsonRequest)
			req.responseChan <- &response{
				result: nil,
				err:    ErrClientDisconnect,
			}
		}
		c.removeAllRequests()
		c.doShutdown()
	}
}

func (c *Client) doDisconnect() bool {
	if c.config.HTTPPostMode {
		return false
	}

	c.mtx.Lock()
	defer c.mtx.Unlock()

	// Nothing to do if already disconnected.
	if c.disconnected {
		return false
	}

	log.Trace(fmt.Sprintf("Disconnecting RPC client %s", c.config.Host))
	close(c.disconnect)
	if c.wsConn != nil {
		c.wsConn.Close()
	}
	c.disconnected = true
	return true
}

func (c *Client) Shutdown() {
	// Do the shutdown under the request lock to prevent clients from
	// adding new requests while the client shutdown process is initiated.
	c.requestLock.Lock()
	defer c.requestLock.Unlock()

	// Ignore the shutdown request if the client is already in the process
	// of shutting down or already shutdown.
	if !c.doShutdown() {
		return
	}

	// Send the ErrClientShutdown error to any pending requests.
	for e := c.requestList.Front(); e != nil; e = e.Next() {
		req := e.Value.(*jsonRequest)
		req.responseChan <- &response{
			result: nil,
			err:    ErrClientShutdown,
		}
	}
	c.removeAllRequests()

	// Disconnect the client if needed.
	c.doDisconnect()
}

func (c *Client) doShutdown() bool {
	// Ignore the shutdown request if the client is already in the process
	// of shutting down or already shutdown.
	select {
	case <-c.shutdown:
		return false
	default:
	}

	log.Trace(fmt.Sprintf("Shutting down RPC client %s", c.config.Host))
	close(c.shutdown)
	return true
}

func (c *Client) WaitForShutdown() {
	c.wg.Wait()
}

func (c *Client) removeAllRequests() {
	c.requestMap = make(map[uint64]*list.Element)
	c.requestList.Init()
}

func (c *Client) wsOutHandler() {
out:
	for {
		// Send any messages ready for send until the client is
		// disconnected closed.
		select {
		case msg := <-c.sendChan:
			err := c.wsConn.WriteMessage(websocket.TextMessage, msg)
			if err != nil {
				c.Disconnect()
				break out
			}

		case <-c.disconnectChan():
			break out
		}
	}

	// Drain any channels before exiting so nothing is left waiting around
	// to send.
cleanup:
	for {
		select {
		case <-c.sendChan:
		default:
			break cleanup
		}
	}
	c.wg.Done()
	log.Trace(fmt.Sprintf("RPC client output handler done for %s", c.config.Host))
}

func (c *Client) wsReconnectHandler() {
out:
	for {
		select {
		case <-c.disconnect:
			// On disconnect, fallthrough to reestablish the
			// connection.

		case <-c.shutdown:
			break out
		}

	reconnect:
		for {
			select {
			case <-c.shutdown:
				break out
			default:
			}

			wsConn, err := dial(c.config)
			if err != nil {
				c.retryCount++
				log.Info(fmt.Sprintf("Failed to connect to %s: %v",
					c.config.Host, err))

				// Scale the retry interval by the number of
				// retries so there is a backoff up to a max
				// of 1 minute.
				scaledInterval := connectionRetryInterval.Nanoseconds() * c.retryCount
				scaledDuration := time.Duration(scaledInterval)
				if scaledDuration > time.Minute {
					scaledDuration = time.Minute
				}
				log.Info(fmt.Sprintf("Retrying connection to %s in "+
					"%s", c.config.Host, scaledDuration))
				time.Sleep(scaledDuration)
				continue reconnect
			}

			log.Info(fmt.Sprintf("Reestablished connection to RPC server %s",
				c.config.Host))

			// Reset the connection state and signal the reconnect
			// has happened.
			c.wsConn = wsConn
			c.retryCount = 0

			c.mtx.Lock()
			c.disconnect = make(chan struct{})
			c.disconnected = false
			c.mtx.Unlock()

			// Start processing input and output for the
			// new connection.
			c.start()

			// Reissue pending requests in another goroutine since
			// the send can block.
			go c.resendRequests()

			// Break out of the reconnect loop back to wait for
			// disconnect again.
			break reconnect
		}
	}
	c.wg.Done()
	log.Trace(fmt.Sprintf("RPC client reconnect handler done for %s", c.config.Host))
}

var ignoreResends = map[string]struct{}{
	"rescan": {},
}

func (c *Client) resendRequests() {
	// Set the notification state back up.  If anything goes wrong,
	// disconnect the client.
	if err := c.reregisterNtfns(); err != nil {
		log.Warn(fmt.Sprintf("Unable to re-establish notification state: %v", err))
		c.Disconnect()
		return
	}

	// Since it's possible to block on send and more requests might be
	// added by the caller while resending, make a copy of all of the
	// requests that need to be resent now and work from the copy.  This
	// also allows the lock to be released quickly.
	c.requestLock.Lock()
	resendReqs := make([]*jsonRequest, 0, c.requestList.Len())
	var nextElem *list.Element
	for e := c.requestList.Front(); e != nil; e = nextElem {
		nextElem = e.Next()

		jReq := e.Value.(*jsonRequest)
		if _, ok := ignoreResends[jReq.method]; ok {
			// If a request is not sent on reconnect, remove it
			// from the request structures, since no reply is
			// expected.
			delete(c.requestMap, jReq.id)
			c.requestList.Remove(e)
		} else {
			resendReqs = append(resendReqs, jReq)
		}
	}
	c.requestLock.Unlock()

	for _, jReq := range resendReqs {
		// Stop resending commands if the client disconnected again
		// since the next reconnect will handle them.
		if c.Disconnected() {
			return
		}

		log.Trace(fmt.Sprintf("Sending command [%s] with id %d", jReq.method,
			jReq.id))
		c.sendMessage(jReq.marshalledJSON)
	}
}

func (c *Client) reregisterNtfns() error {
	// Nothing to do if the caller is not interested in notifications.
	if c.ntfnHandlers == nil {
		return nil
	}

	// In order to avoid holding the lock on the notification state for the
	// entire time of the potentially long running RPCs issued below, make a
	// copy of it and work from that.
	//
	// Also, other commands will be running concurrently which could modify
	// the notification state (while not under the lock of course) which
	// also register it with the remote RPC server, so this prevents double
	// registrations.
	c.ntfnStateLock.Lock()
	stateCopy := c.ntfnState.Copy()
	c.ntfnStateLock.Unlock()

	// Reregister notifyblocks if needed.
	if stateCopy.notifyBlocks {
		log.Debug("Reregistering [notifyblocks]")
		if err := c.NotifyBlocks(); err != nil {
			return err
		}
	}
	if stateCopy.notifyNewTx || stateCopy.notifyNewTxVerbose {
		log.Debug(fmt.Sprintf("Reregistering [notifynewtransactions] (verbose=%v)",
			stateCopy.notifyNewTxVerbose))
		err := c.NotifyNewTransactions(stateCopy.notifyNewTxVerbose)
		if err != nil {
			return err
		}
	}
	// Reregister the combination of all previously registered
	// notifyreceived addresses in one command if needed.
	nrlen := len(stateCopy.notifyReceived)
	if nrlen > 0 {
		addresses := make([]string, 0, nrlen)
		for addr := range stateCopy.notifyReceived {
			addresses = append(addresses, addr)
		}
		log.Debug("Reregistering [notifyreceived] addresses: %v", addresses)
		if err := c.notifyReceivedInternal(addresses).Receive(); err != nil {
			return err
		}
	}

	return nil
}

func New(config *ConnConfig, ntfnHandlers *NotificationHandlers) (*Client, error) {
	// Either open a websocket connection or create an HTTP client depending
	// on the HTTP POST mode.  Also, set the notification handlers to nil
	// when running in HTTP POST mode.
	var wsConn *websocket.Conn
	var httpClient *http.Client
	connEstablished := make(chan struct{})
	var start bool
	if config.HTTPPostMode {
		ntfnHandlers = nil
		start = true

		var err error
		httpClient, err = newHTTPClient(config)
		if err != nil {
			return nil, err
		}
	} else {
		if !config.DisableConnectOnNew {
			var err error
			wsConn, err = dial(config)
			if err != nil {
				return nil, err
			}
			start = true
		}
	}

	client := &Client{
		config:          config,
		wsConn:          wsConn,
		httpClient:      httpClient,
		requestMap:      make(map[uint64]*list.Element),
		requestList:     list.New(),
		ntfnHandlers:    ntfnHandlers,
		ntfnState:       newNotificationState(),
		sendChan:        make(chan []byte, sendBufferSize),
		sendPostChan:    make(chan *sendPostDetails, sendPostBufferSize),
		connEstablished: connEstablished,
		disconnect:      make(chan struct{}),
		shutdown:        make(chan struct{}),
	}

	// Default network is mainnet, no parameters are necessary but if mainnet
	// is specified it will be the param
	switch config.Params {
	case "":
		fallthrough
	case params.MainNetParams.Name:
		client.chainParams = &params.MainNetParams
	case params.TestNetParams.Name:
		client.chainParams = &params.TestNetParams
	case params.PrivNetParams.Name:
		client.chainParams = &params.PrivNetParams
	case params.MixNetParams.Name:
		client.chainParams = &params.MixNetParams
	default:
		return nil, fmt.Errorf("rpcclient.New: Unknown chain %s", config.Params)
	}

	if start {
		log.Info(fmt.Sprintf("Established connection to RPC server %s",
			config.Host))
		close(connEstablished)
		client.start()
		if !client.config.HTTPPostMode && !client.config.DisableAutoReconnect {
			client.wg.Add(1)
			go client.wsReconnectHandler()
		}
	}

	return client, nil
}
