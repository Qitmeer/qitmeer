package peers

import (
	"fmt"
	"github.com/Qitmeer/qitmeer/common/hash"
	"github.com/Qitmeer/qitmeer/core/blockdag"
	"github.com/Qitmeer/qitmeer/core/protocol"
	"github.com/Qitmeer/qitmeer/p2p/qnode"
	"github.com/Qitmeer/qitmeer/p2p/qnr"
	"github.com/libp2p/go-libp2p-core/network"
	"github.com/libp2p/go-libp2p-core/peer"
	ma "github.com/multiformats/go-multiaddr"
	"time"
)

// Peer represents a connected p2p network remote node.
type Peer struct {
	*peerStatus
	pid       peer.ID
	syncPoint *hash.Hash
}

func (p *Peer) SetQNR(record *qnr.Record) {
	p.qnr = record
}

func (p *Peer) UpdateAddrDir(record *qnr.Record, address ma.Multiaddr, direction network.Direction) {
	p.address = address
	p.direction = direction
	if record != nil {
		p.qnr = record
	}
}

func (p *Peer) StatsSnapshot() (*StatsSnap, error) {
	n, err := qnode.New(qnode.ValidSchemes, p.qnr)
	if err != nil {
		return nil, fmt.Errorf("qnode: can't verify local record: %v", err)
	}
	ss := &StatsSnap{
		NodeID:     n.ID().String(),
		PeerID:     p.pid.String(),
		QNR:        n.String(),
		Protocol:   p.ProtocolVersion(),
		Genesis:    p.Genesis(),
		Services:   p.Services(),
		UserAgent:  p.UserAgent(),
		State:      p.peerState,
		Direction:  p.direction,
		GraphState: p.GraphState(),
	}

	return ss, nil
}

func (p *Peer) ProtocolVersion() uint32 {
	if p.chainState == nil {
		return 0
	}
	return p.chainState.ProtocolVersion
}

func (p *Peer) Genesis() *hash.Hash {
	if p.chainState == nil {
		return nil
	}
	genesisHash, err := hash.NewHash(p.chainState.GenesisHash.Hash)
	if err != nil {
		return nil
	}
	return genesisHash
}

func (p *Peer) Services() protocol.ServiceFlag {
	if p.chainState == nil {
		return protocol.Full
	}
	return protocol.ServiceFlag(p.chainState.Services)
}

func (p *Peer) UserAgent() string {
	if p.chainState == nil {
		return ""
	}
	return string(p.chainState.UserAgent)
}

func (p *Peer) GraphState() *blockdag.GraphState {
	if p.chainState == nil {
		return nil
	}
	gs := blockdag.NewGraphState()
	gs.SetTotal(uint(p.chainState.GraphState.Total))
	gs.SetLayer(uint(p.chainState.GraphState.Layer))
	gs.SetMainHeight(uint(p.chainState.GraphState.MainHeight))
	gs.SetMainOrder(uint(p.chainState.GraphState.MainOrder))
	tips := gs.GetTips()
	for _, tip := range p.chainState.GraphState.Tips {
		h, err := hash.NewHash(tip.Hash)
		if err != nil {
			return nil
		}
		tips.Add(h)
	}
	return gs
}

func (p *Peer) Timestamp() time.Time {
	if p.chainState == nil {
		return time.Time{}
	}
	return time.Unix(int64(p.chainState.Timestamp), 0)
}

func NewPeer(pid peer.ID) *Peer {
	return &Peer{
		peerStatus: &peerStatus{},
		pid:        pid,
	}
}
