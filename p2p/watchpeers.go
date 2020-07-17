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
	"context"
	"fmt"
	"github.com/libp2p/go-libp2p-core/host"
	"time"
)

// ensurePeerConnections will attempt to reestablish connection to the peers
// if there are currently no connections to that peer.
func ensurePeerConnections(ctx context.Context, h host.Host, peers ...string) {
	if len(peers) == 0 {
		return
	}
	for _, p := range peers {
		if p == "" {
			continue
		}
		peer, err := MakePeer(p)
		if err != nil {
			log.Error(fmt.Sprintf("Could not make peer: %v", err))
			continue
		}

		c := h.Network().ConnsToPeer(peer.ID)
		if len(c) == 0 {
			log.Debug(fmt.Sprintf("No connections to peer, reconnecting:peer %v", peer.ID))
			ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
			defer cancel()
			if err := h.Connect(ctx, *peer); err != nil {
				log.Error(fmt.Sprintf("Failed to reconnect to peer . peer:%s addrs:%s", peer.ID, peer.Addrs))
				continue
			}
		}
	}
}
