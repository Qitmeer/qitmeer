// Copyright (c) 2017-2018 The nox developers
// Copyright (c) 2013-2016 The btcsuite developers
// Copyright (c) 2016-2018 The Decred developers
// Use of this source code is governed by an ISC
// license that can be found in the LICENSE file.
package peer

import (
	"fmt"
	"github.com/HalalChain/qitmeer-lib/common/hash"
	"github.com/HalalChain/qitmeer/core/blockchain"
	"github.com/HalalChain/qitmeer/core/blockdag"
	"github.com/HalalChain/qitmeer/core/message"
	"github.com/HalalChain/qitmeer/core/protocol"
	"github.com/HalalChain/qitmeer/core/types"
	"github.com/HalalChain/qitmeer/log"
	"math/rand"
	"net"
	"strconv"
	"sync/atomic"
	"time"
)

// ID returns the peer id.
//
// This function is safe for concurrent access.
func (p *Peer) ID() int32 {
	p.flagsMtx.Lock()
	id := p.id
	p.flagsMtx.Unlock()

	return id
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
	log.Trace("Disconnecting ", "peer",p.addr)
	if atomic.LoadInt32(&p.connected) != 0 {
		p.conn.Close()
	}
	close(p.quit)
}

// AssociateConnection associates the given conn to the peer.
// Calling this function when the peer is already connected will
// have no effect.
func (p *Peer) AssociateConnection(conn net.Conn) {
	// Already connected?
	if !atomic.CompareAndSwapInt32(&p.connected, 0, 1) {
		return
	}

	p.conn = conn
	p.timeConnected = time.Now()

	if p.inbound {
		p.addr = p.conn.RemoteAddr().String()

		// Set up a NetAddress for the peer to be used with AddrManager.  We
		// only do this inbound because outbound set this up at connection time
		// and no point recomputing.
		na, err := types.NewNetAddressFailBack(p.conn.RemoteAddr(), p.services)
		if err != nil {
			log.Error("Cannot create remote net address", "error",err)
			p.Disconnect()
			return
		}
		p.na = na
	}

	go func(peer *Peer) {
		if err := peer.start(); err != nil {
			log.Debug("Cannot start peer", "peer",peer.addr, "error",err)
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

// PushAddrMsg sends an addr message to the connected peer using the provided
// addresses.  This function is useful over manually sending the message via
// QueueMessage since it automatically limits the addresses to the maximum
// number allowed by the message and randomizes the chosen addresses when there
// are too many.  It returns the addresses that were actually sent and no
// message will be sent if there are no entries in the provided addresses slice.
//
// This function is safe for concurrent access.
func (p *Peer) PushAddrMsg(addresses []*types.NetAddress) ([]*types.NetAddress, error) {

	// Nothing to send.
	if len(addresses) == 0 {
		return nil, nil
	}

	msg := message.NewMsgAddr()
	msg.AddrList = make([]*types.NetAddress, len(addresses))
	copy(msg.AddrList, addresses)

	// Randomize the addresses sent if there are more than the maximum allowed.
	if len(msg.AddrList) > message.MaxAddrPerMsg {
		// Shuffle the address list.
		for i := range msg.AddrList {
			j := rand.Intn(i + 1)
			msg.AddrList[i], msg.AddrList[j] = msg.AddrList[j], msg.AddrList[i]
		}

		// Truncate it to the maximum size.
		msg.AddrList = msg.AddrList[:message.MaxAddrPerMsg]
	}

	p.QueueMessage(msg, nil)
	return msg.AddrList, nil
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
func (p *Peer) PushGetBlocksMsg(gs *blockdag.GraphState,blocks []*hash.Hash) error {

	isDuplicate:=false
	bs:=blockdag.NewHashSet()


	// Filter duplicate getblocks requests.
	p.prevGetBlocksMtx.Lock()
	if len(blocks)>0 {
		if p.prevGetBlocks==nil {
			p.prevGetBlocks=blockdag.NewHashSet()
			bs.AddList(blocks)
		}else {
			isDuplicate = p.prevGetBlocks.Contain(bs)
			for _,v:=range blocks{
				if !p.prevGetBlocks.Has(v) {
					bs.Add(v)
				}
			}
			if bs.IsEmpty() {
				isDuplicate=true
			}
		}
	}else {
		if p.prevGetGS!=nil {
			isDuplicate = gs.IsEqual(p.prevGetGS)
		}
	}
	p.prevGetBlocksMtx.Unlock()

	if isDuplicate {
		if len(blocks)>0 {
			log.Trace(fmt.Sprintf("Filtering duplicate [getblocks]: "+
				"prev:%d cur:%d", p.prevGetBlocks.Size(),len(blocks)))
		}else {
			log.Trace(fmt.Sprintf("Filtering duplicate [getblocks]: "+
				"prev:%s cur:%s", p.prevGetGS.String(),gs.String()))
		}

		return nil
	}

	// Construct the getblocks request and queue it to be sent.
	msg := message.NewMsgGetBlocks(gs)
	if !bs.IsEmpty() {
		for k:=range bs.GetMap(){
			msg.AddBlockLocatorHash(&k)
		}
	}
	p.QueueMessage(msg, nil)

	// Update the previous getblocks request information for filtering
	// duplicates.
	p.prevGetBlocksMtx.Lock()
	p.prevGetGS=gs
	p.prevGetBlocks.AddSet(bs)
	p.prevGetBlocksMtx.Unlock()
	return nil
}

// PushGetHeadersMsg sends a getblocks message for the provided block locator
// and stop hash.  It will ignore back-to-back duplicate requests.
//
// This function is safe for concurrent access.
func (p *Peer) PushGetHeadersMsg(locator blockchain.BlockLocator, stopHash *hash.Hash) error {
	// Extract the begin hash from the block locator, if one was specified,
	// to use for filtering duplicate getheaders requests.
	var beginHash *hash.Hash
	if len(locator) > 0 {
		beginHash = locator[0]
	}

	// Filter duplicate getheaders requests.
	p.prevGetHdrsMtx.Lock()
	isDuplicate := p.prevGetHdrsStop != nil && p.prevGetHdrsBegin != nil &&
		beginHash != nil && stopHash.IsEqual(p.prevGetHdrsStop) &&
		beginHash.IsEqual(p.prevGetHdrsBegin)
	p.prevGetHdrsMtx.Unlock()

	if isDuplicate {
		log.Trace(fmt.Sprintf("Filtering duplicate [getheaders] with begin hash %v",
			beginHash))
		return nil
	}

	// Construct the getheaders request and queue it to be sent.
	msg := message.NewMsgGetHeaders()
	msg.HashStop = *stopHash
	for _, hash := range locator {
		err := msg.AddBlockLocatorHash(hash)
		if err != nil {
			return err
		}
	}
	p.QueueMessage(msg, nil)

	// Update the previous getheaders request information for filtering
	// duplicates.
	p.prevGetHdrsMtx.Lock()
	p.prevGetHdrsBegin = beginHash
	p.prevGetHdrsStop = stopHash
	p.prevGetHdrsMtx.Unlock()
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
	log.Trace(fmt.Sprintf("Updating last graph state of peer %v from %v to %v",
		p.addr, p.lastGS.String(),newGS.String()))
	p.lastGS.Equal(newGS)
	p.statsMtx.Unlock()
}

// UpdateLastAnnouncedBlock updates meta-data about the last block hash this
// peer is known to have announced.
//
// This function is safe for concurrent access.
func (p *Peer) UpdateLastAnnouncedBlock(blkHash *hash.Hash) {
	log.Trace("Updating last blk for peer", "peer",p.addr, "block hash",blkHash)

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
func (p *Peer) QueueInventoryImmediate(invVect *message.InvVect,gs *blockdag.GraphState) {
	// Don't announce the inventory if the peer is already known to have it.
	if p.knownInventory.Exists(invVect) {
		return
	}

	// Avoid risk of deadlock if goroutine already exited.  The goroutine
	// we will be sending to hangs around until it knows for a fact that
	// it is marked as disconnected and *then* it drains the channels.
	if !p.Connected() {
		return
	}

	// Generate and queue a single inv message with the inventory vector.
	invMsg := message.NewMsgInvSizeHint(1)
	invMsg.GS=gs
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

func (p*Peer) CleanGetBlocksSet() {
	p.statsMtx.RLock()
	if p.prevGetBlocks!=nil {
		p.prevGetBlocks.Clean()
	}
	p.statsMtx.RUnlock()
}