// Copyright 2017-2018 The nox developers
// Use of this source code is governed by an ISC
// license that can be found in the LICENSE file.
package main

import (
	"encoding/hex"
	"fmt"
	"github.com/noxproject/nox/common/encode/base58"
	"github.com/noxproject/nox/common/hash"
)

func ecPubKeyToAddress(version []byte, pubkey string) {
	data, err :=hex.DecodeString(pubkey)
	if err != nil {
		errExit(err)
	}
	h := hash.Hash160(data)

	address := base58.NoxCheckEncode(h, version[:])
	fmt.Printf("%s\n",address)
}
