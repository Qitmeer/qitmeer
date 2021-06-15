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

var MixGeneDataFromImport = []string{}
	addPayout2("XmspWkqJv6a4sziWrPZbSWQ37WoNEmTD1xm",
		Amount{Value: 10000 * AtomsPerCoin, Id: QITID}, "76a914cec9c6fb443a62b5e8d7a00fc3ad0f3cef1f070588ac")
	addPayout2("XmkgU8m4G2GwRjz6rEVskG9HAabT5uUS8Fy",
		Amount{Value: 20000 * AtomsPerCoin, Id: QITID}, "76a914807b61617f12c3166f4531ddeae484b82187ae2f88ac")
	addPayout2("XmnsdQkQYWHih65kMyZPo5bFzRpEyGc3N9x",
		Amount{Value: 30000 * AtomsPerCoin, Id: QITID}, "76a9149887f352a02c4e60d99bcd2eab33c8b7b0198b0488ac")

func main() {
	GeneratePayoutFile(params.MixNetParam.Params, MixGeneData, MixGeneDataFromImport)
}
