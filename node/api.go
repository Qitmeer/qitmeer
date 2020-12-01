// Copyright (c) 2017-2018 The qitmeer developers
// Copyright (c) 2013-2016 The btcsuite developers
// Copyright (c) 2017-2018 The Decred developers
// Use of this source code is governed by an ISC
// license that can be found in the LICENSE file.
package node

import (
	"fmt"
	"github.com/Qitmeer/qitmeer/common/roughtime"
	"github.com/Qitmeer/qitmeer/core/blockdag"
	"github.com/Qitmeer/qitmeer/core/json"
	"github.com/Qitmeer/qitmeer/core/protocol"
	"github.com/Qitmeer/qitmeer/core/types/pow"
	"github.com/Qitmeer/qitmeer/params"
	"github.com/Qitmeer/qitmeer/rpc"
	"github.com/Qitmeer/qitmeer/services/common"
	"github.com/Qitmeer/qitmeer/version"
	"math/big"
	"strconv"
)

func (nf *QitmeerFull) apis() []rpc.API {
	return []rpc.API{
		{
			NameSpace: rpc.DefaultServiceNameSpace,
			Service:   NewPublicBlockChainAPI(nf),
			Public:    true,
		},
		{
			NameSpace: rpc.TestNameSpace,
			Service:   NewPrivateBlockChainAPI(nf),
			Public:    false,
		},
		{
			NameSpace: rpc.LogNameSpace,
			Service:   NewPrivateLogAPI(nf),
			Public:    false,
		},
	}
}

type PublicBlockChainAPI struct {
	node *QitmeerFull
}

func NewPublicBlockChainAPI(node *QitmeerFull) *PublicBlockChainAPI {
	return &PublicBlockChainAPI{node}
}

// Return the node info
func (api *PublicBlockChainAPI) GetNodeInfo() (interface{}, error) {
	best := api.node.blockManager.GetChain().BestSnapshot()
	node := api.node.blockManager.GetChain().BlockIndex().LookupNode(&best.Hash)
	blake2bdNodes := api.node.blockManager.GetChain().GetCurrentPowDiff(*node, pow.BLAKE2BD)
	cuckarooNodes := api.node.blockManager.GetChain().GetCurrentPowDiff(*node, pow.CUCKAROO)
	cuckatooNodes := api.node.blockManager.GetChain().GetCurrentPowDiff(*node, pow.CUCKATOO)
	ret := &json.InfoNodeResult{
		ID:              api.node.node.peerServer.PeerID().String(),
		Version:         int32(1000000*version.Major + 10000*version.Minor + 100*version.Patch),
		BuildVersion:    version.String(),
		ProtocolVersion: int32(protocol.ProtocolVersion),
		TotalSubsidy:    best.TotalSubsidy,
		TimeOffset:      int64(api.node.blockManager.GetChain().TimeSource().Offset().Seconds()),
		Connections:     int32(len(api.node.node.peerServer.Peers().Connected())),
		PowDiff: json.PowDiff{
			Blake2bdDiff: getDifficultyRatio(blake2bdNodes, api.node.node.Params, pow.BLAKE2BD),
			CuckarooDiff: getDifficultyRatio(cuckarooNodes, api.node.node.Params, pow.CUCKAROO),
			CuckatooDiff: getDifficultyRatio(cuckatooNodes, api.node.node.Params, pow.CUCKATOO),
		},
		TestNet:          api.node.node.Config.TestNet,
		MixNet:           api.node.node.Config.MixNet,
		Confirmations:    blockdag.StableConfirmations,
		CoinbaseMaturity: int32(api.node.node.Params.CoinbaseMaturity),
		Modules:          []string{rpc.DefaultServiceNameSpace, rpc.MinerNameSpace, rpc.TestNameSpace, rpc.LogNameSpace},
	}
	ret.GraphState = *getGraphStateResult(best.GraphState)
	hostdns := api.node.node.peerServer.HostDNS()
	if hostdns != nil {
		ret.DNS = hostdns.String()
	}
	if api.node.node.peerServer.Node() != nil {
		ret.QNR = api.node.node.peerServer.Node().String()
	}
	if len(api.node.node.peerServer.HostAddress()) > 0 {
		ret.Addresss = api.node.node.peerServer.HostAddress()
	}
	return ret, nil
}

// getDifficultyRatio returns the proof-of-work difficulty as a multiple of the
// minimum difficulty using the passed bits field from the header of a block.
func getDifficultyRatio(target *big.Int, params *params.Params, powType pow.PowType) float64 {
	instance := pow.GetInstance(powType, 0, []byte{})
	instance.SetParams(params.PowConfig)
	// The minimum difficulty is the max possible proof-of-work limit bits
	// converted back to a number.  Note this is not the same as the proof of
	// work limit directly because the block difficulty is encoded in a block
	// with the compact form which loses precision.
	base := instance.GetSafeDiff(0)
	var difficulty *big.Rat
	if powType == pow.BLAKE2BD {
		difficulty = new(big.Rat).SetFrac(base, target)
	} else {
		difficulty = new(big.Rat).SetFrac(target, base)
	}

	outString := difficulty.FloatString(8)
	diff, err := strconv.ParseFloat(outString, 64)
	if err != nil {
		log.Error(fmt.Sprintf("Cannot get difficulty: %v", err))
		return 0
	}
	return diff
}

