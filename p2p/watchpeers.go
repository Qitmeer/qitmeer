/*
 * Copyright (c) 2017-2020 The qitmeer developers
 */

package p2p

import (
	"fmt"
	"github.com/libp2p/go-libp2p-core/peer"
)

// ensurePeerConnections will attempt to reestablish connection to the peers
// if there are currently no connections to that peer.
func (s *Service) ensurePeerConnections(pes []string) {
	if len(pes) == 0 {
		return
	}
	for _, p := range pes {
		if len(p) <= 0 {
			continue
		}
		peerInfo, err := MakePeer(p)
		if err != nil {
			log.Error(fmt.Sprintf("Could not make peer: %v", err))
			continue
		}
		pe := s.Peers().Get(peerInfo.ID)
		if pe != nil && !pe.CanConnectWithNetwork() {
			continue
		}

		c := s.Host().Network().ConnsToPeer(peerInfo.ID)
		if len(c) == 0 {
			log.Debug(fmt.Sprintf("No connections to peer, reconnecting:peer %v", peerInfo.ID))

			go func(info peer.AddrInfo) {
				if err := s.connectWithPeer(info, true); err != nil {
					log.Trace(fmt.Sprintf("Could not connect with peer %s :%v", info.String(), err))
				}
			}(*peerInfo)
		}
	}
}
