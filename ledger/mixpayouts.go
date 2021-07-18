// This file is ignored during the regular build due to the following build tag.
// It is called by go generate and used to automatically generate pre-computed
// tables used to accelerate operations.
// +build ignore

//go:generate go run mixpayouts.go ledgerpayout.go
package main

import (
	"github.com/Qitmeer/qitmeer/params"
)

var MixGeneData = []GenesisInitPayout{}

// coinid,address,lockAmount,locktype,lockheight
var MixGeneDataFromImport = []string{
	"0,XmCUoNaMxFaKU78BtrNtCtfhuR6AeYLUYts,50000,1,0",
	"0,XmEUDxYWL36NZLFTUxtrkVZuUxeVks78qLo,33762.20532,1,0",
}

func main() {
	GeneratePayoutFile(params.MixNetParam.Params, MixGeneData, MixGeneDataFromImport)
}
