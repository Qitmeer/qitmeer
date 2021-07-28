/*
 * Copyright (c) 2017-2020 The qitmeer developers
 */

package node

import (
	"context"
	"crypto/ecdsa"
	"fmt"
	"github.com/Qitmeer/qitmeer/cmd/crawler/config"
	"github.com/Qitmeer/qitmeer/cmd/crawler/log"
	"github.com/Qitmeer/qitmeer/cmd/crawler/peers"
	"github.com/Qitmeer/qitmeer/cmd/crawler/rpc"
	"github.com/Qitmeer/qitmeer/common/roughtime"
	pv "github.com/Qitmeer/qitmeer/core/protocol"
	"github.com/Qitmeer/qitmeer/crypto/ecc/secp256k1"
	"github.com/Qitmeer/qitmeer/p2p"
	"github.com/Qitmeer/qitmeer/p2p/common"
	"github.com/Qitmeer/qitmeer/p2p/encoder"
	pb "github.com/Qitmeer/qitmeer/p2p/proto/v1"
	"github.com/Qitmeer/qitmeer/p2p/qnode"
	"github.com/Qitmeer/qitmeer/p2p/synch"
	"github.com/Qitmeer/qitmeer/params"
	iaddr "github.com/ipfs/go-ipfs-addr"
	"github.com/libp2p/go-libp2p"
	libp2pcore "github.com/libp2p/go-libp2p-core"
	"github.com/libp2p/go-libp2p-core/crypto"
	"github.com/libp2p/go-libp2p-core/host"
	"github.com/libp2p/go-libp2p-core/network"
	"github.com/libp2p/go-libp2p-core/peer"
	"github.com/libp2p/go-libp2p-kad-dht"
	"github.com/libp2p/go-libp2p-kad-dht/opts"
	"github.com/libp2p/go-libp2p-noise"
	"github.com/libp2p/go-libp2p-secio"
	"github.com/multiformats/go-multiaddr"
	"sync"
	"time"
)

type Service interface {

	// APIs retrieves the list of RPC descriptors the service provides
	APIs() []rpc.API

	// Start is called after all services have been constructed and the networking
	// layer was also initialized to spawn any goroutines required by the service.
	Start() error

	// Stop terminates all goroutines belonging to the service, blocking until they
	// are all terminated.
	Stop() error
}

type Node struct {
	cfg        *config.Config
	ctx        context.Context
	cancel     context.CancelFunc
	privateKey *ecdsa.PrivateKey
	peers      *peers.Peers
	rpcServer  *rpc.RpcServer
	interrupt  chan struct{}
	wg         sync.WaitGroup
	host       host.Host
	service    []Service
}

func (node *Node) Init(cfg *config.Config) error {
	var err error
	log.Log.Info(fmt.Sprintf("Start crawler node..."))
	node.ctx, node.cancel = context.WithCancel(context.Background())
	node.peers, err = peers.NewPeers()
	if err != nil {
		return err
	}
	node.service = []Service{node.peers}
	if err = cfg.Load(); err != nil {
		return err
	}
	node.cfg = cfg
	node.rpcServer, err = rpc.NewRPCServer(cfg)
	if err != nil {
		return err
	}
	pk, err := p2p.PrivateKey(cfg.DataDir, cfg.PrivateKey, 0600)
	if err != nil {
		return err
	}
	node.privateKey = pk
	node.interrupt = make(chan struct{}, 1)
	node.wg = sync.WaitGroup{}
	log.Log.Info(fmt.Sprintf("Load config completed"))
	log.Log.Info(fmt.Sprintf("NetWork:%s  Genesis:%s", params.ActiveNetParams.Name, params.ActiveNetParams.GenesisHash.String()))
	return nil
}

func (node *Node) Exit() error {
	node.rpcServer.Stop()
	for _, ser := range node.service {
		ser.Stop()
	}
	node.cancel()
	log.Log.Info(fmt.Sprintf("Stop crawler node"))
	return nil
}

func (node *Node) startRPC(services []Service) error {
	// Gather all the possible APIs to surface
	apis := []rpc.API{}
	for _, service := range services {
		apis = append(apis, service.APIs()...)
	}

	// Register all the APIs exposed by the services
	for _, api := range apis {
		if err := node.rpcServer.RegisterService(api.NameSpace, api.Service); err != nil {
			return err
		}
	}
	if err := node.rpcServer.Start(); err != nil {
		return err
	}
	return nil
}

