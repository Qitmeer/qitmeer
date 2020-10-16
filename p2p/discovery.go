/*
 * Copyright (c) 2020.
 * Project:qitmeer
 * File:discovery.go
 * Date:7/7/20 3:35 PM
 * Author:Jin
 * Email:lochjin@gmail.com
 */

package p2p

import (
	"crypto/ecdsa"
	"fmt"
	"net"

	"github.com/Qitmeer/qitmeer/p2p/discover"
	"github.com/Qitmeer/qitmeer/p2p/qnode"
	"github.com/Qitmeer/qitmeer/p2p/qnr"
	iaddr "github.com/ipfs/go-ipfs-addr"
	"github.com/libp2p/go-libp2p-core/network"
	"github.com/libp2p/go-libp2p-core/peer"
	ma "github.com/multiformats/go-multiaddr"
	"github.com/pkg/errors"
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
		return nil, errors.Wrap(err, "could not open node's peer database")
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
	pe := s.peers.Get(peerData.ID)
	if pe == nil {
		return false
	}
	if pe.IsBad() {
		return false
	}
	if pe.IsActive() {
		return false
	}
	if s.host.Network().Connectedness(peerData.ID) == network.Connected {
		return false
	}
	nodeQNR := node.Record()
	// Add peer to peer handler.
	s.peers.Add(nodeQNR, peerData.ID, multiAddr, network.DirUnknown)
	return true
}

// This checks our set max peers in our config, and
// determines whether our currently connected and
// active peers are above our set max peer limit.
func (s *Service) isPeerAtLimit() bool {
	numOfConns := len(s.host.Network().Peers())
	maxPeers := int(s.cfg.MaxPeers)
	activePeers := len(s.Peers().Active())

	return activePeers >= maxPeers || numOfConns >= maxPeers
}

func parseBootStrapAddrs(addrs []string) (discv5Nodes []string) {
	discv5Nodes, _ = parseGenericAddrs(addrs)
	if len(discv5Nodes) == 0 {
		log.Warn("No bootstrap addresses supplied")
	}
	return discv5Nodes
}

func parseGenericAddrs(addrs []string) (qnodeString []string, multiAddrString []string) {
	for _, addr := range addrs {
		if addr == "" {
			// Ignore empty entries
			continue
		}
		_, err := qnode.Parse(qnode.ValidSchemes, addr)
		if err == nil {
			qnodeString = append(qnodeString, addr)
			continue
		}
		_, err = multiAddrFromString(addr)
		if err == nil {
			multiAddrString = append(multiAddrString, addr)
			continue
		}
		log.Error(fmt.Sprintf("Invalid address of %s provided: %v", addr, err.Error()))
	}
	return qnodeString, multiAddrString
}

func convertToMultiAddr(nodes []*qnode.Node) []ma.Multiaddr {
	var multiAddrs []ma.Multiaddr
	for _, node := range nodes {
		// ignore nodes with no ip address stored
		if node.IP() == nil {
			continue
		}
		multiAddr, err := convertToSingleMultiAddr(node)
		if err != nil {
			log.Error(fmt.Sprintf("%s Could not convert to multiAddr", err.Error()))
			continue
		}
		multiAddrs = append(multiAddrs, multiAddr)
	}
	return multiAddrs
}

func convertToAddrInfo(node *qnode.Node) (*peer.AddrInfo, ma.Multiaddr, error) {
	multiAddr, err := convertToSingleMultiAddr(node)
	if err != nil {
		return nil, nil, err
	}
	info, err := peer.AddrInfoFromP2pAddr(multiAddr)
	if err != nil {
		return nil, nil, err
	}
	return info, multiAddr, nil
}

func convertToSingleMultiAddr(node *qnode.Node) (ma.Multiaddr, error) {
	ip4 := node.IP().To4()
	if ip4 == nil {
		return nil, errors.Errorf("node doesn't have an ip4 address, it's stated IP is %s", node.IP().String())
	}
	pubkey := node.Pubkey()
	assertedKey := convertToInterfacePubkey(pubkey)
	id, err := peer.IDFromPublicKey(assertedKey)
	if err != nil {
		return nil, errors.Wrap(err, "could not get peer id")
	}
	multiAddrString := fmt.Sprintf("/ip4/%s/tcp/%d/p2p/%s", ip4.String(), node.TCP(), id)
	multiAddr, err := ma.NewMultiaddr(multiAddrString)
	if err != nil {
		return nil, errors.Wrap(err, "could not get multiaddr")
	}
	return multiAddr, nil
}

func peersFromStringAddrs(addrs []string) ([]ma.Multiaddr, error) {
	var allAddrs []ma.Multiaddr
	qnodeString, multiAddrString := parseGenericAddrs(addrs)
	for _, stringAddr := range multiAddrString {
		addr, err := multiAddrFromString(stringAddr)
		if err != nil {
			return nil, errors.Wrapf(err, "Could not get multiaddr from string")
		}
		allAddrs = append(allAddrs, addr)
	}
	for _, stringAddr := range qnodeString {
		qnodeAddr, err := qnode.Parse(qnode.ValidSchemes, stringAddr)
		if err != nil {
			return nil, errors.Wrapf(err, "Could not get qnode from string")
		}
		addr, err := convertToSingleMultiAddr(qnodeAddr)
		if err != nil {
			return nil, errors.Wrapf(err, "Could not get multiaddr")
		}
		allAddrs = append(allAddrs, addr)
	}
	return allAddrs, nil
}

func multiAddrFromString(address string) (ma.Multiaddr, error) {
	addr, err := iaddr.ParseString(address)
	if err != nil {
		return nil, err
	}
	return addr.Multiaddr(), nil
}
