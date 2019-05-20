// Copyright (c) 2017-2018 The nox developers

package address

import (
	"qitmeer/core/types"
	"qitmeer/params"
)

// IsForNetwork returns whether or not the address is associated with the
// passed network.
//TODO, other addr type and ec type check
func IsForNetwork(addr types.Address, p *params.Params) bool {
	switch addr := addr.(type) {
		case *PubKeyHashAddress:
			return addr.netID == p.PubKeyHashAddrID || addr.netID == p.PKHEdwardsAddrID
	}
	return false
}
