// Copyright 2017-2018 The nox developers

package types

type Address string

func (a Address) EncodeAddress() string {
	return "bar"
}

type pubKeyHashAddress struct{
	config *Config
	hash  Hash160          // 160bits pubKey hash
}

func (b pubKeyHashAddress) String() string {
	return "foo"
}

