package coinbase

import (
	"github.com/Qitmeer/qng-core/params"
	"strings"
)

type CoinbaseGenerator struct {
	configs  *params.CoinbaseConfigs
	nodeInfo string
}

func NewCoinbaseGenerator(param *params.Params, nodeInfo string) *CoinbaseGenerator {
	return &CoinbaseGenerator{
		configs:  &param.CoinbaseConfig,
		nodeInfo: nodeInfo,
	}
}

func (cg *CoinbaseGenerator) BuildExtraData(curHeight int64) string {
	current := cg.configs.GetCurrentConfig(curHeight)
	sb := strings.Builder{}
	if current != nil {
		if current.ExtraDataIncludedVer {
			sb.WriteString(current.Version)
		}
		if current.ExtraDataIncludedNodeInfo {
			sb.WriteString(cg.nodeInfo)
		}
	}
	return sb.String()
}

func (cg *CoinbaseGenerator) PeerID() string {
	return cg.nodeInfo
}
