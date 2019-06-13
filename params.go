// Copyright (c) 2017-2018 The nox developers
// Copyright (c) 2013-2016 The btcsuite developers
// Copyright (c) 2015-2016 The Decred developers
// Use of this source code is governed by an ISC
// license that can be found in the LICENSE file.

package main

import "github.com/HalalChain/qitmeer/params"

// activeNetParams is a pointer to the parameters specific to the
// currently active network.
var activeNetParams = &mainNetParams

// netParams is used to group parameters for various networks such as the main
// network and test networks.
type netParams struct {
	*params.Params
	rpcPort string
}

// mainNetParams contains parameters specific to the main network
var mainNetParams = netParams {
	Params:  &params.MainNetParams,
	rpcPort: "8131",
}

// testNetParams contains parameters specific to the test network
var testNetParams = netParams{
	Params:  &params.TestNetParams,
	rpcPort: "18131",
}

// privNetParams contains parameters specific to the private test network
var privNetParams = netParams{
	Params:  &params.PrivNetParams,
	rpcPort: "28131",
}

