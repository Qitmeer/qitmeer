package peers

import (
	"github.com/Qitmeer/qng-core/common/hash"
	"github.com/Qitmeer/qng-core/meerdag"
	"github.com/Qitmeer/qng-core/core/protocol"
	"github.com/Qitmeer/qng-core/params"
	"github.com/libp2p/go-libp2p-core/network"
	"time"
)

// StatsSnap is a snapshot of peer stats at a point in time.
type StatsSnap struct {
	NodeID        string
	PeerID        string
	QNR           string
	Address       string
	Protocol      uint32
	Genesis       *hash.Hash
	Services      protocol.ServiceFlag
	Name          string
	Version       string
	Network       string
	State         PeerConnectionState
	Direction     network.Direction
	GraphState    *meerdag.GraphState
	GraphStateDur time.Duration
	TimeOffset    int64
	ConnTime      time.Duration
	LastSend      time.Time
	LastRecv      time.Time
	BytesSent     uint64
	BytesRecv     uint64
	IsCircuit     bool
	Bads          int
}

func (p *StatsSnap) IsRelay() bool {
	return protocol.HasServices(protocol.ServiceFlag(p.Services), protocol.Relay)
}

func (p *StatsSnap) IsTheSameNetwork() bool {
	return params.ActiveNetParams.Name == p.Network
}
