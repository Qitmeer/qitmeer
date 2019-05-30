package blkmgr

import (
	"fmt"
	"qitmeer/common/hash"
	"qitmeer/core/blockchain"
	"qitmeer/core/message"
	"qitmeer/database"
	"qitmeer/services/mempool"
)
const(

	// maxResendLimit is the maximum number of times a node can resend a
	// block or transaction before it is dropped.
	maxResendLimit = 3

	// minInFlightBlocks is the minimum number of blocks that should be
	// in the request queue for headers-first mode before requesting
	// more.
	minInFlightBlocks = 10
)

// handleBlockMsg handles block messages from all peers.
func (b *BlockManager) handleBlockMsg(bmsg *blockMsg) {
	log.Trace("handleBlockMsg called", "bmsg",bmsg)
	// If we didn't ask for this block then the peer is misbehaving.
	blockHash := bmsg.block.Hash()
	if _, exists := bmsg.peer.RequestedBlocks[*blockHash]; !exists {
		// Check to see if we ever requested this block, since it may
		// have been accidentally sent in duplicate. If it was,
		// increment the counter in the ever requested map and make
		// sure that the node isn't spamming us with these blocks.
		received, exists := b.requestedEverBlocks[*blockHash]
		if exists {
			if received > maxResendLimit {
				log.Warn(fmt.Sprintf("Got duplicate block %v from %s -- "+
					"too many times, disconnecting",
					blockHash, bmsg.peer.Addr()))
				bmsg.peer.Disconnect()
				return
			}
			b.requestedEverBlocks[*blockHash]++
		} else {
			log.Warn(fmt.Sprintf("Got unrequested block %v from %s -- "+
				"disconnecting", blockHash, bmsg.peer.Addr()))
			bmsg.peer.Disconnect()
			return
		}
	}

	// When in headers-first mode, if the block matches the hash of the
	// first header in the list of headers that are being fetched, it's
	// eligible for less validation since the headers have already been
	// verified to link together and are valid up to the next checkpoint.
	// Also, remove the list entry for all blocks except the checkpoint
	// since it is needed to verify the next round of headers links
	// properly.
	isCheckpointBlock := false
	behaviorFlags := blockchain.BFNone
	if b.headersFirstMode {
		firstNodeEl := b.headerList.Front()
		if firstNodeEl != nil {
			firstNode := firstNodeEl.Value.(*headerNode)
			if blockHash.IsEqual(firstNode.hash) {
				behaviorFlags |= blockchain.BFFastAdd
				if firstNode.hash.IsEqual(b.nextCheckpoint.Hash) {
					isCheckpointBlock = true
				} else {
					b.headerList.Remove(firstNodeEl)
				}
			}
		}
	}

	// Remove block from request maps. Either chain will know about it and
	// so we shouldn't have any more instances of trying to fetch it, or we
	// will fail the insert and thus we'll retry next time we get an inv.
	delete(bmsg.peer.RequestedBlocks, *blockHash)
	delete(b.requestedBlocks, *blockHash)

	// Process the block to include validation, best chain selection, orphan
	// handling, etc.
	_, isOrphan, err := b.chain.ProcessBlock(bmsg.block,
		behaviorFlags)
	if err != nil {
		// When the error is a rule error, it means the block was simply
		// rejected as opposed to something actually going wrong, so log
		// it as such.  Otherwise, something really did go wrong, so log
		// it as an actual error.
		if _, ok := err.(blockchain.RuleError); ok {
			log.Info("Rejected block", "hash",blockHash, "peer",
				bmsg.peer, "error",err)
		} else {
			log.Error("Failed to process block", "hash",
				blockHash, "error",err)
		}
		if dbErr, ok := err.(database.Error); ok && dbErr.ErrorCode ==
			database.ErrCorruption {
			log.Error("Critical failure", "error",dbErr.Error())
		}

		// Convert the error into an appropriate reject message and
		// send it.
		code, reason := mempool.ErrToRejectErr(err)
		bmsg.peer.PushRejectMsg(message.CmdBlock, code, reason,
			blockHash, false)
		return
	}

	// Meta-data about the new block this peer is reporting. We use this
	// below to update this peer's lastest block height and the heights of
	// other peers based on their last announced block hash. This allows us
	// to dynamically update the block heights of peers, avoiding stale
	// heights when looking for a new sync peer. Upon acceptance of a block
	// or recognition of an orphan, we also use this information to update
	// the block heights over other peers who's invs may have been ignored
	// if we are actively syncing while the chain is not yet current or
	// who may have lost the lock announcment race.

	// Notify stake difficulty subscribers and prune invalidated
	// transactions.
	best := b.chain.BestSnapshot()
	// Request the parents for the orphan block from the peer that sent it.
	if isOrphan {
		// We've just received an orphan block from a peer. In order
		// to update the height of the peer, we try to extract the
		// block height from the scriptSig of the coinbase transaction.
		// Extraction is only attempted if the block's version is
		// high enough (ver 2+).

		locator:=b.chain.GetOrphanParents(blockHash)
		if len(locator)>0 {
			err = bmsg.peer.PushGetBlocksMsg(best.GraphState,locator)
			if err != nil {
				log.Warn("Failed to push getblocksmsg for the orphan block", "error",err)
			}
		}
	} else {
		// When the block is not an orphan, log information about it and
		// update the chain state.
		b.progressLogger.LogBlockHeight(bmsg.block)

		b.txMemPool.PruneExpiredTx()


		// Clear the rejected transactions.
		b.rejectedTxns = make(map[hash.Hash]struct{})

		// Allow any clients performing long polling via the
		// getblocktemplate RPC to be notified when the new block causes
		// their old block template to become stale.
		// TODO, refactor how bm work with rpc-server
		/*
		rpcServer := b.server.rpcServer
		if rpcServer != nil {
			rpcServer.gbtWorkState.NotifyBlockConnected(blockHash)
		}
		*/
		if !b.current()&&bmsg.peer==b.syncPeer {
			if len(bmsg.peer.RequestedBlocks)==0 {
				err = bmsg.peer.PushGetBlocksMsg(best.GraphState,nil)
				if err != nil {
					log.Warn("Failed to push getblocksmsg for the last block", "error",err)
				}
			}
		}
	}

	// Nothing more to do if we aren't in headers-first mode.
	if !b.headersFirstMode {
		log.Trace("handleBlockMsg done", "headerFist", b.headersFirstMode)
		return
	}

	// This is headers-first mode, so if the block is not a checkpoint
	// request more blocks using the header list when the request queue is
	// getting short.
	if !isCheckpointBlock {
		if b.startHeader != nil &&
			len(bmsg.peer.RequestedBlocks) < minInFlightBlocks {
			b.fetchHeaderBlocks()
		}
		return
	}

	// This is headers-first mode and the block is a checkpoint.  When
	// there is a next checkpoint, get the next round of headers by asking
	// for headers starting from the block after this one up to the next
	// checkpoint.
	prevHeight := b.nextCheckpoint.Height
	prevHash := b.nextCheckpoint.Hash
	b.nextCheckpoint = b.findNextHeaderCheckpoint(prevHeight)
	if b.nextCheckpoint != nil {
		locator := blockchain.BlockLocator([]*hash.Hash{prevHash})
		err := bmsg.peer.PushGetHeadersMsg(locator, b.nextCheckpoint.Hash)
		if err != nil {
			log.Warn("Failed to send getheaders message",
				"peer", bmsg.peer.Addr(),"error",err)
			return
		}
		log.Info(fmt.Sprintf("Downloading headers for blocks %d to %d from "+
			"peer %s", prevHeight+1, b.nextCheckpoint.Height,
			b.syncPeer.Addr()))
		return
	}

	// This is headers-first mode, the block is a checkpoint, and there are
	// no more checkpoints, so switch to normal mode by requesting blocks
	// from the block after this one up to the end of the chain (zero hash).
	b.headersFirstMode = false
	b.headerList.Init()
	log.Info("Reached the final checkpoint -- switching to normal mode")
	err = bmsg.peer.PushGetBlocksMsg(best.GraphState,nil)
	if err != nil {
		log.Warn("Failed to send getblocks message",
			"peer",bmsg.peer.Addr(), "error",err)
		return
	}
}
