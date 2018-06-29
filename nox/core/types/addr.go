// Copyright 2017-2018 The nox developers
package types

import (
	"golang.org/x/crypto/ripemd160"
	"github.com/noxproject/nox/crypto"
)

type Address interface{
	Encode()        string
	Hash160()       *[ripemd160.Size]byte
	SignScheme()    crypto.SignatureScheme
	ScriptAddress() []byte
}

type AddressType byte

const (
	LegerAddress AddressType = 0x01
	ContractAddress AddressType = 0x02
)


