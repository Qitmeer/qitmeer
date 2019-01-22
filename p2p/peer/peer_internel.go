// Copyright (c) 2017-2018 The nox developers
// Copyright (c) 2013-2016 The btcsuite developers
// Copyright (c) 2016-2018 The Decred developers
// Use of this source code is governed by an ISC
// license that can be found in the LICENSE file.
package peer

import (
	"bytes"
	"errors"
	"fmt"
	"github.com/davecgh/go-spew/spew"
	"github.com/noxproject/nox/common/hash"
	"github.com/noxproject/nox/core/message"
	"github.com/noxproject/nox/core/protocol"
	s "github.com/noxproject/nox/core/serialization"
	"github.com/noxproject/nox/core/types"
	"github.com/noxproject/nox/log"
	"github.com/noxproject/nox/p2p/peer/invcache"
	"github.com/noxproject/nox/p2p/peer/nounce"
	"github.com/noxproject/nox/params"
	"net"
	"sync"
	"sync/atomic"
	"time"
)

// outMsg is used to house a message to be sent along with a channel to signal
// when the message has been sent (or won't be sent due to things such as
// shutdown)
type outMsg struct {
	msg      message.Message
	doneChan chan<- struct{}
}

// Peer represents a connected p2p network remote node.
type Peer struct {

	conn net.Conn

	// These fields are set at creation time and never modified, so they are
	// safe to read from concurrently without a mutex.
	addr    string
	cfg     Config
	inbound bool

	// These fields are variables must only be used atomically.
	connected     int32   //connected flag
	disconnect    int32   //disconnect flag

	bytesReceived uint64  //msg bytes read from
	bytesSent     uint64  //msg bytes write to

	lastRecv      int64  //last recv time
	lastSend      int64  //last sent time


	// These fields protects by the flagsMtx mutex
	flagsMtx             sync.Mutex
	// - address
	na                   *types.NetAddress
	// - id
	id                   int32
	// - user-agent
	userAgent            string
	// - version
	versionKnown         bool
	// - services flag
	services             protocol.ServiceFlag
	// - advertised protocol version by remote
	advertisedProtoVer   uint32
	// - negotiated protocol version
	protocolVersion      uint32

	versionSent          bool  // peer sent the version msg
	verAckReceived       bool  // peer received the version ack msg
	sendHeadersPreferred bool  // peer wants header instead of block

	// Inv
	knownInventory     *invcache.InventoryCache

	prevGetBlocksMtx   sync.Mutex
	prevGetBlocksBegin *hash.Hash
	prevGetBlocksStop  *hash.Hash
	prevGetHdrsMtx     sync.Mutex
	prevGetHdrsBegin   *hash.Hash
	prevGetHdrsStop    *hash.Hash

	// These fields keep track of statistics for the peer and are protected
	// by the statsMtx mutex.
	statsMtx           sync.RWMutex

	// - block
	startingHeight     int64
	lastBlock          uint64
	lastAnnouncedBlock *hash.Hash

	// - Time
	timeOffset         int64
	timeConnected      time.Time

	// - Ping/Pong
	lastPingNonce      uint64    // Set to nonce if we have a pending ping.
	lastPingTime       time.Time // Time we sent last ping.
	lastPingMicros     int64     // Time for last ping to return.

	// These fields are chans for peer msg handling
	//  - quit
	quit          chan struct{}
	//  - stall
	stallControl  chan stallControlMsg
	//  - in
	inQuit        chan struct{}
	queueQuit     chan struct{}
	outQuit       chan struct{}
	//  - query
	outputQueue   chan outMsg
	sendQueue     chan outMsg
	sendDoneQueue chan struct{}
	outputInvChan chan *message.InvVect

}

var (
	// nodeCount is the total number of peer connections made since startup
	// and is used to assign an id to a peer.
	nodeCount int32

	// sentNonces houses the unique nonces that are generated when pushing
	// version messages that are used to detect self connections.
	sentNonces = nounce.NewLruNonceCache(50)

	// allowSelfConns is only used to allow the tests to bypass the self
	// connection detecting and disconnect logic since they intentionally
	// do so for testing purposes.
	allowSelfConns bool
)

