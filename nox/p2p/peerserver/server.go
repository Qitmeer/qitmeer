package peerserver

import (
	"github.com/noxproject/nox/params"
	"github.com/noxproject/nox/database"
	"github.com/noxproject/nox/config"
	"github.com/noxproject/nox/log"
	"sync"
	"time"
	"github.com/noxproject/nox/p2p/addmgr"
	"github.com/noxproject/nox/p2p/connmgr"
	"github.com/noxproject/nox/core/message"
)



// Use start to begin accepting connections from peers.
// peer server handling communications to and from nox peers.
type PeerServer struct{

    // address manager caching the peers
	addrManager          *addmgr.AddrManager

	// conn manager handles network connections.
	connManager          *connmgr.ConnManager

	newPeers             chan *serverPeer
	donePeers            chan *serverPeer
	banPeers             chan *serverPeer

	// peer handler chan
	relayInv             chan relayMsg
	broadcast            chan broadcastMsg
	peerHeightsUpdate    chan updatePeerHeightsMsg
	query                chan interface{}
	quit          		 chan struct{}

	wg                   sync.WaitGroup
}

func NewPeerServer(cfg *config.Config, db database.DB, chainParams *params.Params) (*PeerServer, error){
	s := PeerServer{}
	return &s, nil
}

func (p *PeerServer) Start() error {
	log.Debug("Starting P2P server")
	return nil
}
func (p *PeerServer) Stop() error {
	log.Debug("Stopping P2P server")
	return nil
}


// TODO, re-impl peer handler
func (s *PeerServer) peerHandler() {

	s.addrManager.Start()

	log.Trace("Starting peer handler")

	state := &peerState{
		inboundPeers:    make(map[int32]*serverPeer),
		persistentPeers: make(map[int32]*serverPeer),
		outboundPeers:   make(map[int32]*serverPeer),
		banned:          make(map[string]time.Time),
		outboundGroups:  make(map[string]int),
	}

	go s.connManager.Start()

out:
	for {
		select {
		// New peers connected to the server.
		case p := <-s.newPeers:
			s.handleAddPeerMsg(state, p)

		// Disconnected peers.
		case p := <-s.donePeers:
			s.handleDonePeerMsg(state, p)

		// Block accepted in mainchain or orphan, update peer height.
		case umsg := <-s.peerHeightsUpdate:
			s.handleUpdatePeerHeights(state, umsg)

		// Peer to ban.
		case p := <-s.banPeers:
			s.handleBanPeerMsg(state, p)

		// New inventory to potentially be relayed to other peers.
		case invMsg := <-s.relayInv:
			s.handleRelayInvMsg(state, invMsg)

		// Message to broadcast to all connected peers except those
		// which are excluded by the message.
		case bmsg := <-s.broadcast:
			s.handleBroadcastMsg(state, &bmsg)

		case qmsg := <-s.query:
			s.handleQuery(state, qmsg)

		case <-s.quit:
			break out
		}
	}

	s.connManager.Stop()
	s.addrManager.Stop()

	// Drain channels before exiting so nothing is left waiting around
	// to send.
cleanup:
	for {
		select {
		case <-s.newPeers:
		case <-s.donePeers:
		case <-s.peerHeightsUpdate:
		case <-s.relayInv:
		case <-s.broadcast:
		case <-s.query:
		default:
			break cleanup
		}
	}
	s.wg.Done()
	log.Trace("Peer handler done")
}

// handleAddPeerMsg deals with adding new peers.  It is invoked from the
// peerHandler goroutine.
func (s *PeerServer) handleAddPeerMsg(state *peerState, sp *serverPeer) bool {
	return false
}

// handleDonePeerMsg deals with peers that have signalled they are done.  It is
// invoked from the peerHandler goroutine.
func (s *PeerServer) handleDonePeerMsg(state *peerState, sp *serverPeer) {

}

// handleUpdatePeerHeight updates the heights of all peers who were known to
// announce a block we recently accepted.
func (s *PeerServer) handleUpdatePeerHeights(state *peerState, umsg updatePeerHeightsMsg) {

}

// handleBanPeerMsg deals with banning peers.  It is invoked from the
// peerHandler goroutine.
func (s *PeerServer) handleBanPeerMsg(state *peerState, sp *serverPeer) {

}

// handleRelayInvMsg deals with relaying inventory to peers that are not already
// known to have it.  It is invoked from the peerHandler goroutine.
func (s *PeerServer) handleRelayInvMsg(state *peerState, msg relayMsg) {

}

// handleBroadcastMsg deals with broadcasting messages to peers.  It is invoked
// from the peerHandler goroutine.
func (s *PeerServer) handleBroadcastMsg(state *peerState, bmsg *broadcastMsg) {

}

// RelayInventory relays the passed inventory vector to all connected peers
// that are not already known to have it.
func (s *PeerServer) RelayInventory(invVect *message.InvVect, data interface{}) {
	s.relayInv <- relayMsg{invVect: invVect, data: data}
}


