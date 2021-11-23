// Copyright 2017-2018 The qitmeer developers
// Use of this source code is governed by an ISC
// license that can be found in the LICENSE file.
package qx

import (
	"fmt"
	"github.com/Qitmeer/qng-core/crypto/bip32"
	"github.com/Qitmeer/qng-core/params"
)

var (
	QitmeerMainnetBip32Version = bip32.Bip32Version{PrivKeyVersion: params.MainNetParams.HDPrivateKeyID[:], PubKeyVersion: params.MainNetParams.HDPublicKeyID[:]}
	QitmeerTestnetBip32Version = bip32.Bip32Version{PrivKeyVersion: params.TestNetParams.HDPrivateKeyID[:], PubKeyVersion: params.TestNetParams.HDPublicKeyID[:]}
	QitmeerPrivnetBip32Version = bip32.Bip32Version{PrivKeyVersion: params.PrivNetParams.HDPrivateKeyID[:], PubKeyVersion: params.PrivNetParams.HDPublicKeyID[:]}
	QitmeerMixnetBip32Version  = bip32.Bip32Version{PrivKeyVersion: params.MixNetParam.HDPrivateKeyID[:], PubKeyVersion: params.MixNetParam.HDPublicKeyID[:]}
)

type Bip32VersionFlag struct {
	Version bip32.Bip32Version
	flag    string
}

func (v *Bip32VersionFlag) String() string {
	return v.flag
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
	case "mixnet":
		version = QitmeerMixnetBip32Version
	default:
		return fmt.Errorf("unknown bip32 version flag %s", versionFlag)
	}
	v.Version = version
	v.flag = versionFlag
	return nil
}

func GetBip32NetworkInfo(rawVersionByte []byte) string {
	if QitmeerMainnetBip32Version.IsPrivkeyVersion(rawVersionByte) || QitmeerMainnetBip32Version.IsPubkeyVersion(rawVersionByte) {
		return "qx mainet"
	} else if QitmeerTestnetBip32Version.IsPrivkeyVersion(rawVersionByte) || QitmeerTestnetBip32Version.IsPubkeyVersion(rawVersionByte) {
		return "qx testnet"
	} else if QitmeerPrivnetBip32Version.IsPrivkeyVersion(rawVersionByte) || QitmeerPrivnetBip32Version.IsPubkeyVersion(rawVersionByte) {
		return "qx privnet"
	} else if QitmeerMixnetBip32Version.IsPrivkeyVersion(rawVersionByte) || QitmeerMixnetBip32Version.IsPubkeyVersion(rawVersionByte) {
		return "qx mixnet"
	} else if bip32.DefaultBip32Version.IsPrivkeyVersion(rawVersionByte) || bip32.DefaultBip32Version.IsPubkeyVersion(rawVersionByte) {
		return "btc mainnet"
	} else {
		return "unknown"
	}
}
