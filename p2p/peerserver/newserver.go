package peerserver

import (
	"errors"
	"fmt"
	"github.com/HalalChain/qitmeer/common/network"
	"github.com/HalalChain/qitmeer/config"
	"github.com/HalalChain/qitmeer/core/types"
	"github.com/HalalChain/qitmeer/log"
	"github.com/HalalChain/qitmeer/p2p/addmgr"
	"github.com/HalalChain/qitmeer/p2p/connmgr"
	"github.com/HalalChain/qitmeer/params"
	"net"
	"strconv"
	"time"
)

func NewPeerServer(cfg *config.Config,chainParams *params.Params) (*PeerServer, error){

	services := defaultServices

	s := PeerServer{
		services: services,
		cfg: cfg,
		chainParams:          chainParams,
		newPeers:             make(chan *serverPeer, cfg.MaxPeers),
		donePeers:            make(chan *serverPeer, cfg.MaxPeers),
		banPeers:             make(chan *serverPeer, cfg.MaxPeers),
		query:                make(chan interface{}),
		relayInv:             make(chan relayMsg, cfg.MaxPeers),
		broadcast:            make(chan broadcastMsg, cfg.MaxPeers),
		peerHeightsUpdate:    make(chan updatePeerHeightsMsg),
		quit:                 make(chan struct{}),
	}

	amgr := addmgr.New(cfg.DataDir, net.LookupIP)
	var listeners []net.Listener
	var nat NAT
	if !cfg.DisableListen {
		ipv4Addrs, ipv6Addrs, wildcard, err :=
			network.ParseListeners(cfg.Listeners)
		if err != nil {
			return nil, err
		}
		listeners = make([]net.Listener, 0, len(ipv4Addrs)+len(ipv6Addrs))
		discover := true
		if len(cfg.ExternalIPs) != 0 {
			discover = false
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
						log.Warn("Can not parse port for externalip","ip",sip,"error",err)
						continue
					}
					eport = uint16(port)
				}
				na, err := amgr.HostToNetAddress(host, eport, services)
				if err != nil {
					log.Warn("Not adding as externalip","ip", sip, "error",err)
					continue
				}

				err = amgr.AddLocalAddress(na, addmgr.ManualPrio)
				if err != nil {
					log.Warn("Skipping specified external IP", "error", err)
				}
			}
		} else if discover && cfg.Upnp {
			nat, err = Discover()
			if err != nil {
				log.Warn("Can't discover upnp", "error", err)
			}
			// nil nat here is fine, just means no upnp on network.
		}

		// TODO: nonstandard port...
		if wildcard {
			port, err :=
				strconv.ParseUint(chainParams.DefaultPort,
					10, 16)
			if err != nil {
				// I can't think of a cleaner way to do this...
				goto nowc
			}
			addrs, err := net.InterfaceAddrs()
			if err != nil {
				log.Warn("Unable to get interface addresses", "error", err)
			}
			for _, a := range addrs {
				ip, _, err := net.ParseCIDR(a.String())
				if err != nil {
					continue
				}
				na := types.NewNetAddressIPPort(ip,
					uint16(port), services)
				if discover {
					err = amgr.AddLocalAddress(na, addmgr.InterfacePrio)
					if err != nil {
						log.Debug("Skipping local address", "error",err)
					}
				}
			}
		}
	nowc:

		for _, addr := range ipv4Addrs {
			listener, err := net.Listen("tcp4", addr)
			if err != nil {
				log.Warn("Can't listen on","addr",addr, "error",err)
				continue
			}
			listeners = append(listeners, listener)

			if discover {
				if na, err := amgr.DeserializeNetAddress(addr); err == nil {
					err = amgr.AddLocalAddress(na, addmgr.BoundPrio)
					if err != nil {
						log.Warn("Skipping bound address", "addr",addr, "error",err)
					}
				}
			}
		}

		for _, addr := range ipv6Addrs {
			listener, err := net.Listen("tcp6", addr)
			if err != nil {
				log.Warn("Can't listen on", "addr",addr, "error",err)
				continue
			}
			listeners = append(listeners, listener)
			if discover {
				if na, err := amgr.DeserializeNetAddress(addr); err == nil {
					err = amgr.AddLocalAddress(na, addmgr.BoundPrio)
					if err != nil {
						log.Debug("Skipping bound address", "error",err)
					}
				}
			}
		}

		if len(listeners) == 0 {
			return nil, errors.New("no valid listen address")
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
			for tries := 0; tries < 100; tries++ {
				addr := s.addrManager.GetAddress()
				if addr == nil {
					break
				}

				// Address will not be invalid, local or unroutable
				// because addrmanager rejects those on addition.
				// Just check that we don't already have an address
				// in the same group so that we are not connecting
				// to the same network segment at the expense of
				// others.
				key := addmgr.GroupKey(addr.NetAddress())
				if s.OutboundGroupCount(key) != 0 {
					continue
				}

				// only allow recent nodes (10mins) after we failed 30
				// times
				if tries < 30 && time.Since(addr.LastAttempt()) < 10*time.Minute {
					continue
				}

				// allow nondefault ports after 50 failed tries.
				if fmt.Sprintf("%d", addr.NetAddress().Port) !=
					chainParams.DefaultPort && tries < 50 {
					continue
				}

				addrString := addmgr.NetAddressKey(addr.NetAddress())
				return addrStringToNetAddr(addrString)
			}

			return nil, errors.New("no valid connect address")
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
