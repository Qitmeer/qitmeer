// Copyright (c) 2017-2018 The nox developers
// Copyright (c) 2013-2016 The btcsuite developers
// Copyright (c) 2015-2018 The Decred developers
// Use of this source code is governed by an ISC
// license that can be found in the LICENSE file.
package peerserver

import (
	"fmt"
	"github.com/noxproject/nox/core/message"
	"github.com/noxproject/nox/core/protocol"
	"github.com/noxproject/nox/core/types"
	"github.com/noxproject/nox/log"
	"github.com/noxproject/nox/p2p/addmgr"
	"github.com/noxproject/nox/p2p/peer"
	"time"
)

// OnVersion is invoked when a peer receives a version wire message and is used
// to negotiate the protocol version details as well as kick start the
// communications.
func (sp *serverPeer) OnVersion(p *peer.Peer, msg *message.MsgVersion) *message.MsgReject {
	// Update the address manager with the advertised services for outbound
	// connections in case they have changed.  This is not done for inbound
	// connections to help prevent malicious behavior and is skipped when
	// running on the simulation test network since it is only intended to
	// connect to specified peers and actively avoids advertising and
	// connecting to discovered peers.
	//
	// NOTE: This is done before rejecting peers that are too old to ensure
	// it is updated regardless in the case a new minimum protocol version is
	// enforced and the remote node has not upgraded yet.
	isInbound := sp.Inbound()
	remoteAddr := sp.NA()
	addrManager := sp.server.addrManager
	if !sp.server.cfg.PrivNet && !isInbound {
		addrManager.SetServices(remoteAddr, msg.Services)
	}

	// Ignore peers that have a protcol version that is too old.  The peer
	// negotiation logic will disconnect it after this callback returns.
	if msg.ProtocolVersion < int32(protocol.InitialProcotolVersion) {
		return nil
	}

	// Reject outbound peers that are not full nodes.
	wantServices := protocol.Full
	if !isInbound && !protocol.HasServices(msg.Services, wantServices) {
		// missingServices := wantServices & ^msg.Services
		missingServices := protocol.MissingServices(msg.Services, wantServices)
		log.Debug("Rejecting peer %s with services %v due to not "+
			"providing desired services %v", sp.Peer, msg.Services,
			missingServices)
		reason := fmt.Sprintf("required services %#x not offered",
			uint64(missingServices))
		return message.NewMsgReject(msg.Command(), message.RejectNonstandard, reason)
	}

	// Update the address manager and request known addresses from the
	// remote peer for outbound connections.  This is skipped when running
	// on the simulation test network since it is only intended to connect
	// to specified peers and actively avoids advertising and connecting to
	// discovered peers.
	if !sp.server.cfg.PrivNet && !isInbound {
		// Advertise the local address when the server accepts incoming
		// connections and it believes itself to be close to the best
		// known tip.
		if !sp.server.cfg.DisableListen && sp.server.BlockManager.IsCurrent() {
			// Get address that best matches.
			lna := addrManager.GetBestLocalAddress(remoteAddr)
			if addmgr.IsRoutable(lna) {
				// Filter addresses the peer already knows about.
				addresses := []*types.NetAddress{lna}
				sp.pushAddrMsg(addresses)
			}
		}

		// Request known addresses if the server address manager needs
		// more.
		if addrManager.NeedMoreAddresses() {
			p.QueueMessage(message.NewMsgGetAddr(), nil)
		}

		// Mark the address as a known good address.
		addrManager.Good(remoteAddr)
	}

	// Choose whether or not to relay transactions.
	sp.setDisableRelayTx(msg.DisableRelayTx)

	// Add the remote peer time as a sample for creating an offset against
	// the local clock to keep the network time in sync.
	sp.server.TimeSource.AddTimeSample(p.Addr(), msg.Timestamp)

	// Signal the block manager this peer is a new sync candidate.
	sp.server.BlockManager.NewPeer(sp.syncPeer)

	// Add valid peer to the server.
	sp.server.AddPeer(sp)
	return nil
}

// OnGetAddr is invoked when a peer receives a getaddr message and is used
// to provide the peer with known addresses from the address manager.
func (sp *serverPeer) OnGetAddr(p *peer.Peer, msg *message.MsgGetAddr) {
	// Don't return any addresses when running on the simulation test
	// network.  This helps prevent the network from becoming another
	// public test network since it will not be able to learn about other
	// peers that have not specifically been provided.
	if sp.server.cfg.PrivNet {
		return
	}

	// Do not accept getaddr requests from outbound peers.  This reduces
	// fingerprinting attacks.
	if !p.Inbound() {
		return
	}

	// Only respond with addresses once per connection.  This helps reduce
	// traffic and further reduces fingerprinting attacks.
	if sp.addrsSent {
		log.Trace("Ignoring getaddr which already sent", "peer", sp.Peer)
		return
	}
	sp.addrsSent = true

	// Get the current known addresses from the address manager.
	addrCache := sp.server.addrManager.AddressCache()

	// Push the addresses.
	sp.pushAddrMsg(addrCache)
}