func (node *Node) Run() error {
	log.Log.Info(fmt.Sprintf("Run crawler node..."))
	err := node.startRPC(node.service)
	if err != nil {
		log.Log.Error("Failed to start rpc %v", err)
		return err
	}
	for _, ser := range node.service {
		if err = ser.Start(); err != nil {
			log.Log.Error("Failed to start up service %v", err)
			return err
		}
	}

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
		log.Log.Error("Unable to construct multiaddr %v", err)
		return err
	}

	srcMAddr, err := multiaddr.NewMultiaddr(fmt.Sprintf("/ip4/%s/tcp/%s", config.DefaultIP, node.cfg.Port))
	if err != nil {
		log.Log.Error("Unable to construct multiaddr %v", err)
		return err
	}

	opts := []libp2p.Option{
		//libp2p.EnableRelay(relay.OptHop),
		libp2p.ListenAddrs(srcMAddr, eMAddr),
		libp2p.Identity(p2p.ConvertToInterfacePrivkey(node.privateKey)),
	}

	if node.cfg.EnableNoise {
		opts = append(opts, libp2p.Security(noise.ID, noise.New), libp2p.Security(secio.ID, secio.New))
	} else {
		opts = append(opts, libp2p.Security(secio.ID, secio.New))
	}

	node.host, err = libp2p.New(
		node.ctx,
		opts...,
	)
	if err != nil {
		log.Log.Error("Failed to create host %v", err)
		return err
	}

	err = node.registerHandlers()
	if err != nil {
		log.Log.Error(err.Error())
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

	log.Log.Info(fmt.Sprintf("crawler Address: %s/p2p/%s\n", eMAddr.String(), node.host.ID()))
	log.Log.Info("You can copy the crawler address and configure it to the required Qitmeer-Node")

	var peersToWatch []string
	_, bootstrapAddrs := parseGenericAddrs(node.cfg.BootstrapNodeAddr)
	if len(bootstrapAddrs) > 0 {
		peersToWatch = append(peersToWatch, bootstrapAddrs...)
	}
	if len(node.cfg.StaticPeers) > 0 {
		bootstrapAddrs = append(bootstrapAddrs, node.cfg.StaticPeers...)
	}

	if len(bootstrapAddrs) > 0 {
		addrs, err := peersFromStringAddrs(bootstrapAddrs)
		if err != nil {
			log.Log.Error(fmt.Sprintf("Could not connect to static peer: %v", err))
		} else {
			node.connectWithAllPeers(addrs)
		}
	}
	node.wg.Add(1)
	go node.connectWithNewPeers()

	node.wg.Add(1)
	go node.interruptListener()

	node.wg.Wait()
	return nil
}

func (node *Node) connectWithNewPeers() {
	ti := time.NewTicker(30 * time.Second)
	defer func() {
		ti.Stop()
		node.wg.Done()
	}()

	for {
		select {
		case <-node.interrupt:
			log.Log.Info("Shutdown waiting for connection...")
			return
		case <-ti.C:
			peers := node.peers.FindPeerList()
			for _, p := range peers {
				select {
				case <-node.interrupt:
					log.Log.Info("Shutdown connect peer...")
					return
				default:
					node.connectWithAddr(p.Id, p.Addr)
				}
			}
			node.printResult()
		}
	}
}

func (node *Node) connectWithAllPeers(multiAddrs []multiaddr.Multiaddr) {
	addrInfos, err := peer.AddrInfosFromP2pAddrs(multiAddrs...)
	if err != nil {
		log.Log.Error(fmt.Sprintf("Could not convert to peer address info's from multiaddresses: %v", err))
		return
	}
	for _, info := range addrInfos {
		go func(info peer.AddrInfo) {
			if err := node.connectWithPeer(info, false); err != nil {
				log.Log.Trace(fmt.Sprintf("Could not connect with peer %s :%v", info.String(), err))
			}
		}(info)
	}
}

func (node *Node) connectWithPeer(info peer.AddrInfo, force bool) error {
	if info.ID == node.host.ID() {
		return nil
	}
	if err := node.host.Connect(node.ctx, info); err != nil {
		return err
	}
	node.peers.UpdateConnectTime(info.ID.String(), time.Now().Unix())
	return nil
}

func (node *Node) connectWithAddr(id string, addr string) {
	formatStr := getConnPeerAddress(id, addr)
	maAddr, err := multiAddrFromString(formatStr)
	if err != nil {
		log.Log.Trace(fmt.Sprintf("Wrong remote address %s to p2p address, %v", formatStr, err))
		return
	}
	addrInfo, err := peer.AddrInfoFromP2pAddr(maAddr)
	if err != nil {
		log.Log.Trace(fmt.Sprintf("Wrong remote address %s to p2p address, %v", addr, err))
		return
	}
	go func(info peer.AddrInfo) {
		if err := node.connectWithPeer(info, false); err != nil {
			log.Log.Info(fmt.Sprintf("Could not connect with peer %s :%v", info.String(), err))
		}
	}(*addrInfo)
}

