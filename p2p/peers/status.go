package peers

import (
	"errors"
	"github.com/Qitmeer/qitmeer/core/blockdag"
	pb "github.com/Qitmeer/qitmeer/p2p/proto/v1"
	"github.com/Qitmeer/qitmeer/p2p/qnr"
	"github.com/gogo/protobuf/proto"
	"github.com/libp2p/go-libp2p-core/network"
	"github.com/libp2p/go-libp2p-core/peer"
	ma "github.com/multiformats/go-multiaddr"
	"github.com/prysmaticlabs/go-bitfield"
	"sync"
	"time"
)

var (
	// ErrPeerUnknown is returned when there is an attempt to obtain data from a peer that is not known.
	ErrPeerUnknown = errors.New("peer unknown")
)

// Status is the structure holding the peer status information.
type Status struct {
	lock            sync.RWMutex
	maxBadResponses int
	peers           map[peer.ID]*Peer
}

// MaxBadResponses returns the maximum number of bad responses a peer can provide before it is considered bad.
func (p *Status) MaxBadResponses() int {
	return p.maxBadResponses
}

// Bad returns the peers that are bad.
func (p *Status) Bad() []peer.ID {
	p.lock.RLock()
	defer p.lock.RUnlock()
	peers := make([]peer.ID, 0)
	for pid, status := range p.peers {
		if status.badResponses >= p.maxBadResponses {
			peers = append(peers, pid)
		}
	}
	return peers
}

// IsBad states if the peer is to be considered bad.
// If the peer is unknown this will return `false`, which makes using this function easier than returning an error.
func (p *Status) IsBad(pid peer.ID) bool {
	p.lock.RLock()
	defer p.lock.RUnlock()

	if status, ok := p.peers[pid]; ok {
		return status.badResponses >= p.maxBadResponses
	}
	return false
}

// IncrementBadResponses increments the number of bad responses we have received from the given remote peer.
func (p *Status) IncrementBadResponses(pid peer.ID) {
	p.lock.Lock()
	defer p.lock.Unlock()

	status := p.fetch(pid)
	status.badResponses++
}

// fetch is a helper function that fetches a peer, possibly creating it.
func (p *Status) fetch(pid peer.ID) *Peer {
	if _, ok := p.peers[pid]; !ok {
		p.peers[pid] = NewPeer(pid)
	}
	return p.peers[pid]
}

// Add adds a peer.
// If a peer already exists with this ID its address and direction are updated with the supplied data.
func (p *Status) Add(record *qnr.Record, pid peer.ID, address ma.Multiaddr, direction network.Direction) {
	p.lock.Lock()
	defer p.lock.Unlock()

	if pe, ok := p.peers[pid]; ok {
		// Peer already exists, just update its address info.
		pe.UpdateAddrDir(record, address, direction)
		return
	}
	pe := NewPeer(pid)
	pe.UpdateAddrDir(record, address, direction)

	p.peers[pid] = pe
}

// Address returns the multiaddress of the given remote peer.
// This will error if the peer does not exist.
func (p *Status) Address(pid peer.ID) (ma.Multiaddr, error) {
	p.lock.RLock()
	defer p.lock.RUnlock()

	if status, ok := p.peers[pid]; ok {
		return status.address, nil
	}
	return nil, ErrPeerUnknown
}

// Direction returns the direction of the given remote peer.
// This will error if the peer does not exist.
func (p *Status) Direction(pid peer.ID) (network.Direction, error) {
	p.lock.RLock()
	defer p.lock.RUnlock()

	if status, ok := p.peers[pid]; ok {
		return status.direction, nil
	}
	return network.DirUnknown, ErrPeerUnknown
}

// QNR returns the enr for the corresponding peer id.
func (p *Status) QNR(pid peer.ID) (*qnr.Record, error) {
	p.lock.RLock()
	defer p.lock.RUnlock()

	if status, ok := p.peers[pid]; ok {
		return status.qnr, nil
	}
	return nil, ErrPeerUnknown
}

// SetChainState sets the chain state of the given remote peer.
func (p *Status) SetChainState(pid peer.ID, chainState *pb.ChainState) {
	p.lock.Lock()
	defer p.lock.Unlock()

	status := p.fetch(pid)
	status.chainState = chainState
	status.chainStateLastUpdated = time.Now()
}

