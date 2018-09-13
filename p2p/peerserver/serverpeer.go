package peerserver

import (
	"sync"
	"github.com/noxproject/nox/p2p"
	"github.com/noxproject/nox/core/message"
	"github.com/noxproject/nox/common/hash"
	"time"
)

// serverPeer extends the peer to maintain state shared by the server and
// the blockmanager.
type serverPeer struct {
	*p2p.Peer

	persistent      bool
	continueHash    *hash.Hash
	relayMtx        sync.Mutex
	disableRelayTx  bool
	isWhitelisted   bool
	requestQueue    []*message.InvVect
	requestedTxns   map[hash.Hash]struct{}
	requestedBlocks map[hash.Hash]struct{}
	knownAddresses  map[string]struct{}
	quit            chan struct{}

	// addrsSent tracks whether or not the peer has responded to a getaddr
	// request.  It is used to prevent more than one response per connection.
	addrsSent bool

	// The following chans are used to sync blockmanager and server.
	txProcessed    chan struct{}
	blockProcessed chan struct{}
}
// peerState maintains state of inbound, persistent, outbound peers as well
// as banned peers and outbound groups.
type peerState struct {
	inboundPeers    map[int32]*serverPeer
	outboundPeers   map[int32]*serverPeer
	persistentPeers map[int32]*serverPeer
	banned          map[string]time.Time
	outboundGroups  map[string]int
}

