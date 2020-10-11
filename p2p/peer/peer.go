// Copyright (c) 2017-2018 The qitmeer developers
// Copyright (c) 2013-2016 The btcsuite developers
// Copyright (c) 2016-2018 The Decred developers
// Use of this source code is governed by an ISC
// license that can be found in the LICENSE file.
package peer

import (
	"fmt"
	"github.com/Qitmeer/qitmeer/common/hash"
	"github.com/Qitmeer/qitmeer/common/roughtime"
	"github.com/Qitmeer/qitmeer/core/blockdag"
	"github.com/Qitmeer/qitmeer/core/message"
	"github.com/Qitmeer/qitmeer/core/protocol"
	"github.com/Qitmeer/qitmeer/core/types"
	"github.com/Qitmeer/qitmeer/log"
	"github.com/Qitmeer/qitmeer/p2p/connmgr"
	"github.com/satori/go.uuid"
	"net"
	"strconv"
	"sync/atomic"
	"time"
)

// StatsSnap is a snapshot of peer stats at a point in time.
type StatsSnap struct {
	UUID           uuid.UUID
	ID             int32
	Addr           string
	Services       protocol.ServiceFlag
	LastSend       time.Time
	LastRecv       time.Time
	BytesSent      uint64
	BytesRecv      uint64
	ConnTime       time.Time
	TimeOffset     int64
	Version        uint32
	UserAgent      string
	Inbound        bool
	LastPingNonce  uint64
	LastPingTime   time.Time
	LastPingMicros int64
	GraphState     *blockdag.GraphState
}

// ID returns the peer id.
//
// This function is safe for concurrent access.
func (p *Peer) ID() int32 {
	p.flagsMtx.Lock()
	id := p.id
	p.flagsMtx.Unlock()

	return id
}

// UUID returns the peer uuid.
//
// This function is safe for concurrent access.
func (p *Peer) UUID() uuid.UUID {
	p.flagsMtx.Lock()
	uuid := p.uuid
	p.flagsMtx.Unlock()

	return uuid
}

// NA returns the peer network address.
//
// This function is safe for concurrent access.
func (p *Peer) NA() *types.NetAddress {
	p.flagsMtx.Lock()
	na := p.na
	p.flagsMtx.Unlock()

	return na
}

// LastGS returns the last graph state of the peer.
//
// This function is safe for concurrent access.
func (p *Peer) LastGS() *blockdag.GraphState {
	p.statsMtx.RLock()
	lastgs := p.lastGS
	p.statsMtx.RUnlock()

	return lastgs
}

// Inbound returns whether the peer is inbound.
//
// This function is safe for concurrent access.
func (p *Peer) Inbound() bool {
	return p.inbound
}

// Connected returns whether or not the peer is currently connected.
//
// This function is safe for concurrent access.
func (p *Peer) Connected() bool {
	return atomic.LoadInt32(&p.connected) != 0 &&
		atomic.LoadInt32(&p.disconnect) == 0
}

// Disconnect disconnects the peer by closing the connection.  Calling this
// function when the peer is already disconnected or in the process of
// disconnecting will have no effect.
func (p *Peer) Disconnect() {
	if atomic.AddInt32(&p.disconnect, 1) != 1 {
		return
	}
	log.Trace("Disconnecting ", "peer", p.addr)
	if atomic.LoadInt32(&p.connected) != 0 {
		p.conn.Close()
	}
	close(p.quit)
}

// AssociateConnection associates the given conn to the peer.
// Calling this function when the peer is already connected will
// have no effect.
func (p *Peer) AssociateConnection(c *connmgr.ConnReq) {
	// Already connected?
	if !atomic.CompareAndSwapInt32(&p.connected, 0, 1) {
		return
	}

	p.conn = c.Conn()
	p.timeConnected = roughtime.Now()

	if p.inbound {
		p.addr = p.conn.RemoteAddr().String()

		// Set up a NetAddress for the peer to be used with AddrManager.  We
		// only do this inbound because outbound set this up at connection time
		// and no point recomputing.
		na, err := types.NewNetAddressFailBack(p.conn.RemoteAddr(), p.services)
		if err != nil {
			log.Error("Cannot create remote net address", "error", err)
			p.Disconnect()
			return
		}
		p.na = na
	}

	go func(peer *Peer) {
		if err := peer.start(); err != nil {
			log.Debug("Cannot start peer", "peer", peer.addr, "error", err)
			c.Ban = true
			peer.Disconnect()
		}
	}(p)
}

