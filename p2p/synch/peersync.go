/*
 * Copyright (c) 2017-2020 The qitmeer developers
 */

package synch

import (
	"fmt"
	"github.com/Qitmeer/qitmeer/core/blockchain"
	"github.com/Qitmeer/qitmeer/core/blockdag"
	"github.com/Qitmeer/qitmeer/core/protocol"
	"github.com/Qitmeer/qitmeer/core/types"
	"github.com/Qitmeer/qitmeer/p2p/peers"
	pb "github.com/Qitmeer/qitmeer/p2p/proto/v1"
	"github.com/libp2p/go-libp2p-core/peer"
	"math/rand"
	"sync"
	"sync/atomic"
	"time"
)

const (
	// stallSampleInterval the interval at which we will check to see if our
	// sync has stalled.
	stallSampleInterval = 30 * time.Second
)

type PeerSync struct {
	sy *Sync

	splock   sync.RWMutex
	syncPeer *peers.Peer
	lastSync time.Time
	// dag sync
	dagSync *blockdag.DAGSync

	started     int32
	shutdown    int32
	msgChan     chan interface{}
	wg          sync.WaitGroup
	quit        chan struct{}
	longSyncMod bool
}

func (ps *PeerSync) Start() error {
	// Already started?
	if atomic.AddInt32(&ps.started, 1) != 1 {
		return nil
	}

	log.Info("P2P PeerSync Start")
	ps.dagSync = blockdag.NewDAGSync(ps.sy.p2p.BlockChain().BlockDAG())
	ps.longSyncMod = false

	ps.wg.Add(1)
	go ps.handler()
	return nil
}

func (ps *PeerSync) Stop() error {
	if atomic.AddInt32(&ps.shutdown, 1) != 1 {
		log.Warn("PeerSync is already in the process of shutting down")
		return nil
	}
	log.Info("P2P PeerSync Stoping...")

	close(ps.quit)
	ps.wg.Wait()
	log.Info("P2P PeerSync Stoped")
	return nil
}

func (ps *PeerSync) IsRunning() bool {
	if atomic.LoadInt32(&ps.shutdown) != 0 {
		return false
	}
	if atomic.LoadInt32(&ps.started) == 0 {
		return false
	}
	return true
}

func (ps *PeerSync) handler() {
	stallTicker := time.NewTicker(stallSampleInterval)
	defer stallTicker.Stop()

out:
	for {
		select {
		case m := <-ps.msgChan:
			switch msg := m.(type) {
			case pauseMsg:
				// Wait until the sender unpauses the manager.
				<-msg.unpause

			case *ConnectedMsg:
				ps.processConnected(msg)

			case *DisconnectedMsg:
				ps.processDisconnected(msg)
			case *GetBlocksMsg:
				err := ps.processGetBlocks(msg.pe, msg.blocks)
				if err != nil {
					log.Debug(err.Error())
				}
			case *GetBlockDatasMsg:
				err := ps.processGetBlockDatas(msg.pe, msg.blocks)
				if err != nil {
					log.Debug(err.Error())
				}
			case *GetDatasMsg:
				_ = ps.OnGetData(msg.pe, msg.data.Invs)
			case *OnFilterAddMsg:
				ps.OnFilterAdd(msg.pe, msg.data)
			case *OnFilterClearMsg:
				ps.OnFilterClear(msg.pe, msg.data)
			case *OnFilterLoadMsg:
				ps.OnFilterLoad(msg.pe, msg.data)
			case *OnMsgMemPool:
				ps.OnMemPool(msg.pe, msg.data)

			case *UpdateGraphStateMsg:
				log.Trace(fmt.Sprintf("UpdateGraphStateMsg recevied from %v, state=%v ", msg.pe.GetID(), msg.pe.GraphState()))
				err := ps.processUpdateGraphState(msg.pe)
				if err != nil {
					log.Trace(err.Error())
				}
			case *syncDAGBlocksMsg:
				err := ps.processSyncDAGBlocks(msg.pe)
				if err != nil {
					log.Debug(err.Error())
				}
			case *PeerUpdateMsg:
				ps.OnPeerUpdate(msg.pe)

			case *SyncQNRMsg:
				err := ps.processQNR(msg)
				if err != nil {
					log.Warn(err.Error())
				}
			default:
				log.Warn(fmt.Sprintf("Invalid message type in task "+
					"handler: %T", msg))
			}

		case <-stallTicker.C:
			ps.handleStallSample()

		case <-ps.quit:
			break out
		}
	}

	// Drain any wait channels before going away so there is nothing left
	// waiting on this goroutine.
cleanup:
	for {
		select {
		case <-ps.msgChan:
		default:
			break cleanup
		}
	}

	ps.wg.Done()
	log.Trace("Peer Sync handler done")
}

