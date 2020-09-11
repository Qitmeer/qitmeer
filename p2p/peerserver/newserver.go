package peerserver

import (
	"errors"
	"fmt"
	"github.com/Qitmeer/qitmeer/common/network"
	"github.com/Qitmeer/qitmeer/common/roughtime"
	"github.com/Qitmeer/qitmeer/config"
	"github.com/Qitmeer/qitmeer/core/protocol"
	"github.com/Qitmeer/qitmeer/core/types"
	"github.com/Qitmeer/qitmeer/log"
	"github.com/Qitmeer/qitmeer/p2p/addmgr"
	"github.com/Qitmeer/qitmeer/p2p/connmgr"
	"github.com/Qitmeer/qitmeer/params"
	"net"
	"strconv"
	"time"
)

func NewPeerServer(cfg *config.Config, chainParams *params.Params) (*PeerServer, error) {

	services := defaultServices

	s := PeerServer{
		services:    services,
		cfg:         cfg,
		chainParams: chainParams,
		newPeers:    make(chan *serverPeer, cfg.MaxPeers),
		donePeers:   make(chan *serverPeer, cfg.MaxPeers),
		banPeers:    make(chan *BanPeerMsg, cfg.MaxPeers),
		query:       make(chan interface{}),
		relayInv:    make(chan relayMsg, cfg.MaxPeers),
		broadcast:   make(chan broadcastMsg, cfg.MaxPeers),
		quit:        make(chan struct{}),
	}
	if cfg.BanDuration > 0 {
		connmgr.BanDuration = cfg.BanDuration
	}

	if cfg.BanThreshold > 0 {
		connmgr.BanThreshold = cfg.BanThreshold
	}
	amgr := addmgr.New(cfg.DataDir, cfg.GetAddrPercent, net.LookupIP)
	var listeners []net.Listener
	var nat NAT
	if !cfg.DisableListen {
		netAddrs, err := network.ParseListeners(cfg.Listeners)
		if err != nil {
			return nil, err
		}

		listeners = make([]net.Listener, 0, len(netAddrs))
		for _, addr := range netAddrs {
			listener, err := net.Listen(addr.Network(), addr.String())
			if err != nil {
				log.Warn("Can't listen on", "addr", addr, "error", err)
				continue
			}
			listeners = append(listeners, listener)
		}

		if len(listeners) == 0 {
			return nil, errors.New("no valid listen address")
		}

		if len(cfg.ExternalIPs) != 0 {
			// if this fails we have real issues.
			port, _ := strconv.ParseUint(
				chainParams.DefaultPort, 10, 16)

			for _, sip := range cfg.ExternalIPs {
				eport := uint16(port)
				host, portstr, err := net.SplitHostPort(sip)
				if err != nil {
					// no port, use default.
					host = sip
				} else {
					port, err := strconv.ParseUint(
						portstr, 10, 16)
					if err != nil {
						log.Warn("Can not parse port for externalip", "ip", sip, "error", err)
						continue
					}
					eport = uint16(port)
				}
				na, err := amgr.HostToNetAddress(host, eport, services)
				if err != nil {
					log.Warn("Not adding as externalip", "ip", sip, "error", err)
					continue
				}

				err = amgr.AddLocalAddress(na, addmgr.ManualPrio)
				if err != nil {
					log.Warn("Skipping specified external IP", "error", err)
				}
			}
		} else {
			if cfg.Upnp {
				nat, err = Discover()
				if err != nil {
					log.Warn("Can't discover upnp", "error", err)
				}
				// nil nat here is fine, just means no upnp on network.
			}
			// Add bound addresses to address manager to be advertised to peers.
			for _, listener := range listeners {
				addr := listener.Addr().String()
				err := addLocalAddress(amgr, addr, services)
				if err != nil {
					log.Warn(fmt.Sprintf("Skipping bound address %s: %v", addr, err))
				}
			}
		}
	}

	// Only setup a function to return new addresses to connect to when
	// not running in connect-only mode.  The simulation network is always
	// in connect-only mode since it is only intended to connect to
	// specified peers and actively avoid advertising and connecting to
	// discovered peers in order to prevent it from becoming a public test
	// network.
	var newAddressFunc func() (net.Addr, error)
	if !cfg.PrivNet && len(cfg.ConnectPeers) == 0 {
		newAddressFunc = func() (net.Addr, error) {
			addr := s.addrManager.GetAddress()
			if addr == nil {
				//break
				return nil, errors.New("no valid connect address")
			}
			// Address will not be invalid, local or unroutable
			// because addrmanager rejects those on addition.
			// Just check that we don't already have an address
			// in the same group so that we are not connecting
			// to the same network segment at the expense of
			// others.
			key := addmgr.GroupKey(addr.NetAddress())
			if s.OutboundGroupCount(key) != 0 {
				return nil, errors.New("no valid connect address")
			}

			// only allow recent nodes (10mins) after we failed 30
			// times
			if addr.GetAttempts() > 1 && roughtime.Since(addr.LastAttempt()) < 10*time.Minute {
				return nil, errors.New("no valid connect address")
			}
			if s.state.IsBanPeer(addr.NetAddress().IP.String()) {
				return nil, errors.New("no valid connect address")
			}
			addrString := addmgr.NetAddressKey(addr.NetAddress())
			return addrStringToNetAddr(addrString)
		}
	}
	// Create a connection manager.
	targetOutbound := defaultTargetOutbound
	if cfg.MaxPeers < targetOutbound {
		targetOutbound = cfg.MaxPeers
	}
	cmgr, err := connmgr.New(&connmgr.Config{
		Listeners:      listeners,
		OnAccept:       s.inboundPeerConnected,
		RetryDuration:  connectionRetryInterval,
		TargetOutbound: uint32(targetOutbound),
		Dial:           s.Dial,
		OnConnection:   s.outboundPeerConnected,
		GetNewAddress:  newAddressFunc,
	})
	if err != nil {
		return nil, err
	}

	s.addrManager = amgr
	s.connManager = cmgr
	s.nat = nat

	// Start up persistent peers.
	permanentPeers := cfg.ConnectPeers
	if len(permanentPeers) == 0 {
		permanentPeers = cfg.AddPeers
	}
	for _, addr := range permanentPeers {
		tcpAddr, err := addrStringToNetAddr(addr)
		if err != nil {
			return nil, err
		}

		go s.connManager.Connect(&connmgr.ConnReq{
			Addr:      tcpAddr,
			Permanent: true,
		})
	}

	return &s, nil
}