// QueueMessage adds the passed wire message to the peer send queue.
//
// This function is safe for concurrent access.
func (p *Peer) QueueMessage(msg message.Message, doneChan chan<- struct{}) {
	// Avoid risk of deadlock if goroutine already exited.  The goroutine
	// we will be sending to hangs around until it knows for a fact that
	// it is marked as disconnected and *then* it drains the channels.
	if !p.Connected() {
		if doneChan != nil {
			go func() {
				doneChan <- struct{}{}
			}()
		}
		return
	}
	p.outputQueue <- outMsg{msg: msg, doneChan: doneChan}
}

// Addr returns the peer address.
// This function is safe for concurrent access.
func (p *Peer) Addr() string {
	// The address doesn't change after initialization, therefore it is not
	// protected by a mutex.
	return p.addr
}

// WaitForDisconnect waits until the peer has completely disconnected and all
// resources are cleaned up.  This will happen if either the local or remote
// side has been disconnected or the peer is forcibly disconnected via
// Disconnect.
func (p *Peer) WaitForDisconnect() {
	<-p.quit
}

// VersionKnown returns the whether or not the version of a peer is known
// locally.
//
// This function is safe for concurrent access.
func (p *Peer) VersionKnown() bool {
	p.flagsMtx.Lock()
	versionKnown := p.versionKnown
	p.flagsMtx.Unlock()

	return versionKnown
}

// ProtocolVersion returns the negotiated peer protocol version.
//
// This function is safe for concurrent access.
func (p *Peer) ProtocolVersion() uint32 {
	p.flagsMtx.Lock()
	protocolVersion := p.protocolVersion
	p.flagsMtx.Unlock()

	return protocolVersion
}

// VerAckReceived returns whether or not a verack message was received by the
// peer.
//
// This function is safe for concurrent access.
func (p *Peer) VerAckReceived() bool {
	p.flagsMtx.Lock()
	verAckReceived := p.verAckReceived
	p.flagsMtx.Unlock()

	return verAckReceived
}

// PushRejectMsg sends a reject message for the provided command, reject code,
// reject reason, and hash.  The hash will only be used when the command is a tx
// or block and should be nil in other cases.  The wait parameter will cause the
// function to block until the reject message has actually been sent.
//
// This function is safe for concurrent access.
func (p *Peer) PushRejectMsg(command string, code message.RejectCode, reason string, h *hash.Hash, wait bool) {
	msg := message.NewMsgReject(command, code, reason)
	if command == message.CmdTx || command == message.CmdBlock {
		if h == nil {
			log.Warn("Sending a reject message for command "+
				"type %v which should have specified a hash "+
				"but does not", command)
			h = &hash.ZeroHash
		}
		msg.Hash = *h
	}

	// Send the message without waiting if the caller has not requested it.
	if !wait {
		p.QueueMessage(msg, nil)
		return
	}

	// Send the message and block until it has been sent before returning.
	doneChan := make(chan struct{}, 1)
	p.QueueMessage(msg, doneChan)
	<-doneChan
}

// NewInboundPeer returns a new inbound peer. Use Start to begin
// processing incoming and outgoing messages.
func NewInboundPeer(cfg *Config) *Peer {
	return newPeerBase(cfg, true)
}

// NewOutboundPeer returns a new outbound peer.
func NewOutboundPeer(cfg *Config, addr string) (*Peer, error) {
	p := newPeerBase(cfg, false)
	p.addr = addr

	host, portStr, err := net.SplitHostPort(addr)
	if err != nil {
		return nil, err
	}

	port, err := strconv.ParseUint(portStr, 10, 16)
	if err != nil {
		return nil, err
	}

	if cfg.HostToNetAddress != nil {
		na, err := cfg.HostToNetAddress(host, uint16(port), 0)
		if err != nil {
			return nil, err
		}
		p.na = na
	} else {
		p.na = types.NewNetAddressIPPort(net.ParseIP(host), uint16(port), 0)
	}

	return p, nil
}

func (p *Peer) String() string {
	direction := "outbound"
	if p.inbound {
		direction = "inbound"
	}
	return fmt.Sprintf("%s (%s)", p.addr, direction)
}

// UserAgent returns the user agent of the remote peer.
//
// This function is safe for concurrent access.
func (p *Peer) UserAgent() string {
	p.flagsMtx.Lock()
	userAgent := p.userAgent
	p.flagsMtx.Unlock()

	return userAgent
}

// Services returns the services flag of the remote peer.
//
// This function is safe for concurrent access.
func (p *Peer) Services() protocol.ServiceFlag {
	p.flagsMtx.Lock()
	services := p.services
	p.flagsMtx.Unlock()

	return services
}

