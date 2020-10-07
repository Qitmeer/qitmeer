// Copyright (c) 2017-2018 The qitmeer developers
// Copyright (c) 2016 The btcsuite developers
// Copyright (c) 2017 The Decred developers
// Use of this source code is governed by an ISC
// license that can be found in the LICENSE file.
package blkmgr

import (
	"fmt"
	"github.com/Qitmeer/qitmeer/core/message"
	"github.com/Qitmeer/qitmeer/p2p/peer"
	"time"
)

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
	log.Info("Lost peer", "peer", sp)

	b.clearRequestedState(sp)
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
