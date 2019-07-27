// Copyright 2017-2018 The qitmeer developers
// Use of this source code is governed by an ISC
// license that can be found in the LICENSE file.
package qx

import (
	"encoding/hex"
	"fmt"
	"github.com/HalalChain/qitmeer-lib/common/encode/base58"
	"github.com/HalalChain/qitmeer-lib/common/hash"
)

func EcPubKeyToAddress(version []byte, pubkey string) {
	data, err :=hex.DecodeString(pubkey)
	if err != nil {
		ErrExit(err)
	}
	h := hash.Hash160(data)

	address := base58.QitmeerCheckEncode(h, version[:])
	fmt.Printf("%s\n",address)
}
