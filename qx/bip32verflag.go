// Copyright 2017-2018 The qitmeer developers
// Use of this source code is governed by an ISC
// license that can be found in the LICENSE file.
package qx

import (
	"fmt"
	"github.com/HalalChain/qitmeer-lib/crypto/bip32"
	"github.com/HalalChain/qitmeer-lib/params"
)

var (
	QitmeerMainnetBip32Version = bip32.Bip32Version{PrivKeyVersion: params.MainNetParams.HDPrivateKeyID[:], PubKeyVersion: params.MainNetParams.HDPublicKeyID[:]}
	QitmeerTestnetBip32Version = bip32.Bip32Version{PrivKeyVersion: params.TestNetParams.HDPrivateKeyID[:], PubKeyVersion: params.TestNetParams.HDPublicKeyID[:]}
	QitmeerPrivnetBip32Version = bip32.Bip32Version{PrivKeyVersion: params.PrivNetParams.HDPrivateKeyID[:], PubKeyVersion: params.PrivNetParams.HDPublicKeyID[:]}
)

type Bip32VersionFlag struct {
	version bip32.Bip32Version
	flag    string
}

func (v *Bip32VersionFlag) String() string {
	return v.flag
}

func (v *Bip32VersionFlag) Version() bip32.Bip32Version {
	return v.version
}

func (v *Bip32VersionFlag) Set(versionFlag string) error {
	var version bip32.Bip32Version
	switch versionFlag {
	case "bip32", "btc":
		version = bip32.DefaultBip32Version
	case "mainnet":
		version = QitmeerMainnetBip32Version
	case "testnet":
		version = QitmeerTestnetBip32Version
	case "privnet":
		version = QitmeerPrivnetBip32Version
	default:
		return fmt.Errorf("unknown bip32 version flag %s", versionFlag)
	}
	v.version = version
	v.flag = versionFlag
	return nil
}

func GetBip32NetworkInfo(rawVersionByte []byte) string {
	if QitmeerMainnetBip32Version.IsPrivkeyVersion(rawVersionByte) || QitmeerMainnetBip32Version.IsPubkeyVersion(rawVersionByte) {
		return "qitmeer mainet"
	} else if QitmeerTestnetBip32Version.IsPrivkeyVersion(rawVersionByte) || QitmeerTestnetBip32Version.IsPubkeyVersion(rawVersionByte) {
		return "qitmeer testnet"
	} else if QitmeerPrivnetBip32Version.IsPrivkeyVersion(rawVersionByte) || QitmeerPrivnetBip32Version.IsPubkeyVersion(rawVersionByte) {
		return "qitmeer privnet"
	} else if bip32.DefaultBip32Version.IsPrivkeyVersion(rawVersionByte) || bip32.DefaultBip32Version.IsPubkeyVersion(rawVersionByte) {
		return "btc mainnet"
	} else {
		return "unknown"
	}
}

