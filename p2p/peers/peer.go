package peers

import (
	"fmt"
	"github.com/Qitmeer/qitmeer/common/hash"
	"github.com/Qitmeer/qitmeer/core/blockdag"
	"github.com/Qitmeer/qitmeer/core/protocol"
	pb "github.com/Qitmeer/qitmeer/p2p/proto/v1"
	"github.com/Qitmeer/qitmeer/p2p/qnode"
	"github.com/Qitmeer/qitmeer/p2p/qnr"
	"github.com/gogo/protobuf/proto"
	"github.com/libp2p/go-libp2p-core/network"
	"github.com/libp2p/go-libp2p-core/peer"
	ma "github.com/multiformats/go-multiaddr"
	"sync"
	"time"
)

const (
	// maxBadResponses is the maximum number of bad responses from a peer before we stop talking to it.
	maxBadResponses = 5
)

// Peer represents a connected p2p network remote node.
type Peer struct {
	*peerStatus
	pid       peer.ID
	syncPoint *hash.Hash
	// Use to fee filter
	feeFilter int64

	lock *sync.RWMutex
}

func (p *Peer) GetID() peer.ID {
	p.lock.RLock()
	defer p.lock.RUnlock()

	return p.pid
}

// BadResponses obtains the number of bad responses we have received from the given remote peer.
// This will error if the peer does not exist.
func (p *Peer) BadResponses() int {
	p.lock.RLock()
	defer p.lock.RUnlock()

	return p.badResponses
}

// IsBad states if the peer is to be considered bad.
// If the peer is unknown this will return `false`, which makes using this function easier than returning an error.
func (p *Peer) IsBad() bool {
	p.lock.RLock()
	defer p.lock.RUnlock()
	return p.isBad()
}

func (p *Peer) isBad() bool {
	return p.badResponses >= maxBadResponses
}

// IncrementBadResponses increments the number of bad responses we have received from the given remote peer.
func (p *Peer) IncrementBadResponses() {
	p.lock.Lock()
	defer p.lock.Unlock()

	p.badResponses++

	if p.isBad() {
		log.Warn(fmt.Sprintf("I am bad peer:%s", p.pid.String()))
	}
}

func (p *Peer) Decay() {
	p.lock.Lock()
	defer p.lock.Unlock()

	if p.badResponses > 0 {
		p.badResponses--
	}
}

func (p *Peer) ResetBad() {
	p.lock.Lock()
	defer p.lock.Unlock()

	p.badResponses = 0
}

func (p *Peer) UpdateAddrDir(record *qnr.Record, address ma.Multiaddr, direction network.Direction) {
	p.lock.Lock()
	defer p.lock.Unlock()

	p.address = address
	p.direction = direction
	if record != nil {
		p.qnr = record
	}
}

// Address returns the multiaddress of the given remote peer.
// This will error if the peer does not exist.
func (p *Peer) Address() ma.Multiaddr {
	p.lock.RLock()
	defer p.lock.RUnlock()

	return p.address
}

// Direction returns the direction of the given remote peer.
// This will error if the peer does not exist.
func (p *Peer) Direction() network.Direction {
	p.lock.RLock()
	defer p.lock.RUnlock()

	return p.direction
}

// QNR returns the enr for the corresponding peer id.
func (p *Peer) QNR() *qnr.Record {
	p.lock.RLock()
	defer p.lock.RUnlock()

	return p.qnr
}

// ConnectionState gets the connection state of the given remote peer.
// This will error if the peer does not exist.
func (p *Peer) ConnectionState() PeerConnectionState {
	p.lock.RLock()
	defer p.lock.RUnlock()

	return p.peerState
}

// IsActive checks if a peers is active and returns the result appropriately.
func (p *Peer) IsActive() bool {
	p.lock.RLock()
	defer p.lock.RUnlock()

	return p.peerState.IsConnected() || p.peerState.IsConnecting()
}

// SetConnectionState sets the connection state of the given remote peer.
func (p *Peer) SetConnectionState(state PeerConnectionState) {
	p.lock.Lock()
	defer p.lock.Unlock()

	p.peerState = state
}

// SetChainState sets the chain state of the given remote peer.
func (p *Peer) SetChainState(chainState *pb.ChainState) {
	p.lock.Lock()
	defer p.lock.Unlock()

	p.chainState = chainState
	p.chainStateLastUpdated = time.Now()

	log.Trace(fmt.Sprintf("SetChainState(%s) : MainHeight=%d", p.pid.ShortString(), chainState.GraphState.MainHeight))
}

// ChainState gets the chain state of the given remote peer.
// This can return nil if there is no known chain state for the peer.
// This will error if the peer does not exist.
func (p *Peer) ChainState() *pb.ChainState {
	p.lock.RLock()
	defer p.lock.RUnlock()

	return p.chainState
}

// ChainStateLastUpdated gets the last time the chain state of the given remote peer was updated.
// This will error if the peer does not exist.
func (p *Peer) ChainStateLastUpdated() time.Time {
	p.lock.RLock()
	defer p.lock.RUnlock()

	return p.chainStateLastUpdated
}

