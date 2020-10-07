package p2p

import (
	"fmt"
	"github.com/Qitmeer/qitmeer/core/blockchain"
	"github.com/Qitmeer/qitmeer/core/blockdag"
	"github.com/Qitmeer/qitmeer/core/event"
	"github.com/libp2p/go-libp2p-core/peer"
	"sync"
)

type PeerSync struct {
	lock     sync.RWMutex
	service  *Service
	syncPeer peer.ID
	// dag sync
	dagSync *blockdag.DAGSync
}

func (ps *PeerSync) Start() error {
	log.Info("P2P PeerSync Start")
	return nil
}

func (ps *PeerSync) SyncPeer() peer.ID {
	ps.lock.RLock()
	defer ps.lock.RLock()
	return ps.syncPeer
}

func (ps *PeerSync) Stop() error {
	log.Info("P2P PeerSync Stop")
	return nil
}

func (ps *PeerSync) OnPeerConnected(id peer.ID) {
	ps.lock.Lock()
	defer ps.lock.Unlock()

	ti, err := ps.service.peers.Timestamp(id)
	if err == nil && !ti.IsZero() {
		// Add the remote peer time as a sample for creating an offset against
		// the local clock to keep the network time in sync.
		ps.service.TimeSource.AddTimeSample(id.String(), ti)
	}

	if !ps.HasSyncPeer() {
		ps.startSync()
	}
}

func (ps *PeerSync) OnPeerDisconnected(id peer.ID) {

}

func (ps *PeerSync) HasSyncPeer() bool {
	return ps.syncPeer.Validate() == nil
}

func (ps *PeerSync) Chain() *blockchain.BlockChain {
	return ps.service.Chain
}

// startSync will choose the best peer among the available candidate peers to
// download/sync the blockchain from.  When syncing is already running, it
// simply returns.  It also examines the candidates for any which are no longer
// candidates and removes them as needed.
func (ps *PeerSync) startSync() {
	// Return now if we're already syncing.
	if ps.HasSyncPeer() {
		return
	}
	best := ps.Chain().BestSnapshot()
	bestPeer := ps.getBestPeer()
	// Start syncing from the best peer if one was selected.
	if bestPeer.Validate() == nil {
		gs, _ := ps.service.peers.GraphState(bestPeer)

		log.Info(fmt.Sprintf("Syncing to state %s from peer %s cur graph state:%s", gs.String(), bestPeer.String(), best.GraphState.String()))

		// When the current height is less than a known checkpoint we
		// can use block headers to learn about which blocks comprise
		// the chain up to the checkpoint and perform less validation
		// for them.  This is possible since each header contains the
		// hash of the previous header and a merkle root.  Therefore if
		// we validate all of the received headers link together
		// properly and the checkpoint hashes match, we can be sure the
		// hashes for the blocks in between are accurate.  Further, once
		// the full blocks are downloaded, the merkle root is computed
		// and compared against the value in the header which proves the
		// full block hasn't been tampered with.
		//
		// Once we have passed the final checkpoint, or checkpoints are
		// disabled, use standard inv messages learn about the blocks
		// and fully validate them.  Finally, regression test mode does
		// not support the headers-first approach so do normal block
		// downloads when in regression test mode.
		ps.syncPeer = bestPeer
		ps.IntellectSyncBlocks(true)

		ps.dagSync.GSMtx.Lock()
		ps.dagSync.GS = gs
		ps.dagSync.GSMtx.Unlock()
	} else {
		log.Trace("No sync peer candidates available")
	}
}

// getBestPeer
func (ps *PeerSync) getBestPeer() peer.ID {
	best := ps.Chain().BestSnapshot()
	var bestPeer peer.ID
	var bestGS *blockdag.GraphState
	equalPeers := []peer.ID{}
	for _, sp := range ps.service.peers.Connected() {
		// Remove sync candidate peers that are no longer candidates due
		// to passing their latest known block.  NOTE: The < is
		// intentional as opposed to <=.  While techcnically the peer
		// doesn't have a later block when it's equal, it will likely
		// have one soon so it is a reasonable choice.  It also allows
		// the case where both are at 0 such as during regression test.
		gs, err := ps.service.peers.GraphState(sp)
		if err != nil {
			continue
		}
		if best.GraphState.IsExcellent(gs) {
			continue
		}
		// the best sync candidate is the most updated peer
		if len(bestPeer) == 0 {
			bestPeer = sp
			bestGS = gs
			continue
		}
		if gs.IsExcellent(bestGS) {
			bestPeer = sp
			if len(equalPeers) > 0 {
				equalPeers = equalPeers[0:0]
			}
		} else if gs.IsEqual(bestGS) {
			equalPeers = append(equalPeers, sp)
		}
	}
	if len(bestPeer) == 0 {
		return peer.ID("")
	}
	if len(equalPeers) > 0 {
		for _, sp := range equalPeers {
			if sp.String() > bestPeer.String() {
				bestPeer = sp
			}
		}
	}
	return bestPeer
}

// IsCurrent returns true if we believe we are synced with our peers, false if we
// still have blocks to check
func (ps *PeerSync) IsCurrent() bool {
	if !ps.Chain().IsCurrent() {
		return false
	}

	// if blockChain thinks we are current and we have no syncPeer it
	// is probably right.
	if !ps.HasSyncPeer() {
		return true
	}

	// No matter what chain thinks, if we are below the block we are syncing
	// to we are not current.
	gs, err := ps.service.peers.GraphState(ps.syncPeer)
	if err != nil {
		return true
	}
	if gs.IsExcellent(ps.Chain().BestSnapshot().GraphState) {
		log.Trace("comparing the current best vs sync last",
			"current.best", ps.Chain().BestSnapshot().GraphState.String(), "sync.last", gs.String())
		return false
	}

	return true
}

func (ps *PeerSync) IntellectSyncBlocks(refresh bool) {
	if !ps.HasSyncPeer() {
		return
	}

	gs := ps.Chain().BestSnapshot().GraphState
	if ps.Chain().GetOrphansTotal() >= blockchain.MaxOrphanBlocks || refresh {
		ps.Chain().RefreshOrphans()
	}
	allOrphan := ps.Chain().GetRecentOrphansParents()
	if len(allOrphan) > 0 {
		err := peer.PushGetBlocksMsg(gs, allOrphan)
		if err != nil {
			b.PushSyncDAGMsg(peer)
		}
	} else {
		b.PushSyncDAGMsg(peer)
	}
}

func NewPeerSync(service *Service) *PeerSync {
	peerSync := &PeerSync{service: service, syncPeer: peer.ID("")}
	peerSync.dagSync = blockdag.NewDAGSync(service.Chain.BlockDAG())
	best := service.Chain.BestSnapshot()

	peerSync.dagSync.GSMtx.Lock()
	peerSync.dagSync.GS = best.GraphState
	peerSync.dagSync.GSMtx.Unlock()

	return peerSync
}
