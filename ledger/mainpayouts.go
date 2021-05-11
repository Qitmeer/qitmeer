// This file is ignored during the regular build due to the following build tag.
// It is called by go generate and used to automatically generate pre-computed
// tables used to accelerate operations.
// +build ignore

//go:generate go run mainpayouts.go ledgerpayout.go
package main

import (
	"github.com/Qitmeer/qitmeer/params"
)

var MainGeneData = []GenesisInitPayout{}

// coinid,address,lockAmount,locktype,lockheight
var MainGeneDataFromImport = []string{}

func main() {
	GeneratePayoutFile(params.MainNetParam.Params, MainGeneData, MainGeneDataFromImport)
}
