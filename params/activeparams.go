// Copyright (c) 2017-2018 The qitmeer developers
// Copyright (c) 2013-2016 The btcsuite developers
// Copyright (c) 2015-2016 The Decred developers
// Use of this source code is governed by an ISC
// license that can be found in the LICENSE file.

package params

// activeNetParams is a pointer to the parameters specific to the
// currently active network.
var ActiveNetParams = &MainNetParam

// netParams is used to group parameters for various networks such as the main
// network and test networks.
type netParams struct {
	*Params
	RpcPort string
}

// mainNetParams contains parameters specific to the main network
var MainNetParam = netParams{
	Params:  &MainNetParams,
	RpcPort: "8131",
}

// testNetParams contains parameters specific to the test network
var TestNetParam = netParams{
	Params:  &TestNetParams,
	RpcPort: "18131",
}

// privNetParams contains parameters specific to the private test network
var PrivNetParam = netParams{
	Params:  &PrivNetParams,
	RpcPort: "38131",
}

// MixNetParam contains parameters specific to the mix pow test network
var MixNetParam = netParams{
	Params:  &MixNetParams,
	RpcPort: "28131",
}
