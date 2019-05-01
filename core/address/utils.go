// Copyright (c) 2017-2018 The nox developers

package address

import (
	"github.com/noxproject/nox/core/types"
	"github.com/noxproject/nox/params"
)

// IsForNetwork returns whether or not the address is associated with the
// passed network.
//TODO, other addr type and ec type check
func IsForNetwork(addr types.Address, p *params.Params) bool {
	switch addr := addr.(type) {
		case *PubKeyHashAddress:
			return addr.netID == p.PubKeyHashAddrID
	}
	return false
}
