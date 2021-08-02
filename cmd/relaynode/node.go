/*
 * Copyright (c) 2017-2020 The qitmeer developers
 */

package main

import (
	"context"
	"crypto/ecdsa"
	"fmt"
	"github.com/Qitmeer/qitmeer/common/roughtime"
	pv "github.com/Qitmeer/qitmeer/core/protocol"
	"github.com/Qitmeer/qitmeer/p2p"
	"github.com/Qitmeer/qitmeer/p2p/common"
	"github.com/Qitmeer/qitmeer/p2p/encoder"
	"github.com/Qitmeer/qitmeer/p2p/peers"
	pb "github.com/Qitmeer/qitmeer/p2p/proto/v1"
	"github.com/Qitmeer/qitmeer/p2p/synch"
	"github.com/Qitmeer/qitmeer/params"
	ds "github.com/ipfs/go-ds-leveldb"
	"github.com/libp2p/go-libp2p"
	"github.com/libp2p/go-libp2p-circuit"
	libp2pcore "github.com/libp2p/go-libp2p-core"
	"github.com/libp2p/go-libp2p-core/host"
	"github.com/libp2p/go-libp2p-core/network"
	"github.com/libp2p/go-libp2p-core/peer"
	"github.com/libp2p/go-libp2p-kad-dht"
	"github.com/libp2p/go-libp2p-kad-dht/opts"
	"github.com/libp2p/go-libp2p-noise"
	"github.com/libp2p/go-libp2p-peerstore/pstoreds"
	"github.com/libp2p/go-libp2p-secio"
	"github.com/multiformats/go-multiaddr"
	"path"
	"sync"
)

type Node struct {
	cfg        *Config
	ctx        context.Context
	cancel     context.CancelFunc
	privateKey *ecdsa.PrivateKey

	host host.Host

	peerStatus *peers.Status

	hslock sync.RWMutex
}

func (node *Node) init(cfg *Config) error {
	log.Info(fmt.Sprintf("Start relay node..."))
	node.ctx, node.cancel = context.WithCancel(context.Background())

	err := cfg.load()
	if err != nil {
		return err
	}
	node.cfg = cfg

	pk, err := p2p.PrivateKey(cfg.DataDir, cfg.PrivateKey, 0600)
	if err != nil {
		return err
	}
	node.privateKey = pk

	//
	node.peerStatus = peers.NewStatus(nil)

	log.Info(fmt.Sprintf("Load config completed"))
	log.Info(fmt.Sprintf("NetWork:%s  Genesis:%s", params.ActiveNetParams.Name, params.ActiveNetParams.GenesisHash.String()))
	return nil
}

func (node *Node) exit() error {
	node.cancel()
	log.Info(fmt.Sprintf("Stop relay node"))
	return nil
}

