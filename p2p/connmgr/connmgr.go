// Copyright (c) 2017-2018 The qitmeer developers
// Copyright (c) 2016 The btcsuite developers
// Copyright (c) 2017 The Decred developers
// Use of this source code is governed by an ISC
// license that can be found in the LICENSE file.
package connmgr

import (
	"errors"
	"fmt"
	"github.com/Qitmeer/qitmeer/log"
	"net"
	"sync"
	"sync/atomic"
	"time"
)

// ConnManager provides a manager to handle network connections.
type ConnManager struct {
	// The following variables must only be used atomically.
	connReqCount uint64
	start        int32
	stop         int32

	cfg            Config
	wg             sync.WaitGroup
	failedAttempts uint64
	requests       chan interface{}
	quit           chan struct{}
}

// New returns a new connection manager.
// Use Start to start connecting to the network.
func New(cfg *Config) (*ConnManager, error) {
	if cfg.Dial == nil {
		return nil, ErrDialNil
	}
	// Default to sane values
	if cfg.RetryDuration <= 0 {
		cfg.RetryDuration = defaultRetryDuration
	}
	if cfg.TargetOutbound == 0 {
		cfg.TargetOutbound = defaultTargetOutbound
	}
	cm := ConnManager{
		cfg:      *cfg, // Copy so caller can't mutate
		requests: make(chan interface{}),
		quit:     make(chan struct{}),
	}
	return &cm, nil
}

// Start launches the connection manager and begins connecting to the network.
func (cm *ConnManager) Start() {
	// Already started?
	if atomic.AddInt32(&cm.start, 1) != 1 {
		return
	}
	log.Info("Connection manager started")
	cm.wg.Add(1)
	go cm.connHandler()

	// Start all the listeners so long as the caller requested them and
	// provided a callback to be invoked when connections are accepted.
	if cm.cfg.OnAccept != nil {
		for _, listner := range cm.cfg.Listeners {
			cm.wg.Add(1)
			go cm.listenHandler(listner)
		}
	}

	// try open connections : outbound (default, 8)
	for i := atomic.LoadUint64(&cm.connReqCount); i < uint64(cm.cfg.TargetOutbound); i++ {
		go cm.NewConnReq()
	}
}

// Stop gracefully shuts down the connection manager.
func (cm *ConnManager) Stop() {
	if atomic.AddInt32(&cm.stop, 1) != 1 {
		log.Warn("Connection manager already stopped")
		return
	}

	// Stop all the listeners.  There will not be any listeners if
	// listening is disabled.
	for _, listener := range cm.cfg.Listeners {
		// Ignore the error since this is shutdown and there is no way
		// to recover anyways.
		_ = listener.Close()
	}

	close(cm.quit)
	log.Info("Connection manager stopped")
}

// maxFailedAttempts is the maximum number of successive failed connection
// attempts after which network failure is assumed and new connections will
// be delayed by the configured retry duration.
const maxFailedAttempts = 25

var (
	// ErrDialNil is used to indicate that Dial cannot be nil in the configuration.
	ErrDialNil = errors.New("config: dial cannot be nil")

	// maxRetryDuration is the max duration of time retrying of a persistent
	// connection is allowed to grow to.  This is necessary since the retry
	// logic uses a backoff mechanism which increases the interval base times
	// the number of retries that have been done.
	maxRetryDuration = time.Minute * 5

	// defaultRetryDuration is the default duration of time for retrying
	// persistent connections.
	defaultRetryDuration = time.Second * 5

	// defaultTargetOutbound is the default number of outbound connections to
	// maintain.
	defaultTargetOutbound = uint32(8)
)

// ConnState represents the state of the requested connection.
type ConnState uint32

// ConnState can be either pending, established, disconnected or failed.  When
// a new connection is requested, it is attempted and categorized as
// established or failed depending on the connection result.  An established
// connection which was disconnected is categorized as disconnected.
const (
	ConnPending ConnState = iota
	ConnFailing
	ConnCanceled
	ConnEstablished
	ConnDisconnected
)

// ConnReq is the connection request to a network address. If permanent, the
// connection will be retried on disconnection.
type ConnReq struct {
	// The following variables must only be used atomically.
	id    uint64
	state uint32

	retryCount uint32
	conn       net.Conn
	Addr       net.Addr
	Permanent  bool
	// This connect is illegal
	Ban bool
}

