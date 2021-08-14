/*
 * Copyright (c) 2017-2020 The qitmeer developers
 */

package main

import (
	"github.com/Qitmeer/qitmeer/core/blockdag"
	"github.com/Qitmeer/qitmeer/core/json"
	"github.com/Qitmeer/qitmeer/core/protocol"
	"github.com/Qitmeer/qitmeer/node"
	"github.com/Qitmeer/qitmeer/params"
	"github.com/Qitmeer/qitmeer/rpc"
	"github.com/Qitmeer/qitmeer/rpc/client/cmds"
	"github.com/Qitmeer/qitmeer/version"
	"time"
)

func (node *Node) api() rpc.API {
	return rpc.API{
		NameSpace: cmds.DefaultServiceNameSpace,
		Service:   NewPublicRelayAPI(node),
		Public:    true,
	}
}

type PublicRelayAPI struct {
	node *Node
}

func NewPublicRelayAPI(node *Node) *PublicRelayAPI {
	return &PublicRelayAPI{node}
}

// Return the RPC info
func (api *PublicRelayAPI) GetRpcInfo() (interface{}, error) {
	rs := api.node.rpcServer.ReqStatus
	jrs := []*cmds.JsonRequestStatus{}
	for _, v := range rs {
		jrs = append(jrs, v.ToJson())
	}
	return jrs, nil
}

// Return the peer info
func (api *PublicRelayAPI) GetPeerInfo(verbose *bool, network *string) (interface{}, error) {
	vb := false
	if verbose != nil {
		vb = *verbose
	}
	networkName := ""
	if network != nil {
		networkName = *network
	}
	if len(networkName) <= 0 {
		networkName = params.ActiveNetParams.Name
	}
	peers := api.node.peerStatus.StatsSnapshots()
	infos := make([]*json.GetPeerInfoResult, 0, len(peers))
	for _, p := range peers {

		if len(networkName) != 0 && networkName != "all" {
			if p.Network != networkName {
				continue
			}
		}

		if !vb {
			if !p.State.IsConnected() {
				continue
			}
		}
		info := &json.GetPeerInfoResult{
			ID:        p.PeerID,
			Name:      p.Name,
			Address:   p.Address,
			BytesSent: p.BytesSent,
			BytesRecv: p.BytesRecv,
		}
		info.Protocol = p.Protocol
		info.Services = p.Services.String()
		if p.Genesis != nil {
			info.Genesis = p.Genesis.String()
		}
		if p.IsTheSameNetwork() {
			info.State = p.State.String()
		}
		if len(p.Version) > 0 {
			info.Version = p.Version
		}
		if len(p.Network) > 0 {
			info.Network = p.Network
		}

		if p.State.IsConnected() {
			info.TimeOffset = p.TimeOffset
			if p.Genesis != nil {
				info.Genesis = p.Genesis.String()
			}
			info.Direction = p.Direction.String()
			if p.GraphState != nil {
				info.GraphState = node.GetGraphStateResult(p.GraphState)
			}
			info.ConnTime = p.ConnTime.Truncate(time.Second).String()

			info.GSUpdate = p.GraphStateDur.Truncate(time.Second).String()
		}
		if !p.LastSend.IsZero() {
			info.LastSend = p.LastSend.String()
		}
		if !p.LastRecv.IsZero() {
			info.LastRecv = p.LastRecv.String()
		}
		if len(p.QNR) > 0 {
			info.QNR = p.QNR
		}
		infos = append(infos, info)
	}
	return infos, nil
}

// Return the node info
func (api *PublicRelayAPI) GetNodeInfo() (interface{}, error) {
	ret := &json.InfoNodeResult{
		ID:              api.node.Host().ID().String(),
		Version:         int32(1000000*version.Major + 10000*version.Minor + 100*version.Patch),
		BuildVersion:    version.String(),
		ProtocolVersion: int32(protocol.ProtocolVersion),
		Connections:     int32(len(api.node.peerStatus.Connected())),
		Network:         params.ActiveNetParams.Name,
		Confirmations:   blockdag.StableConfirmations,
		Modules:         []string{cmds.DefaultServiceNameSpace},
	}
	hostdns := api.node.HostDNS()
	if hostdns != nil {
		ret.DNS = hostdns.String()
	}

	hostaddrs := api.node.HostAddress()
	if len(hostaddrs) > 0 {
		ret.Addresss = hostaddrs
	}

	return ret, nil
}

func (api *PublicRelayAPI) GetNetworkInfo() (interface{}, error) {
	peers := api.node.peerStatus.StatsSnapshots()
	nstat := &json.NetworkStat{Infos: []*json.NetworkInfo{}}
	infos := map[string]*json.NetworkInfo{}

	for _, p := range peers {
		nstat.TotalPeers++

		if p.Services&protocol.Relay > 0 {
			nstat.TotalRelays++
		}
		//
		if len(p.Network) <= 0 {
			continue
		}

		info, ok := infos[p.Network]
		if !ok {
			info = &json.NetworkInfo{Name: p.Network}
			infos[p.Network] = info
			nstat.Infos = append(nstat.Infos, info)
		}
		info.Peers++
		if p.State.IsConnected() {
			info.Connecteds++
			nstat.TotalConnected++
		}
		if p.Services&protocol.Relay > 0 {
			info.Relays++
		}
	}
	return nstat, nil
}
