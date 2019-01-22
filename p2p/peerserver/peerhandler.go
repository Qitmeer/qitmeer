package peerserver

import (
	"fmt"
	"github.com/noxproject/nox/log"
	"github.com/noxproject/nox/p2p/addmgr"
	"net"
	"sync/atomic"
	"time"
)

// handleAddPeerMsg deals with adding new peers.  It is invoked from the
// peerHandler goroutine.
func (s *PeerServer) handleAddPeerMsg(state *peerState, sp *serverPeer) bool {
	if sp == nil {
		return false
	}

	// Ignore new peers if we're shutting down.
	if atomic.LoadInt32(&s.shutdown) != 0 {
		log.Info(fmt.Sprintf("New peer %s ignored - server is shutting down", sp))
		sp.Disconnect()
		return false
	}

	// Disconnect banned peers.
	host, _, err := net.SplitHostPort(sp.Addr())
	if err != nil {
		log.Debug("can't split hostport", "error",err)
		sp.Disconnect()
		return false
	}
	if banEnd, ok := state.banned[host]; ok {
		if time.Now().Before(banEnd) {
			log.Debug(fmt.Sprintf("Peer %s is banned for another %v - disconnecting",
				host, time.Until(banEnd)))
			sp.Disconnect()
			return false
		}

		log.Info("Peer is no longer banned","peer", host)
		delete(state.banned, host)
	}

	// TODO: Check for max peers from a single IP.

	// Limit max number of total peers.
	if state.Count() >= s.cfg.MaxPeers {
		log.Info(fmt.Sprintf("Max peers reached [%d] - disconnecting peer %s",
			s.cfg.MaxPeers, sp))
		sp.Disconnect()
		// TODO: how to handle permanent peers here?
		// they should be rescheduled.
		return false
	}

	// Add the new peer and start it.
	log.Debug("New peer", "peer",sp)
	if sp.Inbound() {
		state.inboundPeers[sp.ID()] = sp
	} else {
		state.outboundGroups[addmgr.GroupKey(sp.NA())]++
		if sp.persistent {
			state.persistentPeers[sp.ID()] = sp
		} else {
			state.outboundPeers[sp.ID()] = sp
		}
	}

	return true
}

// handleUpdatePeerHeight updates the heights of all peers who were known to
// announce a block we recently accepted.
func (s *PeerServer) handleUpdatePeerHeights(state *peerState, umsg updatePeerHeightsMsg) {
	log.Trace("TODO handleUpdatePeerHeights()")
}
