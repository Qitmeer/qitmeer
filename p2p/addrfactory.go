/*
 * Copyright (c) 2017-2020 The qitmeer developers
 */

package p2p

import (
	"fmt"
	"github.com/Qitmeer/qitmeer/p2p/qnode"
	"github.com/Qitmeer/qitmeer/p2p/qnr"
	iaddr "github.com/ipfs/go-ipfs-addr"
	"github.com/libp2p/go-libp2p-core/peer"
	ma "github.com/multiformats/go-multiaddr"
	"strings"
)

// withRelayAddrs returns an AddrFactory which will return Multiaddr via
// specified relay string in addition to existing MultiAddr.
func withRelayAddrs(relay string, addrs []ma.Multiaddr) []ma.Multiaddr {
	if relay == "" {
		return addrs
	}

	var relayAddrs []ma.Multiaddr

	for _, a := range addrs {
		if strings.Contains(a.String(), "/p2p-circuit") {
			continue
		}
		relayAddr, err := ma.NewMultiaddr(relay + "/p2p-circuit" + a.String())
		if err != nil {
			log.Error(fmt.Sprintf("Failed to create multiaddress for relay node: %v", err))
		} else {
			relayAddrs = append(relayAddrs, relayAddr)
		}
	}

	if len(relayAddrs) == 0 {
		log.Warn("Addresses via relay node are zero - using non-relay addresses")
		return addrs
	}
	return append(addrs, relayAddrs...)
}

func parseBootStrapAddrs(addrs []string) (discv5Nodes []string) {
	discv5Nodes, discvNodes := parseGenericAddrs(addrs)
	if len(discv5Nodes) == 0 && len(discvNodes) <= 0 {
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
		_, err = MultiAddrFromString(addr)
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

func convertQNRToMultiAddr(qnrStr string) (ma.Multiaddr, error) {
	node, err := qnode.Parse(qnode.ValidSchemes, qnrStr)
	if err != nil {
		return nil, err
	}
	// do not dial bootnodes with their tcp ports not set
	if err := node.Record().Load(qnr.WithEntry("tcp", new(qnr.TCP))); err != nil {
		if !qnr.IsNotFound(err) {
			err = fmt.Errorf("Could not retrieve tcp port:%v\n", err)
		}
		return nil, err
	}

	return convertToSingleMultiAddr(node)
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
		return nil, fmt.Errorf("node doesn't have an ip4 address, it's stated IP is %s", node.IP().String())
	}
	pubkey := node.Pubkey()
	assertedKey := convertToInterfacePubkey(pubkey)
	id, err := peer.IDFromPublicKey(assertedKey)
	if err != nil {
		return nil, fmt.Errorf("could not get peer id:%w", err)
	}
	multiAddrString := fmt.Sprintf("/ip4/%s/tcp/%d/p2p/%s", ip4.String(), node.TCP(), id)
	multiAddr, err := ma.NewMultiaddr(multiAddrString)
	if err != nil {
		return nil, fmt.Errorf("could not get multiaddr:%w", err)
	}
	return multiAddr, nil
}

func peersFromStringAddrs(addrs []string) ([]ma.Multiaddr, error) {
	var allAddrs []ma.Multiaddr
	qnodeString, multiAddrString := parseGenericAddrs(addrs)
	for _, stringAddr := range multiAddrString {
		addr, err := MultiAddrFromString(stringAddr)
		if err != nil {
			return nil, fmt.Errorf("Could not get multiaddr from string:%w", err)
		}
		allAddrs = append(allAddrs, addr)
	}
	for _, stringAddr := range qnodeString {
		qnodeAddr, err := qnode.Parse(qnode.ValidSchemes, stringAddr)
		if err != nil {
			return nil, fmt.Errorf("Could not get qnode from string:%w", err)
		}
		addr, err := convertToSingleMultiAddr(qnodeAddr)
		if err != nil {
			return nil, fmt.Errorf("Could not get multiaddr:%w", err)
		}
		allAddrs = append(allAddrs, addr)
	}
	return allAddrs, nil
}

func MultiAddrFromString(address string) (ma.Multiaddr, error) {
	addr, err := iaddr.ParseString(address)
	if err != nil {
		return nil, err
	}
	return addr.Multiaddr(), nil
}