// updateState updates the state of the connection request.
func (c *ConnReq) updateState(state ConnState) {
	atomic.StoreUint32(&c.state, uint32(state))
}

// ID returns a unique identifier for the connection request.
func (c *ConnReq) ID() uint64 {
	return atomic.LoadUint64(&c.id)
}

// State is the connection state of the requested connection.
func (c *ConnReq) State() ConnState {
	return ConnState(atomic.LoadUint32(&c.state))
}

// String returns a human-readable string for the connection request.
func (c *ConnReq) String() string {
	if c.Addr == nil || c.Addr.String() == "" {
		return fmt.Sprintf("reqid %d", atomic.LoadUint64(&c.id))
	}
	return fmt.Sprintf("%s (reqid %d)", c.Addr, atomic.LoadUint64(&c.id))
}

// conn
func (c *ConnReq) Conn() net.Conn {
	return c.conn
}

// set conn
func (c *ConnReq) SetConn(conn net.Conn) {
	c.conn = conn
}

// NewConnReq
func NewConnReq() *ConnReq {
	c := ConnReq{}
	c.id = 0
	c.Ban = false
	return &c
}

// Config holds the configuration options related to the connection manager.
type Config struct {
	// Listeners defines a slice of listeners for which the connection
	// manager will take ownership of and accept connections.  When a
	// connection is accepted, the OnAccept handler will be invoked with the
	// connection.  Since the connection manager takes ownership of these
	// listeners, they will be closed when the connection manager is
	// stopped.
	//
	// This field will not have any effect if the OnAccept field is not
	// also specified.  It may be nil if the caller does not wish to listen
	// for incoming connections.
	Listeners []net.Listener

	// OnAccept is a callback that is fired when an inbound connection is
	// accepted.  It is the caller's responsibility to close the connection.
	// Failure to close the connection will result in the connection manager
	// believing the connection is still active and thus have undesirable
	// side effects such as still counting toward maximum connection limits.
	//
	// This field will not have any effect if the Listeners field is not
	// also specified since there couldn't possibly be any accepted
	// connections in that case.
	OnAccept func(*ConnReq)

	// TargetOutbound is the number of outbound network connections to
	// maintain. Defaults to 8.
	TargetOutbound uint32

	// RetryDuration is the duration to wait before retrying connection
	// requests. Defaults to 5s.
	RetryDuration time.Duration

	// OnConnection is a callback that is fired when a new outbound
	// connection is established.
	OnConnection func(*ConnReq)

	// OnDisconnection is a callback that is fired when an outbound
	// connection is disconnected.
	OnDisconnection func(*ConnReq)

	// GetNewAddress is a way to get an address to make a network connection
	// to.  If nil, no new connections will be made automatically.
	GetNewAddress func() (net.Addr, error)

	// Dial connects to the address on the named network. It cannot be nil.
	Dial func(network, addr string) (net.Conn, error)
}

// registerPending is used to register a pending connection attempt. By
// registering pending connection attempts we allow callers to cancel pending
// connection attempts before their successful or in the case they're not
// longer wanted.
type registerPending struct {
	c    *ConnReq
	done chan struct{}
}

// handleConnected is used to queue a successful connection.
type handleConnected struct {
	c    *ConnReq
	conn net.Conn
}

// handleDisconnected is used to remove a connection.
type handleDisconnected struct {
	id    uint64
	retry bool
}

// handleFailed is used to remove a pending connection.
type handleFailed struct {
	c   *ConnReq
	err error
}

