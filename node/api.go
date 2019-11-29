// Copyright (c) 2017-2018 The qitmeer developers
// Copyright (c) 2013-2016 The btcsuite developers
// Copyright (c) 2017-2018 The Decred developers
// Use of this source code is governed by an ISC
// license that can be found in the LICENSE file.
package node

import (
	"fmt"
	"github.com/Qitmeer/qitmeer/core/blockdag"
	"github.com/Qitmeer/qitmeer/core/json"
	"github.com/Qitmeer/qitmeer/core/message"
	"github.com/Qitmeer/qitmeer/core/protocol"
	"github.com/Qitmeer/qitmeer/core/types/pow"
	"github.com/Qitmeer/qitmeer/params"
	"github.com/Qitmeer/qitmeer/rpc"
	"github.com/Qitmeer/qitmeer/version"
	"math/big"
	"strconv"
	"time"
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
		UUID:            message.UUID.String(),
		Version:         int32(1000000*version.Major + 10000*version.Minor + 100*version.Patch),
		ProtocolVersion: int32(protocol.ProtocolVersion),
		TotalSubsidy:    best.TotalSubsidy,
		TimeOffset:      int64(api.node.blockManager.GetChain().TimeSource().Offset().Seconds()),
		Connections:     api.node.node.peerServer.ConnectedCount(),
		PowDiff: json.PowDiff{
			Blake2bdDiff: getDifficultyRatio(blake2bdNodes, api.node.node.Params, pow.BLAKE2BD),
			CuckarooDiff: getDifficultyRatio(cuckarooNodes, api.node.node.Params, pow.CUCKAROO),
			CuckatooDiff: getDifficultyRatio(cuckatooNodes, api.node.node.Params, pow.CUCKATOO),
		},
		TestNet:          api.node.node.Config.TestNet,
		Confirmations:    blockdag.StableConfirmations,
		CoinbaseMaturity: int32(api.node.node.Params.CoinbaseMaturity),
		Modules:          []string{rpc.DefaultServiceNameSpace, rpc.MinerNameSpace, rpc.TestNameSpace},
	}
	ret.GraphState = *getGraphStateResult(best.GraphState)
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
func (api *PublicBlockChainAPI) GetPeerInfo() (interface{}, error) {
	peers := api.node.node.peerServer.ConnectedPeers()
	syncPeerID := api.node.blockManager.SyncPeerID()
	infos := make([]*json.GetPeerInfoResult, 0, len(peers))
	for _, p := range peers {
		statsSnap := p.StatsSnapshot()
		info := &json.GetPeerInfoResult{
			UUID:       statsSnap.UUID.String(),
			ID:         statsSnap.ID,
			Addr:       statsSnap.Addr,
			AddrLocal:  p.LocalAddr().String(),
			Services:   fmt.Sprintf("%08d", uint64(statsSnap.Services)),
			RelayTxes:  !p.IsTxRelayDisabled(),
			LastSend:   statsSnap.LastSend.Unix(),
			LastRecv:   statsSnap.LastRecv.Unix(),
			BytesSent:  statsSnap.BytesSent,
			BytesRecv:  statsSnap.BytesRecv,
			ConnTime:   statsSnap.ConnTime.Unix(),
			PingTime:   float64(statsSnap.LastPingMicros),
			TimeOffset: statsSnap.TimeOffset,
			Version:    statsSnap.Version,
			SubVer:     statsSnap.UserAgent,
			Inbound:    statsSnap.Inbound,
			BanScore:   int32(p.BanScore()),
			SyncNode:   statsSnap.ID == syncPeerID,
		}
		if statsSnap.GraphState != nil {
			info.GraphState = *getGraphStateResult(statsSnap.GraphState)
		}
		if p.LastPingNonce() != 0 {
			wait := float64(time.Since(statsSnap.LastPingTime).Nanoseconds())
			// We actually want microseconds.
			info.PingWait = wait / 1000
		}
		infos = append(infos, info)
	}
	return infos, nil
}

func getGraphStateResult(gs *blockdag.GraphState) *json.GetGraphStateResult {
	if gs != nil {
		tips := []string{}
		for k := range gs.GetTips().GetMap() {
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
