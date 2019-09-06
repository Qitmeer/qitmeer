// Copyright (c) 2017-2018 The qitmeer developers
// Copyright (c) 2013-2016 The btcsuite developers
// Copyright (c) 2015-2018 The Decred developers
// Use of this source code is governed by an ISC
// license that can be found in the LICENSE file.
package peerserver

import (
	"github.com/Qitmeer/qitmeer-lib/core/message"
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