// handleFailedConn handles a connection failed due to a disconnect or any
// other failure. If permanent, it retries the connection after the configured
// retry duration. Otherwise, if required, it makes a new connection request.
// After maxFailedConnectionAttempts new connections will be retried after the
// configured retry duration.
func (cm *ConnManager) handleFailedConn(c *ConnReq) {
	if atomic.LoadInt32(&cm.stop) != 0 {
		return
	}
	if c.Permanent {
		c.retryCount++
		d := time.Duration(c.retryCount) * cm.cfg.RetryDuration
		if d > maxRetryDuration {
			d = maxRetryDuration
		}
		log.Debug("Retrying connection", "conn req", c, "wait time", d)
		time.AfterFunc(d, func() {
			cm.Connect(c)
		})
	} else if cm.cfg.GetNewAddress != nil {
		cm.failedAttempts++
		if c.Ban {
			retryDuration := cm.cfg.RetryDuration * 2
			log.Trace(fmt.Sprintf("Retrying ban connection in: %v", retryDuration))
			time.AfterFunc(retryDuration, func() {
				c.Ban = false
				cm.NewConnReq()
			})
		} else if cm.failedAttempts >= maxFailedAttempts {
			//log.Trace(fmt.Sprintf("Max %d failed connection attempts reached", maxFailedAttempts))
			//log.Trace(fmt.Sprintf("Retrying connection in: %v", cm.cfg.RetryDuration),
			//	"failed", cm.failedAttempts)
			time.AfterFunc(cm.cfg.RetryDuration, func() {
				cm.NewConnReq()
			})
		} else {
			//log.Trace("Trying new connection", "failed", cm.failedAttempts)
			go cm.NewConnReq()
		}
	}
}

// connHandler handles all connection related requests.  It must be run as a
// goroutine.
//
// The connection handler makes sure that we maintain a pool of active outbound
// connections so that we remain connected to the network.  Connection requests
// are processed and mapped by their assigned ids.
func (cm *ConnManager) connHandler() {
	// pending holds all registered conn requests that have yet to
	// succeed.
	pending := make(map[uint64]*ConnReq)
	conns := make(map[uint64]*ConnReq, cm.cfg.TargetOutbound)
out:
	for {
		select {
		case req := <-cm.requests:
			switch msg := req.(type) {

			case registerPending:
				connReq := msg.c
				connReq.updateState(ConnPending)
				pending[msg.c.id] = connReq
				close(msg.done)

			case handleConnected:
				connReq := msg.c
				if _, ok := pending[connReq.id]; !ok {
					if msg.conn != nil {
						msg.conn.Close()
					}
					log.Debug(fmt.Sprintf("Ignoring connection for "+
						"canceled connreq=%v", connReq))
					continue
				}
				connReq.updateState(ConnEstablished)
				connReq.conn = msg.conn
				conns[connReq.id] = connReq
				log.Debug(fmt.Sprintf("Connected to %v", connReq))
				connReq.retryCount = 0
				cm.failedAttempts = 0

				delete(pending, connReq.id)

				if cm.cfg.OnConnection != nil {
					go cm.cfg.OnConnection(connReq)
				}

			case handleDisconnected:
				connReq, ok := conns[msg.id]
				if !ok {
					connReq, ok = pending[msg.id]
					if !ok {
						log.Error(fmt.Sprintf("Unknown connid=%d",
							msg.id))
						continue
					}

					// Pending connection was found, remove
					// it from pending map if we should
					// ignore a later, successful
					// connection.
					connReq.updateState(ConnCanceled)
					log.Debug(fmt.Sprintf("Canceling: %v", connReq))
					delete(pending, msg.id)
					continue

				}
				// An existing connection was located, mark as
				// disconnected and execute disconnection
				// callback.
				log.Debug(fmt.Sprintf("Disconnected from %v", connReq))
				delete(conns, msg.id)

				if connReq.conn != nil {
					connReq.conn.Close()
				}

				if cm.cfg.OnDisconnection != nil {
					go cm.cfg.OnDisconnection(connReq)
				}

				// All internal state has been cleaned up, if
				// this connection is being removed, we will
				// make no further attempts with this request.
				if !msg.retry {
					connReq.updateState(ConnDisconnected)
					continue
				}

				// Otherwise, we will attempt a reconnection if
				// we do not have enough peers, or if this is a
				// persistent peer. The connection request is
				// re added to the pending map, so that
				// subsequent processing of connections and
				// failures do not ignore the request.
				if uint32(len(conns)) < cm.cfg.TargetOutbound ||
					connReq.Permanent {

					connReq.updateState(ConnPending)
					log.Debug(fmt.Sprintf("Reconnecting to %v",
						connReq))
					pending[msg.id] = connReq
					cm.handleFailedConn(connReq)
				}
			case handleFailed:
				connReq := msg.c
				if _, ok := pending[connReq.id]; !ok {
					log.Debug(fmt.Sprintf("Ignoring connection for "+
						"canceled conn req: %v", connReq))
					continue
				}
				connReq.updateState(ConnFailing)
				//log.Debug("Failed to connect", "con", connReq, "error", msg.err)
				cm.handleFailedConn(connReq)
			}

		case <-cm.quit:
			break out
		}
	}

	cm.wg.Done()
	log.Trace("Connection handler done")
}

