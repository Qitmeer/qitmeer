// Copyright 2017-2018 The nox developers
// Use of this source code is governed by an ISC
// license that can be found in the LICENSE file.
package main

import (
	"encoding/hex"
	"qitmeer/params"
	"errors"
)

type noxBase58checkVersionFlag struct {
	ver []byte
	flag string
}
func (n *noxBase58checkVersionFlag) Set(s string) error {
	n.ver = []byte{}
	switch (s) {
	case "mainnet":
		n.ver = append(n.ver, params.MainNetParams.PubKeyHashAddrID[0:]...)
	case "privnet":
		n.ver = append(n.ver, params.PrivNetParams.PubKeyHashAddrID[0:]...)
	case "testnet":
		n.ver = append(n.ver, params.TestNetParams.PubKeyHashAddrID[0:]...)
	default:
		v, err := hex.DecodeString(s)
		if err!=nil {
			errExit(err)
		}
		n.ver = append(n.ver,v...)
	}
	n.flag = s
	return nil
}
func (n *noxBase58checkVersionFlag) SetCurve(s string) error {
	n.ver = []byte{}
	addrID := [2]byte{}
	p := params.Params{}
	switch n.flag {
	case "mainnet":
		p = params.MainNetParams
	case "privnet":
		p = params.PrivNetParams
	case "testnet":
		p = params.TestNetParams
	}
	switch curve {
	case "secp256k1":
		addrID = p.PubKeyHashAddrID
	case "ed25519":
		addrID = p.PKHEdwardsAddrID
	default:
		errExit(errors.New("curve not support:"+curve))
	}
	n.ver = addrID[:]
	return nil
}

func (n *noxBase58checkVersionFlag) String() string{
	return n.flag
}
