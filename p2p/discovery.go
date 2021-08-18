/*
 * Copyright (c) 2017-2020 The qitmeer developers
 */

package p2p

import (
	"crypto/ecdsa"
	"fmt"
	"github.com/Qitmeer/qitmeer/p2p/discover"
	"github.com/Qitmeer/qitmeer/p2p/qnode"
	"github.com/Qitmeer/qitmeer/p2p/qnr"
	"github.com/libp2p/go-libp2p-core/network"
	"github.com/libp2p/go-libp2p-core/protocol"
	"github.com/libp2p/go-libp2p-kad-dht"
	"github.com/libp2p/go-libp2p-kad-dht/opts"
	"net"
)

const (
	ProtocolDHT protocol.ID = "/qitmeer/kad/1.0.0"
)

// Listener defines the discovery V5 network interface that is used
// to communicate with other peers.
type Listener interface {
	Self() *qnode.Node
	Close()
	Lookup(qnode.ID) []*qnode.Node
	Resolve(*qnode.Node) *qnode.Node
	RandomNodes() qnode.Iterator
	Ping(*qnode.Node) error
	RequestQNR(*qnode.Node) (*qnode.Node, error)
	LocalNode() *qnode.LocalNode
}

func (s *Service) createListener(
	ipAddr net.IP,
	privKey *ecdsa.PrivateKey,
) *discover.UDPv5 {
	udpAddr := &net.UDPAddr{
		IP:   ipAddr,
		Port: int(s.cfg.UDPPort),
	}
	// assume ip is either ipv4 or ipv6
	networkVersion := ""
	if ipAddr.To4() != nil {
		networkVersion = "udp4"
	} else {
		networkVersion = "udp6"
	}
	conn, err := net.ListenUDP(networkVersion, udpAddr)
	if err != nil {
		log.Error(err.Error())
		return nil
	}
	localNode, err := s.createLocalNode(
		privKey,
		ipAddr,
		int(s.cfg.UDPPort),
		int(s.cfg.TCPPort),
	)
	if err != nil {
		log.Error(err.Error())
		return nil
	}
	if s.cfg.HostAddress != "" {
		hostIP := net.ParseIP(s.cfg.HostAddress)
		if hostIP.To4() == nil && hostIP.To16() == nil {
			log.Error(fmt.Sprintf("Invalid host address given: %s", hostIP.String()))
		} else {
			localNode.SetFallbackIP(hostIP)
			localNode.SetStaticIP(hostIP)
		}
	}
	dv5Cfg := discover.Config{
		PrivateKey: privKey,
	}
	dv5Cfg.Bootnodes = []*qnode.Node{}
	for _, addr := range s.cfg.Discv5BootStrapAddr {
		bootNode, err := qnode.Parse(qnode.ValidSchemes, addr)
		if err != nil {
			log.Error(err.Error())
			return nil
		}
		dv5Cfg.Bootnodes = append(dv5Cfg.Bootnodes, bootNode)
	}

	network, err := discover.ListenV5(conn, localNode, dv5Cfg)
	if err != nil {
		log.Error(err.Error())
		return nil
	}
	return network
}

func (s *Service) createLocalNode(
	privKey *ecdsa.PrivateKey,
	ipAddr net.IP,
	udpPort int,
	tcpPort int,
) (*qnode.LocalNode, error) {
	db, err := qnode.OpenDB("")
	if err != nil {
		return nil, fmt.Errorf("could not open node's peer database:%w", err)
	}
	localNode := qnode.NewLocalNode(db, privKey)
	ipEntry := qnr.IP(ipAddr)
	udpEntry := qnr.UDP(udpPort)
	tcpEntry := qnr.TCP(tcpPort)
	localNode.Set(ipEntry)
	localNode.Set(udpEntry)
	localNode.Set(tcpEntry)
	localNode.SetFallbackIP(ipAddr)
	localNode.SetFallbackUDP(udpPort)

	return localNode, nil
}

func (s *Service) startDiscoveryV5(
	addr net.IP,
	privKey *ecdsa.PrivateKey,
) (*discover.UDPv5, error) {
	listener := s.createListener(addr, privKey)
	record := listener.Self()
	log.Info(fmt.Sprintf("Started discovery v5:QNR(%s)", record.String()))
	return listener, nil
}

// filterPeer validates each node that we retrieve from our dht. We
// try to ascertain that the peer can be a valid protocol peer.
// Validity Conditions:
// 1) The local node is still actively looking for peers to
//    connect to.
// 2) Peer has a valid IP and TCP port set in their qnr.
// 3) Peer hasn't been marked as 'bad'
// 4) Peer is not currently active or connected.

func (s *Service) filterPeer(node *qnode.Node) bool {
	// ignore nodes with no ip address stored.
	if node.IP() == nil {
		return false
	}
	// do not dial nodes with their tcp ports not set
	if err := node.Record().Load(qnr.WithEntry("tcp", new(qnr.TCP))); err != nil {
		if !qnr.IsNotFound(err) {
			log.Debug(fmt.Sprintf("%s Could not retrieve tcp port", err.Error()))
		}
		return false
	}
	peerData, multiAddr, err := convertToAddrInfo(node)
	if err != nil {
		log.Debug(fmt.Sprintf("%s Could not convert to peer data", err.Error()))
		return false
	}
	nodeQNR := node.Record()
	pe := s.Peers().Fetch(peerData.ID)
	pe.SetQNR(nodeQNR)

	if pe.IsBad() {
		return false
	}
	if pe.IsActive() {
		return false
	}
	if s.host.Network().Connectedness(peerData.ID) == network.Connected {
		return false
	}
	// Add peer to peer handler.
	s.Peers().Add(nodeQNR, peerData.ID, multiAddr, network.DirUnknown)
	return true
}

// This checks our set max peers in our config, and
// determines whether our currently connected and
// active peers are above our set max peer limit.
func (s *Service) isPeerAtLimit() bool {
	return s.sy.IsPeerAtLimit()
}

func (s *Service) isInboundPeerAtLimit() bool {
	return s.sy.IsInboundPeerAtLimit()
}

func (s *Service) startKademliaDHT() error {
	kademliaDHT, err := dht.New(s.ctx, s.host, dhtopts.Protocols(ProtocolDHT))
	if err != nil {
		return err
	}
	s.kademliaDHT = kademliaDHT

	err = kademliaDHT.Bootstrap(s.ctx)
	if err != nil {
		return err
	}
	return nil
}
