// Copyright (c) 2017-2018 The qitmeer developers
// Copyright (c) 2016 The btcsuite developers
// Copyright (c) 2017 The Decred developers
// Use of this source code is governed by an ISC
// license that can be found in the LICENSE file.
package blkmgr

import (
	"fmt"
	"github.com/Qitmeer/qitmeer/common/hash"
	"github.com/Qitmeer/qitmeer/core/message"
	"github.com/Qitmeer/qitmeer/core/protocol"
	"github.com/Qitmeer/qitmeer/p2p/peer"
	"math/rand"
	"sync/atomic"
	"time"
)

// handleNewPeerMsg deals with new peers that have signalled they may
// be considered as a sync peer (they have already successfully negotiated).  It
// also starts syncing if needed.  It is invoked from the syncHandler goroutine.
func (b *BlockManager) handleNewPeerMsg(sp *peer.ServerPeer) {
	// Ignore if in the process of shutting down.
	if atomic.LoadInt32(&b.shutdown) != 0 {
		return
	}

	log.Info(fmt.Sprintf("New valid peer: %s,user-agent:%s",sp,sp.UserAgent()))

	sp.SyncCandidate=b.isSyncCandidate(sp)
	b.peers[sp.Peer]=sp

	// Start syncing by choosing the best candidate if needed.
	if sp.SyncCandidate && b.syncPeer == nil {
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
func (b *BlockManager) handleDonePeerMsg(sp *peer.ServerPeer) {
	// Remove the peer from the list of candidate peers.
	peer, exists := b.peers[sp.Peer]
	if !exists {
		log.Warn(fmt.Sprintf("Received done peer message for unknown peer %s", sp))
		return
	}
	delete(b.peers, peer.Peer)
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
	equalPeers:=[]*peer.ServerPeer{}

	for _,sp:=range b.peers {
		if !sp.SyncCandidate {
			continue
		}
		// Remove sync candidate peers that are no longer candidates due
		// to passing their latest known block.  NOTE: The < is
		// intentional as opposed to <=.  While techcnically the peer
		// doesn't have a later block when it's equal, it will likely
		// have one soon so it is a reasonable choice.  It also allows
		// the case where both are at 0 such as during regression test.
		if best.GraphState.IsExcellent(sp.LastGS()) {
			sp.SyncCandidate=false
			continue
		}
		// the best sync candidate is the most updated peer
		if bestPeer == nil {
			bestPeer = sp
			continue
		}
		if sp.LastGS().IsExcellent(bestPeer.LastGS()) {
			bestPeer = sp
			if len(equalPeers)>0 {
				equalPeers=equalPeers[0:0]
			}
		}else if sp.LastGS().IsEqual(bestPeer.LastGS()) {
			equalPeers=append(equalPeers,sp)
		}
	}
	if len(equalPeers)>0 {
		equalPeers=append(equalPeers,bestPeer)
		bestPeer = equalPeers[rand.Intn(len(equalPeers))]
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

