// Copyright 2017-2018 The nox developers
// Use of this source code is governed by an ISC
// license that can be found in the LICENSE file.
package main

import (
	"fmt"
	"qitmeer/common/encode/base58"
	"qitmeer/common/hash"
)

func ecPubKeyToAddress(version []byte, pubkey string) {
	h := hash.Hash160([]byte(pubkey))
	address := base58.NoxCheckEncode(h, version[:])
	fmt.Printf("%s\n",address)
}