// SetMetadata sets the metadata of the given remote peer.
func (p *Peer) SetMetadata(metaData *pb.MetaData) {
	p.lock.Lock()
	defer p.lock.Unlock()

	p.metaData = metaData
}

// Metadata returns a copy of the metadata corresponding to the provided
// peer id.
func (p *Peer) Metadata() *pb.MetaData {
	p.lock.RLock()
	defer p.lock.RUnlock()

	return proto.Clone(p.metaData).(*pb.MetaData)
}

// CommitteeIndices retrieves the committee subnets the peer is subscribed to.
func (p *Peer) CommitteeIndices() []uint64 {
	p.lock.RLock()
	defer p.lock.RUnlock()

	if p.qnr == nil || p.metaData == nil {
		return []uint64{}
	}
	return retrieveIndicesFromBitfield(p.metaData.Subnets)
}

func (p *Peer) StatsSnapshot() (*StatsSnap, error) {
	p.lock.RLock()
	defer p.lock.RUnlock()

	if p.qnr == nil {
		return nil, fmt.Errorf("no qnr")
	}
	n, err := qnode.New(qnode.ValidSchemes, p.qnr)
	if err != nil {
		return nil, fmt.Errorf("qnode: can't verify local record: %v", err)
	}
	ss := &StatsSnap{
		NodeID:     n.ID().String(),
		PeerID:     p.pid.String(),
		QNR:        n.String(),
		Protocol:   p.protocolVersion(),
		Genesis:    p.genesis(),
		Services:   p.services(),
		UserAgent:  p.userAgent(),
		State:      p.peerState,
		Direction:  p.direction,
		GraphState: p.graphState(),
	}

	return ss, nil
}

func (p *Peer) Timestamp() time.Time {
	p.lock.RLock()
	defer p.lock.RUnlock()

	if p.chainState == nil {
		return time.Time{}
	}
	return time.Unix(int64(p.chainState.Timestamp), 0)
}

func (p *Peer) SetQNR(record *qnr.Record) {
	p.lock.Lock()
	defer p.lock.Unlock()

	p.qnr = record
}

func (p *Peer) protocolVersion() uint32 {
	if p.chainState == nil {
		return 0
	}
	return p.chainState.ProtocolVersion
}

func (p *Peer) genesis() *hash.Hash {
	if p.chainState == nil {
		return nil
	}
	genesisHash, err := hash.NewHash(p.chainState.GenesisHash.Hash)
	if err != nil {
		return nil
	}
	return genesisHash
}

func (p *Peer) services() protocol.ServiceFlag {
	if p.chainState == nil {
		return protocol.Full
	}
	return protocol.ServiceFlag(p.chainState.Services)
}

func (p *Peer) userAgent() string {
	if p.chainState == nil {
		return ""
	}
	return string(p.chainState.UserAgent)
}

func (p *Peer) graphState() *blockdag.GraphState {
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

func (p *Peer) GraphState() *blockdag.GraphState {
	p.lock.RLock()
	defer p.lock.RUnlock()

	return p.graphState()
}

func (p *Peer) UpdateGraphState(gs *pb.GraphState) {
	p.lock.Lock()
	defer p.lock.Unlock()

	if p.chainState == nil {
		p.chainState = &pb.ChainState{}
		//per.chainState.GraphState=&pb.GraphState{}
	}
	p.chainState.GraphState = gs
	log.Trace(fmt.Sprintf("UpdateGraphState(%s) : MainHeight=%d", p.pid.ShortString(), gs.MainHeight))
	/*	per.chainState.GraphState.Total=uint32(gs.GetTotal())
		per.chainState.GraphState.Layer=uint32(gs.GetLayer())
		per.chainState.GraphState.MainOrder=uint32(gs.GetMainOrder())
		per.chainState.GraphState.MainHeight=uint32(gs.GetMainHeight())
		per.chainState.GraphState.Tips=[]*pb.Hash{}
		for h:=range gs.GetTips().GetMap() {
			per.chainState.GraphState.Tips=append(per.chainState.GraphState.Tips,&pb.Hash{Hash:h.Bytes()})
		}*/
}

func (p *Peer) UpdateSyncPoint(point *hash.Hash) {
	p.lock.Lock()
	defer p.lock.Unlock()

	p.syncPoint = point
}

func (p *Peer) SyncPoint() *hash.Hash {
	p.lock.RLock()
	defer p.lock.RUnlock()

	return p.syncPoint
}

func (p *Peer) DisableRelayTx() bool {
	p.lock.RLock()
	defer p.lock.RUnlock()

	if p.chainState == nil {
		return false
	}
	return p.chainState.DisableRelayTx
}

func (p *Peer) FeeFilter() int64 {
	p.lock.RLock()
	defer p.lock.RUnlock()

	return p.feeFilter
}

func NewPeer(pid peer.ID, point *hash.Hash) *Peer {
	return &Peer{
		peerStatus: &peerStatus{
			peerState: PeerDisconnected,
		},
		pid:       pid,
		lock:      &sync.RWMutex{},
		syncPoint: point,
	}
}