func (ps *PeerSync) handleStallSample() {
	if atomic.LoadInt32(&ps.shutdown) != 0 {
		return
	}
	if ps.HasSyncPeer() {
		if time.Since(ps.lastSync) >= ps.sy.PeerInterval {
			ps.updateSyncPeer(true)
		}
	}
}

func (ps *PeerSync) Pause() chan<- struct{} {
	c := make(chan struct{})
	ps.msgChan <- pauseMsg{c}
	return c
}

func (ps *PeerSync) SyncPeer() *peers.Peer {
	ps.splock.RLock()
	defer ps.splock.RUnlock()

	return ps.syncPeer
}

func (ps *PeerSync) SetSyncPeer(pe *peers.Peer) {
	ps.splock.Lock()
	defer ps.splock.Unlock()

	ps.syncPeer = pe

	if pe != nil {
		ps.lastSync = time.Now()
	}
}

func (ps *PeerSync) OnPeerConnected(pe *peers.Peer) {
	if !ps.IsRunning() {
		return
	}
	ti := pe.Timestamp()
	if !ti.IsZero() {
		// Add the remote peer time as a sample for creating an offset against
		// the local clock to keep the network time in sync.
		ps.sy.p2p.TimeSource().AddTimeSample(pe.GetID().String(), ti)
	}

	ps.updateSyncPeer(false)
}

func (ps *PeerSync) OnPeerDisconnected(pe *peers.Peer) {

	if ps.HasSyncPeer() {
		if ps.isSyncPeer(pe) {
			ps.updateSyncPeer(true)
		}
	}
}

func (ps *PeerSync) isSyncPeer(pe *peers.Peer) bool {
	ps.splock.RLock()
	defer ps.splock.RUnlock()

	if ps.syncPeer == nil || pe == nil {
		return false
	}
	if pe == ps.syncPeer || pe.GetID() == ps.syncPeer.GetID() {
		return true
	}
	return false
}

func (ps *PeerSync) PeerUpdate(pe *peers.Peer, immediately bool) {
	// Ignore if we are shutting down.
	if atomic.LoadInt32(&ps.shutdown) != 0 {
		return
	}

	if immediately {
		ps.msgChan <- &PeerUpdateMsg{pe: pe}
		return
	}

	pe.RunRate(PeerUpdate, DefaultRateTaskTime, func() {
		ps.msgChan <- &PeerUpdateMsg{pe: pe}
	})

}

func (ps *PeerSync) OnPeerUpdate(pe *peers.Peer) {
	log.Trace(fmt.Sprintf("OnPeerUpdate peer=%v", pe.GetID()))
	ps.updateSyncPeer(false)
}

func (ps *PeerSync) HasSyncPeer() bool {
	return ps.SyncPeer() != nil
}

func (ps *PeerSync) Chain() *blockchain.BlockChain {
	return ps.sy.p2p.BlockChain()
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
	if bestPeer != nil {
		gs := bestPeer.GraphState()

		log.Info(fmt.Sprintf("Syncing to state %s from peer %s cur graph state:%s", gs.String(), bestPeer.GetID().String(), best.GraphState.String()))

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
		ps.SetSyncPeer(bestPeer)
		ps.IntellectSyncBlocks(true, bestPeer)
		ps.dagSync.SetGraphState(gs)

	} else {
		log.Trace("You're already up to date, no synchronization is required.")
	}
}

// getBestPeer
func (ps *PeerSync) getBestPeer() *peers.Peer {
	best := ps.Chain().BestSnapshot()
	var bestPeer *peers.Peer
	equalPeers := []*peers.Peer{}
	for _, sp := range ps.sy.peers.ConnectedPeers() {
		// Remove sync candidate peers that are no longer candidates due
		// to passing their latest known block.  NOTE: The < is
		// intentional as opposed to <=.  While techcnically the peer
		// doesn't have a later block when it's equal, it will likely
		// have one soon so it is a reasonable choice.  It also allows
		// the case where both are at 0 such as during regression test.
		gs := sp.GraphState()
		if gs == nil {
			continue
		}
		if !gs.IsExcellent(best.GraphState) {
			continue
		}
		// the best sync candidate is the most updated peer
		if bestPeer == nil {
			bestPeer = sp
			equalPeers = []*peers.Peer{sp}
			continue
		}
		if gs.IsExcellent(bestPeer.GraphState()) {
			bestPeer = sp
			equalPeers = []*peers.Peer{sp}
		} else if gs.IsEqual(bestPeer.GraphState()) {
			equalPeers = append(equalPeers, sp)
		}
	}
	if len(equalPeers) <= 0 {
		return nil
	}
	if len(equalPeers) == 1 {
		return equalPeers[0]
	}

	index := int(rand.Int63n(int64(len(equalPeers))))
	if index >= len(equalPeers) {
		index = 0
	}
	return equalPeers[index]
}

