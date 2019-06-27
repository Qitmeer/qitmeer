// Copyright (c) 2017-2018 The nox developers
// Copyright (c) 2013-2017 The btcsuite developers
// Use of this source code is governed by an ISC
// license that can be found in the LICENSE file.

package btcaddr

import (
	"golang.org/x/crypto/ripemd160"
	"github.com/HalalChain/qitmeer-lib/common/encode/bech32"
	"github.com/HalalChain/qitmeer-lib/common/encode/base58"
	"github.com/HalalChain/qitmeer-lib/crypto/ecc"
)


// TstAddressPubKeyHash makes an AddressPubKeyHash, setting the
// unexported fields with the parameters hash and netID.
func TstAddressPubKeyHash(hash [ripemd160.Size]byte,
	netID byte) *AddressPubKeyHash {

	return &AddressPubKeyHash{
		hash:  hash,
		netID: netID,
	}
}

// TstAddressScriptHash makes an AddressScriptHash, setting the
// unexported fields with the parameters hash and netID.
func TstAddressScriptHash(hash [ripemd160.Size]byte,
	netID byte) *AddressScriptHash {

	return &AddressScriptHash{
		hash:  hash,
		netID: netID,
	}
}

// TstAddressWitnessPubKeyHash creates an AddressWitnessPubKeyHash, initiating
// the fields as given.
func TstAddressWitnessPubKeyHash(version byte, program [20]byte,
	hrp string) *AddressWitnessPubKeyHash {

	return &AddressWitnessPubKeyHash{
		hrp:            hrp,
		witnessVersion: version,
		witnessProgram: program,
	}
}

// TstAddressWitnessScriptHash creates an AddressWitnessScriptHash, initiating
// the fields as given.
func TstAddressWitnessScriptHash(version byte, program [32]byte,
	hrp string) *AddressWitnessScriptHash {

	return &AddressWitnessScriptHash{
		hrp:            hrp,
		witnessVersion: version,
		witnessProgram: program,
	}
}

// TstAddressPubKey makes an AddressPubKey, setting the unexported fields with
// the parameters.
func TstAddressPubKey(serializedPubKey []byte, pubKeyFormat PubKeyFormat,
	netID byte) *AddressPubKey {
	pubKey, _ := ecc.Secp256k1.ParsePubKey(serializedPubKey)
	return &AddressPubKey{
		pubKeyFormat: pubKeyFormat,
		pubKey:       pubKey,
		pubKeyHashID: netID,
	}
}

// TstAddressSAddr returns the expected script address bytes for
// P2PKH and P2SH bitcoin addresses.
func TstAddressSAddr(addr string) []byte {
	decoded := base58.Decode(addr)
	return decoded[1 : 1+ripemd160.Size]
}

// TstAddressSegwitSAddr returns the expected witness program bytes for
// bech32 encoded P2WPKH and P2WSH bitcoin addresses.
func TstAddressSegwitSAddr(addr string) []byte {
	_, data, err := bech32.DecodeBech32(addr)
	if err != nil {
		return []byte{}
	}

	// First byte is version, rest is base 32 encoded data.
	data, err = bech32.ConvertBits(data[1:], 5, 8, false)
	if err != nil {
		return []byte{}
	}
	return data
}
