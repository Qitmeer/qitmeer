package qx

import (
	"encoding/hex"
	"github.com/HalalChain/qitmeer-lib/common/encode/base58"
	"github.com/HalalChain/qitmeer-lib/common/hash"
	"github.com/HalalChain/qitmeer-lib/params"
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

	address := base58.NoxCheckEncode(h, ver[:])
	return address, nil
}