func (node *Node) run() error {
	log.Info(fmt.Sprintf("Run relay node..."))

	var exip string
	if len(node.cfg.ExternalIP) > 0 {
		exip = node.cfg.ExternalIP
	} else {
		eip := p2p.IpAddr()
		if eip == nil {
			return fmt.Errorf("Can't get IP")
		}
		exip = eip.String()
	}

	eMAddr, err := multiaddr.NewMultiaddr(fmt.Sprintf("/ip4/%s/tcp/%s", exip, node.cfg.Port))
	if err != nil {
		log.Error("Unable to construct multiaddr %v", err)
		return err
	}

	srcMAddr, err := multiaddr.NewMultiaddr(fmt.Sprintf("/ip4/%s/tcp/%s", defaultIP, node.cfg.Port))
	if err != nil {
		log.Error("Unable to construct multiaddr %v", err)
		return err
	}

	opts := []libp2p.Option{
		libp2p.EnableRelay(relay.OptHop),
		libp2p.ListenAddrs(srcMAddr, eMAddr),
		libp2p.Identity(p2p.ConvertToInterfacePrivkey(node.privateKey)),
	}

	if node.cfg.EnableNoise {
		opts = append(opts, libp2p.Security(noise.ID, noise.New), libp2p.Security(secio.ID, secio.New))
	} else {
		opts = append(opts, libp2p.Security(secio.ID, secio.New))
	}

	if node.cfg.HostDNS != "" {
		opts = append(opts, libp2p.AddrsFactory(func(addrs []multiaddr.Multiaddr) []multiaddr.Multiaddr {
			external, err := multiaddr.NewMultiaddr(fmt.Sprintf("/dns4/%s/tcp/%s", node.cfg.HostDNS, node.cfg.Port))
			if err != nil {
				log.Error(fmt.Sprintf("Unable to create external multiaddress:%v", err))
			} else {
				addrs = append(addrs, external)
			}
			return addrs
		}))
	}

	if node.cfg.UsePeerStore {
		ps, err := node.initPeerStore()
		if err != nil {
			log.Error(err.Error())
			return err
		}
		opts = append(opts, ps)
	}

	node.host, err = libp2p.New(
		node.ctx,
		opts...,
	)
	if err != nil {
		log.Error("Failed to create host %v", err)
		return err
	}

	err = node.registerHandlers()
	if err != nil {
		log.Error(err.Error())
		return err
	}

	kademliaDHT, err := dht.New(node.ctx, node.host, dhtopts.Protocols(p2p.ProtocolDHT))
	if err != nil {
		return err
	}

	err = kademliaDHT.Bootstrap(node.ctx)
	if err != nil {
		return err
	}

	log.Info(fmt.Sprintf("Relay Address: %s/p2p/%s\n", eMAddr.String(), node.host.ID()))
	log.Info("You can copy the relay address and configure it to the required Qitmeer-Node")
	if len(node.cfg.HostDNS) > 0 {
		logExternalDNSAddr(node.host.ID(), node.cfg.HostDNS, node.cfg.Port)
	}
	interrupt := interruptListener()
	<-interrupt
	return nil
}

func (node *Node) initPeerStore() (libp2p.Option, error) {
	dsPath := path.Join(node.cfg.DataDir, p2p.PeerStore)
	peerDS, err := ds.NewDatastore(dsPath, nil)
	if err != nil {
		return nil, err
	}
	log.Info(fmt.Sprintf("Start Peers from:%s", dsPath))

	ps, err := pstoreds.NewPeerstore(node.ctx, peerDS, pstoreds.DefaultOpts())
	if err != nil {
		return nil, err
	}
	return libp2p.Peerstore(ps), nil
}

func (node *Node) registerHandlers() error {

	node.host.Network().Notify(&network.NotifyBundle{
		ConnectedF: func(net network.Network, conn network.Conn) {
			remotePeer := conn.RemotePeer()
			go node.processConnected(remotePeer, conn)
		},
	})

	node.host.Network().Notify(&network.NotifyBundle{
		DisconnectedF: func(net network.Network, conn network.Conn) {
			remotePeer := conn.RemotePeer()
			go node.processDisconnected(remotePeer, conn)
		},
	})
	//

	synch.RegisterRPC(
		node,
		synch.RPCChainState,
		&pb.ChainState{},
		node.chainStateHandler,
	)

	return nil
}

func (node *Node) Encoding() encoder.NetworkEncoding {
	return &encoder.SszNetworkEncoder{UseSnappyCompression: true}
}

func (node *Node) Host() host.Host {
	return node.host
}

func (node *Node) Context() context.Context {
	return node.ctx
}

func (node *Node) Disconnect(pid peer.ID) error {
	return node.host.Network().ClosePeer(pid)
}

func (node *Node) IncreaseBytesSent(pid peer.ID, size int) {
}

func (node *Node) IncreaseBytesRecv(pid peer.ID, size int) {
}

