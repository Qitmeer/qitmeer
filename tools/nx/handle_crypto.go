// Copyright 2017-2018 The nox developers
// Use of this source code is governed by an ISC
// license that can be found in the LICENSE file.
package main

import (
	"encoding/hex"
	"fmt"
	"github.com/noxproject/nox/common/encode/base58"
	"github.com/noxproject/nox/crypto/bip32"
	"github.com/noxproject/nox/crypto/seed"
)

func newSeed(size uint) {
	s,err :=seed.GenerateSeed(uint16(size))
	if err!=nil {
		errExit(err)
	}
	fmt.Printf("%x\n",s)
}

func hdNewMasterPrivateKey(version string, seedStr string){
	seed, err := hex.DecodeString(seedStr)
	if err!=nil {
		errExit(err)
	}
	masterKey, err := bip32.NewMasterKey(seed)
	if err !=nil {
		errExit(err)
	}
	fmt.Printf("%s\n",masterKey)
}

func hdPrivateKeyToHdPublicKey(privateKeyStr string){
	data := base58.Decode(privateKeyStr)
	masterKey, err :=bip32.Deserialize(data)
	if err !=nil {
		errExit(err)
	}
	pubKey := masterKey.PublicKey()
	fmt.Printf("%s\n",pubKey)
}
