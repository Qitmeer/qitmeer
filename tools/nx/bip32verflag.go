// Copyright 2017-2018 The nox developers
// Use of this source code is governed by an ISC
// license that can be found in the LICENSE file.
package main

import (
	"fmt"
	"qitmeer/crypto/bip32"
	"qitmeer/params"
)

var (
	NoxMainnetBip32Version = bip32.Bip32Version{PrivKeyVersion: params.MainNetParams.HDPrivateKeyID[:], PubKeyVersion: params.MainNetParams.HDPublicKeyID[:]}
	NoxTestnetBip32Version = bip32.Bip32Version{PrivKeyVersion: params.TestNetParams.HDPrivateKeyID[:], PubKeyVersion: params.TestNetParams.HDPublicKeyID[:]}
	NoxPrivnetBip32Version = bip32.Bip32Version{PrivKeyVersion: params.PrivNetParams.HDPrivateKeyID[:], PubKeyVersion: params.PrivNetParams.HDPublicKeyID[:]}
)

type bip32VersionFlag struct {
	version bip32.Bip32Version
	flag    string
}

func (v *bip32VersionFlag) String() string {
	return v.flag
}

func (v *bip32VersionFlag) Set(versionFlag string) error {
	var version bip32.Bip32Version
	switch versionFlag {
	case "bip32", "btc":
		version = bip32.DefaultBip32Version
	case "mainnet":
		version = NoxMainnetBip32Version
	case "testnet":
		version = NoxTestnetBip32Version
	case "privnet":
		version = NoxPrivnetBip32Version
	default:
		return fmt.Errorf("unknown bip32 version flag %s", versionFlag)
	}
	v.version = version
	v.flag = versionFlag
	return nil
}

func getBip32NetworkInfo(rawVersionByte []byte) string {
	if NoxMainnetBip32Version.IsPrivkeyVersion(rawVersionByte) || NoxMainnetBip32Version.IsPubkeyVersion(rawVersionByte) {
		return "nox mainet"
	} else if NoxTestnetBip32Version.IsPrivkeyVersion(rawVersionByte) || NoxTestnetBip32Version.IsPubkeyVersion(rawVersionByte) {
		return "nox testnet"
	} else if NoxPrivnetBip32Version.IsPrivkeyVersion(rawVersionByte) || NoxPrivnetBip32Version.IsPubkeyVersion(rawVersionByte) {
		return "nox privnet"
	} else if bip32.DefaultBip32Version.IsPrivkeyVersion(rawVersionByte) || bip32.DefaultBip32Version.IsPubkeyVersion(rawVersionByte) {
		return "btc mainnet"
	} else {
		return "unknown"
	}
}
