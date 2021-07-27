// Copyright (c) 2017-2018 The qitmeer developers
// Copyright (c) 2013-2016 The btcsuite developers
// Copyright (c) 2017-2018 The Decred developers
// Use of this source code is governed by an ISC
// license that can be found in the LICENSE file.
package node

import (
	"fmt"
	"github.com/Qitmeer/qitmeer/common/roughtime"
	"github.com/Qitmeer/qitmeer/core/blockchain"
	"github.com/Qitmeer/qitmeer/core/blockdag"
	"github.com/Qitmeer/qitmeer/core/json"
	"github.com/Qitmeer/qitmeer/core/protocol"
	"github.com/Qitmeer/qitmeer/core/types/pow"
	"github.com/Qitmeer/qitmeer/params"
	"github.com/Qitmeer/qitmeer/rpc"
	"github.com/Qitmeer/qitmeer/rpc/client/cmds"
	"github.com/Qitmeer/qitmeer/services/common"
	"github.com/Qitmeer/qitmeer/version"
	"math/big"
	"strconv"
	"time"
)

func (nf *QitmeerFull) apis() []rpc.API {
	return []rpc.API{
		{
			NameSpace: cmds.DefaultServiceNameSpace,
			Service:   NewPublicBlockChainAPI(nf),
			Public:    true,
		},
		{
			NameSpace: cmds.TestNameSpace,
			Service:   NewPrivateBlockChainAPI(nf),
			Public:    false,
		},
		{
			NameSpace: cmds.LogNameSpace,
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
	node := api.node.blockManager.GetChain().BlockDAG().GetBlock(&best.Hash)
	powNodes := api.node.blockManager.GetChain().GetCurrentPowDiff(node, pow.MEERXKECCAKV1)
	ret := &json.InfoNodeResult{
		ID:              api.node.node.peerServer.PeerID().String(),
		Version:         int32(1000000*version.Major + 10000*version.Minor + 100*version.Patch),
		BuildVersion:    version.String(),
		ProtocolVersion: int32(protocol.ProtocolVersion),
		TotalSubsidy:    best.TotalSubsidy,
		TimeOffset:      int64(api.node.blockManager.GetChain().TimeSource().Offset().Seconds()),
		Connections:     int32(len(api.node.node.peerServer.Peers().Connected())),
		PowDiff: json.PowDiff{
			CurrentDiff: getDifficultyRatio(powNodes, api.node.node.Params, pow.MEERXKECCAKV1),
		},
		Network:          params.ActiveNetParams.Name,
		Confirmations:    blockdag.StableConfirmations,
		CoinbaseMaturity: int32(api.node.node.Params.CoinbaseMaturity),
		Modules:          []string{cmds.DefaultServiceNameSpace, cmds.MinerNameSpace, cmds.TestNameSpace, cmds.LogNameSpace},
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

	// soft forks
	ret.ConsensusDeployment = make(map[string]*json.ConsensusDeploymentDesc)
	for deployment, deploymentDetails := range params.ActiveNetParams.Deployments {
		// Map the integer deployment ID into a human readable
		// fork-name.
		var forkName string
		switch deployment {
		case params.DeploymentTestDummy:
			forkName = "dummy"

		case params.DeploymentToken:
			forkName = "token"

		default:
			return nil, fmt.Errorf("Unknown deployment %v detected\n", deployment)
		}

		// Query the chain for the current status of the deployment as
		// identified by its deployment ID.
		deploymentStatus, err := api.node.blockManager.GetChain().ThresholdState(uint32(deployment))
		if err != nil {
			return nil, fmt.Errorf("Failed to obtain deployment status\n")
		}

		// Finally, populate the soft-fork description with all the
		// information gathered above.
		ret.ConsensusDeployment[forkName] = &json.ConsensusDeploymentDesc{
			Status:    deploymentStatus.HumanString(),
			Bit:       deploymentDetails.BitNumber,
			StartTime: int64(deploymentDetails.StartTime),
			Timeout:   int64(deploymentDetails.ExpireTime),
		}

		if deploymentDetails.PerformTime != 0 {
			ret.ConsensusDeployment[forkName].Perform = int64(deploymentDetails.PerformTime)
		}

		if deploymentDetails.StartTime >= blockchain.CheckerTimeThreshold {
			if time.Unix(int64(deploymentDetails.ExpireTime), 0).After(best.MedianTime) {
				startTime := time.Unix(int64(deploymentDetails.StartTime), 0)
				ret.ConsensusDeployment[forkName].Since = best.MedianTime.Sub(startTime).String()
			}
		}

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
	if powType == pow.BLAKE2BD || powType == pow.MEERXKECCAKV1 ||
		powType == pow.QITMEERKECCAK256 ||
		powType == pow.X8R16 ||
		powType == pow.X16RV3 ||
		powType == pow.CRYPTONIGHT {
		if target.Cmp(big.NewInt(0)) > 0 {
			difficulty = new(big.Rat).SetFrac(base, target)
		}
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
	vb := false
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
			ID:        p.PeerID,
			Name:      p.GetName(),
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
		version := p.GetVersion()
		if len(version) > 0 {
			info.Version = version
		}
		network := p.GetNetwork()
		if len(version) > 0 {
			info.Network = network
		}

		if p.State.IsConnected() {
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

// Return the RPC info
func (api *PublicBlockChainAPI) GetRpcInfo() (interface{}, error) {
	rs := api.node.node.rpcServer.ReqStatus
	jrs := []*cmds.JsonRequestStatus{}
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
