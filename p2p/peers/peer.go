package peers

import (
	"github.com/Qitmeer/qitmeer/p2p/qnr"
	"github.com/libp2p/go-libp2p-core/network"
	"github.com/libp2p/go-libp2p-core/peer"
	ma "github.com/multiformats/go-multiaddr"
)

// Peer represents a connected p2p network remote node.
type Peer struct {
	*peerStatus
	pid peer.ID
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

func NewPeer(pid peer.ID) *Peer {
	return &Peer{
		peerStatus: &peerStatus{},
		pid:        pid,
	}
}
