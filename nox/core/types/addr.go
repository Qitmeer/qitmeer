// Copyright 2017-2018 The nox developers

package types

import "fmt"

type Address interface{
	Encode() (string, error)
	getHash() []byte
}

type AddressType byte

type pubKeyHashAddress struct{
	pkhash  Hash         // pubKey hash
	addrType AddressType
}

type pubKeyAddress struct{
	pk      []byte
	addrType AddressType
}

func (a pubKeyHashAddress) Encode() (string, error) {
	// TODO encode pkhash
	return "", fmt.Errorf("unsupport encode for %v",a.addrType)
}


