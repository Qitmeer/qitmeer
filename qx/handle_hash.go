// Copyright 2017-2018 The qitmeer developers
// Use of this source code is governed by an ISC
// license that can be found in the LICENSE file.
package qx

import (
	"crypto"
	"encoding/hex"
	"fmt"
	"github.com/Qitmeer/qng-core/common/hash"
	"github.com/Qitmeer/qng-core/common/hash/btc"
	"github.com/Qitmeer/qng-core/common/hash/dcr"
)

func Sha256(input string) {
	data, err := hex.DecodeString(input)
	if err != nil {
		ErrExit(err)
	}
	fmt.Printf("%x\n", btc.HashB(data))
}

func Blake256(input string) {
	data, err := hex.DecodeString(input)
	if err != nil {
		ErrExit(err)
	}
	fmt.Printf("%x\n", dcr.HashB(data))
}

func Blake2b256(input string) {
	data, err := hex.DecodeString(input)
	if err != nil {
		ErrExit(err)
	}
	fmt.Printf("%x\n", hash.HashB(data))
}

func Blake2b512(input string) {
	data, err := hex.DecodeString(input)
	if err != nil {
		ErrExit(err)
	}
	fmt.Printf("%x\n", hash.Hash512B(data))
}

func Sha3_256(input string) {
	data, err := hex.DecodeString(input)
	if err != nil {
		ErrExit(err)
	}
	fmt.Printf("%x\n", hash.CalcHash(data, hash.GetHasher(hash.SHA3_256)))
}

func Keccak256(input string) {
	data, err := hex.DecodeString(input)
	if err != nil {
		ErrExit(err)
	}
	fmt.Printf("%x\n", hash.CalcHash(data, hash.GetHasher(hash.Keccak_256)))
}

func Ripemd160(input string) {
	data, err := hex.DecodeString(input)
	if err != nil {
		ErrExit(err)
	}
	hasher := crypto.RIPEMD160.New()
	hasher.Write(data)
	hash := hasher.Sum(nil)
	fmt.Printf("%x\n", hash[:])
}

func Bitcoin160(input string) {
	data, err := hex.DecodeString(input)
	if err != nil {
		ErrExit(err)
	}
	fmt.Printf("%x\n", btc.Hash160(data))
}

func Hash160(input string) {
	data, err := hex.DecodeString(input)
	if err != nil {
		ErrExit(err)
	}
	fmt.Printf("%x\n", hash.Hash160(data))
}
