/*
 * Copyright (c) 2017-2020 The qitmeer developers
 */

package synch

import (
	"fmt"
	"github.com/Qitmeer/qitmeer/core/blockchain"
	"github.com/Qitmeer/qitmeer/core/blockdag"
	"github.com/Qitmeer/qitmeer/core/types"
	"github.com/Qitmeer/qitmeer/p2p/peers"
	pb "github.com/Qitmeer/qitmeer/p2p/proto/v1"
	"sync"
	"sync/atomic"
	"time"
)

const (
	// stallSampleInterval the interval at which we will check to see if our
	// sync has stalled.
	stallSampleInterval = 300 * time.Second
)

type PeerSync struct {
	sy *Sync

	splock   sync.RWMutex
	syncPeer *peers.Peer
	// dag sync
	dagSync *blockdag.DAGSync

	started  int32
	shutdown int32
	msgChan  chan interface{}
	wg       sync.WaitGroup
	quit     chan struct{}
}

func (ps *PeerSync) Start() error {
	// Already started?
	if atomic.AddInt32(&ps.started, 1) != 1 {
		return nil
	}

	log.Info("P2P PeerSync Start")
	ps.dagSync = blockdag.NewDAGSync(ps.sy.p2p.BlockChain().BlockDAG())

	ps.wg.Add(1)
	go ps.handler()
	return nil
}

func (ps *PeerSync) Stop() error {
	if atomic.AddInt32(&ps.shutdown, 1) != 1 {
		log.Warn("PeerSync is already in the process of shutting down")
		return nil
	}
	log.Info("P2P PeerSync Stop")

	close(ps.quit)
	ps.wg.Wait()

	return nil
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
				fmt.Println("ConnectedMsg")
				ps.processConnected(msg)
				fmt.Println("ConnectedMsg end")

			case *DisconnectedMsg:
				fmt.Println("DisconnectedMsg")
				ps.processDisconnected(msg)
				fmt.Println("DisconnectedMsg end")
			case *GetBlocksMsg:
				fmt.Println("GetBlocksMsg")
				err := ps.processGetBlocks(msg.pe, msg.blocks)
				if err != nil {
					log.Error(err.Error())
				}
				fmt.Println("GetBlocksMsg end")
			case *GetBlockDatasMsg:
				fmt.Println("GetBlockDatasMsg")
				err := ps.processGetBlockDatas(msg.pe, msg.blocks)
				if err != nil {
					log.Error(err.Error())
				}
				fmt.Println("GetBlockDatasMsg end")
			case *UpdateGraphStateMsg:
				fmt.Println("UpdateGraphStateMsg")
				err := ps.processUpdateGraphState(msg.pe)
				if err != nil {
					log.Error(err.Error())
				}
				fmt.Println("UpdateGraphStateMsg end")
			case *syncDAGBlocksMsg:
				fmt.Println("syncDAGBlocksMsg")
				err := ps.processSyncDAGBlocks(msg.pe)
				if err != nil {
					log.Error(err.Error())
				}
				fmt.Println("syncDAGBlocksMsg end")
			case *PeerUpdateMsg:
				fmt.Println("PeerUpdateMsg")
				ps.OnPeerUpdate(msg.pe, msg.orphan)
				fmt.Println("PeerUpdateMsg end")
			case *getTxsMsg:
				fmt.Println("getTxsMsg")
				err := ps.processGetTxs(msg.pe, msg.txs)
				if err != nil {
					log.Error(err.Error())
				}
				fmt.Println("getTxsMsg end")
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
}

func (ps *PeerSync) OnPeerConnected(pe *peers.Peer) {

	ti := pe.Timestamp()
	if !ti.IsZero() {
		// Add the remote peer time as a sample for creating an offset against
		// the local clock to keep the network time in sync.
		ps.sy.p2p.TimeSource().AddTimeSample(pe.GetID().String(), ti)
	}

	if !ps.HasSyncPeer() {
		ps.startSync()
	}
}

func (ps *PeerSync) OnPeerDisconnected(pe *peers.Peer) {

	if ps.HasSyncPeer() {
		if ps.isSyncPeer(pe) {
			ps.updateSyncPeer(true)
		}
	}
}