func (node *Node) chainStateHandler(ctx context.Context, msg interface{}, stream libp2pcore.Stream) *common.Error {
	pe := node.peerStatus.Get(stream.Conn().RemotePeer())
	if pe == nil {
		return synch.ErrPeerUnknown
	}
	log.Trace(fmt.Sprintf("chainStateHandler:%s", pe.GetID()))

	ctx, cancel := context.WithTimeout(ctx, synch.HandleTimeout)
	defer cancel()

	m, ok := msg.(*pb.ChainState)
	if !ok {
		return synch.ErrMessage(fmt.Errorf("message is not type *pb.ChainState"))
	}

	pe.SetChainState(m)

	return synch.EncodeResponseMsg(node, stream, node.getChainState(), common.ErrNone)
}

func (node *Node) processConnected(pid peer.ID, conn network.Conn) {
	node.hslock.Lock()
	defer node.hslock.Unlock()

	peerInfoStr := fmt.Sprintf("peer:%s", pid)
	remotePe := node.peerStatus.Fetch(pid)
	peerConnectionState := remotePe.ConnectionState()
	if remotePe.IsActive() {
		log.Trace(fmt.Sprintf("%s currentState:%d reason:already active, Ignoring connection request", peerInfoStr, peerConnectionState))
		return
	}
	node.peerStatus.Add(nil /* QNR */, pid, conn.RemoteMultiaddr(), conn.Stat().Direction)
	if remotePe.IsBad() {
		log.Trace(fmt.Sprintf("%s reason bad peer, Ignoring connection request.", peerInfoStr))
		node.Disconnect(pid)
		return
	}
	remotePe.SetConnectionState(peers.PeerConnected)
	// Go through the handshake process.
	multiAddr := fmt.Sprintf("%s/p2p/%s", remotePe.Address().String(), remotePe.GetID().String())

	log.Info(fmt.Sprintf("%s direction:%s multiAddr:%s",
		remotePe.GetID(), remotePe.Direction(), multiAddr))
}

func (node *Node) processDisconnected(pid peer.ID, conn network.Conn) {
	node.hslock.Lock()
	defer node.hslock.Unlock()

	peerInfoStr := fmt.Sprintf("peer:%s", pid)
	pe := node.peerStatus.Get(pid)
	if pe == nil {
		return
	}
	if pe.ConnectionState().IsDisconnected() {
		return
	}
	// Exit early if we are still connected to the peer.
	if node.Host().Network().Connectedness(pid) == network.Connected {
		return
	}
	priorState := pe.ConnectionState()

	pe.SetConnectionState(peers.PeerDisconnected)
	// Only log disconnections if we were fully connected.
	if priorState == peers.PeerConnected {
		log.Info(fmt.Sprintf("%s Peer Disconnected", peerInfoStr))
	}
}

func (node *Node) getChainState() *pb.ChainState {
	genesisHash := params.ActiveNetParams.GenesisHash

	gs := &pb.GraphState{
		Total:      1,
		Layer:      0,
		MainHeight: 0,
		MainOrder:  0,
		Tips:       []*pb.Hash{},
	}
	gs.Tips = append(gs.Tips, &pb.Hash{Hash: genesisHash.Bytes()})

	return &pb.ChainState{
		GenesisHash:     &pb.Hash{Hash: genesisHash.Bytes()},
		ProtocolVersion: pv.ProtocolVersion,
		Timestamp:       uint64(roughtime.Now().Unix()),
		Services:        uint64(pv.Relay),
		GraphState:      gs,
		UserAgent:       []byte(p2p.BuildUserAgent("Qitmeer-relay")),
		DisableRelayTx:  true,
	}
}

func closeSteam(stream libp2pcore.Stream) {
	if err := stream.Close(); err != nil {
		log.Error(fmt.Sprintf("Failed to close stream:%v", err))
	}
}

func logExternalDNSAddr(id peer.ID, addr string, port string) {
	if addr != "" {
		log.Info(fmt.Sprintf("Relay node started external p2p server:multiAddr=%s", "/dns4/"+addr+"/tcp/"+port+"/p2p/"+id.String()))
	}
}
