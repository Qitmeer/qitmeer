// Copyright 2017-2018 The nox developers
// Use of this source code is governed by an ISC
// license that can be found in the LICENSE file.
package main

import (
	"encoding/hex"
	"fmt"
	"github.com/noxproject/nox/common/encode/base58"
	"github.com/noxproject/nox/crypto/bip32"
	"github.com/noxproject/nox/crypto/bip39"
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


func mnemonicNew(entropyStr string) {
	entropy, err := hex.DecodeString(entropyStr)
	if err!=nil {
		errExit(err)
	}
	mnemonic, err := bip39.NewMnemonic(entropy)
	if err!=nil {
		errExit(err)
	}
	fmt.Printf("%s\n",mnemonic)
}

func mnemonicToEntropy(mnemonicStr string) {
	entropy, err :=bip39.EntropyFromMnemonic(mnemonicStr)
	if err!=nil {
		errExit(err)
	}
	fmt.Printf("%x\n",entropy)
}

func mnemonicToSeed(passphrase string, mnemonicStr string) {
	seed, err :=bip39.NewSeedWithErrorChecking(mnemonicStr, passphrase)
	if err!=nil {
		errExit(err)
	}
	fmt.Printf("%x\n",seed)
}