// IsCurrent returns true if we believe we are synced with our peers, false if we
// still have blocks to check
func (ps *PeerSync) IsCurrent() bool {
	if !ps.Chain().IsCurrent() {
		return false
	}

	return ps.IsCompleteForSyncPeer()
}

func (ps *PeerSync) IsCompleteForSyncPeer() bool {
	// if blockChain thinks we are current and we have no syncPeer it
	// is probably right.
	sp := ps.SyncPeer()
	if sp == nil {
		return true
	}

	// No matter what chain thinks, if we are below the block we are syncing
	// to we are not current.
	gs := sp.GraphState()
	if gs == nil {
		return true
	}
	if gs.IsExcellent(ps.Chain().BestSnapshot().GraphState) {
		//log.Trace("comparing the current best vs sync last",
		//	"current.best", ps.Chain().BestSnapshot().GraphState.String(), "sync.last", gs.String())
		return false
	}

	return true
}

func (ps *PeerSync) IntellectSyncBlocks(refresh bool, pe *peers.Peer) {
	if pe == nil {
		log.Trace(fmt.Sprintf("IntellectSyncBlocks has not sync peer, return directly"))
		return
	}

	if ps.Chain().GetOrphansTotal() >= blockchain.MaxOrphanBlocks || refresh {
		err := ps.Chain().RefreshOrphans()
		if err != nil {
			log.Trace(fmt.Sprintf("IntellectSyncBlocks failed to refresh orphans, err=%v", err.Error()))
		}
	}
	allOrphan := ps.Chain().GetRecentOrphansParents()

	if len(allOrphan) > 0 {
		log.Trace(fmt.Sprintf("IntellectSyncBlocks do ps.GetBlock, peer=%v,allOrphan=%v ", pe.GetID(), allOrphan))
		go ps.GetBlocks(pe, allOrphan)
	} else {
		log.Trace(fmt.Sprintf("IntellectSyncBlocks do ps.syncDAGBlocks, peer=%v ", pe.GetID()))
		go ps.syncDAGBlocks(pe)
	}
}

func (ps *PeerSync) updateSyncPeer(force bool) {
	if !ps.IsRunning() {
		return
	}
	log.Debug("Updating sync peer")
	if force {
		ps.SetSyncPeer(nil)
	}
	ps.startSync()
}

func (ps *PeerSync) continueSync(orphan bool) {
	if !ps.IsRunning() {
		return
	}
	log.Debug("Continue sync peer")
	sp := ps.SyncPeer()
	if sp != nil {
		spgs := sp.GraphState()
		if !sp.IsConnected() || spgs == nil {
			ps.updateSyncPeer(true)
			return
		}
		bestPeer := ps.getBestPeer()
		if bestPeer == nil {
			ps.updateSyncPeer(true)
			return
		}
		if bestPeer != sp {
			ps.updateSyncPeer(true)
			return
		}

		ps.IntellectSyncBlocks(orphan, sp)
		return
	}
	ps.updateSyncPeer(false)
}

