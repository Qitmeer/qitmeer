// Copyright (c) 2017-2018 The nox developers
// Copyright (c) 2013-2016 The btcsuite developers
// Copyright (c) 2015-2018 The Decred developers
// Use of this source code is governed by an ISC
// license that can be found in the LICENSE file.
package peerserver

import (
	"errors"
	"fmt"
	"github.com/noxproject/nox/common/hash"
	"github.com/noxproject/nox/config"
	"github.com/noxproject/nox/core/blockchain"
	"github.com/noxproject/nox/core/message"
	"github.com/noxproject/nox/core/protocol"
	"github.com/noxproject/nox/core/types"
	"github.com/noxproject/nox/log"
	"github.com/noxproject/nox/p2p/addmgr"
	"github.com/noxproject/nox/p2p/connmgr"
	"github.com/noxproject/nox/p2p/peer"
	"github.com/noxproject/nox/params"
	"github.com/noxproject/nox/services/blkmgr"
	"github.com/noxproject/nox/services/mempool"
	"github.com/noxproject/nox/version"
	"net"
	"strconv"
	"sync"
	"sync/atomic"
	"time"
)

const (
	// the default services supported by the node
	defaultServices = protocol.Full| protocol.CF

	// the default services that are required to be supported
	defaultRequiredServices = protocol.Full

	// defaultTargetOutbound is the default number of outbound peers to
	// target.
	defaultTargetOutbound = 8

	// connectionRetryInterval is the base amount of time to wait in between
	// retries when connecting to persistent peers.  It is adjusted by the
	// number of retries such that there is a retry backoff.
	connectionRetryInterval = time.Second * 5

	// maxProtocolVersion is the max protocol version the server supports.
	maxProtocolVersion = peer.MaxProtocolVersion
)

var (
	// userAgentName is the user agent name and is used to help identify
	// ourselves to other peers.
	userAgentName = "nox"

	// userAgentVersion is the user agent version and is used to help
	// identify ourselves to other peers.
	userAgentVersion = fmt.Sprintf("%d.%d.%d", version.Major, version.Minor,
		version.Patch)
)


// Use start to begin accepting connections from peers.
// peer server handling communications to and from nox peers.
type PeerServer struct {
	// These fields are variables must only be used atomically.
	bytesReceived uint64 // Total bytes received from all peers since start.
	bytesSent     uint64 // Total bytes sent by all peers since start.

	started       int32  // p2p server start flag
	shutdown      int32  // p2p server stop flag

    // address manager caching the peers
	addrManager *addmgr.AddrManager

	// conn manager handles network connections.
	connManager *connmgr.ConnManager
	nat         NAT

	newPeers  chan *serverPeer
	donePeers chan *serverPeer
	banPeers  chan *serverPeer

	// peer handler chan
	relayInv          chan relayMsg
	broadcast         chan broadcastMsg
	peerHeightsUpdate chan updatePeerHeightsMsg
	query             chan interface{}
	quit              chan struct{}

	wg sync.WaitGroup

	chainParams *params.Params
	cfg         *config.Config

	TimeSource   blockchain.MedianTimeSource
	BlockManager *blkmgr.BlockManager
	txMemPool    *mempool.TxPool

	services             protocol.ServiceFlag
}

// OutboundGroupCount returns the number of peers connected to the given
// outbound group key.
func (s *PeerServer) OutboundGroupCount(key string) int {
	replyChan := make(chan int)
	s.query <- getOutboundGroup{key: key, reply: replyChan}
	return <-replyChan
}

// inboundPeerConnected is invoked by the connection manager when a new inbound
// connection is established.  It initializes a new inbound server peer
// instance, associates it with the connection, and starts a goroutine to wait
// for disconnection.
func (s *PeerServer) inboundPeerConnected(conn net.Conn) {
	sp := newServerPeer(s,false)
	sp.isWhitelisted = isWhitelisted(s.cfg, conn.RemoteAddr())
	sp.Peer = peer.NewInboundPeer(newPeerConfig(sp))
	sp.AssociateConnection(conn)
	sp.syncPeer.Peer = sp.Peer
	go s.peerDoneHandler(sp)
	go sp.syncPeerHandler()
}

// outboundPeerConnected is invoked by the connection manager when a new
// outbound connection is established.  It initializes a new outbound server
// peer instance, associates it with the relevant state such as the connection
// request instance and the connection itself, and finally notifies the address
// manager of the attempt.
func (s *PeerServer) outboundPeerConnected(c *connmgr.ConnReq, conn net.Conn) {
	sp := newServerPeer(s, c.Permanent)
	p, err := peer.NewOutboundPeer(newPeerConfig(sp), c.Addr.String())
	if err != nil {
		log.Debug("Cannot create outbound peer %s: %v", c.Addr, err)
		s.connManager.Disconnect(c.ID())
	}
	sp.Peer = p
	sp.syncPeer.Peer = sp.Peer
	sp.connReq = c
	sp.isWhitelisted = isWhitelisted(s.cfg, conn.RemoteAddr())
	sp.AssociateConnection(conn)
	go s.peerDoneHandler(sp)
	go sp.syncPeerHandler()
	s.addrManager.Attempt(sp.NA())
}