// ChainState gets the chain state of the given remote peer.
// This can return nil if there is no known chain state for the peer.
// This will error if the peer does not exist.
func (p *Status) ChainState(pid peer.ID) (*pb.ChainState, error) {
	p.lock.RLock()
	defer p.lock.RUnlock()

	if status, ok := p.peers[pid]; ok {
		return status.chainState, nil
	}
	return nil, ErrPeerUnknown
}

// IsActive checks if a peers is active and returns the result appropriately.
func (p *Status) IsActive(pid peer.ID) bool {
	p.lock.RLock()
	defer p.lock.RUnlock()

	status, ok := p.peers[pid]
	return ok && (status.peerState == PeerConnected || status.peerState == PeerConnecting)
}

// SetMetadata sets the metadata of the given remote peer.
func (p *Status) SetMetadata(pid peer.ID, metaData *pb.MetaData) {
	p.lock.Lock()
	defer p.lock.Unlock()

	status := p.fetch(pid)
	status.metaData = metaData
}

// Metadata returns a copy of the metadata corresponding to the provided
// peer id.
func (p *Status) Metadata(pid peer.ID) (*pb.MetaData, error) {
	p.lock.RLock()
	defer p.lock.RUnlock()

	if status, ok := p.peers[pid]; ok {
		return proto.Clone(status.metaData).(*pb.MetaData), nil
	}
	return nil, ErrPeerUnknown
}

// CommitteeIndices retrieves the committee subnets the peer is subscribed to.
func (p *Status) CommitteeIndices(pid peer.ID) ([]uint64, error) {
	p.lock.RLock()
	defer p.lock.RUnlock()

	if status, ok := p.peers[pid]; ok {
		if status.qnr == nil || status.metaData == nil {
			return []uint64{}, nil
		}
		return retrieveIndicesFromBitfield(status.metaData.Subnets), nil
	}
	return nil, ErrPeerUnknown
}

// SubscribedToSubnet retrieves the peers subscribed to the given
// committee subnet.
func (p *Status) SubscribedToSubnet(index uint64) []peer.ID {
	p.lock.RLock()
	defer p.lock.RUnlock()

	peers := make([]peer.ID, 0)
	for pid, status := range p.peers {
		// look at active peers
		connectedStatus := status.peerState == PeerConnecting || status.peerState == PeerConnected
		if connectedStatus && status.metaData != nil && status.metaData.Subnets != nil {
			indices := retrieveIndicesFromBitfield(status.metaData.Subnets)
			for _, idx := range indices {
				if idx == index {
					peers = append(peers, pid)
					break
				}
			}
		}
	}
	return peers
}

// SetConnectionState sets the connection state of the given remote peer.
func (p *Status) SetConnectionState(pid peer.ID, state PeerConnectionState) {
	p.lock.Lock()
	defer p.lock.Unlock()

	status := p.fetch(pid)
	status.peerState = state
}

// ConnectionState gets the connection state of the given remote peer.
// This will error if the peer does not exist.
func (p *Status) ConnectionState(pid peer.ID) (PeerConnectionState, error) {
	p.lock.RLock()
	defer p.lock.RUnlock()

	if status, ok := p.peers[pid]; ok {
		return status.peerState, nil
	}
	return PeerDisconnected, ErrPeerUnknown
}

// ChainStateLastUpdated gets the last time the chain state of the given remote peer was updated.
// This will error if the peer does not exist.
func (p *Status) ChainStateLastUpdated(pid peer.ID) (time.Time, error) {
	p.lock.RLock()
	defer p.lock.RUnlock()

	if status, ok := p.peers[pid]; ok {
		return status.chainStateLastUpdated, nil
	}
	return time.Now(), ErrPeerUnknown
}

// BadResponses obtains the number of bad responses we have received from the given remote peer.
// This will error if the peer does not exist.
func (p *Status) BadResponses(pid peer.ID) (int, error) {
	p.lock.RLock()
	defer p.lock.RUnlock()

	if status, ok := p.peers[pid]; ok {
		return status.badResponses, nil
	}
	return -1, ErrPeerUnknown
}

