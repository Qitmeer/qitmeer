package p2p

import (
	"context"
	"fmt"
	"github.com/Qitmeer/qitmeer/common/roughtime"
	"github.com/Qitmeer/qitmeer/p2p/peers"
	"github.com/Qitmeer/qitmeer/p2p/runutil"
	"github.com/libp2p/go-libp2p-core/network"
	"github.com/libp2p/go-libp2p-core/peer"
)

// maintainPeerStatuses by infrequently polling peers for their latest status.
func (s *Service) maintainPeerStatuses() {
	interval := s.BlockManager.ChainParams().TargetTimePerBlock * 2
	runutil.RunEvery(s.ctx, interval, func() {
		for _, pid := range s.Peers().Connected() {
			go func(id peer.ID) {
				// If our peer status has not been updated correctly we disconnect over here
				// and set the connection state over here instead.
				if s.Host().Network().Connectedness(id) != network.Connected {
					s.Peers().SetConnectionState(id, peers.PeerDisconnecting)
					if err := s.Disconnect(id); err != nil {
						log.Error(fmt.Sprintf("Error when disconnecting with peer: %v", err))
					}
					s.Peers().SetConnectionState(id, peers.PeerDisconnected)
					return
				}
				if s.Peers().IsBad(id) {
					if err := s.sendGoodByeAndDisconnect(s.ctx, codeGenericError, id); err != nil {
						log.Error(fmt.Sprintf("Error when disconnecting with bad peer: %v", err))
					}
					return
				}
				// If the status hasn't been updated in the recent interval time.
				lastUpdated, err := s.Peers().ChainStateLastUpdated(id)
				if err != nil {
					// Peer has vanished; nothing to do.
					return
				}
				if roughtime.Now().After(lastUpdated.Add(interval)) {
					if err := s.reValidatePeer(s.ctx, id); err != nil {
						log.Error(fmt.Sprintf("Failed to revalidate peer (%v), peer:%s", err, id))
						s.Peers().IncrementBadResponses(id)
					}
				}
			}(pid)
		}
	})
}

func (s *Service) reValidatePeer(ctx context.Context, id peer.ID) error {
	if err := s.sendChainStateRequest(ctx, id); err != nil {
		return err
	}

	// Do not return an error for ping requests.
	if err := s.sendPingRequest(ctx, id); err != nil {
		log.Debug(fmt.Sprintf("Could not ping peer:%v", err))
	}
	return nil
}