func (node *Node) registerHandlers() error {

	node.host.Network().Notify(&network.NotifyBundle{
		ConnectedF: func(net network.Network, conn network.Conn) {
			remotePeer := conn.RemotePeer()
			//log.Log.Info(fmt.Sprintf("Connected:%s (%s)", remotePeer, conn.RemoteMultiaddr()))
			node.peers.Add(remotePeer.String(), conn.RemoteMultiaddr().String())
		},
	})

	node.host.Network().Notify(&network.NotifyBundle{
		DisconnectedF: func(net network.Network, conn network.Conn) {
			remotePeer := conn.RemotePeer()
			node.peers.UpdateUnConnected(remotePeer.String())
			log.Log.Info(fmt.Sprintf("Disconnected:%s (%s)", remotePeer, conn.RemoteMultiaddr()))
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
	pid := stream.Conn().RemotePeer()
	log.Log.Trace(fmt.Sprintf("chainStateHandler:%s", pid))

	ctx, cancel := context.WithTimeout(ctx, synch.HandleTimeout)
	defer cancel()

	genesisHash := params.ActiveNetParams.GenesisHash

	gs := &pb.GraphState{
		Total:      1,
		Layer:      0,
		MainHeight: 0,
		MainOrder:  0,
		Tips:       []*pb.Hash{},
	}
	gs.Tips = append(gs.Tips, &pb.Hash{Hash: genesisHash.Bytes()})

	resp := &pb.ChainState{
		GenesisHash:     &pb.Hash{Hash: genesisHash.Bytes()},
		ProtocolVersion: pv.ProtocolVersion,
		Timestamp:       uint64(roughtime.Now().Unix()),
		Services:        uint64(pv.Observer),
		GraphState:      gs,
		UserAgent:       []byte("qitmeer-crawler"),
		DisableRelayTx:  true,
	}
	return synch.EncodeResponseMsg(node, stream, resp, common.ErrNone)
}

func (node *Node) printResult() {
	all := node.peers.All()

	log.Log.Info(fmt.Sprintf("Find %d peers", len(all)))
	for _, peer := range all {
		log.Log.Info(fmt.Sprintf("Peer:%s, connected time:%s, connected:%v",
			getConnPeerAddress(peer.Id, peer.Addr),
			time.Unix(peer.ConnectTime, 0).String(),
			peer.Connected))
	}
}

func (node *Node) interruptListener() {
	interrupt := interruptListener()
	<-interrupt
	close(node.interrupt)
	node.wg.Done()
}

func closeSteam(stream libp2pcore.Stream) {
	if err := stream.Close(); err != nil {
		log.Log.Error(fmt.Sprintf("Failed to close stream:%v", err))
	}
}

func parseBootStrapAddrs(addrs []string) (discv5Nodes []string) {
	discv5Nodes, discvNodes := parseGenericAddrs(addrs)
	if len(discv5Nodes) == 0 && len(discvNodes) <= 0 {
		log.Log.Warn("No bootstrap addresses supplied")
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
		log.Log.Error(fmt.Sprintf("Invalid address of %s provided: %v", addr, err.Error()))
	}
	return qnodeString, multiAddrString
}

func multiAddrFromString(address string) (multiaddr.Multiaddr, error) {
	addr, err := iaddr.ParseString(address)
	if err != nil {
		return nil, err
	}
	return addr.Multiaddr(), nil
}

func peersFromStringAddrs(addrs []string) ([]multiaddr.Multiaddr, error) {
	var allAddrs []multiaddr.Multiaddr
	qnodeString, multiAddrString := parseGenericAddrs(addrs)
	for _, stringAddr := range multiAddrString {
		addr, err := multiAddrFromString(stringAddr)
		if err != nil {
			return nil, fmt.Errorf("could not get multiaddr from string : %w", err)
		}
		allAddrs = append(allAddrs, addr)
	}
	for _, stringAddr := range qnodeString {
		qnodeAddr, err := qnode.Parse(qnode.ValidSchemes, stringAddr)
		if err != nil {
			return nil, fmt.Errorf("could not get qnode from string : %w", err)
		}
		addr, err := convertToSingleMultiAddr(qnodeAddr)
		if err != nil {
			return nil, fmt.Errorf("could not get multiaddr : %w", err)
		}
		allAddrs = append(allAddrs, addr)
	}
	return allAddrs, nil
}

func convertToSingleMultiAddr(node *qnode.Node) (multiaddr.Multiaddr, error) {
	ip4 := node.IP().To4()
	if ip4 == nil {
		return nil, fmt.Errorf("node doesn't have an ip4 address, it's stated IP is %s", node.IP().String())
	}
	pubkey := node.Pubkey()
	assertedKey := convertToInterfacePubkey(pubkey)
	id, err := peer.IDFromPublicKey(assertedKey)
	if err != nil {
		return nil, fmt.Errorf("could not get peer id : %w", err)
	}
	multiAddrString := fmt.Sprintf("/ip4/%s/tcp/%d/p2p/%s", ip4.String(), node.TCP(), id)
	multiAddr, err := multiaddr.NewMultiaddr(multiAddrString)
	if err != nil {
		return nil, fmt.Errorf("could not get multiaddr : %w", err)
	}
	return multiAddr, nil
}

func ConvertToInterfacePrivkey(privkey *ecdsa.PrivateKey) crypto.PrivKey {
	typeAssertedKey := crypto.PrivKey((*crypto.Secp256k1PrivateKey)((*secp256k1.PrivateKey)(privkey)))
	return typeAssertedKey
}

func convertToInterfacePubkey(pubkey *ecdsa.PublicKey) crypto.PubKey {
	typeAssertedKey := crypto.PubKey((*crypto.Secp256k1PublicKey)((*secp256k1.PublicKey)(pubkey)))
	return typeAssertedKey
}

func getConnPeerAddress(id, addr string) string {
	return fmt.Sprintf("%s/p2p/%s", addr, id)
}