// readMessage reads the next wire message from the peer with logging.
func (p *Peer) readMessage() (message.Message, []byte, error) {
	n, msg, buf, err := message.ReadMessageN(p.conn, p.ProtocolVersion(),
		p.cfg.ChainParams.Net)
	atomic.AddUint64(&p.bytesReceived, uint64(n))
	if p.cfg.Listeners.OnRead != nil {
		p.cfg.Listeners.OnRead(p, n, msg, err)
	}
	if err != nil {
		return nil, nil, err
	}

	// Use closures to log expensive operations so they are only run when
	// the logging level requires it.
	log.Debug(fmt.Sprintf("%v",log.LogClosure(func() string {
		// Debug summary of message.
		summary := message.Summary(msg)
		if len(summary) > 0 {
			summary = " (" + summary + ")"
		}
		return fmt.Sprintf("Received %v%s from %s",
			msg.Command(), summary, p.addr)
	})))
	log.Trace(fmt.Sprintf("%v",log.LogClosure(func() string {
		return spew.Sdump(msg)
	})))
	log.Trace(fmt.Sprintf("%v",log.LogClosure(func() string {
		return spew.Sdump(buf)
	})))

	return msg, buf, nil
}

// writeMessage sends a wire message to the peer with logging.
func (p *Peer) writeMessage(msg message.Message) error {
	// Don't do anything if we're disconnecting.
	if atomic.LoadInt32(&p.disconnect) != 0 {
		return nil
	}

	// Use closures to log expensive operations so they are only run when
	// the logging level requires it.
	log.Debug(fmt.Sprintf("%v", log.LogClosure(func() string {
		// Debug summary of message.
		summary := message.Summary(msg)
		if len(summary) > 0 {
			summary = " (" + summary + ")"
		}
		return fmt.Sprintf("Sending %v%s to %s", msg.Command(),
			summary, p.addr)
	})))
	log.Trace(fmt.Sprintf("%v", log.LogClosure(func() string {
		return spew.Sdump(msg)
	})))
	log.Trace(fmt.Sprintf("%v", log.LogClosure(func() string {
		var buf bytes.Buffer
		err := message.WriteMessage(&buf, msg, p.ProtocolVersion(),
			p.cfg.ChainParams.Net)
		if err != nil {
			return err.Error()
		}
		return spew.Sdump(buf.Bytes())
	})))

	// Write the message to the peer.
	n, err := message.WriteMessageN(p.conn, msg, p.ProtocolVersion(),
		p.cfg.ChainParams.Net)
	atomic.AddUint64(&p.bytesSent, uint64(n))
	if p.cfg.Listeners.OnWrite != nil {
		p.cfg.Listeners.OnWrite(p, n, msg, err)
	}
	return err
}

