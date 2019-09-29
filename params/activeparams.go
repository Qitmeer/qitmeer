// Copyright (c) 2017-2018 The qitmeer developers
// Copyright (c) 2013-2016 The btcsuite developers
// Copyright (c) 2015-2016 The Decred developers
// Use of this source code is governed by an ISC
// license that can be found in the LICENSE file.

package params

import "github.com/Qitmeer/qitmeer-lib/params"

// activeNetParams is a pointer to the parameters specific to the
// currently active network.
var ActiveNetParams = &MainNetParams

// netParams is used to group parameters for various networks such as the main
// network and test networks.
type netParams struct {
	*params.Params
	RpcPort string
}

// mainNetParams contains parameters specific to the main network
var MainNetParams = netParams {
	Params:  &params.MainNetParams,
	RpcPort: "8131",
}

// testNetParams contains parameters specific to the test network
var TestNetParams = netParams{
	Params:  &params.TestNetParams,
	RpcPort: "18131",
}

// privNetParams contains parameters specific to the private test network
var PrivNetParams = netParams{
	Params:  &params.PrivNetParams,
	RpcPort: "28131",
}

