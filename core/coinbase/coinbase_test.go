package coinbase

import (
	"github.com/Qitmeer/qng-core/engine/txscript"
	"github.com/Qitmeer/qng-core/params"
	"github.com/stretchr/testify/assert"
	"strings"
	"testing"
)

func TestNewCoinbaseCheck(t *testing.T) {
	cs := &params.Params{
		CoinbaseConfig: params.CoinbaseConfigs{
			{
				Height:  0,
				Version: "0.10.1",
			},
			{
				Height:                    10,
				Version:                   "0.10.2",
				ExtraDataIncludedVer:      true,
				ExtraDataIncludedNodeInfo: true,
			},
			{
				Height:                    20,
				ExtraDataIncludedVer:      false,
				ExtraDataIncludedNodeInfo: true,
				Version:                   "0.10.3",
			},
			{
				Height:                    50,
				ExtraDataIncludedVer:      true,
				ExtraDataIncludedNodeInfo: false,
				Version:                   "0.10.4",
			},
		},
	}
	type expect struct {
		Height        int64
		ExpectVersion string
		ExpectPeerID  string
		VerResult     bool
		NodeResult    bool
	}
	tests := []expect{{
		Height:        2,
		ExpectVersion: "0.10.1",
		ExpectPeerID:  "16Uiu2HAmLeQEdD9oSrYti91eA8U38FhSPz7W4hEVrjLvNs7dfu92",
		VerResult:     false,
		NodeResult:    false,
	}, {
		Height:        11,
		ExpectVersion: "0.10.2",
		ExpectPeerID:  "16Uiu2HAmLeQEdD9oSrYti91eA8U38FhSPz7W4hEVrjLvNs7dfu93",
		VerResult:     true,
		NodeResult:    true,
	}, {
		Height:        22,
		ExpectVersion: "0.10.3",
		ExpectPeerID:  "16Uiu2HAmLeQEdD9oSrYti91eA8U38FhSPz7W4hEVrjLvNs7dfu95",
		VerResult:     false,
		NodeResult:    true,
	}, {
		Height:        1000,
		ExpectVersion: "0.10.4",
		ExpectPeerID:  "16Uiu2HAmLeQEdD9oSrYti91eA8U38FhSPz7W4hEVrjLvNs7dfu94",
		VerResult:     true,
		NodeResult:    false,
	}}
	for _, test := range tests {
		cg := NewCoinbaseGenerator(cs, test.ExpectPeerID)
		nextBlockHeight := test.Height
		extraNonce := 1
		CoinbaseFlags := "/qitmeer/"
		scriptBuilder := txscript.NewScriptBuilder().AddInt64(nextBlockHeight).
			AddInt64(int64(extraNonce)).AddData([]byte(CoinbaseFlags))
		extraData := cg.BuildExtraData(nextBlockHeight)
		if extraData != "" {
			scriptBuilder = scriptBuilder.AddData([]byte(extraData))
		}
		b, _ := scriptBuilder.Script()

		assert.Equal(t, strings.Contains(string(b), test.ExpectPeerID), test.NodeResult)
		assert.Equal(t, strings.Contains(string(b), test.ExpectVersion), test.VerResult)
	}
}
