// This file is ignored during the regular build due to the following build tag.
// It is called by go generate and used to automatically generate pre-computed
// tables used to accelerate operations.
// +build ignore

//go:generate go run privpayouts.go ledgerpayout.go

package main

import (
	"github.com/Qitmeer/qitmeer/params"
)

var PrivGeneData = []GenesisInitPayout{}

// coinid,address,lockAmount,locktype,lockheight
var PrivGeneDataFromImport = []string{
	"0,RmBKxMWg4C4EMzYowisDEGSBwmnR6tPgjLs,5000000000000,0,0",
	"0,RmBKxMWg4C4EMzYowisDEGSBwmnR6tPgjLs,50000000000,0,0",
	"0,RmHFARk5xmoMNUVJ6UCHFiWQML1vxwUhw1b,10000000000,2,2",
	"0,RmBKxMWg4C4EMzYowisDEGSBwmnR6tPgjLs,5000000000000,1,0",
	"0,RmHFARk5xmoMNUVJ6UCHFiWQML1vxwUhw1b,125400000000,1,0",
	"0,Rm7B35PAP24GkZW1Za2gKrSiigZQ7M46KfL,100000000000,1,0",
	"0,RmCM99PchggcoZWkMVBWqAMHBcsn3T6VetG,200000000000,1,0",
}

func main() {
	GeneratePayoutFile(params.PrivNetParam.Params, PrivGeneData, PrivGeneDataFromImport)
}
