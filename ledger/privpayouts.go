// This file is ignored during the regular build due to the following build tag.
// It is called by go generate and used to automatically generate pre-computed
// tables used to accelerate operations.
// +build ignore

//go:generate go run privpayouts.go ledgerpayout.go

package main

import (
	"github.com/Qitmeer/qitmeer/core/types"
	"github.com/Qitmeer/qitmeer/params"
)

var PrivGeneData = []GenesisInitPayout{
	{
		types.MEERID, "RmBKxMWg4C4EMzYowisDEGSBwmnR6tPgjLs", 5000, GENE_PAYOUT_TYPE_STANDARD, 0,
	},
	{
		types.MEERID, "RmBKxMWg4C4EMzYowisDEGSBwmnR6tPgjLs", 500, GENE_PAYOUT_TYPE_STANDARD, 0,
	},
	{
		types.MEERID, "RmHFARk5xmoMNUVJ6UCHFiWQML1vxwUhw1b", 100, GENE_PAYOUT_TYPE_LOCK_WITH_HEIGHT, 2,
	},
}

// coinid,address,lockAmount,locktype,lockheight
var PrivGeneDataFromImport = []string{
	"0,RmBKxMWg4C4EMzYowisDEGSBwmnR6tPgjLs,50000,1,0",
	"0,RmHFARk5xmoMNUVJ6UCHFiWQML1vxwUhw1b,1254.345,1,0",
}

func main() {
	GeneratePayoutFile(params.PrivNetParam.Params, PrivGeneData, PrivGeneDataFromImport)
}
