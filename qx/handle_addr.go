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