// Copyright (c) 2017-2018 The qitmeer developers
// Copyright (c) 2013-2016 The btcsuite developers
// Copyright (c) 2016-2018 The Decred developers
// Use of this source code is governed by an ISC
// license that can be found in the LICENSE file.
package peer

import (
	"github.com/Qitmeer/qitmeer/common/hash"
	"github.com/Qitmeer/qitmeer/core/message"
)

// ServerPeer extends the peer to maintain state shared by the p2p server and
// the blockmanager.
type ServerPeer struct {
	// The following chans are used to sync blockmanager and server.
	*Peer
	TxProcessed     chan struct{}
	BlockProcessed  chan int
	RequestedBlocks map[hash.Hash]struct{}
	RequestedTxns   map[hash.Hash]struct{}
	RequestQueue    []*message.InvVect
	SyncCandidate   bool
}
