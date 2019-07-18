// Copyright (c) 2017-2018 The nox developers
// Copyright (c) 2013-2016 The btcsuite developers
// Copyright (c) 2015-2018 The Decred developers
// Use of this source code is governed by an ISC
// license that can be found in the LICENSE file.
package peerserver

import (
	"fmt"
	"github.com/HalalChain/qitmeer-lib/common/hash"
	"github.com/HalalChain/qitmeer-lib/core/message"
	"github.com/HalalChain/qitmeer-lib/core/protocol"
	"github.com/HalalChain/qitmeer-lib/core/types"
	"github.com/HalalChain/qitmeer-lib/log"
	"github.com/HalalChain/qitmeer-lib/params/dcr/types"
	"github.com/HalalChain/qitmeer/p2p/addmgr"
	"github.com/HalalChain/qitmeer/p2p/peer"
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
		log.Debug(fmt.Sprintf("Rejecting peer %s with services %v due to not "+
			"providing desired services %v", sp.Peer, msg.Services,
			missingServices))
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
	log.Trace("OnVersion -> NewPeer send to blkMgr msgChan", "peer", sp.syncPeer)
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
	log.Trace("OnBlock called", "peer",p,  "block", msg)
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
	log.Trace("OnBlock done, sp.syncPeer.BlockProcessed")
}

