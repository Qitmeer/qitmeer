package qx

import (
	"encoding/hex"
	"fmt"
	"github.com/Qitmeer/qitmeer/common/encode/base58"
	"github.com/Qitmeer/qitmeer/common/hash"
	"github.com/Qitmeer/qitmeer/params"
)

func EcPubKeyToAddress(version string, pubkey string) (string, error) {
	ver := []byte{}

	switch version {
	case "mainnet":
		ver = append(ver, params.MainNetParams.PubKeyHashAddrID[0:]...)
	case "privnet":
		ver = append(ver, params.PrivNetParams.PubKeyHashAddrID[0:]...)
	case "testnet":
		ver = append(ver, params.TestNetParams.PubKeyHashAddrID[0:]...)
	case "mixnet":
		ver = append(ver, params.MixNetParam.PubKeyHashAddrID[0:]...)
	default:
		v, err := hex.DecodeString(version)
		if err != nil {
			return "", err
		}
		ver = append(ver, v...)
	}

	data, err := hex.DecodeString(pubkey)
	if err != nil {
		return "", err
	}
	h := hash.Hash160(data)

	address := base58.QitmeerCheckEncode(h, ver[:])
	return address, nil
}

func EcScriptKeyToAddress(version string, pubkey string) (string, error) {
	ver := []byte{}

	switch version {
	case "mainnet":
		ver = append(ver, params.MainNetParams.ScriptHashAddrID[0:]...)
	case "privnet":
		ver = append(ver, params.PrivNetParams.ScriptHashAddrID[0:]...)
	case "testnet":
		ver = append(ver, params.TestNetParams.ScriptHashAddrID[0:]...)
	case "mixnet":
		ver = append(ver, params.MixNetParam.ScriptHashAddrID[0:]...)
	default:
		v, err := hex.DecodeString(version)
		if err != nil {
			return "", err
		}
		ver = append(ver, v...)
	}

	data, err := hex.DecodeString(pubkey)
	if err != nil {
		return "", err
	}
	h := hash.Hash160(data)

	address := base58.QitmeerCheckEncode(h, ver[:])
	return address, nil
}

func PKHEdwardsAddrIDToAddress(version string, pubkey string) (string, error) {
	ver := []byte{}

	switch version {
	case "mainnet":
		ver = append(ver, params.MainNetParams.PKHEdwardsAddrID[0:]...)
	case "privnet":
		ver = append(ver, params.PrivNetParams.PKHEdwardsAddrID[0:]...)
	case "testnet":
		ver = append(ver, params.TestNetParams.PKHEdwardsAddrID[0:]...)
	case "mixnet":
		ver = append(ver, params.MixNetParam.PKHEdwardsAddrID[0:]...)
	default:
		v, err := hex.DecodeString(version)
		if err != nil {
			return "", err
		}
		ver = append(ver, v...)
	}

	data, err := hex.DecodeString(pubkey)
	if err != nil {
		return "", err
	}
	h := hash.Hash160(data)

	address := base58.QitmeerCheckEncode(h, ver[:])
	return address, nil
}

func PKHSchnorrAddrIDToAddress(version string, pubkey string) (string, error) {
	ver := []byte{}

	switch version {
	case "mainnet":
		ver = append(ver, params.MainNetParams.PKHSchnorrAddrID[0:]...)
	case "privnet":
		ver = append(ver, params.PrivNetParams.PKHSchnorrAddrID[0:]...)
	case "testnet":
		ver = append(ver, params.TestNetParams.PKHSchnorrAddrID[0:]...)
	case "mixnet":
		ver = append(ver, params.MixNetParam.PKHSchnorrAddrID[0:]...)
	default:
		v, err := hex.DecodeString(version)
		if err != nil {
			return "", err
		}
		ver = append(ver, v...)
	}

	data, err := hex.DecodeString(pubkey)
	if err != nil {
		return "", err
	}
	h := hash.Hash160(data)

	address := base58.QitmeerCheckEncode(h, ver[:])
	return address, nil
}

func PrivateKeyIDToAddress(version string, pubkey string) (string, error) {
	ver := []byte{}

	switch version {
	case "mainnet":
		ver = append(ver, params.MainNetParams.PrivateKeyID[0:]...)
	case "privnet":
		ver = append(ver, params.PrivNetParams.PrivateKeyID[0:]...)
	case "testnet":
		ver = append(ver, params.TestNetParams.PrivateKeyID[0:]...)
	case "mixnet":
		ver = append(ver, params.MixNetParam.PrivateKeyID[0:]...)
	default:
		v, err := hex.DecodeString(version)
		if err != nil {
			return "", err
		}
		ver = append(ver, v...)
	}

	data, err := hex.DecodeString(pubkey)
	if err != nil {
		return "", err
	}
	h := hash.Hash160(data)

	address := base58.QitmeerCheckEncode(h, ver[:])
	return address, nil
}

func EcPublicKeyToAddress(version string, pubkey string) (string, error) {
	ver := []byte{}

	switch version {
	case "mainnet":
		ver = append(ver, params.MainNetParams.PubKeyAddrID[0:]...)
	case "privnet":
		ver = append(ver, params.PrivNetParams.PubKeyAddrID[0:]...)
	case "testnet":
		ver = append(ver, params.TestNetParams.PubKeyAddrID[0:]...)
	case "mixnet":
		ver = append(ver, params.MixNetParam.PubKeyAddrID[0:]...)
	default:
		v, err := hex.DecodeString(version)
		if err != nil {
			return "", err
		}
		ver = append(ver, v...)
	}

	data, err := hex.DecodeString(pubkey)
	if err != nil {
		return "", err
	}
	h := hash.Hash160(data)

	address := base58.QitmeerCheckEncode(h, ver[:])
	return address, nil
}

func EcPubKeyToAddressSTDO(version []byte, pubkey string) {
	data, err :=hex.DecodeString(pubkey)
	if err != nil {
		ErrExit(err)
	}
	h := hash.Hash160(data)

	address := base58.QitmeerCheckEncode(h, version[:])
	fmt.Printf("%s\n",address)
}