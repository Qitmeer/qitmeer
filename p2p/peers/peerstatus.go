package peers

import (
	pb "github.com/Qitmeer/qitmeer/p2p/proto/v1"
	"github.com/Qitmeer/qitmeer/p2p/qnr"
	"github.com/libp2p/go-libp2p-core/network"
	ma "github.com/multiformats/go-multiaddr"
	"time"
)

// PeerConnectionState is the state of the connection.
type PeerConnectionState int32

const (
	// PeerDisconnected means there is no connection to the peer.
	PeerDisconnected PeerConnectionState = iota
	// PeerDisconnecting means there is an on-going attempt to disconnect from the peer.
	PeerDisconnecting
	// PeerConnected means the peer has an active connection.
	PeerConnected
	// PeerConnecting means there is an on-going attempt to connect to the peer.
	PeerConnecting
)

func (pcs PeerConnectionState) String() string {
	switch pcs {
	case PeerDisconnected:
		return "disconnected"
	case PeerDisconnecting:
		return "disconnecting"
	case PeerConnected:
		return "connected"
	case PeerConnecting:
		return "connecting"
	}
	return ""
}

func (pcs PeerConnectionState) IsConnected() bool {
	return pcs == PeerConnected
}

func (pcs PeerConnectionState) IsConnecting() bool {
	return pcs == PeerConnecting
}

func (pcs PeerConnectionState) IsDisconnected() bool {
	return pcs == PeerDisconnected
}

func (pcs PeerConnectionState) IsDisconnecting() bool {
	return pcs == PeerDisconnecting
}

// peerStatus is the status of an individual peer at the protocol level.
type peerStatus struct {
	address               ma.Multiaddr
	direction             network.Direction
	peerState             PeerConnectionState
	qnr                   *qnr.Record
	metaData              *pb.MetaData
	chainState            *pb.ChainState
	chainStateLastUpdated time.Time
	badResponses          int
}
