package peerserver

import "github.com/noxproject/nox/log"

// handleBroadcastMsg deals with broadcasting messages to peers.  It is invoked
// from the peerHandler goroutine.
func (s *PeerServer) handleBroadcastMsg(state *peerState, bmsg *broadcastMsg) {
	log.Trace("TODO handleBroadcastMsg()")
}