// NewConnReq creates a new connection request and connects to the
// corresponding address.
func (cm *ConnManager) NewConnReq() {
	//log.Trace("Open new Connection")
	if atomic.LoadInt32(&cm.stop) != 0 {
		return
	}
	if cm.cfg.GetNewAddress == nil {
		return
	}

	c := NewConnReq()
	atomic.StoreUint64(&c.id, atomic.AddUint64(&cm.connReqCount, 1))
	// Submit a request of a pending connection attempt to the connection
	// manager. By registering the id before the connection is even
	// established, we'll be able to later cancel the connection via the
	// Remove method.
	done := make(chan struct{})
	select {
	case cm.requests <- registerPending{c, done}:
	case <-cm.quit:
		return
	}

	// Wait for the registration to successfully add the pending conn req to
	// the conn manager's internal state.
	select {
	case <-done:
	case <-cm.quit:
		return
	}

	addr, err := cm.cfg.GetNewAddress()
	if err != nil {
		select {
		case cm.requests <- handleFailed{c, err}:
		case <-cm.quit:
		}
		return
	}

	c.Addr = addr

	cm.Connect(c)
}

// Connect assigns an id and dials a connection to the address of the
// connection request.
func (cm *ConnManager) Connect(c *ConnReq) {
	if atomic.LoadInt32(&cm.stop) != 0 {
		return
	}

	// During the time we wait for retry there is a chance that
	// this connection was already cancelled
	if c.State() == ConnCanceled {
		log.Debug(fmt.Sprintf("Ignoring connect for canceled connreq=%v", c))
		return
	}

	if atomic.LoadUint64(&c.id) == 0 {
		atomic.StoreUint64(&c.id, atomic.AddUint64(&cm.connReqCount, 1))

		// Submit a request of a pending connection attempt to the
		// connection manager. By registering the id before the
		// connection is even established, we'll be able to later
		// cancel the connection via the Remove method.
		done := make(chan struct{})
		select {
		case cm.requests <- registerPending{c, done}:
		case <-cm.quit:
			return
		}

		// Wait for the registration to successfully add the pending
		// conn req to the conn manager's internal state.
		select {
		case <-done:
		case <-cm.quit:
			return
		}
	}
	log.Debug(fmt.Sprintf("Attempting to connect to %v", c))
	conn, err := cm.cfg.Dial(c.Addr.Network(), c.Addr.String())
	if err != nil {
		select {
		case cm.requests <- handleFailed{c, err}:
		case <-cm.quit:
		}
		return
	}

	select {
	case cm.requests <- handleConnected{c, conn}:
	case <-cm.quit:
	}
}

// Disconnect disconnects the connection corresponding to the given connection
// id. If permanent, the connection will be retried with an increasing backoff
// duration.
func (cm *ConnManager) Disconnect(id uint64) {
	if atomic.LoadInt32(&cm.stop) != 0 {
		return
	}
	cm.requests <- handleDisconnected{id, true}
}

// Remove removes the connection corresponding to the given connection
// id from known connections.
func (cm *ConnManager) Remove(id uint64) {
	if atomic.LoadInt32(&cm.stop) != 0 {
		return
	}
	cm.requests <- handleDisconnected{id, false}
}

// listenHandler accepts incoming connections on a given listener.  It must be
// run as a goroutine.
func (cm *ConnManager) listenHandler(listener net.Listener) {
	log.Info("P2P Server listening on", "addr", listener.Addr())
	for atomic.LoadInt32(&cm.stop) == 0 {
		conn, err := listener.Accept()
		if err != nil {
			// Only log the error if not forcibly shutting down.
			if atomic.LoadInt32(&cm.stop) == 0 {
				log.Error("Can't accept connection", "error", err)
			}
			continue
		}
		c := NewConnReq()
		c.conn = conn
		go cm.cfg.OnAccept(c)
	}

	cm.wg.Done()
	log.Trace("Listener handler done", "addr", listener.Addr())
}

// Wait blocks until the connection manager halts gracefully.
func (cm *ConnManager) Wait() {
	cm.wg.Wait()
}
