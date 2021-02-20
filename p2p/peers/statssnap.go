package peers

import (
	"github.com/Qitmeer/qitmeer/common/hash"
	"github.com/Qitmeer/qitmeer/core/blockdag"
	"github.com/Qitmeer/qitmeer/core/protocol"
	"github.com/libp2p/go-libp2p-core/network"
	"time"
)

// StatsSnap is a snapshot of peer stats at a point in time.
type StatsSnap struct {
	NodeID     string
	PeerID     string
	QNR        string
	Address    string
	Protocol   uint32
	Genesis    *hash.Hash
	Services   protocol.ServiceFlag
	UserAgent  string
	State      PeerConnectionState
	Direction  network.Direction
	GraphState *blockdag.GraphState
	TimeOffset int64
	ConnTime   time.Duration
	LastSend   time.Time
	LastRecv   time.Time
	BytesSent  uint64
	BytesRecv  uint64
}

func (p *StatsSnap) IsRelay() bool {
	return protocol.HasServices(protocol.ServiceFlag(p.Services), protocol.Relay)
}