func (ps *PeerSync) isSyncPeer(pe *peers.Peer) bool {
	if !ps.HasSyncPeer() {
		return false
	}
	if pe == ps.SyncPeer() || pe.GetID() == ps.SyncPeer().GetID() {
		return true
	}
	return false
}

func (ps *PeerSync) PeerUpdate(pe *peers.Peer, orphan bool) {
	// Ignore if we are shutting down.
	if atomic.LoadInt32(&ps.shutdown) != 0 {
		return
	}

	ps.msgChan <- &PeerUpdateMsg{pe: pe, orphan: orphan}
}

func (ps *PeerSync) OnPeerUpdate(pe *peers.Peer, orphan bool) {

	if ps.HasSyncPeer() {
		spgs := ps.SyncPeer().GraphState()
		if !ps.SyncPeer().IsActive() || spgs == nil {
			ps.updateSyncPeer(true)
			return
		}
		if pe != nil {
			pegs := pe.GraphState()
			if pegs != nil {
				if pegs.IsExcellent(spgs) {
					ps.updateSyncPeer(true)
					return
				}
			}

		}
		ps.IntellectSyncBlocks(orphan)
		return
	}
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
		ps.IntellectSyncBlocks(true)
		ps.dagSync.SetGraphState(gs)

	} else {
		log.Trace("No synchronization is required.")
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
		if best.GraphState.IsExcellent(gs) {
			continue
		}
		// the best sync candidate is the most updated peer
		if bestPeer == nil {
			bestPeer = sp
			continue
		}
		if gs.IsExcellent(bestPeer.GraphState()) {
			bestPeer = sp
			if len(equalPeers) > 0 {
				equalPeers = equalPeers[0:0]
			}
		} else if gs.IsEqual(bestPeer.GraphState()) {
			equalPeers = append(equalPeers, sp)
		}
	}
	if bestPeer == nil {
		return nil
	}
	if len(equalPeers) > 0 {
		for _, sp := range equalPeers {
			if sp.GetID().String() > bestPeer.GetID().String() {
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
	gs := ps.SyncPeer().GraphState()
	if gs == nil {
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

	if ps.Chain().GetOrphansTotal() >= blockchain.MaxOrphanBlocks || refresh {
		ps.Chain().RefreshOrphans()
	}
	allOrphan := ps.Chain().GetRecentOrphansParents()

	if len(allOrphan) > 0 {
		go ps.GetBlocks(ps.SyncPeer(), allOrphan)
	} else {
		go ps.syncDAGBlocks(ps.SyncPeer())
	}
}

func (ps *PeerSync) updateSyncPeer(force bool) {
	log.Debug("Updating sync peer")
	if force {
		ps.SetSyncPeer(nil)
	}
	ps.startSync()
}

func (ps *PeerSync) RelayInventory(data interface{}) {
	ps.sy.Peers().ForPeers(peers.PeerConnected, func(pe *peers.Peer) {
		msg := &pb.Inventory{Invs: []*pb.InvVect{}}
		switch value := data.(type) {
		case []*types.TxDesc:
			// Don't relay the transaction to the peer when it has
			// transaction relaying disabled.
			if pe.DisableRelayTx() {
				return
			}
			for _, tx := range value {
				feeFilter := pe.FeeFilter()
				if feeFilter > 0 && tx.FeePerKB < feeFilter {
					return
				}
				msg.Invs = append(msg.Invs, NewInvVect(InvTypeTx, tx.Tx.Hash()))
			}
		case types.BlockHeader:
			blockHash := value.BlockHash()
			msg.Invs = append(msg.Invs, NewInvVect(InvTypeBlock, &blockHash))
		}
		go ps.sy.sendInventoryRequest(ps.sy.p2p.Context(), pe, msg)
	})
}

func NewPeerSync(sy *Sync) *PeerSync {
	peerSync := &PeerSync{
		sy:      sy,
		msgChan: make(chan interface{}),
		quit:    make(chan struct{}),
	}

	return peerSync
}
