package peerserver

import (
	"fmt"
	"github.com/Qitmeer/qitmeer/log"
	"net"
	"time"
)

// BanPeer bans a peer that has already been connected to the server by ip.
func (s *PeerServer) BanPeer(sp *serverPeer) {
	s.banPeers <- sp
}

// handleBanPeerMsg deals with banning peers.  It is invoked from the
// peerHandler goroutine.
func (s *PeerServer) handleBanPeerMsg(state *peerState, sp *serverPeer) {
	host, _, err := net.SplitHostPort(sp.Addr())
	if err != nil {
		log.Debug(fmt.Sprintf("can't split ban peer %s %v", sp.Addr(), err))
		return
	}
	direction := directionString(sp.Inbound())
	log.Info(fmt.Sprintf("Banned peer %s (%s) for %v", host, direction,
		s.cfg.BanDuration))
	state.banned[host] = time.Now().Add(s.cfg.BanDuration)
}

// addBanScore increases the persistent and decaying ban score fields by the
// values passed as parameters. If the resulting score exceeds half of the ban
// threshold, a warning is logged including the reason provided. Further, if
// the score is above the ban threshold, the peer will be banned and
// disconnected.
func (sp *serverPeer) addBanScore(persistent, transient uint32, reason string) {
	// No warning is logged and no score is calculated if banning is disabled.
	if sp.server.cfg.DisableBanning {
		return
	}
	if sp.isWhitelisted {
		log.Debug("Misbehaving whitelisted peer", "sp",sp, "reason",reason)
		return
	}

	warnThreshold := sp.server.cfg.BanThreshold >> 1
	if transient == 0 && persistent == 0 {
		// The score is not being increased, but a warning message is still
		// logged if the score is above the warn threshold.
		score := sp.banScore.Int()
		if score > warnThreshold {
			log.Warn(fmt.Sprintf("Misbehaving peer %s: %s -- ban score is %d, "+
				"it was not increased this time", sp, reason, score))
		}
		return
	}
	score := sp.banScore.Increase(persistent, transient)
	if score > warnThreshold {
		log.Warn("Misbehaving peer %s: %s -- ban score increased to %d",
			sp, reason, score)
		if score > sp.server.cfg.BanThreshold {
			log.Warn("Misbehaving peer -- banning and disconnecting","peer", sp)
			sp.server.BanPeer(sp)
			sp.Disconnect()
		}
	}
}

// directionString is a helper function that returns a string that represents
// the direction of a connection (inbound or outbound).
func directionString(inbound bool) string {
	if inbound {
		return "inbound"
	}
	return "outbound"
}




