/*
 * Copyright (c) 2017-2020 The qitmeer developers
 */

package main

import (
	"github.com/Qitmeer/qitmeer/core/json"
	"github.com/Qitmeer/qitmeer/node"
	"github.com/Qitmeer/qitmeer/params"
	"github.com/Qitmeer/qitmeer/rpc"
	"github.com/Qitmeer/qitmeer/rpc/client/cmds"
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
