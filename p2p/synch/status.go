/*
 * Copyright (c) 2017-2020 The qitmeer developers
 */

package synch

import (
	"context"
	"fmt"
	"github.com/Qitmeer/qitmeer/common/roughtime"
	"github.com/Qitmeer/qitmeer/p2p/runutil"
	"github.com/libp2p/go-libp2p-core/network"
	"github.com/libp2p/go-libp2p-core/peer"
	"time"
)

const (
	ReconnectionTime = time.Second * 30
)

// maintainPeerStatuses by infrequently polling peers for their latest status.
func (s *Sync) maintainPeerStatuses() {
	interval := s.p2p.BlockChain().ChainParams().TargetTimePerBlock * 5
	runutil.RunEvery(s.p2p.Context(), interval, func() {
		for _, pid := range s.Peers().Connected() {
			pe := s.peers.Get(pid)
			if pe == nil {
				continue
			}
			go func(id peer.ID) {
				// If our peer status has not been updated correctly we disconnect over here
				// and set the connection state over here instead.
				if s.p2p.Host().Network().Connectedness(id) != network.Connected {
					s.peerSync.Disconnect(pe)
					return
				}

				if pe.IsBad() {
					if err := s.sendGoodByeAndDisconnect(s.p2p.Context(), codeGenericError, id); err != nil {
						log.Error(fmt.Sprintf("Error when disconnecting with bad peer: %v", err))
					}
					return
				}
				// If the status hasn't been updated in the recent interval time.

				if roughtime.Now().After(pe.ChainStateLastUpdated().Add(interval)) {
					if err := s.reValidatePeer(s.p2p.Context(), id); err != nil {
						log.Error(fmt.Sprintf("Failed to revalidate peer (%v), peer:%s", err, id))
						s.Peers().IncrementBadResponses(id)
					}
				}
			}(pid)
		}

		for _, pid := range s.Peers().Disconnected() {
			pe := s.peers.Get(pid)
			if pe == nil {
				continue
			}
			node := pe.Node()
			if node == nil ||
				time.Since(pe.ConnectionTime()) < ReconnectionTime ||
				pe.IsBad() {
				continue
			}
			s.p2p.Resolve(node)
		}
	})
}

func (s *Sync) reValidatePeer(ctx context.Context, id peer.ID) error {
	if err := s.sendChainStateRequest(ctx, id); err != nil {
		return err
	}

	// Do not return an error for ping requests.
	if err := s.SendPingRequest(ctx, id); err != nil {
		log.Debug(fmt.Sprintf("Could not ping peer:%v", err))
	}
	return nil
}