// newPeerConfig returns the configuration for the given serverPeer.
func newPeerConfig(sp *serverPeer) *peer.Config {

	return &peer.Config{
		Listeners: peer.MessageListeners{
			OnVersion:        sp.OnVersion,
			OnGetAddr:        sp.OnGetAddr,
			OnAddr:           sp.OnAddr,
			OnRead:           sp.OnRead,
			OnWrite:          sp.OnWrite,
			OnGetBlocks:      sp.OnGetBlocks,
			OnBlock:          sp.OnBlock,
			OnGetData:        sp.OnGetData,
			OnInv:            sp.OnInv,
			OnGetMiningState: sp.OnGetMiningState,
			OnMiningState:    sp.OnMiningState,
			//OnMemPool:        sp.OnMemPool,
			//OnTx:             sp.OnTx,
			//OnHeaders:        sp.OnHeaders,
			//OnGetHeaders:     sp.OnGetHeaders,
			//OnGetCFilter:     sp.OnGetCFilter,
			//OnGetCFHeaders:   sp.OnGetCFHeaders,
			//OnGetCFTypes:     sp.OnGetCFTypes,
		},
		NewestBlock:       sp.newestBlock,
		HostToNetAddress:  sp.server.addrManager.HostToNetAddress,
		UserAgentName:     userAgentName,
		UserAgentVersion:  userAgentVersion,
		ChainParams:       sp.server.chainParams,
		Services:          sp.server.services,
		DisableRelayTx:    sp.server.cfg.BlocksOnly,
		ProtocolVersion:   maxProtocolVersion,
	}
}

// isWhitelisted returns whether the IP address is included in the whitelisted
// networks and IPs.
func isWhitelisted(cfg *config.Config, addr net.Addr) bool {
	if len(cfg.GetWhitelists()) == 0 {
		return false
	}

	host, _, err := net.SplitHostPort(addr.String())
	if err != nil {
		log.Warn("Unable to SplitHostPort on '%s': %v", addr, err)
		return false
	}
	ip := net.ParseIP(host)
	if ip == nil {
		log.Warn("Unable to parse IP '%s'", addr)
		return false
	}

	for _, ipnet := range cfg.GetWhitelists() {
		if ipnet.Contains(ip) {
			return true
		}
	}
	return false
}

// addrStringToNetAddr takes an address in the form of 'host:port' and returns
// a net.Addr which maps to the original address with any host names resolved
// to IP addresses.
func addrStringToNetAddr(addr string) (net.Addr, error) {
	host, strPort, err := net.SplitHostPort(addr)
	if err != nil {
		return nil, err
	}

	// Attempt to look up an IP address associated with the parsed host.
	ips, err := net.LookupIP(host)
	if err != nil {
		return nil, err
	}
	if len(ips) == 0 {
		return nil, fmt.Errorf("no addresses found for %s", host)
	}

	port, err := strconv.Atoi(strPort)
	if err != nil {
		return nil, err
	}

	return &net.TCPAddr{
		IP:   ips[0],
		Port: port,
	}, nil
}

func (p *PeerServer) Start() error {

	// Already started?
	if atomic.AddInt32(&p.started, 1) != 1 {
		return errors.New("p2p server already started")
	}

	log.Debug("Starting P2P server")

	// Start the peer handler which in turn starts the address and block
	// managers.
	p.wg.Add(1)
	go p.peerHandler()

	if p.nat != nil {
		p.wg.Add(1)
		go p.upnpUpdateThread()
	}
	return nil
}
func (p *PeerServer) Stop() error {
	// Make sure this only happens once.
	if atomic.AddInt32(&p.shutdown, 1) != 1 {
		log.Info("P2P Server is already in the process of shutting down")
		return nil
	}
	log.Warn("Stopping P2P Server")

	// Signal the remaining goroutines to quit.
	close(p.quit)
	log.Warn("wait for P2P stop ...")
	p.wg.Wait()
	log.Warn("P2P Server stopped")
	return nil
}

// newestBlock returns the current best block hash and height using the format
// required by the configuration for the peer package.
func (sp *serverPeer) newestBlock() (*hash.Hash, uint64, error) {
	best := sp.server.BlockManager.GetChain().BestSnapshot()
	return &best.Hash, best.Height, nil
}

// AddPeer adds a new peer that has already been connected to the server.
func (s *PeerServer) AddPeer(sp *serverPeer) {
	s.newPeers <- sp
}

