/*
 * Copyright (c) 2020.
 * Project:qitmeer
 * File:handshake.go
 * Date:7/19/20 10:37 AM
 * Author:Jin
 * Email:lochjin@gmail.com
 */

package p2p

import (
	"context"
	"fmt"
	"github.com/Qitmeer/qitmeer/p2p/peers"
	"io"
	"sync"
	"time"

	"github.com/libp2p/go-libp2p-core/network"
	"github.com/libp2p/go-libp2p-core/peer"
)

const (
	// The time to wait for a chain state request.
	timeForChainState = 10 * time.Second
)

// AddConnectionHandler adds a callback function which handles the connection with a
// newly added peer. It performs a handshake with that peer by sending a hello request
// and validating the response from the peer.
func (s *Service) AddConnectionHandler(reqFunc func(ctx context.Context, id peer.ID) error,
	goodbyeFunc func(ctx context.Context, id peer.ID) error) {

	// Peer map and lock to keep track of current connection attempts.
	peerMap := make(map[peer.ID]bool)
	peerLock := new(sync.Mutex)

	// This is run at the start of each connection attempt, to ensure
	// that there aren't multiple inflight connection requests for the
	// same peer at once.
	peerHandshaking := func(id peer.ID) bool {
		peerLock.Lock()
		defer peerLock.Unlock()

		if peerMap[id] {
			return true
		}

		peerMap[id] = true
		return false
	}

	peerFinished := func(id peer.ID) {
		peerLock.Lock()
		defer peerLock.Unlock()

		delete(peerMap, id)
	}

	s.host.Network().Notify(&network.NotifyBundle{
		ConnectedF: func(net network.Network, conn network.Conn) {
			peerInfoStr := fmt.Sprintf("peer:%s", conn.RemotePeer().Pretty())
			remotePeer := conn.RemotePeer()
			remotePe := s.peers.Fetch(remotePeer)

			disconnectFromPeer := func() {
				remotePe.SetConnectionState(peers.PeerDisconnecting)
				if err := s.Disconnect(remotePeer); err != nil {
					log.Error(fmt.Sprintf("%s Unable to disconnect from peer:%v", peerInfoStr, err))
				}
				remotePe.SetConnectionState(peers.PeerDisconnected)
			}
			// Connection handler must be non-blocking as part of libp2p design.
			go func() {
				if peerHandshaking(remotePeer) {
					// Exit this if there is already another connection
					// request in flight.
					return
				}
				defer peerFinished(remotePeer)
				// Handle the various pre-existing conditions that will result in us not handshaking.
				peerConnectionState := remotePe.ConnectionState()
				if remotePe.IsActive() {
					log.Trace(fmt.Sprintf("%s currentState:%d reason:already active, Ignoring connection request", peerInfoStr, peerConnectionState))
					return
				}
				s.peers.Add(nil /* QNR */, remotePeer, conn.RemoteMultiaddr(), conn.Stat().Direction)
				if remotePe.IsBad() {
					log.Trace(fmt.Sprintf("%s reason bad peer, Ignoring connection request.", peerInfoStr))
					disconnectFromPeer()
					return
				}
				validPeerConnection := func() {
					remotePe.SetConnectionState(peers.PeerConnected)
					// Go through the handshake process.
					multiAddr := fmt.Sprintf("%s/p2p/%s", conn.RemoteMultiaddr().String(), conn.RemotePeer().String())
					log.Info(fmt.Sprintf("%s direction:%s multiAddr:%s activePeers:%d Peer Connected",
						peerInfoStr, conn.Stat().Direction, multiAddr, len(s.peers.Active())))

					s.peerSync.OnPeerConnected(remotePe)
				}

				// Do not perform handshake on inbound dials.
				if conn.Stat().Direction == network.DirInbound {
					currentTime := time.Now()

					// Wait for peer to initiate handshake
					time.Sleep(timeForChainState)

					// Exit if we are disconnected with the peer.
					if s.host.Network().Connectedness(remotePeer) != network.Connected {
						return
					}

					// If peer hasn't sent a status request, we disconnect with them
					if remotePe.ChainState() == nil {
						disconnectFromPeer()
						return
					}
					updated := remotePe.ChainStateLastUpdated()
					// exit if we don't receive any current status messages from
					// peer.
					if updated.IsZero() || !updated.After(currentTime) {
						disconnectFromPeer()
						return
					}
					validPeerConnection()
					return
				}

				remotePe.SetConnectionState(peers.PeerConnecting)
				if err := reqFunc(context.Background(), conn.RemotePeer()); err != nil && err != io.EOF {
					log.Trace(fmt.Sprintf("%s Handshake failed", peerInfoStr))
					disconnectFromPeer()
					return
				}
				validPeerConnection()
			}()
		},
	})
}

// AddDisconnectionHandler disconnects from peers.  It handles updating the peer status.
// This also calls the handler responsible for maintaining other parts of the sync or p2p system.
func (s *Service) AddDisconnectionHandler(handler func(ctx context.Context, id peer.ID) error) {
	s.host.Network().Notify(&network.NotifyBundle{
		DisconnectedF: func(net network.Network, conn network.Conn) {
			peerInfoStr := fmt.Sprintf("peer:%s", conn.RemotePeer().Pretty())
			// Must be handled in a goroutine as this callback cannot be blocking.
			pe := s.peers.Get(conn.RemotePeer())
			if pe == nil {
				return
			}
			go func() {
				// Exit early if we are still connected to the peer.
				if net.Connectedness(conn.RemotePeer()) == network.Connected {
					return
				}
				priorState := pe.ConnectionState()

				pe.SetConnectionState(peers.PeerDisconnecting)
				ctx := context.Background()
				if err := handler(ctx, conn.RemotePeer()); err != nil {
					log.Error(fmt.Sprintf("%s Disconnect handler failed", peerInfoStr))
				}
				pe.SetConnectionState(peers.PeerDisconnected)
				// Only log disconnections if we were fully connected.
				if priorState == peers.PeerConnected {
					log.Info(fmt.Sprintf("%s Peer Disconnected,activePeers:%d", peerInfoStr, len(s.peers.Active())))
					s.peerSync.OnPeerDisconnected(pe)
				}
			}()
		},
	})
}
