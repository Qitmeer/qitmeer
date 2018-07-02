// Copyright (c) 2017-2018 The nox developers
// Copyright (c) 2015 The Decred developers
// Copyright (c) 2016-2017 The btcsuite developers
// Use of this source code is governed by an ISC
// license that can be found in the LICENSE file.

package btc

import (
	"crypto/sha256"
	"github.com/noxproject/nox/common/hash"
)

// HashB calculates hash(b) and returns the resulting bytes.
func HashB(b []byte) []byte {
	hash := sha256.Sum256(b)
	return hash[:]
}

// HashH calculates hash(b) and returns the resulting bytes as a Hash.
func HashH(b []byte) hash.Hash {
	return hash.Hash(sha256.Sum256(b))
}

// DoubleHashB calculates hash(hash(b)) and returns the resulting bytes.
func DoubleHashB(b []byte) []byte {
	first := sha256.Sum256(b)
	second := sha256.Sum256(first[:])
	return second[:]
}

// DoubleHashH calculates hash(hash(b)) and returns the resulting bytes as a
// Hash.
func DoubleHashH(b []byte) hash.Hash {
	first := sha256.Sum256(b)
	return hash.Hash(sha256.Sum256(first[:]))
}

func Hash160(buf []byte) []byte {
	return hash.CalcHash(HashB(buf), hash.GetHasher(hash.Ripemd160))
	//return hash.CalcHash(DoubleHashB(buf), hash.GetHasher(hash.Ripemd160))
}
