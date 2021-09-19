/*
 * Copyright (c) 2017-2020 The qitmeer developers
 */

package synch

import (
	"bufio"
	"fmt"
	"github.com/Qitmeer/qitmeer/core/protocol"
	"github.com/Qitmeer/qitmeer/p2p/peers"
	"github.com/multiformats/go-multistream"
	"io"
	"net"
	"strings"
	"sync/atomic"
	"time"

	"github.com/libp2p/go-libp2p-core/network"
	"github.com/libp2p/go-libp2p-core/peer"
)

const (
	// The time to wait for a chain state request.
	timeForChainState = 10 * time.Second

	timeForBidirChan = 4 * time.Second

	timeForBidirChanLife = 10 * time.Minute
)

func (ps *PeerSync) Connected(pid peer.ID, conn network.Conn) {
	// Ignore if we are shutting down.
	if atomic.LoadInt32(&ps.shutdown) != 0 {
		return
	}

	//ps.msgChan <- &ConnectedMsg{ID: pid, Conn: conn}
	go ps.processConnected(&ConnectedMsg{ID: pid, Conn: conn})
}

func (ps *PeerSync) processConnected(msg *ConnectedMsg) {
	remotePe := ps.sy.peers.Fetch(msg.ID)

	remotePe.HSlock.Lock()
	defer remotePe.HSlock.Unlock()

	peerInfoStr := fmt.Sprintf("peer:%s", msg.ID)
	remotePeer := msg.ID
	conn := msg.Conn
	// Handle the various pre-existing conditions that will result in us not handshaking.
	peerConnectionState := remotePe.ConnectionState()
	if remotePe.IsActive() {
		log.Trace(fmt.Sprintf("%s currentState:%d reason:already active, Ignoring connection request", peerInfoStr, peerConnectionState))
		return
	}
	ps.sy.peers.Add(nil /* QNR */, remotePeer, conn.RemoteMultiaddr(), conn.Stat().Direction)
	if remotePe.IsBad() && !ps.sy.IsWhitePeer(remotePeer) {
		log.Trace(fmt.Sprintf("%s reason bad peer, Ignoring connection request.", peerInfoStr))
		ps.Disconnect(remotePe)
		return
	}
	if time.Since(remotePe.ConnectionTime()) <= time.Second {
		ps.sy.Peers().IncrementBadResponses(remotePeer, "Connection is too frequent")
		log.Debug(fmt.Sprintf("%s is too frequent, so I'll deduct you points", remotePeer))
	}
	remotePe.SetConnectionState(peers.PeerConnecting)

	// Do not perform handshake on inbound dials.
	if conn.Stat().Direction == network.DirInbound {
		return
	}

	if err := ps.sy.reValidatePeer(ps.sy.p2p.Context(), remotePeer); err != nil && err != io.EOF {
		log.Trace(fmt.Sprintf("%s Handshake failed (%s)", peerInfoStr, err))
		ps.Disconnect(remotePe)
		return
	}
	ps.Connection(remotePe)
}

func (ps *PeerSync) immediatelyConnected(pe *peers.Peer) {
	pe.HSlock.Lock()
	defer pe.HSlock.Unlock()

	if !pe.ConnectionState().IsConnecting() {
		go ps.PeerUpdate(pe, true)
		return
	}
	ps.Connection(pe)
}

func (ps *PeerSync) Connection(pe *peers.Peer) {
	if pe.ConnectionState().IsConnected() {
		return
	}
	pe.SetConnectionState(peers.PeerConnected)
	// Go through the handshake process.
	multiAddr := fmt.Sprintf("%s/p2p/%s", pe.Address().String(), pe.GetID().String())

	if !pe.IsConsensus() {
		log.Info(fmt.Sprintf("%s direction:%s multiAddr:%s  (%s)",
			pe.GetID(), pe.Direction(), multiAddr, pe.Services().String()))
		return
	}
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
	if !pe.IsConsensus() {
		if pe.Services() == protocol.Unknown {
			log.Trace(fmt.Sprintf("Disconnect:%v ", pe.IDWithAddress()))
		} else {
			log.Trace(fmt.Sprintf("Disconnect:%v (%s)", pe.IDWithAddress(), pe.Services().String()))
		}
		return
	}

	log.Trace(fmt.Sprintf("Disconnect:%v ", pe.IDWithAddress()))
	ps.OnPeerDisconnected(pe)
}

// AddConnectionHandler adds a callback function which handles the connection with a
// newly added peer. It performs a handshake with that peer by sending a hello request
// and validating the response from the peer.
func (s *Sync) AddConnectionHandler() {
	s.connectionNotify = &network.NotifyBundle{
		ConnectedF: func(net network.Network, conn network.Conn) {
			remotePeer := conn.RemotePeer()
			if !s.connectionGater(remotePeer, conn) {
				return
			}
			log.Trace(fmt.Sprintf("ConnectedF:%s, %v ", remotePeer, conn.RemoteMultiaddr()))
			s.peerSync.Connected(remotePeer, conn)
		},
	}
	s.p2p.Host().Network().Notify(s.connectionNotify)
}

