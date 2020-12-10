// Copyright 2017-2018 The qitmeer developers
package types

import (
	"github.com/Qitmeer/qitmeer/crypto/ecc"
)

type Address interface {
	// String returns the string encoding of the transaction output
	// destination.
	//
	// Please note that String differs subtly from EncodeAddress: String
	// will return the value as a string without any conversion, while
	// EncodeAddress may convert destination types (for example,
	// converting pubkeys to P2PKH addresses) before encoding as a
	// payment address string.
	String() string

	// with encode
	Encode() string

	// Hash160 returns the Hash160(data) where data is the data normally
	// hashed to 160 bits from the respective address type.
	Hash160() *[20]byte

	EcType() ecc.EcType

	// raw byte in script, aka the hash in the most case
	Script() []byte

	// TODO, revisit the design of address type decision
	// IsForNetwork returns whether or not the address is associated with the
	// passed network.
	// IsForNetwork(hashID [2]byte) bool
}

type AddressType byte

const (
	LegerAddress    AddressType = 0x01
	ContractAddress AddressType = 0x02
)
