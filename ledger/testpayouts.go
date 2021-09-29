// This file is ignored during the regular build due to the following build tag.
// It is called by go generate and used to automatically generate pre-computed
// tables used to accelerate operations.
// +build ignore

//go:generate go run testpayouts.go ledgerpayout.go
package main

import (
	"github.com/Qitmeer/qitmeer/params"
)

// coinid,address,lockAmount,locktype,lockheight
var TestGeneData = []GenesisInitPayout{}

// coinid,address,lockAmount,locktype,lockheight
var TestGeneDataFromImport = []string{
	"0,TnNbgxLpoPJCLTcsJbHCzpzcHUouTtfbP8c,5000000000000,1,0",
	"0,TnW8Lm56EyS5ax183uy4vKtm3snG2fhkffn,3300000000000,1,0",
	"0,TnNFxCA8a9KRUukU2JsWdMjT7BKeGUXjTdX,5000000000000,1,0",
	"0,TnMTUwUDxJHH2HhM1xABRtTSXvGNvU4uDbw,5000000000000,1,0",
}

func main() {
	GeneratePayoutFile(params.TestNetParam.Params, TestGeneData, TestGeneDataFromImport)
}
