// Copyright (c) 2017-2018 The qitmeer developers
// Copyright (c) 2016 The btcsuite developers
// Copyright (c) 2017 The Decred developers
// Use of this source code is governed by an ISC
// license that can be found in the LICENSE file.
package blkmgr

import (
	"container/list"
	"fmt"
	"github.com/HalalChain/qitmeer-lib/common/hash"
	"github.com/HalalChain/qitmeer-lib/core/message"
	"github.com/HalalChain/qitmeer-lib/core/protocol"
	"github.com/HalalChain/qitmeer/p2p/peer"
	"sync/atomic"
	"time"
)

// handleNewPeerMsg deals with new peers that have signalled they may
// be considered as a sync peer (they have already successfully negotiated).  It
// also starts syncing if needed.  It is invoked from the syncHandler goroutine.
func (b *BlockManager) handleNewPeerMsg(peers *list.List, sp *peer.ServerPeer) {
	// Ignore if in the process of shutting down.
	if atomic.LoadInt32(&b.shutdown) != 0 {
		return
	}

	log.Info(fmt.Sprintf("New valid peer: %s,user-agent:%s",sp,sp.UserAgent()))

	// Ignore the peer if it's not a sync candidate.
	if !b.isSyncCandidate(sp) {
		return
	}

	// Add the peer as a candidate to sync from.
	peers.PushBack(sp)

	// Start syncing by choosing the best candidate if needed.
	if b.syncPeer == nil {
		b.startSync()
	}
	// Grab the mining state from this peer after we're synced.
	if b.config.MiningStateSync {
		b.syncMiningStateAfterSync(sp)
	}
}

// handleDonePeerMsg deals with peers that have signalled they are done.  It
// removes the peer as a candidate for syncing and in the case where it was
// the current sync peer, attempts to select a new best peer to sync from.  It
// is invoked from the syncHandler goroutine.
func (b *BlockManager) handleDonePeerMsg(peers *list.List, sp *peer.ServerPeer) {
	// Remove the peer from the list of candidate peers.
	hasRemove:=false
	for e := peers.Front(); e != nil; e = e.Next() {
		if e.Value == sp {
			peers.Remove(e)
			hasRemove=true
			break
		}
	}
	if !hasRemove {
		log.Warn(fmt.Sprintf("Received done peer message for unknown peer %s", sp))
		return
	}
	log.Info("Lost peer", "peer",sp)

	b.clearRequestedState(sp)

	if b.syncPeer == sp {
		// Update the sync peer. The server has already disconnected the
		// peer before signaling to the sync manager.
		b.updateSyncPeer(false)
	}
}

// isSyncCandidate returns whether or not the peer is a candidate to consider
// syncing from.
func (b *BlockManager) isSyncCandidate(sp *peer.ServerPeer) bool {
	// The peer is not a candidate for sync if it's not a full node.
	return sp.Services()&protocol.Full == protocol.Full
}

// syncMiningStateAfterSync polls the blockMananger for the current sync
// state; if the mananger is synced, it executes a call to the peer to
// sync the mining state to the network.
func (b *BlockManager) syncMiningStateAfterSync(sp *peer.ServerPeer) {
	go func() {
		for {
			time.Sleep(3 * time.Second)
			if !sp.Connected() {
				return
			}
			if b.IsCurrent() {
				msg := message.NewMsgGetMiningState()
				sp.QueueMessage(msg, nil)
				return
			}
		}
	}()
}

// getSyncPeerMsg is a message type to be sent across the message channel for
// retrieving the current sync peer.
type getSyncPeerMsg struct {
	reply chan int32
}

// startSync will choose the best peer among the available candidate peers to
// download/sync the blockchain from.  When syncing is already running, it
// simply returns.  It also examines the candidates for any which are no longer
// candidates and removes them as needed.
func (b *BlockManager) startSync() {
	// Return now if we're already syncing.
	if b.syncPeer != nil {
		return
	}

	best := b.chain.BestSnapshot()
	var bestPeer *peer.ServerPeer
	var enext *list.Element
	for e := b.candidatePeers.Front(); e != nil; e = enext {
		enext = e.Next()
		sp := e.Value.(*peer.ServerPeer)

		// Remove sync candidate peers that are no longer candidates due
		// to passing their latest known block.  NOTE: The < is
		// intentional as opposed to <=.  While techcnically the peer
		// doesn't have a later block when it's equal, it will likely
		// have one soon so it is a reasonable choice.  It also allows
		// the case where both are at 0 such as during regression test.
		if best.GraphState.IsExcellent(sp.LastGS()) {
			b.candidatePeers.Remove(e)
			continue
		}

		// the best sync candidate is the most updated peer
		if bestPeer == nil {
			bestPeer = sp
		}
		if sp.LastGS().IsExcellent(bestPeer.LastGS()) {
			bestPeer = sp
		}
	}

	// Start syncing from the best peer if one was selected.
	if bestPeer != nil {
		// Clear the requestedBlocks if the sync peer changes, otherwise
		// we may ignore blocks we need that the last sync peer failed
		// to send.
		b.requestedBlocks = make(map[hash.Hash]struct{})

		log.Info(fmt.Sprintf("Syncing to state %s from peer %s cur graph state:%s",bestPeer.LastGS().String(), bestPeer.Addr(),best.GraphState.String()))

		// When the current height is less than a known checkpoint we
		// can use block headers to learn about which blocks comprise
		// the chain up to the checkpoint and perform less validation
		// for them.  This is possible since each header contains the
		// hash of the previous header and a merkle root.  Therefore if
		// we validate all of the received headers link together
		// properly and the checkpoint hashes match, we can be sure the
		// hashes for the blocks in between are accurate.  Further, once
		// the full blocks are downloaded, the merkle root is computed
		// and compared against the value in the header which proves the
		// full block hasn't been tampered with.
		//
		// Once we have passed the final checkpoint, or checkpoints are
		// disabled, use standard inv messages learn about the blocks
		// and fully validate them.  Finally, regression test mode does
		// not support the headers-first approach so do normal block
		// downloads when in regression test mode.
		err := bestPeer.PushGetBlocksMsg(best.GraphState,nil)
		if err != nil {
			log.Error(fmt.Sprintf("Failed to push getblocksmsg for the "+
				"latest GS: %v", err))
			return
		}
		b.syncPeer = bestPeer
		// Reset the last progress time now that we have a non-nil
		// syncPeer to avoid instantly detecting it as stalled in the
		// event the progress time hasn't been updated recently.
		b.lastProgressTime = time.Now()

		b.syncGSMtx.Lock()
		b.syncGS = bestPeer.LastGS()
		b.syncGSMtx.Unlock()
	} else {
		log.Warn("No sync peer candidates available")
	}
}

