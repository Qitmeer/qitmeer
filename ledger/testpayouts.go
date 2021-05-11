// This file is ignored during the regular build due to the following build tag.
// It is called by go generate and used to automatically generate pre-computed
// tables used to accelerate operations.
// +build ignore

//go:generate go run testpayouts.go ledgerpayout.go
package main

import (
	"github.com/Qitmeer/qitmeer/params"
)

var TestGeneData = []GenesisInitPayout{}

// coinid,address,lockAmount,locktype,lockheight
var TestGeneDataFromImport = []string{}

func main() {
	GeneratePayoutFile(params.TestNetParam.Params, TestGeneData, TestGeneDataFromImport)
}
