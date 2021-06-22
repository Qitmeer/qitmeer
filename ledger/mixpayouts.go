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

/**
// mainnet : MmRfhzqvkimi8VrSTPXfFr8W36KNhvyLzBx
// mixnet : Xmf3VqRkCYuuZnsVNgVxq6vzUfitw2sQy2t
// testnet : TmV2ThmkxspY1un3PZRmwJawYWfEetHYmmc
// privnet : Rm7B35PAP24GkZW1Za2gKrSiigZQ7M46KfL
*/

/**
[mainnet] MmLVbvqUS49N5VqhfUNpkYDwa9zzmrX1CtJ
[mixnet] XmZsPmRHstHZWnrkamM8Ko2S1jQX1748nxx
[testnet] TmPrMdmJeDCBxumJbeGwRzgP5aLriqGeE5o
[privnet] RmCM99PchggcoZWkMVBWqAMHBcsn3T6VetG
*/

// coinid,address,lockAmount,locktype,lockheight
var MixGeneDataFromImport = []string{
	"0,Xmf3VqRkCYuuZnsVNgVxq6vzUfitw2sQy2t,5000,1,0",
	"0,XmZsPmRHstHZWnrkamM8Ko2S1jQX1748nxx,5000,1,0",
	"0,XmkgU8m4G2GwRjz6rEVskG9HAabT5uUS8Fy,50000,1,0", // will release 6 days
}

func main() {
	GeneratePayoutFile(params.MixNetParam.Params, MixGeneData, MixGeneDataFromImport)
}
