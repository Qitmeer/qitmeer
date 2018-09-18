// Copyright 2017-2018 The nox developers
// Use of this source code is governed by an ISC
// license that can be found in the LICENSE file.
package main

import (
	"fmt"
	"github.com/noxproject/nox/common/encode/base58"
	"testing"
)

func TestNoxBase58CheckEncode(t *testing.T) {
	// Encode example data with the Base58Check encoding scheme.
	data := []byte{ 0x64, 0xe2, 0x0e, 0xb6, 0x07, 0x55, 0x61, 0xd3, 0x0c, 0x23, 0xa5, 0x17,
		0xc5, 0xb7, 0x3b, 0xad, 0xbc, 0x12, 0x0f, 0x05}

	for i:=byte(0); i<byte(0xff); i++{
		for j:=byte(0); j< byte(0xff); j++{
			ver := [2]byte{i,j}
			encoded := base58.CheckEncode(data, ver)
			// Show the encoded data.
			fmt.Printf("Encoded Data ver[%x,%x] : %s\n",i,j, encoded)
		}
	}
}
