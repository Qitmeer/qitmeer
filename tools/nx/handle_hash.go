// Copyright 2017-2018 The nox developers
// Use of this source code is governed by an ISC
// license that can be found in the LICENSE file.
package main

import (
	"encoding/hex"
	"fmt"
	"github.com/noxproject/nox/common/hash"
	"github.com/noxproject/nox/common/hash/btc"
)

func sha256(input string){
	data, err :=hex.DecodeString(input)
	if err != nil {
		errExit(err)
	}
	fmt.Printf("%x\n",btc.HashB(data))
}

func blake2b256(input string){
	data, err :=hex.DecodeString(input)
	if err != nil {
		errExit(err)
	}
	fmt.Printf("%x\n",hash.HashB(data))
}