// OnAddr is invoked when a peer receives an addr message and is used to
// notify the server about advertised addresses.
func (sp *serverPeer) OnAddr(p *peer.Peer, msg *message.MsgAddr) {
	// Ignore addresses when running on the simulation test network.  This
	// helps prevent the network from becoming another public test network
	// since it will not be able to learn about other peers that have not
	// specifically been provided.
	if sp.server.cfg.PrivNet {
		return
	}

	// A message that has no addresses is invalid.
	if len(msg.AddrList) == 0 {
		log.Error("Command does not contain any addresses",
			"command",msg.Command(),"peer", p)
		p.Disconnect()
		return
	}

	now := time.Now()
	for _, na := range msg.AddrList {
		// Don't add more address if we're disconnecting.
		if !p.Connected() {
			return
		}

		// Set the timestamp to 5 days ago if it's more than 24 hours
		// in the future so this address is one of the first to be
		// removed when space is needed.
		if na.Timestamp.After(now.Add(time.Minute * 10)) {
			na.Timestamp = now.Add(-1 * time.Hour * 24 * 5)
		}

		// Add address to known addresses for this peer.
		sp.addKnownAddresses([]*types.NetAddress{na})
	}

	// Add addresses to server address manager.  The address manager handles
	// the details of things such as preventing duplicate addresses, max
	// addresses, and last seen updates.
	// TODO, if need to add a time penalty
	sp.server.addrManager.AddAddresses(msg.AddrList, p.NA())
}

// OnRead is invoked when a peer receives a message and it is used to update
// the bytes received by the server.
func (sp *serverPeer) OnRead(p *peer.Peer, bytesRead int, msg message.Message, err error) {
	sp.server.AddBytesReceived(uint64(bytesRead))
}

// OnWrite is invoked when a peer sends a message and it is used to update
// the bytes sent by the server.
func (sp *serverPeer) OnWrite(p *peer.Peer, bytesWritten int, msg message.Message, err error) {
	sp.server.AddBytesSent(uint64(bytesWritten))
}

// OnBlock is invoked when a peer receives a block wire message.  It blocks
// until the network block has been fully processed.
func (sp *serverPeer) OnBlock(p *peer.Peer, msg *message.MsgBlock, buf []byte) {
	// Convert the raw MsgBlock to a types.Block which provides some
	// convenience methods and things such as hash caching.

	block := types.NewBlockFromBlockAndBytes(msg.Block, buf)

	// Add the block to the known inventory for the peer.
	iv := message.NewInvVect(message.InvTypeBlock, block.Hash())
	p.AddKnownInventory(iv)

	// Queue the block up to be handled by the block manager and
	// intentionally block further receives until the network block is fully
	// processed and known good or bad.  This helps prevent a malicious peer
	// from queuing up a bunch of bad blocks before disconnecting (or being
	// disconnected) and wasting memory.  Additionally, this behavior is
	// depended on by at least the block acceptance test tool as the
	// reference implementation processes blocks in the same thread and
	// therefore blocks further messages until the network block has been
	// fully processed.
	sp.server.BlockManager.QueueBlock(block, sp.syncPeer)
	<-sp.syncPeer.BlockProcessed
}

// OnGetBlocks is invoked when a peer receives a getblocks wire message.
func (sp *serverPeer) OnGetBlocks(p *peer.Peer, msg *message.MsgGetBlocks) {
	// Find the most recent known block in the best chain based on the block
	// locator and fetch all of the block hashes after it until either
	// wire.MaxBlocksPerMsg have been fetched or the provided stop hash is
	// encountered.
	//
	// Use the block after the genesis block if no other blocks in the
	// provided locator are known.  This does mean the client will start
	// over with the genesis block if unknown block locators are provided.
	chain := sp.server.BlockManager.GetChain()
	hashList := chain.LocateBlocks(msg.BlockLocatorHashes, &msg.HashStop,
		message.MaxBlocksPerMsg)

	// Generate inventory message.
	invMsg := message.NewMsgInv()
	for i := range hashList {
		iv := message.NewInvVect(message.InvTypeBlock, &hashList[i])
		invMsg.AddInvVect(iv)
	}

	// Send the inventory message if there is anything to send.
	if len(invMsg.InvList) > 0 {
		invListLen := len(invMsg.InvList)
		if invListLen == message.MaxBlocksPerMsg {
			// Intentionally use a copy of the final hash so there
			// is not a reference into the inventory slice which
			// would prevent the entire slice from being eligible
			// for GC as soon as it's sent.
			continueHash := invMsg.InvList[invListLen-1].Hash
			sp.continueHash = &continueHash
		}
		p.QueueMessage(invMsg, nil)
	}
}

// OnInv is invoked when a peer receives an inv  message and is used to
// examine the inventory being advertised by the remote peer and react
// accordingly.  We pass the message down to blockmanager which will call
// QueueMessage with any appropriate responses.
func (sp *serverPeer) OnInv(p *peer.Peer, msg *message.MsgInv) {
	if !sp.server.cfg.BlocksOnly {
		if len(msg.InvList) > 0 {
			sp.server.BlockManager.QueueInv(msg, sp.syncPeer)
		}
		return
	}

	newInv := message.NewMsgInvSizeHint(uint(len(msg.InvList)))
	for _, invVect := range msg.InvList {
		if invVect.Type == message.InvTypeTx {
			log.Info(fmt.Sprintf("Peer %v is announcing transactions -- "+
				"disconnecting", p))
			p.Disconnect()
			return
		}
		err := newInv.AddInvVect(invVect)
		if err != nil {
			log.Error("Failed to add inventory vector", "error",err)
			break
		}
	}

	if len(newInv.InvList) > 0 {
		sp.server.BlockManager.QueueInv(newInv, sp.syncPeer)
	}
}