// AddBytesReceived adds the passed number of bytes to the total bytes received
// counter for the server.  It is safe for concurrent access.
func (s *PeerServer) AddBytesReceived(bytesReceived uint64) {
	atomic.AddUint64(&s.bytesReceived, bytesReceived)
}

// AddBytesSent adds the passed number of bytes to the total bytes sent counter
// for the server.  It is safe for concurrent access.
func (s *PeerServer) AddBytesSent(bytesSent uint64) {
	atomic.AddUint64(&s.bytesSent, bytesSent)
}


// peerDoneHandler handles peer disconnects by notifiying the server that it's
// done.
func (s *PeerServer) peerDoneHandler(sp *serverPeer) {
	log.Trace("start peerDoneHandler")
	sp.WaitForDisconnect()
	s.donePeers <- sp

	// Only tell block manager we are gone if we ever told it we existed.
	if sp.VersionKnown() {
		log.Trace("peerDoneHandler send blkmgr donePeerMsg ")
		s.BlockManager.DonePeer(sp.syncPeer)
	}
	close(sp.quit)
	log.Trace("stop peerDoneHandler")
}
// peerHandler is used to handle peer operations such as adding and removing
// peers to and from the server, banning peers, and broadcasting messages to
// peers.  It must be run in a goroutine.
func (s *PeerServer) peerHandler() {

	s.addrManager.Start()

	log.Trace("Starting peer handler")

	state := &peerState{
		inboundPeers:    make(map[int32]*serverPeer),
		persistentPeers: make(map[int32]*serverPeer),
		outboundPeers:   make(map[int32]*serverPeer),
		banned:          make(map[string]time.Time),
		outboundGroups:  make(map[string]int),
	}

	if !s.cfg.DisableDNSSeed {
		// Add peers discovered through DNS to the address manager.
		connmgr.SeedFromDNS(s.chainParams, defaultRequiredServices, net.LookupIP, func(addrs []*types.NetAddress) {
			// Bitcoind uses a lookup of the dns seeder here. This
			// is rather strange since the values looked up by the
			// DNS seed lookups will vary quite a lot.
			// to replicate this behaviour we put all addresses as
			// having come from the first one.
			s.addrManager.AddAddresses(addrs, addrs[0])
		})
	}
	go s.connManager.Start()

out:
	for {
		select {
		// New peers connected to the server.
		case p := <-s.newPeers:
			s.handleAddPeerMsg(state, p)

		// Disconnected peers.
		case p := <-s.donePeers:
			s.handleDonePeerMsg(state, p)

		// Block accepted in mainchain or orphan, update peer height.
		case umsg := <-s.peerHeightsUpdate:
			s.handleUpdatePeerHeights(state, umsg)

		// Peer to ban.
		case p := <-s.banPeers:
			s.handleBanPeerMsg(state, p)

		// New inventory to potentially be relayed to other peers.
		case invMsg := <-s.relayInv:
			s.handleRelayInvMsg(state, invMsg)

		// Message to broadcast to all connected peers except those
		// which are excluded by the message.
		case bmsg := <-s.broadcast:
			s.handleBroadcastMsg(state, &bmsg)

		case qmsg := <-s.query:
			s.handleQuery(state, qmsg)

		case <-s.quit:
			// Disconnect all peers on server shutdown.
			state.forAllPeers(func(sp *serverPeer) {
				log.Trace("Shutdown peer", "peer",sp)
				sp.Disconnect()
			})
			break out
		}
	}

	s.connManager.Stop()
	s.addrManager.Stop()

	// Drain channels before exiting so nothing is left waiting around
	// to send.
cleanup:
	for {
		select {
		case <-s.newPeers:
		case <-s.donePeers:
		case <-s.peerHeightsUpdate:
		case <-s.relayInv:
		case <-s.broadcast:
		case <-s.query:
		default:
			break cleanup
		}
	}
	s.wg.Done()
	log.Trace("Peer handler done")
}


// RelayInventory relays the passed inventory vector to all connected peers
// that are not already known to have it.
func (s *PeerServer) RelayInventory(invVect *message.InvVect, data interface{}) {
	s.relayInv <- relayMsg{invVect: invVect, data: data}
}

// UpdatePeerHeights updates the heights of all peers who have have announced
// the latest connected main chain block, or a recognized orphan. These height
// updates allow us to dynamically refresh peer heights, ensuring sync peer
// selection has access to the latest block heights for each peer.
func (s *PeerServer) UpdatePeerHeights(latestBlkHash *hash.Hash, latestHeight uint64, updateSource *serverPeer) {
	s.peerHeightsUpdate <- updatePeerHeightsMsg{
		newHash:    latestBlkHash,
		newHeight:  latestHeight,
		originPeer: updateSource,
	}
}

