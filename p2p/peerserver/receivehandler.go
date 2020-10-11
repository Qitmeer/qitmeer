// Copyright (c) 2017-2018 The qitmeer developers
// Copyright (c) 2013-2016 The btcsuite developers
// Copyright (c) 2015-2018 The Decred developers
// Use of this source code is governed by an ISC
// license that can be found in the LICENSE file.
package peerserver

import (
	"fmt"
	"github.com/Qitmeer/qitmeer/common/hash"
	"github.com/Qitmeer/qitmeer/core/blockdag"
	"github.com/Qitmeer/qitmeer/core/message"
	"github.com/Qitmeer/qitmeer/core/protocol"
	"github.com/Qitmeer/qitmeer/core/types"
	"github.com/Qitmeer/qitmeer/log"
	"github.com/Qitmeer/qitmeer/p2p/connmgr"
	"github.com/Qitmeer/qitmeer/p2p/peer"
)

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

// OnGetBlocks is invoked when a peer receives a getblocks wire message.
func (sp *serverPeer) OnGetBlocks(p *peer.Peer, msg *message.MsgGetBlocks) {
	if msg.GS.IsGenesis() && !msg.GS.GetTips().HasOnly(sp.server.chainParams.GenesisHash) {
		sp.addBanScore(0, connmgr.SeriousScore, "ongetblocks")
		log.Warn(fmt.Sprintf("Wrong genesis(%s) from peer(%s),your genesis is %s",
			msg.GS.GetTips().List()[0].String(), p.String(), sp.server.chainParams.GenesisHash.String()))
		return
	}
	// Find the most recent known block in the best chain based on the block
	// locator and fetch all of the block hashes after it until either
	// wire.MaxBlocksPerMsg have been fetched or the provided stop hash is
	// encountered.
	//
	sp.UpdateLastGS(p, msg.GS)
	// Use the block after the genesis block if no other blocks in the
	// provided locator are known.  This does mean the client will start
	// over with the genesis block if unknown block locators are provided.
	chain := sp.server.BlockManager.GetChain()
	dagSync := sp.server.BlockManager.DAGSync()
	gs := chain.BestSnapshot().GraphState
	blocks, _ := dagSync.CalcSyncBlocks(msg.GS, msg.BlockLocatorHashes, blockdag.DirectMode, message.MaxBlockLocatorsPerMsg)
	hsLen := len(blocks)
	if hsLen == 0 {
		log.Trace(fmt.Sprintf("Sorry, there are not these blocks for %s", p.String()))

		rMsg := message.NewMsgSyncResult(gs.Clone(), blockdag.SubDAGMode)
		p.QueueMessage(rMsg, nil)
		return
	}

	invMsg := message.NewMsgInv()
	invMsg.GS = gs
	for i := 0; i < hsLen; i++ {
		iv := message.NewInvVect(message.InvTypeBlock, blocks[i])
		invMsg.AddInvVect(iv)
	}
	if len(invMsg.InvList) > 0 {
		p.QueueMessage(invMsg, nil)
	}

}

// OnGetHeaders is invoked when a peer receives a getheaders
// message.
func (sp *serverPeer) OnGetHeaders(p *peer.Peer, msg *message.MsgGetHeaders) {
	// Ignore getheaders requests if not in sync.
	if !sp.server.BlockManager.IsCurrent() {
		return
	}

	sp.UpdateLastGS(p, msg.GS)
	chain := sp.server.BlockManager.GetChain()
	hashSlice := []*hash.Hash{}
	if len(msg.BlockLocatorHashes) > 0 {
		for _, v := range msg.BlockLocatorHashes {
			if chain.BlockDAG().HasBlock(v) {
				hashSlice = append(hashSlice, v)
			}
		}
		if len(hashSlice) >= 2 {
			hashSlice = chain.BlockDAG().SortBlock(hashSlice)
		}
	}
	hsLen := len(hashSlice)
	if hsLen == 0 {
		log.Trace(fmt.Sprintf("Sorry, there are not these blocks for %s", p.String()))
		return
	}

	headersMsg := message.NewMsgHeaders(chain.BestSnapshot().GraphState)
	for i := 0; i < hsLen; i++ {
		blockHead, err := chain.HeaderByHash(hashSlice[i])
		if err != nil {
			log.Trace(fmt.Sprintf("Sorry, there are not these blocks %s for %s", hashSlice[i].String(), p.String()))
			return
		}
		headersMsg.AddBlockHeader(&blockHead)
	}
	if len(headersMsg.Headers) > 0 {
		p.QueueMessage(headersMsg, nil)
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
	sp.addBanScore(0, uint32(length)*99/message.MaxInvPerMsg, "getdata")

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
		default:
			log.Warn("Unknown type in inventory request", "type", iv.Type)
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

	blockHashes := children // TODO, the children should be sorted by rules
	numBlocks := len(blockHashes)
	if numBlocks == 0 {
		return
	}
	if numBlocks > message.MaxMSBlocksAtHeadPerMsg {
		blockHashes = blockHashes[:message.MaxMSBlocksAtHeadPerMsg]
	}

	err = sp.pushMiningStateMsg(uint32(best.GraphState.GetTotal()), blockHashes)
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
		log.Warn("couldn't handle mining state message", "error", err.Error())
	}
}

// OnTx is invoked when a peer receives a tx message.  It blocks until the
// transaction has been fully processed.  Unlock the block handler this does not
// serialize all transactions through a single thread transactions don't rely on
// the previous one in a linear fashion like blocks.
func (sp *serverPeer) OnTx(p *peer.Peer, msg *message.MsgTx) {
	log.Trace("OnTx called, peer received tx message", "peer", p, "msg", msg)
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

// OnMemPool
func (sp *serverPeer) OnMemPool(_ *peer.Peer, msg *message.MsgMemPool) {
	if sp.server.services&protocol.Bloom != protocol.Bloom {
		log.Debug(fmt.Sprintf("peer %v sent mempool request with bloom filtering disabled -- disconnecting", sp))
		sp.Disconnect()
		return
	}

	sp.addBanScore(0, 33, "mempool")

	txMemPool := sp.server.TxMemPool
	txDescs := txMemPool.TxDescs()
	invMsg := message.NewMsgInvSizeHint(uint(len(txDescs)))

	for _, txDesc := range txDescs {
		iv := message.NewInvVect(message.InvTypeTx, txDesc.Tx.Hash())
		invMsg.AddInvVect(iv)
		if len(invMsg.InvList)+1 > message.MaxInvPerMsg {
			break
		}
	}

	if len(invMsg.InvList) > 0 {
		sp.QueueMessage(invMsg, nil)
	}
}
