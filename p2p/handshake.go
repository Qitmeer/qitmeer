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
	// The time to wait for a status request.
	timeForStatus = 10 * time.Second
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
			disconnectFromPeer := func() {
				s.peers.SetConnectionState(remotePeer, peers.PeerDisconnecting)
				if err := s.Disconnect(remotePeer); err != nil {
					log.Error(fmt.Sprintf("%s Unable to disconnect from peer:%v", peerInfoStr, err))
				}
				s.peers.SetConnectionState(remotePeer, peers.PeerDisconnected)
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
				peerConnectionState, err := s.peers.ConnectionState(remotePeer)
				if err == nil && (peerConnectionState == peers.PeerConnected || peerConnectionState == peers.PeerConnecting) {
					log.Trace(fmt.Sprintf("%s currentState:%d reason:already active, Ignoring connection request", peerInfoStr, peerConnectionState))
					return
				}
				s.peers.Add(nil /* QNR */, remotePeer, conn.RemoteMultiaddr(), conn.Stat().Direction)
				if s.peers.IsBad(remotePeer) {
					log.Trace(fmt.Sprintf("%s reason bad peer, Ignoring connection request.", peerInfoStr))
					disconnectFromPeer()
					return
				}
				validPeerConnection := func() {
					s.peers.SetConnectionState(conn.RemotePeer(), peers.PeerConnected)
					// Go through the handshake process.
					multiAddr := fmt.Sprintf("%s/p2p/%s", conn.RemoteMultiaddr().String(), conn.RemotePeer().String())
					log.Info(fmt.Sprintf("%s direction:%s multiAddr:%s activePeers:%d Peer Connected",
						peerInfoStr, conn.Stat().Direction, multiAddr, len(s.peers.Active())))
				}

				// Do not perform handshake on inbound dials.
				if conn.Stat().Direction == network.DirInbound {
					_, err := s.peers.ChainState(remotePeer)
					peerExists := err == nil
					//currentTime := time.Now()

					// Wait for peer to initiate handshake
					time.Sleep(timeForStatus)

					// Exit if we are disconnected with the peer.
					if s.host.Network().Connectedness(remotePeer) != network.Connected {
						return
					}

					// If peer hasn't sent a status request, we disconnect with them
					if _, err := s.peers.ChainState(remotePeer); err == peers.ErrPeerUnknown {
						disconnectFromPeer()
						return
					}
					if peerExists {
						/*updated*/ _, err := s.peers.ChainStateLastUpdated(remotePeer)
						if err != nil {
							disconnectFromPeer()
							return
						}
						// TODO Optimize Block DAG State
						// exit if we don't receive any current status messages from
						// peer.
						/*if updated.IsZero() || !updated.After(currentTime) {
							disconnectFromPeer()
							return
						}*/
					}
					validPeerConnection()
					return
				}

				s.peers.SetConnectionState(conn.RemotePeer(), peers.PeerConnecting)
				if err := reqFunc(context.Background(), conn.RemotePeer()); err != nil && err != io.EOF {
					log.Trace(fmt.Sprintf("%s Handshake failed", peerInfoStr))
					if err.Error() == "protocol not supported" {
						// This is only to ensure the smooth running of our testnets. This will not be
						// used in production.
						log.Debug("Not disconnecting peer with unsupported protocol. This maybe the relay node.")
						s.peers.SetConnectionState(conn.RemotePeer(), peers.PeerDisconnected)
						return
					}
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
			go func() {
				// Exit early if we are still connected to the peer.
				if net.Connectedness(conn.RemotePeer()) == network.Connected {
					return
				}
				priorState, err := s.peers.ConnectionState(conn.RemotePeer())
				if err != nil {
					// Can happen if the peer has already disconnected, so...
					priorState = peers.PeerDisconnected
				}
				s.peers.SetConnectionState(conn.RemotePeer(), peers.PeerDisconnecting)
				ctx := context.Background()
				if err := handler(ctx, conn.RemotePeer()); err != nil {
					log.Error(fmt.Sprintf("%s Disconnect handler failed", peerInfoStr))
				}
				s.peers.SetConnectionState(conn.RemotePeer(), peers.PeerDisconnected)
				// Only log disconnections if we were fully connected.
				if priorState == peers.PeerConnected {
					log.Info(fmt.Sprintf("%s Peer Disconnected,activePeers:%d", peerInfoStr, len(s.peers.Active())))
				}
			}()
		},
	})
}