// Connecting returns the peers that are connecting.
func (p *Status) Connecting() []peer.ID {
	p.lock.RLock()
	defer p.lock.RUnlock()
	peers := make([]peer.ID, 0)
	for pid, status := range p.peers {
		if status.peerState == PeerConnecting {
			peers = append(peers, pid)
		}
	}
	return peers
}

// Connected returns the peers that are connected.
func (p *Status) Connected() []peer.ID {
	p.lock.RLock()
	defer p.lock.RUnlock()
	peers := make([]peer.ID, 0)
	for pid, status := range p.peers {
		if status.peerState == PeerConnected {
			peers = append(peers, pid)
		}
	}
	return peers
}

// Active returns the peers that are connecting or connected.
func (p *Status) Active() []peer.ID {
	p.lock.RLock()
	defer p.lock.RUnlock()
	peers := make([]peer.ID, 0)
	for pid, status := range p.peers {
		if status.peerState == PeerConnecting || status.peerState == PeerConnected {
			peers = append(peers, pid)
		}
	}
	return peers
}

// Disconnecting returns the peers that are disconnecting.
func (p *Status) Disconnecting() []peer.ID {
	p.lock.RLock()
	defer p.lock.RUnlock()
	peers := make([]peer.ID, 0)
	for pid, status := range p.peers {
		if status.peerState == PeerDisconnecting {
			peers = append(peers, pid)
		}
	}
	return peers
}

// Disconnected returns the peers that are disconnected.
func (p *Status) Disconnected() []peer.ID {
	p.lock.RLock()
	defer p.lock.RUnlock()
	peers := make([]peer.ID, 0)
	for pid, status := range p.peers {
		if status.peerState == PeerDisconnected {
			peers = append(peers, pid)
		}
	}
	return peers
}

// Inactive returns the peers that are disconnecting or disconnected.
func (p *Status) Inactive() []peer.ID {
	p.lock.RLock()
	defer p.lock.RUnlock()
	peers := make([]peer.ID, 0)
	for pid, status := range p.peers {
		if status.peerState == PeerDisconnecting || status.peerState == PeerDisconnected {
			peers = append(peers, pid)
		}
	}
	return peers
}

// All returns all the peers regardless of state.
func (p *Status) All() []peer.ID {
	p.lock.RLock()
	defer p.lock.RUnlock()
	pids := make([]peer.ID, 0, len(p.peers))
	for pid := range p.peers {
		pids = append(pids, pid)
	}
	return pids
}

func (p *Status) StatsSnapshots() []*StatsSnap {
	p.lock.RLock()
	defer p.lock.RUnlock()

	pes := make([]*StatsSnap, 0, len(p.peers))
	for _, pe := range p.peers {
		ss, err := pe.StatsSnapshot()
		if err != nil {
			continue
		}
		pes = append(pes, ss)
	}
	return pes
}

// Decay reduces the bad responses of all peers, giving reformed peers a chance to join the network.
// This can be run periodically, although note that each time it runs it does give all bad peers another chance as well to clog up
// the network with bad responses, so should not be run too frequently; once an hour would be reasonable.
func (p *Status) Decay() {
	p.lock.Lock()
	defer p.lock.Unlock()
	for _, status := range p.peers {
		if status.badResponses > 0 {
			status.badResponses--
		}
	}
}

func (p *Status) Timestamp(pid peer.ID) (time.Time, error) {
	p.lock.RLock()
	defer p.lock.RUnlock()

	if status, ok := p.peers[pid]; ok {
		return status.Timestamp(), nil
	}
	return time.Time{}, ErrPeerUnknown
}

func (p *Status) GraphState(pid peer.ID) (*blockdag.GraphState, error) {
	p.lock.RLock()
	defer p.lock.RUnlock()

	if status, ok := p.peers[pid]; ok {
		return status.GraphState(), nil
	}
	return nil, ErrPeerUnknown
}

// NewStatus creates a new status entity.
func NewStatus(maxBadResponses int) *Status {
	return &Status{
		maxBadResponses: maxBadResponses,
		peers:           make(map[peer.ID]*Peer),
	}
}

func retrieveIndicesFromBitfield(bitV bitfield.Bitvector64) []uint64 {
	committeeIdxs := []uint64{}
	for i := uint64(0); i < 64; i++ {
		if bitV.BitAt(i) {
			committeeIdxs = append(committeeIdxs, i)
		}
	}
	return committeeIdxs
}
