// Copyright (c) 2017-2018 The nox developers
// Copyright (c) 2013-2016 The btcsuite developers
// Copyright (c) 2015-2018 The Decred developers
// Use of this source code is governed by an ISC
// license that can be found in the LICENSE file.
package peerserver

import (
	"github.com/noxproject/nox/common/hash"
	"github.com/noxproject/nox/core/message"
)

// broadcastMsg provides the ability to house a  message to be broadcast
// to all connected peers except specified excluded peers.
type broadcastMsg struct {
	message      message.Message
	excludePeers []*serverPeer
}

// broadcastInventoryAdd is a type used to declare that the InvVect it contains
// needs to be added to the rebroadcast map
type broadcastInventoryAdd relayMsg

// broadcastInventoryDel is a type used to declare that the InvVect it contains
// needs to be removed from the rebroadcast map
type broadcastInventoryDel *message.InvVect

// relayMsg packages an inventory vector along with the newly discovered
// inventory and a flag that determines if the relay should happen immediately
// (it will be put into a trickle queue if false) so the relay has access to
// that information.
type relayMsg struct {
	invVect   *message.InvVect
	data      interface{}
	immediate bool
}

// updatePeerHeightsMsg is a message sent from the blockmanager to the server
// after a new block has been accepted. The purpose of the message is to update
// the heights of peers that were known to announce the block before we
// connected it to the main chain or recognized it as an orphan. With these
// updates, peer heights will be kept up to date, allowing for fresh data when
// selecting sync peer candidacy.
type updatePeerHeightsMsg struct {
	newHash    *hash.Hash
	newHeight  uint64
	originPeer *serverPeer
}