// OnGetBlocks is invoked when a peer receives a getblocks wire message.
func (sp *serverPeer) OnGetBlocks(p *peer.Peer, msg *message.MsgGetBlocks) {
	// Find the most recent known block in the best chain based on the block
	// locator and fetch all of the block hashes after it until either
	// wire.MaxBlocksPerMsg have been fetched or the provided stop hash is
	// encountered.
	//
	p.UpdateLastGS(msg.GS)
	// Use the block after the genesis block if no other blocks in the
	// provided locator are known.  This does mean the client will start
	// over with the genesis block if unknown block locators are provided.
	chain := sp.server.BlockManager.GetChain()
	hashSlice:=[]*hash.Hash{}
	if len(msg.BlockLocatorHashes)>0 {
		for _,v:=range msg.BlockLocatorHashes{
			if chain.BlockDAG().HasBlock(v) {
				hashSlice=append(hashSlice,v)
			}
		}
	}else {
		hashSlice = chain.LocateBlocks(msg.GS,message.MaxBlocksPerMsg)
	}

	if len(hashSlice)==0 {
		return
	}

	hashSlice=chain.BlockDAG().SortBlock(hashSlice)
	// Generate inventory message.
	invMsg := message.NewMsgInv()
	invMsg.GS=chain.BestSnapshot().GraphState
	for i := range hashSlice {
		iv := message.NewInvVect(message.InvTypeBlock, hashSlice[i])
		invMsg.AddInvVect(iv)
	}

	// Send the inventory message if there is anything to send.
	if len(invMsg.InvList) > 0 {
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
	newInv.GS=msg.GS
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

// handleGetData is invoked when a peer receives a getdata wire message and is
// used to deliver block and transaction information.
func (sp *serverPeer) OnGetData(p *peer.Peer, msg *message.MsgGetData) {
	// Ignore empty getdata messages.
	if len(msg.InvList) == 0 {
		return
	}

	numAdded := 0
	notFound := message.NewMsgNotFound()

	length := len(msg.InvList)
	// A decaying ban score increase is applied to prevent exhausting resources
	// with unusually large inventory queries.
	// Requesting more than the maximum inventory vector length within a short
	// period of time yields a score above the default ban threshold. Sustained
	// bursts of small requests are not penalized as that would potentially ban
	// peers performing IBD.
	// This incremental score decays each minute to half of its value.
	sp.addBanScore(0, uint32(length)*99/wire.MaxInvPerMsg, "getdata")

	// We wait on this wait channel periodically to prevent queuing
	// far more data than we can send in a reasonable time, wasting memory.
	// The waiting occurs after the database fetch for the next one to
	// provide a little pipelining.
	var waitChan chan struct{}
	doneChan := make(chan struct{}, 1)

	for i, iv := range msg.InvList {
		var c chan struct{}
		// If this will be the last message we send.
		if i == length-1 && len(notFound.InvList) == 0 {
			c = doneChan
		} else if (i+1)%3 == 0 {
			// Buffered so as to not make the send goroutine block.
			c = make(chan struct{}, 1)
		}
		var err error
		switch iv.Type {
		case message.InvTypeTx:
			err = sp.server.pushTxMsg(sp, &iv.Hash, c, waitChan)
		case message.InvTypeBlock:
			err = sp.server.pushBlockMsg(sp, &iv.Hash, c, waitChan)
		default:
			log.Warn("Unknown type in inventory request", "type",iv.Type)
			continue
		}
		if err != nil {
			notFound.AddInvVect(iv)

			// When there is a failure fetching the final entry
			// and the done channel was sent in due to there
			// being no outstanding not found inventory, consume
			// it here because there is now not found inventory
			// that will use the channel momentarily.
			if i == len(msg.InvList)-1 && c != nil {
				<-c
			}
		}
		numAdded++
		waitChan = c
	}
	if len(notFound.InvList) != 0 {
		p.QueueMessage(notFound, doneChan)
	}

	// Wait for messages to be sent. We can send quite a lot of data at this
	// point and this will keep the peer busy for a decent amount of time.
	// We don't process anything else by them in this time so that we
	// have an idea of when we should hear back from them - else the idle
	// timeout could fire when we were only half done sending the blocks.
	if numAdded > 0 {
		<-doneChan
	}
}

// OnGetMiningState is invoked when a peer receives a getminings wire message.
// It constructs a list of the current best blocks and votes that should be
// mined on and pushes a miningstate wire message back to the requesting peer.
func (sp *serverPeer) OnGetMiningState(p *peer.Peer, msg *message.MsgGetMiningState) {
	// Access the block manager and get the list of best blocks to mine on.
	bm := sp.server.BlockManager
	best := bm.GetChain().BestSnapshot()

	// Obtain the entire generation of blocks stemming from the parent of
	// the current tip.
	children, err := bm.TipGeneration()
	if err != nil {
		log.Warn(fmt.Sprintf("failed to access block manager to get the generation "+
			"for a mining state request (block: %v): %v", best.Hash, err))
		return
	}

	// Get the list of blocks of blocks that are eligible to built on and
	// limit the list to the maximum number of allowed eligible block hashes
	// per mining state message.  There is nothing to send when there are no
	// eligible blocks.

	blockHashes := children   // TODO, the children should be sorted by rules
	numBlocks := len(blockHashes)
	if numBlocks == 0 {
		return
	}
	if numBlocks > message.MaxMSBlocksAtHeadPerMsg {
		blockHashes = blockHashes[:message.MaxMSBlocksAtHeadPerMsg]
	}

	err = sp.pushMiningStateMsg(uint32(best.Order), blockHashes)
	if err != nil {
		log.Warn(fmt.Sprintf("unexpected error while pushing data for "+
			"mining state request: %v", err.Error()))
	}
}

// OnMiningState is invoked when a peer receives a miningstate wire message.  It
// requests the data advertised in the message from the peer.
func (sp *serverPeer) OnMiningState(p *peer.Peer, msg *message.MsgMiningState) {
	err := sp.server.BlockManager.RequestFromPeer(sp.syncPeer, msg.BlockHashes)
	if err != nil {
		log.Warn("couldn't handle mining state message", "error",err.Error())
	}
}

// OnTx is invoked when a peer receives a tx message.  It blocks until the
// transaction has been fully processed.  Unlock the block handler this does not
// serialize all transactions through a single thread transactions don't rely on
// the previous one in a linear fashion like blocks.
func (sp *serverPeer) OnTx(p *peer.Peer, msg *message.MsgTx) {
	log.Trace("OnTx called, peer received tx message", "peer",p, "msg",msg)
	if sp.server.cfg.BlocksOnly {
		log.Trace(fmt.Sprintf("Ignoring tx %v from %v - blocksonly enabled",
			msg.Tx.TxHash(), p))
		return
	}

	// Add the transaction to the known inventory for the peer.
	// Convert the raw MsgTx to a dcrutil.Tx which provides some convenience
	// methods and things such as hash caching.
	tx := types.NewTx(msg.Tx)
	iv := message.NewInvVect(message.InvTypeTx, tx.Hash())
	p.AddKnownInventory(iv)

	// Queue the transaction up to be handled by the block manager and
	// intentionally block further receives until the transaction is fully
	// processed and known good or bad.  This helps prevent a malicious peer
	// from queuing up a bunch of bad transactions before disconnecting (or
	// being disconnected) and wasting memory.
	sp.server.BlockManager.QueueTx(tx, sp.syncPeer)
	<-sp.syncPeer.TxProcessed
}
