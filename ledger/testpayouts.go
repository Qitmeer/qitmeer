// This file is ignored during the regular build due to the following build tag.
// It is called by go generate and used to automatically generate pre-computed
// tables used to accelerate operations.
// +build ignore

//go:generate go run testpayouts.go ledgerpayout.go
package main

import (
	"github.com/Qitmeer/qitmeer/core/types"
	"github.com/Qitmeer/qitmeer/params"
)

// coinid,address,lockAmount,locktype,lockheight
var TestGeneData = []GenesisInitPayout{
	{
		types.MEERID, "TnGgUt7QtoZ5KygP9qx8yNvPsvaPbqveAfN", 100000, GENE_PAYOUT_TYPE_STANDARD, 0,
	},
	{
		types.MEERID, "TnMdRWtpBYdyhzQZe5XGRtvewd7C9FjPFUd", 100000, GENE_PAYOUT_TYPE_STANDARD, 0,
	},
	{
		types.MEERID, "TnLvcEP8TfQFPp8RG5bPQxcSBiRVcYgVra7", 100000, GENE_PAYOUT_TYPE_STANDARD, 0,
	},
	{
		types.MEERID, "TnYBtsooefyKDFTnX2mbsioaSQai6xfn1p5", 100000, GENE_PAYOUT_TYPE_STANDARD, 0,
	},
}

// coinid,address,lockAmount,locktype,lockheight
var TestGeneDataFromImport = []string{
	"0,TnNbgxLpoPJCLTcsJbHCzpzcHUouTtfbP8c,50000,1,0",
	"0,TnW8Lm56EyS5ax183uy4vKtm3snG2fhkffn,33000,1,0",
	"0,TnNFxCA8a9KRUukU2JsWdMjT7BKeGUXjTdX,50000,1,0",
	"0,TnMTUwUDxJHH2HhM1xABRtTSXvGNvU4uDbw,50000,1,0",
}

func main() {
	GeneratePayoutFile(params.TestNetParam.Params, TestGeneData, TestGeneDataFromImport)
}
