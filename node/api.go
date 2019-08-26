// Copyright (c) 2017-2018 The qitmeer developers
// Copyright (c) 2013-2016 The btcsuite developers
// Copyright (c) 2017-2018 The Decred developers
// Use of this source code is governed by an ISC
// license that can be found in the LICENSE file.
package node

import (
	"fmt"
	"github.com/Qitmeer/qitmeer-lib/core/dag"
	"github.com/Qitmeer/qitmeer-lib/core/json"
	"github.com/Qitmeer/qitmeer-lib/core/protocol"
	"github.com/Qitmeer/qitmeer-lib/params"
	"github.com/Qitmeer/qitmeer-lib/rpc"
	"github.com/Qitmeer/qitmeer/core/blockchain"
	"github.com/Qitmeer/qitmeer/version"
	"math/big"
	"strconv"
	"time"
)

func (nf *QitmeerFull) API() rpc.API {
	return rpc.API{
		NameSpace: rpc.DefaultServiceNameSpace,
		Service:   NewPublicBlockChainAPI(nf),
		Public:    true,
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
	ret := &json.InfoNodeResult{
		Version:         int32(1000000*version.Major + 10000*version.Minor + 100*version.Patch),
		ProtocolVersion: int32(protocol.ProtocolVersion),
		TimeOffset:      int64(api.node.blockManager.GetChain().TimeSource().Offset().Seconds()),
		Connections:     api.node.node.peerServer.ConnectedCount(),
		Difficulty:      getDifficultyRatio(best.Bits,api.node.node.Params),
		TestNet:         api.node.node.Config.TestNet,
		Modules:         []string{rpc.DefaultServiceNameSpace,rpc.MinerNameSpace},
	}
	ret.GraphState=*getGraphStateResult(best.GraphState)
	return ret, nil
}

// getDifficultyRatio returns the proof-of-work difficulty as a multiple of the
// minimum difficulty using the passed bits field from the header of a block.
func getDifficultyRatio(bits uint32, params *params.Params) float64 {
	// The minimum difficulty is the max possible proof-of-work limit bits
	// converted back to a number.  Note this is not the same as the proof of
	// work limit directly because the block difficulty is encoded in a block
	// with the compact form which loses precision.
	max := blockchain.CompactToBig(params.PowLimitBits)
	target := blockchain.CompactToBig(bits)

	difficulty := new(big.Rat).SetFrac(max, target)
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
	peersM:=map[string]bool{}
	for _, p := range peers {
		statsSnap := p.StatsSnapshot()
		if _,ok:=peersM[statsSnap.Addr];ok {
			continue
		}
		peersM[statsSnap.Addr]=true
		info := &json.GetPeerInfoResult{
			ID:             statsSnap.ID,
			Addr:           statsSnap.Addr,
			AddrLocal:      p.LocalAddr().String(),
			Services:       fmt.Sprintf("%08d", uint64(statsSnap.Services)),
			RelayTxes:      !p.IsTxRelayDisabled(),
			LastSend:       statsSnap.LastSend.Unix(),
			LastRecv:       statsSnap.LastRecv.Unix(),
			BytesSent:      statsSnap.BytesSent,
			BytesRecv:      statsSnap.BytesRecv,
			ConnTime:       statsSnap.ConnTime.Unix(),
			PingTime:       float64(statsSnap.LastPingMicros),
			TimeOffset:     statsSnap.TimeOffset,
			Version:        statsSnap.Version,
			SubVer:         statsSnap.UserAgent,
			Inbound:        statsSnap.Inbound,
			BanScore:       int32(p.BanScore()),
			SyncNode:       statsSnap.ID == syncPeerID,
		}
		if statsSnap.GraphState!=nil {
			info.GraphState=*getGraphStateResult(statsSnap.GraphState)
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

func getGraphStateResult(gs *dag.GraphState) *json.GetGraphStateResult{
	if gs!=nil {
		tips:=[]string{}
		for k:=range gs.GetTips().GetMap(){
			tips=append(tips,k.String())
		}
		return &json.GetGraphStateResult{
			Tips:tips,
			Total:uint32(gs.GetTotal()),
			Layer:uint32(gs.GetLayer()),
			MainHeight:uint32(gs.GetMainHeight()),
		}
	}
	return nil
}