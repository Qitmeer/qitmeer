package peerserver

import (
	"fmt"
	"github.com/Qitmeer/qitmeer/common/roughtime"
	"github.com/Qitmeer/qitmeer/log"
	"time"
)

// peerState maintains state of inbound, persistent, outbound peers as well
// as banned peers and outbound groups.
type peerState struct {
	inboundPeers    map[int32]*serverPeer
	outboundPeers   map[int32]*serverPeer
	persistentPeers map[int32]*serverPeer
	banned          map[string]time.Time
	outboundGroups  map[string]int
}

// Count returns the count of all known peers.
func (ps *peerState) Count() int {
	return len(ps.inboundPeers) + len(ps.outboundPeers) +
		len(ps.persistentPeers)
}

// forAllPeers is a helper function that runs closure on all peers known to
// peerState.
func (ps *peerState) forAllPeers(closure func(sp *serverPeer)) {
	for _, e := range ps.inboundPeers {
		closure(e)
	}
	ps.forAllOutboundPeers(closure)
}

// forAllOutboundPeers is a helper function that runs closure on all outbound
// peers known to peerState.
func (ps *peerState) forAllOutboundPeers(closure func(sp *serverPeer)) {
	for _, e := range ps.outboundPeers {
		closure(e)
	}
	for _, e := range ps.persistentPeers {
		closure(e)
	}
}

func (ps *peerState) IsBanPeer(host string) bool {
	if banEnd, ok := ps.banned[host]; ok {
		if roughtime.Now().Before(banEnd) {
			log.Debug(fmt.Sprintf("Peer %s is banned for another %v - disconnecting",
				host, roughtime.Until(banEnd)))
			return true
		}
		log.Info("Peer is no longer banned", "peer", host)
		delete(ps.banned, host)
	}
	return false
}

func (ps *peerState) IsMaxInboundPeer(sp *serverPeer) bool {
	if !sp.Inbound() {
		return false
	}
	host := sp.NA().IP.String()
	inshost := map[string]int{}
	for _, e := range ps.inboundPeers {
		_, ok := inshost[e.NA().IP.String()]
		if ok {
			inshost[e.NA().IP.String()]++
			continue
		}
		inshost[e.NA().IP.String()] = 1
	}
	total, ok := inshost[host]
	if !ok {
		return false
	}
	return total >= sp.server.cfg.MaxInbound
}