// PushGetBlocksMsg sends a getblocks message for the provided block locator
// and stop hash.  It will ignore back-to-back duplicate requests.
//
// This function is safe for concurrent access.
func (p *Peer) PushGetBlocksMsg(sgs *blockdag.GraphState, blocks []*hash.Hash) error {
	gs := sgs.Clone()
	ok, bs := p.PrevGet.CheckBlocks(p, gs, blocks)
	if !ok {
		return fmt.Errorf("duplicate")
	}
	// Construct the getblocks request and queue it to be sent.
	msg := message.NewMsgGetBlocks(gs)
	if !bs.IsEmpty() {
		for k := range bs.GetMap() {
			ha := k
			msg.AddBlockLocatorHash(&ha)
		}

	}
	p.QueueMessage(msg, nil)
	// Update the previous getblocks request information for filtering
	// duplicates.
	p.PrevGet.UpdateBlocks(blocks)
	return nil
}

// PushGetHeadersMsg sends a getblocks message
//
// This function is safe for concurrent access.
func (p *Peer) PushGetHeadersMsg(sgs *blockdag.GraphState, blocks []*hash.Hash) error {
	gs := sgs.Clone()
	ok, bs := p.prevGetHdrs.CheckBlocks(p, gs, blocks)
	if !ok {
		return nil
	}
	// Construct the getblocks request and queue it to be sent.
	msg := message.NewMsgGetHeaders(gs)
	if !bs.IsEmpty() {
		for k := range bs.GetMap() {
			ha := k
			msg.AddBlockLocatorHash(&ha)
		}

	}
	p.QueueMessage(msg, nil)
	// Update the previous getblocks request information for filtering
	// duplicates.
	p.prevGetHdrs.UpdateBlocks(blocks)
	return nil
}

// AddKnownInventory adds the passed inventory to the cache of known inventory
// for the peer.
//
// This function is safe for concurrent access.
func (p *Peer) AddKnownInventory(invVect *message.InvVect) {
	p.knownInventory.Add(invVect)
}

// UpdateLastGS updates the last known graph state for the peer.
//
// This function is safe for concurrent access.
func (p *Peer) UpdateLastGS(newGS *blockdag.GraphState) {
	p.statsMtx.Lock()
	if !p.lastGS.IsEqual(newGS) {
		log.Trace(fmt.Sprintf("Updating last graph state of peer %v from %v to %v",
			p.addr, p.lastGS.String(), newGS.String()))
		p.lastGS.Equal(newGS)
	}
	p.statsMtx.Unlock()
}

// UpdateLastAnnouncedBlock updates meta-data about the last block hash this
// peer is known to have announced.
//
// This function is safe for concurrent access.
func (p *Peer) UpdateLastAnnouncedBlock(blkHash *hash.Hash) {
	log.Trace("Updating last blk for peer", "peer", p.addr, "block hash", blkHash)

	p.statsMtx.Lock()
	p.lastAnnouncedBlock = blkHash
	p.statsMtx.Unlock()
}

// WantsHeaders returns if the peer wants header messages instead of
// inventory vectors for blocks.
//
// This function is safe for concurrent access.
func (p *Peer) WantsHeaders() bool {
	p.flagsMtx.Lock()
	sendHeadersPreferred := p.sendHeadersPreferred
	p.flagsMtx.Unlock()

	return sendHeadersPreferred
}

// QueueInventory adds the passed inventory to the inventory send queue which
// might not be sent right away, rather it is trickled to the peer in batches.
// Inventory that the peer is already known to have is ignored.
//
// This function is safe for concurrent access.
func (p *Peer) QueueInventory(invVect *message.InvVect) {
	// Don't add the inventory to the send queue if the peer is already
	// known to have it.
	if p.knownInventory.Exists(invVect) {
		return
	}

	// Avoid risk of deadlock if goroutine already exited.  The goroutine
	// we will be sending to hangs around until it knows for a fact that
	// it is marked as disconnected and *then* it drains the channels.
	if !p.Connected() {
		return
	}

	p.outputInvChan <- invVect
}

