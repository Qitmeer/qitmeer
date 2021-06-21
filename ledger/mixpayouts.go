// This file is ignored during the regular build due to the following build tag.
// It is called by go generate and used to automatically generate pre-computed
// tables used to accelerate operations.
// +build ignore

//go:generate go run mixpayouts.go ledgerpayout.go
package main

import (
	"github.com/Qitmeer/qitmeer/core/types"
	"github.com/Qitmeer/qitmeer/params"
)

var MixGeneData = []GenesisInitPayout{
	{
		types.MEERID, "XmspWkqJv6a4sziWrPZbSWQ37WoNEmTD1xm", 10000, GENE_PAYOUT_TYPE_STANDARD, 0,
	},
	{
		types.MEERID, "XmspWkqJv6a4sziWrPZbSWQ37WoNEmTD1xm", 10000, GENE_PAYOUT_TYPE_STANDARD, 0,
	},
	{
		types.MEERID, "XmspWkqJv6a4sziWrPZbSWQ37WoNEmTD1xm", 10000, GENE_PAYOUT_TYPE_STANDARD, 0,
	},

	{
		types.MEERID, "XmkgU8m4G2GwRjz6rEVskG9HAabT5uUS8Fy", 20000, GENE_PAYOUT_TYPE_STANDARD, 0,
	},
	{
		types.MEERID, "XmkgU8m4G2GwRjz6rEVskG9HAabT5uUS8Fy", 20000, GENE_PAYOUT_TYPE_STANDARD, 0,
	},
	{
		types.MEERID, "XmkgU8m4G2GwRjz6rEVskG9HAabT5uUS8Fy", 20000, GENE_PAYOUT_TYPE_STANDARD, 0,
	},

	{
		types.MEERID, "XmnsdQkQYWHih65kMyZPo5bFzRpEyGc3N9x", 30000, GENE_PAYOUT_TYPE_STANDARD, 0,
	},
	{
		types.MEERID, "XmnsdQkQYWHih65kMyZPo5bFzRpEyGc3N9x", 30000, GENE_PAYOUT_TYPE_STANDARD, 0,
	},
	{
		types.MEERID, "XmnsdQkQYWHih65kMyZPo5bFzRpEyGc3N9x", 30000, GENE_PAYOUT_TYPE_STANDARD, 0,
	},
}

// coinid,address,lockAmount,locktype,lockheight
var MixGeneDataFromImport = []string{
	"0,XmnsdQkQYWHih65kMyZPo5bFzRpEyGc3N9x,500,1,0",
}

func main() {
	GeneratePayoutFile(params.MixNetParam.Params, MixGeneData, MixGeneDataFromImport)
}