func addLocalAddress(addrMgr *addmgr.AddrManager, addr string, services protocol.ServiceFlag) error {
	host, portStr, err := net.SplitHostPort(addr)
	if err != nil {
		return err
	}
	port, err := strconv.ParseUint(portStr, 10, 16)
	if err != nil {
		return err
	}

	if ip := net.ParseIP(host); ip != nil && ip.IsUnspecified() {
		// If bound to unspecified address, advertise all local interfaces
		addrs, err := net.InterfaceAddrs()
		if err != nil {
			return err
		}

		for _, addr := range addrs {
			ifaceIP, _, err := net.ParseCIDR(addr.String())
			if err != nil {
				continue
			}

			// If bound to 0.0.0.0, do not add IPv6 interfaces and if bound to
			// ::, do not add IPv4 interfaces.
			if (ip.To4() == nil) != (ifaceIP.To4() == nil) {
				continue
			}

			netAddr := types.NewNetAddressIPPort(ifaceIP, uint16(port), services)
			addrMgr.AddLocalAddress(netAddr, addmgr.BoundPrio)
		}
	} else {
		netAddr, err := addrMgr.HostToNetAddress(host, uint16(port), services)
		if err != nil {
			return err
		}

		addrMgr.AddLocalAddress(netAddr, addmgr.BoundPrio)
	}

	return nil
}
