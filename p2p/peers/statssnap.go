package peers

import (
	"github.com/Qitmeer/qitmeer/common/hash"
	"github.com/Qitmeer/qitmeer/core/blockdag"
	"github.com/Qitmeer/qitmeer/core/protocol"
	"github.com/Qitmeer/qitmeer/params"
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

func (p *StatsSnap) GetName() string {
	err, name, _, _ := ParseUserAgent(p.UserAgent)
	if err != nil {
		return p.UserAgent
	}
	return name
}

func (p *StatsSnap) GetVersion() string {
	err, _, version, _ := ParseUserAgent(p.UserAgent)
	if err != nil {
		return ""
	}
	return version
}

func (p *StatsSnap) GetNetwork() string {
	err, _, _, network := ParseUserAgent(p.UserAgent)
	if err != nil {
		return ""
	}
	return network
}

func (p *StatsSnap) IsTheSameNetwork() bool {
	return params.ActiveNetParams.Name == p.GetNetwork()
}
