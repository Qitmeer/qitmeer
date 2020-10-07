package peers

import (
	"github.com/Qitmeer/qitmeer/common/hash"
	"github.com/Qitmeer/qitmeer/core/blockdag"
	"github.com/Qitmeer/qitmeer/core/protocol"
	"github.com/libp2p/go-libp2p-core/network"
)

// StatsSnap is a snapshot of peer stats at a point in time.
type StatsSnap struct {
	NodeID     string
	PeerID     string
	QNR        string
	Protocol   uint32
	Genesis    *hash.Hash
	Services   protocol.ServiceFlag
	UserAgent  string
	State      PeerConnectionState
	Direction  network.Direction
	GraphState *blockdag.GraphState
}