func (ps *PeerSync) RelayInventory(data interface{}, filters []peer.ID) {
	filtersM := map[peer.ID]struct{}{}
	if len(filters) > 0 {
		for _, f := range filters {
			filtersM[f] = struct{}{}
		}
	}
	ps.sy.Peers().ForPeers(peers.PeerConnected, func(pe *peers.Peer) {
		_, ok := filtersM[pe.GetID()]
		if ok {
			return
		}
		msg := &pb.Inventory{Invs: []*pb.InvVect{}}

		switch value := data.(type) {
		case *types.TxDesc:
			if pe.HasBroadcast(value.Tx.Hash().String()) {
				return
			}
			// Don't relay the transaction to the peer when it has
			// transaction relaying disabled.
			if pe.DisableRelayTx() {
				return
			}
			feeFilter := pe.FeeFilter()
			if feeFilter > 0 && value.FeePerKB < feeFilter {
				return
			}
			// Don't relay the transaction if there is a bloom
			// filter loaded and the transaction doesn't match it.
			filter := pe.Filter()
			if filter.IsLoaded() {
				if !filter.MatchTxAndUpdate(value.Tx) {
					return
				}
			}
			msg.Invs = append(msg.Invs, NewInvVect(InvTypeTx, value.Tx.Hash()))
			log.Trace(fmt.Sprintf("Relay inventory tx(%s) to peer(%s)", value.Tx.Hash().String(), pe.GetID().String()))
			pe.Broadcast(value.Tx.Hash().String(), value)

		case types.BlockHeader:
			blockHash := value.BlockHash()
			msg.Invs = append(msg.Invs, NewInvVect(InvTypeBlock, &blockHash))
			log.Trace(fmt.Sprintf("Relay inventory block(%s) to peer(%s)", blockHash.String(), pe.GetID().String()))
		}

		if len(msg.Invs) <= 0 {
			return
		}

		go ps.sy.sendInventoryRequest(ps.sy.p2p.Context(), pe, msg)
	})
}

// EnforceNodeBloomFlag disconnects the peer if the server is not configured to
// allow bloom filters.  Additionally, if the peer has negotiated to a protocol
// version  that is high enough to observe the bloom filter service support bit,
// it will be banned since it is intentionally violating the protocol.
func (ps *PeerSync) EnforceNodeBloomFlag(sp *peers.Peer) bool {
	services := sp.Services()
	if services&protocol.Bloom != protocol.Bloom {
		// Disconnect the peer regardless of protocol version or banning
		// state.
		log.Debug(fmt.Sprintf("%s sent a filterclear request with no "+
			"filter loaded -- disconnecting", sp.GetID().String()))
		ps.Disconnect(sp)
		return false
	}

	return true
}

// OnFilterAdd is invoked when a peer receives a filteradd qitmeer
// message and is used by remote peers to add data to an already loaded bloom
// filter.  The peer will be disconnected if a filter is not loaded when this
// message is received or the server is not configured to allow bloom filters.
func (ps *PeerSync) OnFilterAdd(sp *peers.Peer, msg *types.MsgFilterAdd) {
	// Disconnect and/or ban depending on the node bloom services flag and
	// negotiated protocol version.
	if !ps.EnforceNodeBloomFlag(sp) {
		return
	}
	filter := sp.Filter()
	if !filter.IsLoaded() {
		log.Debug(fmt.Sprintf("%s sent a filterclear request with no "+
			"filter loaded -- disconnecting", sp.GetID().String()))
		ps.Disconnect(sp)
		return
	}

	filter.Add(msg.Data)
}

// OnFilterClear is invoked when a peer receives a filterclear qitmeer
// message and is used by remote peers to clear an already loaded bloom filter.
// The peer will be disconnected if a filter is not loaded when this message is
// received  or the server is not configured to allow bloom filters.
func (ps *PeerSync) OnFilterClear(sp *peers.Peer, msg *types.MsgFilterClear) {
	// Disconnect and/or ban depending on the node bloom services flag and
	// negotiated protocol version.
	if !ps.EnforceNodeBloomFlag(sp) {
		return
	}
	filter := sp.Filter()

	if !filter.IsLoaded() {
		log.Debug(fmt.Sprintf("%s sent a filterclear request with no "+
			"filter loaded -- disconnecting", sp.GetID().String()))
		ps.Disconnect(sp)
		return
	}

	filter.Unload()
}

// OnFilterLoad is invoked when a peer receives a filterload qitmeer
// message and it used to load a bloom filter that should be used for
// delivering merkle blocks and associated transactions that match the filter.
// The peer will be disconnected if the server is not configured to allow bloom
// filters.
func (ps *PeerSync) OnFilterLoad(sp *peers.Peer, msg *types.MsgFilterLoad) {
	// Disconnect and/or ban depending on the node bloom services flag and
	// negotiated protocol version.
	if !ps.EnforceNodeBloomFlag(sp) {
		return
	}
	filter := sp.Filter()
	sp.DisableRelayTx()

	filter.Reload(msg)
}

func NewPeerSync(sy *Sync) *PeerSync {
	peerSync := &PeerSync{
		sy:      sy,
		msgChan: make(chan interface{}),
		quit:    make(chan struct{}),
	}

	return peerSync
}
