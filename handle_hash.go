// Copyright 2017-2018 The qitmeer developers
// Use of this source code is governed by an ISC
// license that can be found in the LICENSE file.
package main

import (
	"crypto"
	"encoding/hex"
	"fmt"
	"github.com/HalalChain/qitmeer-lib/common/hash"
	"github.com/HalalChain/qitmeer-lib/common/hash/btc"
	"github.com/HalalChain/qitmeer-lib/common/hash/dcr"
)

func sha256(input string){
	data, err :=hex.DecodeString(input)
	if err != nil {
		errExit(err)
	}
	fmt.Printf("%x\n",btc.HashB(data))
}

func blake256(input string){
	data, err :=hex.DecodeString(input)
	if err != nil {
		errExit(err)
	}
	fmt.Printf("%x\n",dcr.HashB(data))
}

func blake2b256(input string){
	data, err :=hex.DecodeString(input)
	if err != nil {
		errExit(err)
	}
	fmt.Printf("%x\n",hash.HashB(data))
}

func blake2b512(input string){
	data, err :=hex.DecodeString(input)
	if err != nil {
		errExit(err)
	}
	fmt.Printf("%x\n",hash.Hash512B(data))
}

func sha3_256(input string){
	data, err :=hex.DecodeString(input)
	if err != nil {
		errExit(err)
	}
	fmt.Printf("%x\n",hash.CalcHash(data,hash.GetHasher(hash.SHA3_256)))
}

func keccak256(input string){
	data, err :=hex.DecodeString(input)
	if err != nil {
		errExit(err)
	}
	fmt.Printf("%x\n",hash.CalcHash(data,hash.GetHasher(hash.Keccak_256)))
}

func ripemd160(input string){
	data, err :=hex.DecodeString(input)
	if err != nil {
		errExit(err)
	}
	hasher := crypto.RIPEMD160.New()
	hasher.Write(data)
	hash := hasher.Sum(nil)
	fmt.Printf("%x\n",hash[:])
}

func bitcoin160(input string){
	data, err :=hex.DecodeString(input)
	if err != nil {
		errExit(err)
	}
	fmt.Printf("%x\n",btc.Hash160(data))
}

func hash160(input string){
	data, err :=hex.DecodeString(input)
	if err != nil {
		errExit(err)
	}
	fmt.Printf("%x\n",hash.Hash160(data))
}

