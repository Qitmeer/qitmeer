/*
 * Copyright (c) 2020.
 * Project:qitmeer
 * File:watchpeers.go
 * Date:7/15/20 5:03 PM
 * Author:Jin
 * Email:lochjin@gmail.com
 */

package p2p

import (
	"fmt"
	"github.com/libp2p/go-libp2p-core/peer"
	ma "github.com/multiformats/go-multiaddr"
)

// ensurePeerConnections will attempt to reestablish connection to the peers
// if there are currently no connections to that peer.
func (s *Service) ensurePeerConnections(pes []ma.Multiaddr) {
	if len(pes) == 0 {
		return
	}
	for _, p := range pes {
		if p == nil {
			continue
		}
		peerInfo, err := peer.AddrInfoFromP2pAddr(p)
		if err != nil {
			log.Error(fmt.Sprintf("Could not make peer: %v", err))
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
