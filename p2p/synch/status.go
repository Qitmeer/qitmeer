/*
 * Copyright (c) 2017-2020 The qitmeer developers
 */

package synch

import (
	"context"
	"fmt"
	"github.com/Qitmeer/qitmeer/common/roughtime"
	"github.com/Qitmeer/qitmeer/p2p/common"
	"github.com/Qitmeer/qitmeer/p2p/peers"
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
	runutil.RunEvery(s.p2p.Context(), s.PeerInterval, func() {
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

				if (pe.IsBad() && !s.IsWhitePeer(id)) || !pe.CanConnectWithNetwork() {
					if err := s.sendGoodByeAndDisconnect(s.p2p.Context(), common.ErrBadPeer, id); err != nil {
						log.Debug(fmt.Sprintf("Error when disconnecting with bad peer: %v", err))
					}
					return
				}
				if !pe.IsConsensus() {
					return
				}
				// If the status hasn't been updated in the recent interval time.
				if roughtime.Now().After(pe.ChainStateLastUpdated().Add(s.PeerInterval)) {
					if err := s.reValidatePeer(s.p2p.Context(), id); err != nil {
						log.Debug(fmt.Sprintf("Failed to revalidate peer (%v), peer:%s", err, id))
						s.Peers().IncrementBadResponses(id, "maintain peer to reValidatePeer")
					}
				}

				if pe.QNR() == nil && time.Since(pe.ConnectionTime()) > ReconnectionTime && s.p2p.Node() != nil {
					s.peerSync.SyncQNR(pe, s.p2p.Node().String())
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
			s.LookupNode(nil, node)
		}
	})
}

func (s *Sync) reValidatePeer(ctx context.Context, id peer.ID) error {
	if err := s.sendChainStateRequest(ctx, id); err != nil {
		return err
	}
	pe := s.peers.Get(id)
	if pe == nil {
		return peers.ErrPeerUnknown
	}
	if !pe.IsConsensus() {
		return nil
	}
	// Do not return an error for ping requests.
	if err := s.SendPingRequest(ctx, id); err != nil {
		log.Debug(fmt.Sprintf("Could not ping peer:%v", err))
	}
	return nil
}
