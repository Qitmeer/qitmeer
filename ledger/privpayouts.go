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
	"0,Rm7B35PAP24GkZW1Za2gKrSiigZQ7M46KfL,1000.23456,1,0",
	"0,RmCM99PchggcoZWkMVBWqAMHBcsn3T6VetG,2000.23456,1,0",
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

func main() {
	GeneratePayoutFile(params.PrivNetParam.Params, PrivGeneData, PrivGeneDataFromImport)
}