// readRemoteVersionMsg waits for the next message to arrive from the remote
// peer.  If the next message is not a version message or the version is not
// acceptable then return an error.
func (p *Peer) readRemoteVersionMsg() error {
	// Read their version message.
	remoteMsg, _, err := p.readMessage()
	if err != nil {
		return err
	}

	// Notify and disconnect clients if the first message is not a version
	// message.
	msg, ok := remoteMsg.(*message.MsgVersion)
	if !ok {
		reason := "a version message must precede all others"
		rejectMsg := message.NewMsgReject(msg.Command(), message.RejectMalformed,
			reason)
		_ = p.writeMessage(rejectMsg)
		return errors.New(reason)
	}

	// Detect self connections.
	if !allowSelfConns && sentNonces.Exists(msg.Nonce) {
		return errors.New("disconnecting peer connected to self")
	}

	// Negotiate the protocol version and set the services to what the remote
	// peer advertised.
	p.flagsMtx.Lock()
	p.advertisedProtoVer = uint32(msg.ProtocolVersion)
	// negotiated protocol version is the minor version of remote version and local version
	if (p.protocolVersion > p.advertisedProtoVer) {
		p.protocolVersion = p.advertisedProtoVer
	}
	p.versionKnown = true
	p.services = msg.Services
	p.na.Services = msg.Services
	p.flagsMtx.Unlock()
	log.Debug("Negotiated protocol version", "ver",p.protocolVersion,"peer", p.addr)

	// Updating a bunch of stats.
	p.statsMtx.Lock()
	p.lastBlock = uint64(msg.LastBlock)
	p.startingHeight = int64(msg.LastBlock)

	// Set the peer's time offset.
	p.timeOffset = msg.Timestamp.Unix() - time.Now().Unix()
	p.statsMtx.Unlock()

	// Set the peer's ID and user agent.
	p.flagsMtx.Lock()
	p.id = atomic.AddInt32(&nodeCount, 1)
	p.userAgent = msg.UserAgent
	p.flagsMtx.Unlock()

	// Invoke the callback if specified.  In the case the callback returns a
	// reject message, notify and disconnect the peer accordingly.
	if p.cfg.Listeners.OnVersion != nil {
		rejectMsg := p.cfg.Listeners.OnVersion(p, msg)
		if rejectMsg != nil {
			_ = p.writeMessage(rejectMsg)
			return errors.New(rejectMsg.Reason)
		}
	}

	// Notify and disconnect clients that have a protocol version that is
	// too old.
	if msg.ProtocolVersion < int32(protocol.InitialProcotolVersion) {
		// Send a reject message indicating the protocol version is
		// obsolete and wait for the message to be sent before
		// disconnecting.
		reason := fmt.Sprintf("protocol version must be %d or greater",
			protocol.InitialProcotolVersion)
		rejectMsg := message.NewMsgReject(msg.Command(), message.RejectObsolete,
			reason)
		_ = p.writeMessage(rejectMsg)
		return errors.New(reason)
	}

	return nil
}

// localVersionMsg creates a version message that can be used to send to the
// remote peer.
func (p *Peer) localVersionMsg() (*message.MsgVersion, error) {
	var blockNum uint64
	if p.cfg.NewestBlock != nil {
		var err error
		_, blockNum, err = p.cfg.NewestBlock()
		if err != nil {
			return nil, err
		}
	}

	theirNA := p.na

	// If we are behind a proxy and the connection comes from the proxy then
	// we return an unroutable address as their address. This is to prevent
	// leaking the tor proxy address.
	if p.cfg.Proxy != "" {
		proxyaddress, _, err := net.SplitHostPort(p.cfg.Proxy)
		// invalid proxy means poorly configured, be on the safe side.
		if err != nil || p.na.IP.String() == proxyaddress {
			theirNA = types.NewNetAddressIPPort(net.IP([]byte{0, 0, 0, 0}), 0,
				theirNA.Services)
		}
	}

	// Create a wire.NetAddress with only the services set to use as the
	// "addrme" in the version message.
	//
	// Older nodes previously added the IP and port information to the
	// address manager which proved to be unreliable as an inbound
	// connection from a peer didn't necessarily mean the peer itself
	// accepted inbound connections.
	//
	// Also, the timestamp is unused in the version message.
	ourNA := &types.NetAddress{
		Services: p.cfg.Services,
	}

	// Generate a unique nonce for this peer so self connections can be
	// detected.  This is accomplished by adding it to a size-limited map of
	// recently seen nonces.
	nonce, err := s.RandomUint64()
	if err != nil {
		return nil, err
	}

	// sentNonces houses the unique nonces that are generated when pushing
	// version messages that are used to detect self connections.
	sentNonces.Add(nonce)

	// Version message.
	msg := message.NewMsgVersion(ourNA, theirNA, nonce, int32(blockNum))
	msg.AddUserAgent(p.cfg.UserAgentName, p.cfg.UserAgentVersion,
		p.cfg.UserAgentComments...)

	// Advertise local services.
	msg.Services = p.cfg.Services

	// Advertise our max supported protocol version.
	msg.ProtocolVersion = int32(p.ProtocolVersion())

	// Advertise if inv messages for transactions are desired.
	msg.DisableRelayTx = p.cfg.DisableRelayTx

	return msg, nil
}

