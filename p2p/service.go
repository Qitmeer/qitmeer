/*
 * Copyright (c) 2020.
 * Project:qitmeer
 * File:Service.go
 * Date:7/2/20 8:04 PM
 * Author:Jin
 * Email:lochjin@gmail.com
 */
package p2p

import (
	"context"
	"crypto/ecdsa"
	"encoding/base64"
	"fmt"
	"github.com/Qitmeer/qitmeer/common/hash"
	"github.com/Qitmeer/qitmeer/config"
	"github.com/Qitmeer/qitmeer/core/blockchain"
	"github.com/Qitmeer/qitmeer/core/event"
	"github.com/Qitmeer/qitmeer/core/message"
	pv "github.com/Qitmeer/qitmeer/core/protocol"
	"github.com/Qitmeer/qitmeer/node/notify"
	"github.com/Qitmeer/qitmeer/p2p/common"
	"github.com/Qitmeer/qitmeer/p2p/encoder"
	"github.com/Qitmeer/qitmeer/p2p/peers"
	pb "github.com/Qitmeer/qitmeer/p2p/proto/v1"
	"github.com/Qitmeer/qitmeer/p2p/qnode"
	"github.com/Qitmeer/qitmeer/p2p/qnr"
	"github.com/Qitmeer/qitmeer/p2p/runutil"
	"github.com/Qitmeer/qitmeer/p2p/synch"
	"github.com/Qitmeer/qitmeer/services/mempool"
	"github.com/Qitmeer/qitmeer/version"
	"github.com/dgraph-io/ristretto"
	"github.com/gogo/protobuf/proto"
	"github.com/libp2p/go-libp2p"
	"github.com/libp2p/go-libp2p-core/host"
	"github.com/libp2p/go-libp2p-core/peer"
	"github.com/libp2p/go-libp2p-pubsub"
	"github.com/libp2p/go-libp2p-pubsub/pb"
	"github.com/multiformats/go-multiaddr"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

const (
	// the default services supported by the node
	defaultServices = pv.Full | pv.CF
)

var (
	// In the event that we are at our peer limit, we
	// stop looking for new peers and instead poll
	// for the current peer limit status for the time period
	// defined below.
	pollingPeriod = 6 * time.Second

	// Refresh rate of QNR
	refreshRate = time.Hour
)

type Service struct {
	cfg           *common.Config
	ctx           context.Context
	cancel        context.CancelFunc
	exclusionList *ristretto.Cache
	started       bool
	isPreGenesis  bool
	privKey       *ecdsa.PrivateKey
	metaData      *pb.MetaData
	addrFilter    *multiaddr.Filters
	host          host.Host
	pubsub        *pubsub.PubSub

	dv5Listener Listener
	events      *event.Feed
	sy          *synch.Sync

	Chain      *blockchain.BlockChain
	TimeSource blockchain.MedianTimeSource
	TxMemPool  *mempool.TxPool
	Notify     notify.Notify
}

func (s *Service) Start() error {
	if s.started {
		return fmt.Errorf("Attempted to start p2p service when it was already started")
	}
	log.Info("P2P Service Start")

	err := s.sy.Start()
	if err != nil {
		return err
	}

	s.isPreGenesis = false
	s.started = true

	var peersToWatch []string
	if s.cfg.RelayNodeAddr != "" {
		peersToWatch = append(peersToWatch, s.cfg.RelayNodeAddr)
		if err := dialRelayNode(s.ctx, s.host, s.cfg.RelayNodeAddr); err != nil {
			log.Warn(fmt.Sprintf("Could not dial relay node:%v", err))
		}
	}

	if !s.cfg.NoDiscovery {
		ipAddr := ipAddr()
		listener, err := s.startDiscoveryV5(
			ipAddr,
			s.privKey,
		)
		if err != nil {
			log.Error(fmt.Sprintf("Failed to start discovery:%v", err))
			return err
		}
		err = s.connectToBootnodes()
		if err != nil {
			log.Error(fmt.Sprintf("Could not add bootnode to the exclusion list:%v", err))
			return err
		}
		s.dv5Listener = listener
		go s.listenForNewNodes()
	}

	s.started = true

	if len(s.cfg.StaticPeers) > 0 {
		addrs, err := peersFromStringAddrs(s.cfg.StaticPeers)
		if err != nil {
			log.Error(fmt.Sprintf("Could not connect to static peer: %v", err))
		} else {
			s.connectWithAllPeers(addrs)
		}
	}

	// Periodic functions.
	runutil.RunEvery(s.ctx, 5*time.Second, func() {
		ensurePeerConnections(s.ctx, s.host, peersToWatch...)
	})
	runutil.RunEvery(s.ctx, time.Hour, s.Peers().Decay)
	runutil.RunEvery(s.ctx, refreshRate, func() {
		s.RefreshQNR()
	})

	multiAddrs := s.host.Network().ListenAddresses()
	logIPAddr(s.host.ID(), multiAddrs...)

	p2pHostAddress := s.cfg.HostAddress
	p2pTCPPort := s.cfg.TCPPort

	if p2pHostAddress != "" {
		logExternalIPAddr(s.host.ID(), p2pHostAddress, p2pTCPPort)
		verifyConnectivity(p2pHostAddress, p2pTCPPort, "tcp")
	}

	p2pHostDNS := s.cfg.HostDNS
	if p2pHostDNS != "" {
		logExternalDNSAddr(s.host.ID(), p2pHostDNS, p2pTCPPort)
	}

	return nil
}

// Started returns true if the p2p service has successfully started.
func (s *Service) Started() bool {
	return s.started
}

func (s *Service) Stop() error {
	log.Info("P2P Service Stop")

	defer s.cancel()
	s.started = false
	if s.dv5Listener != nil {
		s.dv5Listener.Close()
	}
	return s.sy.Stop()
}

func (s *Service) connectToBootnodes() error {
	nodes := make([]*qnode.Node, 0, len(s.cfg.Discv5BootStrapAddr))
	for _, addr := range s.cfg.Discv5BootStrapAddr {
		bootNode, err := qnode.Parse(qnode.ValidSchemes, addr)
		if err != nil {
			return err
		}
		// do not dial bootnodes with their tcp ports not set
		if err := bootNode.Record().Load(qnr.WithEntry("tcp", new(qnr.TCP))); err != nil {
			if !qnr.IsNotFound(err) {
				log.Error("Could not retrieve tcp port:%v", err)
			}
			continue
		}
		nodes = append(nodes, bootNode)
	}
	multiAddresses := convertToMultiAddr(nodes)
	s.connectWithAllPeers(multiAddresses)
	return nil
}

func (s *Service) connectWithAllPeers(multiAddrs []multiaddr.Multiaddr) {
	addrInfos, err := peer.AddrInfosFromP2pAddrs(multiAddrs...)
	if err != nil {
		log.Error("Could not convert to peer address info's from multiaddresses: %v", err)
		return
	}
	for _, info := range addrInfos {
		// make each dial non-blocking
		go func(info peer.AddrInfo) {
			if err := s.connectWithPeer(info); err != nil {
				log.Trace(fmt.Sprintf("Could not connect with peer %s :%v", info.String(), err))
			}
		}(info)
	}
}

func (s *Service) connectWithPeer(info peer.AddrInfo) error {
	if info.ID == s.host.ID() {
		return nil
	}
	pe := s.Peers().Get(info.ID)
	if pe == nil {
		return nil
	}
	if pe.IsBad() {
		return nil
	}
	if err := s.host.Connect(s.ctx, info); err != nil {
		s.Peers().IncrementBadResponses(info.ID)
		return err
	}
	return nil
}

// Peers returns the peer status interface.
func (s *Service) Peers() *peers.Status {
	return s.sy.Peers()
}

// listen for new nodes watches for new nodes in the network and adds them to the peerstore.
func (s *Service) listenForNewNodes() {
	iterator := s.dv5Listener.RandomNodes()
	iterator = qnode.Filter(iterator, s.filterPeer)
	defer iterator.Close()
	for {
		// Exit if service's context is canceled
		if s.ctx.Err() != nil {
			break
		}
		if s.isPeerAtLimit() {
			// Pause the main loop for a period to stop looking
			// for new peers.
			log.Trace("Not looking for peers, at peer limit")
			time.Sleep(pollingPeriod)
			continue
		}
		exists := iterator.Next()
		if !exists {
			break
		}
		node := iterator.Node()
		peerInfo, _, err := convertToAddrInfo(node)
		if err != nil {
			log.Error(fmt.Sprintf("Could not convert to peer info:%v", err))
			continue
		}
		go func(info *peer.AddrInfo) {
			if err := s.connectWithPeer(*info); err != nil {
				log.Trace(fmt.Sprintf("Could not connect with peer %s  :%v", info.String(), err))
			}
		}(peerInfo)
	}
}

func (s *Service) RefreshQNR() {
	// return early if discv5 isnt running
	if s.dv5Listener == nil {
		return
	}
	// ping all peers to inform them of new metadata
	//s.pingPeers()
}

func (s *Service) pingPeers() {
	for _, pid := range s.Peers().Connected() {
		go func(id peer.ID) {
			if err := s.sy.SendPingRequest(s.ctx, id); err != nil {
				log.Error("Failed to ping peer:id=%s  %v", id, err)
			}
		}(pid)
	}
}

// PubSub returns the p2p pubsub framework.
func (s *Service) PubSub() *pubsub.PubSub {
	return s.pubsub
}

// Host returns the currently running libp2p
// host of the service.
func (s *Service) Host() host.Host {
	return s.host
}

// PeerID returns the Peer ID of the local peer.
func (s *Service) PeerID() peer.ID {
	return s.host.ID()
}

// Disconnect from a peer.
func (s *Service) Disconnect(pid peer.ID) error {
	return s.host.Network().ClosePeer(pid)
}

// Connect to a specific peer.
func (s *Service) Connect(pi peer.AddrInfo) error {
	return s.host.Connect(s.ctx, pi)
}

// ENR returns the local node's current ENR.
func (s *Service) QNR() *qnr.Record {
	if s.dv5Listener == nil {
		return nil
	}
	return s.dv5Listener.Self().Record()
}

// Metadata returns a copy of the peer's metadata.
func (s *Service) Metadata() *pb.MetaData {
	return proto.Clone(s.metaData).(*pb.MetaData)
}

// MetadataSeq returns the metadata sequence number.
func (s *Service) MetadataSeq() uint64 {
	return s.metaData.SeqNumber
}

// Encoding returns the configured networking encoding.
func (s *Service) Encoding() encoder.NetworkEncoding {
	encoding := s.cfg.Encoding
	switch encoding {
	case encoder.SSZ:
		return &encoder.SszNetworkEncoder{}
	case encoder.SSZSnappy:
		return &encoder.SszNetworkEncoder{UseSnappyCompression: true}
	default:
		panic("Invalid Network Encoding Flag Provided")
	}
}

func (s *Service) GetGenesisHash() *hash.Hash {
	return s.Chain.BlockDAG().GetGenesisHash()
}

func (s *Service) BlockChain() *blockchain.BlockChain {
	return s.Chain
}

func (s *Service) Context() context.Context {
	return s.ctx
}

func NewService(cfg *config.Config, events *event.Feed) (*Service, error) {
	var err error
	ctx, cancel := context.WithCancel(context.Background())
	cache, err := ristretto.NewCache(&ristretto.Config{
		NumCounters: 1000,
		MaxCost:     1000,
		BufferItems: 64,
	})
	if err != nil {
		return nil, err
	}

	bootnodesTemp := cfg.BootstrapNodes
	bootnodeAddrs := make([]string, 0) //dest of final list of nodes
	for _, addr := range bootnodesTemp {
		if filepath.Ext(addr) == ".yaml" {
			fileNodes, err := readbootNodes(addr)
			if err != nil {
				return nil, err
			}
			bootnodeAddrs = append(bootnodeAddrs, fileNodes...)
		} else {
			bootnodeAddrs = append(bootnodeAddrs, addr)
		}
	}

	s := &Service{
		cfg: &common.Config{
			NoDiscovery:          cfg.NoDiscovery,
			EnableUPnP:           cfg.Upnp,
			StaticPeers:          cfg.AddPeers,
			BootstrapNodeAddr:    bootnodeAddrs,
			DataDir:              cfg.DataDir,
			MaxPeers:             uint(cfg.MaxPeers),
			ReadWritePermissions: 0600, //-rw------- Read and Write permissions for user
			MetaDataDir:          cfg.MetaDataDir,
			TCPPort:              uint(cfg.P2PTCPPort),
			UDPPort:              uint(cfg.P2PUDPPort),
			Encoding:             "ssz-snappy",
			ProtocolVersion:      pv.ProtocolVersion,
			Services:             defaultServices,
			UserAgent:            fmt.Sprintf("qitmeer(%s)", version.String()),
			DisableRelayTx:       cfg.BlocksOnly,
			MaxOrphanTxs:         cfg.MaxOrphanTxs,
		},
		ctx:           ctx,
		cancel:        cancel,
		exclusionList: cache,
		isPreGenesis:  true,
		events:        events,
	}

	dv5Nodes := parseBootStrapAddrs(s.cfg.BootstrapNodeAddr)
	s.cfg.Discv5BootStrapAddr = dv5Nodes

	ipAddr := ipAddr()
	s.privKey, err = privKey(s.cfg)
	if err != nil {
		log.Error(fmt.Sprintf("Failed to generate p2p private key:%v", err))
		return nil, err
	}
	s.metaData, err = metaDataFromConfig(s.cfg)
	if err != nil {
		log.Error(fmt.Sprintf("Failed to create peer metadata:%v", err))
		return nil, err
	}
	s.addrFilter, err = configureFilter(s.cfg)
	if err != nil {
		log.Error(fmt.Sprintf("Failed to create address filter:%v", err))
		return nil, err
	}
	opts := s.buildOptions(ipAddr, s.privKey)
	h, err := libp2p.New(s.ctx, opts...)
	if err != nil {
		log.Error("Failed to create p2p host")
		return nil, err
	}

	s.host = h

	psOpts := []pubsub.Option{
		pubsub.WithMessageSigning(false),
		pubsub.WithStrictSignatureVerification(false),
		pubsub.WithMessageIdFn(msgIDFunction),
	}

	gs, err := pubsub.NewGossipSub(s.ctx, s.host, psOpts...)
	if err != nil {
		log.Error(fmt.Sprintf("Failed to start pubsub:%v", err))
		return nil, err
	}
	s.pubsub = gs

	s.sy = synch.NewSync(s)
	return s, nil
}

func readbootNodes(fileName string) ([]string, error) {
	fileContent, err := ioutil.ReadFile(fileName)
	if err != nil {
		return nil, err
	}
	listNodes := make([]string, 0)
	err = yaml.Unmarshal(fileContent, &listNodes)
	if err != nil {
		return nil, err
	}
	return listNodes, nil
}

func msgIDFunction(pmsg *pubsub_pb.Message) string {
	h := FastSum256(pmsg.Data)
	return base64.URLEncoding.EncodeToString(h[:])
}

func logIPAddr(id peer.ID, addrs ...multiaddr.Multiaddr) {
	var correctAddr multiaddr.Multiaddr
	for _, addr := range addrs {
		if strings.Contains(addr.String(), "/ip4/") || strings.Contains(addr.String(), "/ip6/") {
			correctAddr = addr
			break
		}
	}
	if correctAddr != nil {
		log.Info(fmt.Sprintf("Node started p2p server:multiAddr=%s", correctAddr.String()+"/p2p/"+id.String()))
	}
}

func logExternalIPAddr(id peer.ID, addr string, port uint) {
	if addr != "" {
		multiAddr, err := multiAddressBuilder(addr, port)
		if err != nil {
			log.Error(fmt.Sprintf("Could not create multiaddress: %v", err))
			return
		}
		log.Info(fmt.Sprintf("Node started external p2p server:multiAddr=%s", multiAddr.String()+"/p2p/"+id.String()))
	}
}

func logExternalDNSAddr(id peer.ID, addr string, port uint) {
	if addr != "" {
		p := strconv.FormatUint(uint64(port), 10)
		log.Info(fmt.Sprintf("Node started external p2p server:multiAddr=%s", "/dns4/"+addr+"/tcp/"+p+"/p2p/"+id.String()))
	}
}

// TODO
func (s *Service) GetBanlist() map[string]time.Time {
	return nil
}

func (s *Service) RemoveBan(host string) {

}

func (s *Service) RelayInventory(invVect *message.InvVect, data interface{}) {

}

func (s *Service) BroadcastMessage(msg message.Message) {

}
