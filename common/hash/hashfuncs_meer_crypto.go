// Copyright (c) 2017-2018 The qitmeer developers
package hash

import (
	qitmeerSha3 "github.com/Qitmeer/crypto/sha3"
)

// Meer Crypto calculates hash(b) and returns the resulting bytes as a Hash.
func HashMeerCrypto(input []byte) Hash {
	h := qitmeerSha3.NewLegacyKeccak512()
	// first round
	h.Write(input)
	input = h.Sum(nil)
	// second round
	h.Write(input)
	input = h.Sum(nil)
	// first byte ^1
	input[0] ^= 1
	// third round
	h1 := qitmeerSha3.NewQitmeerKeccak256()
	h1.Write(input)
	r := h.Sum(nil)
	hashR := [32]byte{}
	copy(hashR[:32], r[:32])
	return hashR
}