func (ps *PeerSync) Disconnected(pid peer.ID, conn network.Conn) {
	// Ignore if we are shutting down.
	if atomic.LoadInt32(&ps.shutdown) != 0 {
		return
	}

	//ps.msgChan <- &DisconnectedMsg{ID: pid, Conn: conn}
	go ps.processDisconnected(&DisconnectedMsg{ID: pid, Conn: conn})
}

func (ps *PeerSync) processDisconnected(msg *DisconnectedMsg) {
	// Must be handled in a goroutine as this callback cannot be blocking.
	pe := ps.sy.peers.Get(msg.ID)
	if pe == nil {
		return
	}

	pe.HSlock.Lock()
	defer pe.HSlock.Unlock()

	peerInfoStr := fmt.Sprintf("peer:%s", msg.ID)

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
	s.disconnectionNotify = &network.NotifyBundle{
		DisconnectedF: func(net network.Network, conn network.Conn) {
			remotePeer := conn.RemotePeer()
			log.Trace(fmt.Sprintf("DisconnectedF:%s", remotePeer))
			s.peerSync.Disconnected(remotePeer, conn)
		},
	}
	s.p2p.Host().Network().Notify(s.disconnectionNotify)
}

func (s *Sync) bidirectionalChannelCapacity(pe *peers.Peer, conn network.Conn) bool {
	if conn.Stat().Direction == network.DirOutbound {
		pe.SetBidChanCap(time.Now())
		return true
	}
	if s.p2p.Config().IsCircuit || s.IsWhitePeer(pe.GetID()) {
		pe.SetBidChanCap(time.Now())
		return true
	}

	bidChanLife := pe.GetBidChanCap()
	if !bidChanLife.IsZero() {
		if time.Since(bidChanLife) < timeForBidirChanLife {
			return true
		}
	}

	//
	peAddr := conn.RemoteMultiaddr()
	ipAddr := ""
	protocol := ""
	port := ""
	ps := peAddr.Protocols()
	if len(ps) >= 1 {
		ia, err := peAddr.ValueForProtocol(ps[0].Code)
		if err != nil {
			log.Debug(err.Error())
			pe.SetBidChanCap(time.Time{})
			return false
		}
		ipAddr = ia
	}
	if len(ps) >= 2 {
		protocol = ps[1].Name
		po, err := peAddr.ValueForProtocol(ps[1].Code)
		if err != nil {
			log.Debug(err.Error())
			pe.SetBidChanCap(time.Time{})
			return false
		}
		port = po
	}
	if len(ipAddr) <= 0 ||
		len(protocol) <= 0 ||
		len(port) <= 0 {
	}
	bidConn, err := net.DialTimeout(protocol, fmt.Sprintf("%s:%s", ipAddr, port), timeForBidirChan)
	if err != nil {
		log.Debug(err.Error())
		pe.SetBidChanCap(time.Time{})
		return false
	}
	reply, err := bufio.NewReader(bidConn).ReadString('\n')
	if err != nil {
		log.Debug(err.Error())
		pe.SetBidChanCap(time.Time{})
		return false
	}
	if !strings.Contains(reply, multistream.ProtocolID) {
		log.Debug(fmt.Sprintf("BidChan protocol is error"))
		pe.SetBidChanCap(time.Time{})
		return false
	}
	log.Debug(fmt.Sprintf("Bidirectional channel capacity:%s", pe.GetID().String()))
	bidConn.Write([]byte(fmt.Sprintf("%s\n", multistream.ProtocolID)))
	bidConn.Close()

	pe.SetBidChanCap(time.Now())
	return true
}

func (s *Sync) IsWhitePeer(pid peer.ID) bool {
	_, ok := s.LANPeers[pid]
	return ok
}

func (s *Sync) IsPeerAtLimit() bool {
	//numOfConns := len(s.p2p.Host().Network().Peers())
	maxPeers := int(s.p2p.Config().MaxPeers)
	activePeers := len(s.Peers().Active())

	return activePeers >= maxPeers
}

func (s *Sync) IsInboundPeerAtLimit() bool {
	return len(s.Peers().DirInbound()) >= s.p2p.Config().MaxInbound
}

func (s *Sync) connectionGater(pid peer.ID, conn network.Conn) bool {
	ret := true
	if s.IsWhitePeer(pid) {
		return ret
	}
	if s.IsPeerAtLimit() {
		log.Trace(fmt.Sprintf("connectionGater  peer:%s reason:at peer max limit", pid.String()))
		ret = false
	}
	if ret {
		if conn.Stat().Direction == network.DirInbound {
			if s.IsInboundPeerAtLimit() {
				log.Trace(fmt.Sprintf("peer:%s reason:at peer limit,Not accepting inbound dial", pid.String()))
				ret = false
			}
		}
	}

	if !ret {
		if err := s.p2p.Disconnect(pid); err != nil {
			log.Error(fmt.Sprintf("%s Unable to disconnect from peer:%v", pid, err))
		}
	}
	return true
}
