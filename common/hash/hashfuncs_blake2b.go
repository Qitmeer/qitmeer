// Copyright (c) 2017-2018 The qitmeer developers
package hash

import (
	"github.com/Qitmeer/qitmeer/crypto/x16rv3"
	"golang.org/x/crypto/blake2b"
)

// HashB using blake2b calculates 256 bits hash and returns the resulting bytes.
func HashB(b []byte) []byte {
	hash := blake2b.Sum256(b)
	return hash[:]
}

// Hash512B using blake2b calculates 512 bits hash and returns the resulting bytes.
func Hash512B(b []byte) []byte {
	hash := blake2b.Sum512(b)
	return hash[:]
}

// HashH calculates hash(b) and returns the resulting bytes as a Hash.
func HashH(b []byte) Hash {
	return Hash(blake2b.Sum256(b))
}

// HashX16rv3 calculates hash(b) and returns the resulting bytes as a Hash.
func HashX16rv3(b []byte) Hash {
	return Hash(x16rv3.Sum256(b))
}

// HashX8r16 calculates hash(b) and returns the resulting bytes as a Hash.
func HashX8r16(b []byte) Hash {
	return Hash(x16rv3.Sum256(b))
}

// DoubleHashB calculates hash(hash(b)) and returns the resulting bytes.
func DoubleHashB(b []byte) []byte {
	first := blake2b.Sum256(b)
	second := blake2b.Sum256(first[:])
	return second[:]
}

// DoubleHashH calculates hash(hash(b)) and returns the resulting bytes as a
// Hash.
func DoubleHashH(b []byte) Hash {
	first := blake2b.Sum256(b)
	return Hash(blake2b.Sum256(first[:]))
}
