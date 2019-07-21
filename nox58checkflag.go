// Copyright 2017-2018 The qitmeer developers
// Use of this source code is governed by an ISC
// license that can be found in the LICENSE file.
package main

import (
	"encoding/hex"
	"github.com/HalalChain/qitmeer-lib/params"
)

type qitmeerBase58checkVersionFlag struct {
	ver []byte
	flag string
}
func (n *qitmeerBase58checkVersionFlag) Set(s string) error {
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

func (n *qitmeerBase58checkVersionFlag) String() string{
	return n.flag
}
