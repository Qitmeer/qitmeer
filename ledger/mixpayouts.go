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
	// PMEER and HLC mapping data
	"0,Mma1MhE6ETFLNNTcS6PFnHtmoesaXNBC6kr,2100000000,1,0",
}

func main() {
	GeneratePayoutFile(params.MixNetParam.Params, MixGeneData, MixGeneDataFromImport)
}
