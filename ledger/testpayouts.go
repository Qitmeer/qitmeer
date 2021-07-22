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
var TestGeneDataFromImport = []string{
	"0,XmLbv8fvAvrBFtQu2e6PwAxqrkMnuwW4Wbg,50000,1,0",
	"0,XmU8ZwQBcWz4WNo9mxnFrfrzd9L9UjShGD9,33000,1,0",
	"0,XmLGBNVDwgsQQLYVkMghZhhggSsXiUuY2SK,50000,1,0",
}

func main() {
	GeneratePayoutFile(params.TestNetParam.Params, TestGeneData, TestGeneDataFromImport)
}