// QueueInventoryImmediate adds the passed inventory to the send queue to be
// sent immediately.  This should typically only be used for inventory that is
// time sensitive such as new tip blocks or votes.  Normal inventory should be
// announced via QueueInventory which instead trickles it to the peer in
// batches.  Inventory that the peer is already known to have is ignored.
//
// This function is safe for concurrent access.
func (p *Peer) QueueInventoryImmediate(invVect *message.InvVect, gs *blockdag.GraphState) {
	// Don't announce the inventory if the peer is already known to have it.
	if p.knownInventory.Exists(invVect) {
		return
	}

	// Avoid risk of deadlock if goroutine already exited.  The goroutine
	// we will be sending to hangs around until it knows for a fact that
	// it is marked as disconnected and *then* it drains the channels.
	if !p.Connected() || gs == nil {
		return
	}

	// Generate and queue a single inv message with the inventory vector.
	invMsg := message.NewMsgInvSizeHint(1)
	invMsg.GS = gs
	invMsg.AddInvVect(invVect)
	p.AddKnownInventory(invVect)
	p.outputQueue <- outMsg{msg: invMsg, doneChan: nil}
}

// LastAnnouncedBlock returns the last announced block of the remote peer.
//
// This function is safe for concurrent access.
func (p *Peer) LastAnnouncedBlock() *hash.Hash {
	p.statsMtx.RLock()
	lastAnnouncedBlock := p.lastAnnouncedBlock
	p.statsMtx.RUnlock()

	return lastAnnouncedBlock
}

func (p *Peer) CleanGetBlocksSet() {
	p.statsMtx.RLock()
	p.PrevGet.Clean()
	p.prevGetHdrs.Clean()
	p.statsMtx.RUnlock()
}

// LastSend returns the last send time of the peer.
//
// This function is safe for concurrent access.
func (p *Peer) LastSend() time.Time {
	return time.Unix(atomic.LoadInt64(&p.lastSend), 0)
}

// LastRecv returns the last recv time of the peer.
//
// This function is safe for concurrent access.
func (p *Peer) LastRecv() time.Time {
	return time.Unix(atomic.LoadInt64(&p.lastRecv), 0)
}

// BytesSent returns the total number of bytes sent by the peer.
//
// This function is safe for concurrent access.
func (p *Peer) BytesSent() uint64 {
	return atomic.LoadUint64(&p.bytesSent)
}

// BytesReceived returns the total number of bytes received by the peer.
//
// This function is safe for concurrent access.
func (p *Peer) BytesReceived() uint64 {
	return atomic.LoadUint64(&p.bytesReceived)
}

// LocalAddr returns the local address of the connection.
//
// This function is safe fo concurrent access.
func (p *Peer) LocalAddr() net.Addr {
	var localAddr net.Addr
	if atomic.LoadInt32(&p.connected) != 0 {
		localAddr = p.conn.LocalAddr()
	}
	return localAddr
}

// StatsSnapshot returns a snapshot of the current peer flags and statistics.
//
// This function is safe for concurrent access.
func (p *Peer) StatsSnapshot() *StatsSnap {
	p.statsMtx.RLock()

	p.flagsMtx.Lock()
	id := p.id
	uuid := p.uuid
	addr := p.addr
	userAgent := p.userAgent
	services := p.services
	protocolVersion := p.advertisedProtoVer
	p.flagsMtx.Unlock()

	// Get a copy of all relevant flags and stats.
	statsSnap := &StatsSnap{
		UUID:           uuid,
		ID:             id,
		Addr:           addr,
		UserAgent:      userAgent,
		Services:       services,
		LastSend:       p.LastSend(),
		LastRecv:       p.LastRecv(),
		BytesSent:      p.BytesSent(),
		BytesRecv:      p.BytesReceived(),
		ConnTime:       p.timeConnected,
		TimeOffset:     p.timeOffset,
		Version:        protocolVersion,
		Inbound:        p.inbound,
		LastPingNonce:  p.lastPingNonce,
		LastPingMicros: p.lastPingMicros,
		LastPingTime:   p.lastPingTime,
		GraphState:     p.lastGS,
	}

	p.statsMtx.RUnlock()
	return statsSnap
}

// LastPingNonce returns the last ping nonce of the remote peer.
func (p *Peer) LastPingNonce() uint64 {
	p.statsMtx.RLock()
	lastPingNonce := p.lastPingNonce
	p.statsMtx.RUnlock()

	return lastPingNonce
}

func (p *Peer) Cfg() *Config {
	return &p.cfg
}

func (p *Peer) PushGraphStateMsg(gs *blockdag.GraphState) error {
	msg := message.NewMsgGraphState(gs)
	p.QueueMessage(msg, nil)
	return nil
}

func (p *Peer) PushSyncDAGMsg(sgs *blockdag.GraphState, mainLocator []*hash.Hash) error {
	gs := sgs.Clone()
	msg := message.NewMsgSyncDAG(gs, mainLocator)
	p.QueueMessage(msg, nil)
	p.PrevGet.UpdateGS(gs, mainLocator)
	return nil
}
