// Copyright (c) 2017-2018 The nox developers
// Copyright (c) 2013-2016 The btcsuite developers
// Copyright (c) 2015-2017 The Decred developers
// Use of this source code is governed by an ISC
// license that can be found in the LICENSE file.

package protocol

import (
	"fmt"
)

const (
	// InitialProcotolVersion is the initial protocol version for the
	// network.
	InitialProcotolVersion uint32 = 1

	// ProtocolVersion is the latest protocol version this package supports.
	ProtocolVersion uint32 = 1
)

// Network represents which nox network a message belongs to.
type Network uint32

// Constants used to indicate the message of network.  They can also be
// used to seek to the next message when a stream's state is unknown, but
// this package does not provide that functionality since it's generally a
// better idea to simply disconnect clients that are misbehaving over TCP.
const (
	// MainNet represents the main network.
	MainNet Network = 0xb4c3dce8

	// TestNet2 represents the test network.
	TestNet Network = 0x35e0c424

	// PrivNet represents the private test network.
	PrivNet Network = 0xf1eb0001
)

// bnStrings is a map of networks back to their constant names for
// pretty printing.
var bnStrings = map[Network]string{
	MainNet:  "MainNet",
	TestNet:  "TestNet",
	PrivNet:  "PirvNet",
}

// String returns the CurrencyNet in human-readable form.
func (n Network) String() string {
	if s, ok := bnStrings[n]; ok {
		return s
	}
	return fmt.Sprintf("Unknown Network (%d)", uint32(n))
}

// ServiceFlag identifies services supported by a peer node.
type ServiceFlag uint64

const (
	//  full node.
	Full ServiceFlag = 1 << iota

	// light node
	Light

	// a peer supports bloom filtering.
	Bloom

	// a peer supports committed filters (CFs).
	CF
)