// writeLocalVersionMsg writes our version message to the remote peer.
func (p *Peer) writeLocalVersionMsg() error {
	localVerMsg, err := p.localVersionMsg()
	if err != nil {
		return err
	}

	if err := p.writeMessage(localVerMsg); err != nil {
		return err
	}

	p.flagsMtx.Lock()
	p.versionSent = true
	p.flagsMtx.Unlock()
	return nil
}

// negotiateInboundProtocol waits to receive a version message from the peer
// then sends our version message. If the events do not occur in that order then
// it returns an error.
func (p *Peer) negotiateInboundProtocol() error {
	if err := p.readRemoteVersionMsg(); err != nil {
		return err
	}

	return p.writeLocalVersionMsg()
}

// negotiateOutboundProtocol sends our version message then waits to receive a
// version message from the peer.  If the events do not occur in that order then
// it returns an error.
func (p *Peer) negotiateOutboundProtocol() error {
	if err := p.writeLocalVersionMsg(); err != nil {
		return err
	}

	return p.readRemoteVersionMsg()
}

// start begins processing input and output messages.
func (p *Peer) start() error {
	log.Trace("Starting peer", "peer", p.addr)

	negotiateErr := make(chan error, 1)
	go func() {
		if p.inbound {
			negotiateErr <- p.negotiateInboundProtocol()
		} else {
			negotiateErr <- p.negotiateOutboundProtocol()
		}
	}()

	// Negotiate the protocol within the specified negotiateTimeout.
	select {
	case err := <-negotiateErr:
		if err != nil {
			p.Disconnect()
			return err
		}
	case <-time.After(negotiateTimeout):
		p.Disconnect()
		return errors.New("protocol negotiation timeout")
	}
	log.Debug("Connected to ", "peer",p.Addr())

	// The protocol has been negotiated successfully so start processing input
	// and output messages.
	go p.stallHandler()
	go p.inHandler()
	go p.queueHandler()
	go p.outHandler()

	// Send our verack message now that the IO processing machinery has started.
	p.QueueMessage(message.NewMsgVerAck(), nil)
	return nil
}

// newPeerBase returns a new base peer based on the inbound flag.  This
// is used by the NewInboundPeer and NewOutboundPeer functions to perform base
// setup needed by both types of peers.
func newPeerBase(cfg *Config, inbound bool) *Peer {
	// Default to the max supported protocol version.  Override to the
	// version specified by the caller if configured.
	protocolVersion := MaxProtocolVersion
	if cfg.ProtocolVersion != 0 {
		protocolVersion = cfg.ProtocolVersion
	}

	// Set the chain parameters to testnet if the caller did not specify any.
	if cfg.ChainParams == nil {
		cfg.ChainParams = &params.TestNetParams
	}

	p := Peer{
		inbound:         inbound,
		knownInventory:  invcache.NewLruInventoryCache(maxKnownInventory),
		stallControl:    make(chan stallControlMsg, 1), // nonblocking sync
		outputQueue:     make(chan outMsg, outputBufferSize),
		sendQueue:       make(chan outMsg, 1),   // nonblocking sync
		sendDoneQueue:   make(chan struct{}, 1), // nonblocking sync
		outputInvChan:   make(chan *message.InvVect, outputBufferSize),
		inQuit:          make(chan struct{}),
		queueQuit:       make(chan struct{}),
		outQuit:         make(chan struct{}),
		quit:            make(chan struct{}),
		cfg:             *cfg, // Copy so caller can't mutate.
		services:        cfg.Services,
		protocolVersion: protocolVersion,
	}
	return &p
}
