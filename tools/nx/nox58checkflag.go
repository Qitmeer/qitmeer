// Copyright 2017-2018 The nox developers
// Use of this source code is governed by an ISC
// license that can be found in the LICENSE file.
package main

import (
	"encoding/hex"
	"qitmeer/params"
)

type noxBase58checkVersionFlag struct {
	ver  []byte
	flag string
	mode string
}

//btcmainnet|btctestnet|btcregressionnet
func (n *noxBase58checkVersionFlag) Set(s string) error {
	n.ver = []byte{}
	switch s {
	case "mainnet":
		n.ver = append(n.ver, params.MainNetParams.PubKeyHashAddrID[0:]...)
		n.mode = "nox"
	case "privnet":
		n.ver = append(n.ver, params.PrivNetParams.PubKeyHashAddrID[0:]...)
		n.mode = "nox"
	case "testnet":
		n.ver = append(n.ver, params.TestNetParams.PubKeyHashAddrID[0:]...)
		n.mode = "nox"
	default:
		v, err := hex.DecodeString(s)
		if err != nil {
			errExit(err)
		}
		n.ver = append(n.ver, v...)
	}
	n.flag = s
	return nil
}

func (n *noxBase58checkVersionFlag) String() string {
	return n.flag
}
