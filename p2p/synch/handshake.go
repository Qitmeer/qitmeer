/*
 * Copyright (c) 2017-2020 The qitmeer developers
 */

package synch

import (
	"context"
	"fmt"
	"github.com/Qitmeer/qitmeer/p2p/peers"
	"io"
	"sync/atomic"
	"time"

	"github.com/libp2p/go-libp2p-core/network"
	"github.com/libp2p/go-libp2p-core/peer"
)

const (
	// The time to wait for a chain state request.
	timeForChainState = 10 * time.Second
)

func (ps *PeerSync) Connected(pid peer.ID, conn network.Conn) {
	// Ignore if we are shutting down.
	if atomic.LoadInt32(&ps.shutdown) != 0 {
		return
	}

	ps.msgChan <- &ConnectedMsg{ID: pid, Conn: conn}
}

func (ps *PeerSync) processConnected(msg *ConnectedMsg) {
	peerInfoStr := fmt.Sprintf("peer:%s", msg.ID)
	remotePeer := msg.ID
	conn := msg.Conn
	remotePe := ps.sy.peers.Fetch(remotePeer)
	// Handle the various pre-existing conditions that will result in us not handshaking.
	peerConnectionState := remotePe.ConnectionState()
	if remotePe.IsActive() {
		log.Trace(fmt.Sprintf("%s currentState:%d reason:already active, Ignoring connection request", peerInfoStr, peerConnectionState))
		return
	}
	ps.sy.peers.Add(nil /* QNR */, remotePeer, conn.RemoteMultiaddr(), conn.Stat().Direction)
	if remotePe.IsBad() {
		log.Trace(fmt.Sprintf("%s reason bad peer, Ignoring connection request.", peerInfoStr))
		ps.Disconnect(remotePe)
		return
	}

	// Do not perform handshake on inbound dials.
	if conn.Stat().Direction == network.DirInbound {
		currentTime := time.Now()

		// Wait for peer to initiate handshake
		time.Sleep(timeForChainState)

		// Exit if we are disconnected with the peer.
		if ps.sy.p2p.Host().Network().Connectedness(remotePeer) != network.Connected {
			return
		}

		// If peer hasn't sent a status request, we disconnect with them
		if remotePe.ChainState() == nil {
			ps.Disconnect(remotePe)
			return
		}
		updated := remotePe.ChainStateLastUpdated()
		// exit if we don't receive any current status messages from
		// peer.
		if updated.IsZero() || !updated.After(currentTime) {
			ps.Disconnect(remotePe)
			return
		}
		ps.Connection(remotePe)
		return
	}

	remotePe.SetConnectionState(peers.PeerConnecting)
	if err := ps.sy.reValidatePeer(context.Background(), remotePeer); err != nil && err != io.EOF {
		log.Trace(fmt.Sprintf("%s Handshake failed", peerInfoStr))
		ps.Disconnect(remotePe)
		return
	}
	ps.Connection(remotePe)
}

func (ps *PeerSync) Connection(pe *peers.Peer) {
	if pe.ConnectionState().IsConnected() {
		return
	}
	pe.SetConnectionState(peers.PeerConnected)
	// Go through the handshake process.
	multiAddr := fmt.Sprintf("%s/p2p/%s", pe.Address().String(), pe.GetID().String())
	log.Info(fmt.Sprintf("%s direction:%s multiAddr:%s activePeers:%d Peer Connected",
		pe.GetID(), pe.Direction(), multiAddr, len(ps.sy.peers.Active())))

	ps.OnPeerConnected(pe)
}

func (ps *PeerSync) Disconnect(pe *peers.Peer) {
	if !pe.IsActive() {
		return
	}
	pe.SetConnectionState(peers.PeerDisconnecting)
	if err := ps.sy.p2p.Disconnect(pe.GetID()); err != nil {
		log.Error(fmt.Sprintf("%s Unable to disconnect from peer:%v", pe.GetID(), err))
	}
	// TODO some handle
	pe.SetConnectionState(peers.PeerDisconnected)
	log.Trace(fmt.Sprintf("Disconnect:%v", pe.GetID()))
	ps.OnPeerDisconnected(pe)
}

// AddConnectionHandler adds a callback function which handles the connection with a
// newly added peer. It performs a handshake with that peer by sending a hello request
// and validating the response from the peer.
func (s *Sync) AddConnectionHandler() {
	s.p2p.Host().Network().Notify(&network.NotifyBundle{
		ConnectedF: func(net network.Network, conn network.Conn) {
			remotePeer := conn.RemotePeer()
			log.Trace(fmt.Sprintf("ConnectedF:%s", remotePeer))
			s.peerSync.Connected(remotePeer, conn)
		},
	})
}

func (ps *PeerSync) Disconnected(pid peer.ID, conn network.Conn) {
	// Ignore if we are shutting down.
	if atomic.LoadInt32(&ps.shutdown) != 0 {
		return
	}

	ps.msgChan <- &DisconnectedMsg{ID: pid, Conn: conn}
}

func (ps *PeerSync) processDisconnected(msg *DisconnectedMsg) {

	peerInfoStr := fmt.Sprintf("peer:%s", msg.ID)
	// Must be handled in a goroutine as this callback cannot be blocking.
	pe := ps.sy.peers.Get(msg.ID)
	if pe == nil {
		return
	}
	if pe.ConnectionState().IsDisconnected() {
		return
	}
	// Exit early if we are still connected to the peer.
	if ps.sy.p2p.Host().Network().Connectedness(msg.ID) == network.Connected {
		return
	}
	priorState := pe.ConnectionState()

	pe.SetConnectionState(peers.PeerDisconnected)
	// Only log disconnections if we were fully connected.
	if priorState == peers.PeerConnected {
		log.Info(fmt.Sprintf("%s Peer Disconnected,activePeers:%d", peerInfoStr, len(ps.sy.peers.Active())))
		ps.OnPeerDisconnected(pe)
	}
}

// AddDisconnectionHandler disconnects from peers.  It handles updating the peer status.
// This also calls the handler responsible for maintaining other parts of the sync or p2p system.
func (s *Sync) AddDisconnectionHandler() {
	s.p2p.Host().Network().Notify(&network.NotifyBundle{
		DisconnectedF: func(net network.Network, conn network.Conn) {
			remotePeer := conn.RemotePeer()
			log.Trace(fmt.Sprintf("DisconnectedF:%s", remotePeer))
			s.peerSync.Disconnected(remotePeer, conn)
		},
	})
}