// Return the peer info
func (api *PublicBlockChainAPI) GetPeerInfo(verbose *bool) (interface{}, error) {
	vb := true
	if verbose != nil {
		vb = *verbose
	}
	ps := api.node.node.peerServer
	peers := ps.Peers().StatsSnapshots()
	infos := make([]*json.GetPeerInfoResult, 0, len(peers))
	for _, p := range peers {
		if !vb {
			if p.State.IsDisconnected() || p.State.IsDisconnecting() {
				continue
			}
		}
		info := &json.GetPeerInfoResult{
			ID:      p.PeerID,
			State:   p.State.String(),
			Address: p.Address,
		}
		if p.State.IsConnected() {
			info.Protocol = p.Protocol
			info.Services = p.Services.String()
			info.UserAgent = p.UserAgent
			info.TimeOffset = p.TimeOffset
			if p.Genesis != nil {
				info.Genesis = p.Genesis.String()
			}
			info.Direction = p.Direction.String()
			if p.GraphState != nil {
				info.GraphState = getGraphStateResult(p.GraphState)
			}
			if ps.PeerSync().SyncPeer() != nil {
				info.SyncNode = p.PeerID == ps.PeerSync().SyncPeer().GetID().String()
			} else {
				info.SyncNode = false
			}
		}
		if len(p.QNR) > 0 {
			info.QNR = p.QNR
		}
		infos = append(infos, info)
	}
	return infos, nil
}

// Return the RPC info
func (api *PublicBlockChainAPI) GetRpcInfo() (interface{}, error) {
	rs := api.node.node.rpcServer.ReqStatus
	jrs := []*rpc.JsonRequestStatus{}
	for _, v := range rs {
		jrs = append(jrs, v.ToJson())
	}
	return jrs, nil
}

func getGraphStateResult(gs *blockdag.GraphState) *json.GetGraphStateResult {
	if gs != nil {
		mainTip := gs.GetMainChainTip()
		tips := []string{mainTip.String() + " main"}
		for k := range gs.GetTips().GetMap() {
			if k.IsEqual(mainTip) {
				continue
			}
			tips = append(tips, k.String())
		}
		return &json.GetGraphStateResult{
			Tips:       tips,
			MainOrder:  uint32(gs.GetMainOrder()),
			Layer:      uint32(gs.GetLayer()),
			MainHeight: uint32(gs.GetMainHeight()),
		}
	}
	return nil
}

func (api *PublicBlockChainAPI) GetTimeInfo() (interface{}, error) {
	return fmt.Sprintf("Now:%s offset:%s", roughtime.Now(), roughtime.Offset()), nil
}

type PrivateBlockChainAPI struct {
	node *QitmeerFull
}

func NewPrivateBlockChainAPI(node *QitmeerFull) *PrivateBlockChainAPI {
	return &PrivateBlockChainAPI{node}
}

// Stop the node
func (api *PrivateBlockChainAPI) Stop() (interface{}, error) {
	select {
	case api.node.node.rpcServer.RequestedProcessShutdown() <- struct{}{}:
	default:
	}
	return "Qitmeer stopping.", nil
}

// Banlist
func (api *PrivateBlockChainAPI) Banlist() (interface{}, error) {
	bl := api.node.node.peerServer.GetBanlist()
	bls := []*json.GetBanlistResult{}
	for k, v := range bl {
		bls = append(bls, &json.GetBanlistResult{ID: k, Bads: v})
	}
	return bls, nil
}

// RemoveBan
func (api *PrivateBlockChainAPI) RemoveBan(id *string) (interface{}, error) {
	ho := ""
	if id != nil {
		ho = *id
	}
	api.node.node.peerServer.RemoveBan(ho)
	return true, nil
}

// SetRpcMaxClients
func (api *PrivateBlockChainAPI) SetRpcMaxClients(max int) (interface{}, error) {
	if max <= 0 {
		err := fmt.Errorf("error:Must greater than 0 (cur max =%d)", api.node.node.Config.RPCMaxClients)
		return api.node.node.Config.RPCMaxClients, err
	}
	api.node.node.Config.RPCMaxClients = max
	return api.node.node.Config.RPCMaxClients, nil
}

type PrivateLogAPI struct {
	node *QitmeerFull
}

func NewPrivateLogAPI(node *QitmeerFull) *PrivateLogAPI {
	return &PrivateLogAPI{node}
}

// set log
func (api *PrivateLogAPI) SetLogLevel(level string) (interface{}, error) {
	err := common.ParseAndSetDebugLevels(level)
	if err != nil {
		return nil, err
	}
	return level, nil
}
