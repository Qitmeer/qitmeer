// Copyright (c) 2017-2018 The qitmeer developers
package hash

import (
	qitmeerSha3 "github.com/Qitmeer/crypto/sha3"
)

// Meer Crypto calculates hash(b) and returns the resulting bytes as a Hash.
// 2 round of NewLegacyKeccak512
// 1 round of NewQitmeerKeccak256
func HashMeerCrypto(input []byte) Hash {
	// input length 117 bytes
	h := qitmeerSha3.NewLegacyKeccak512()
	// first round
	h.Write(input)
	// result length 64 bytes
	input = h.Sum(nil)
	// second round
	h.Write(input)
	// result length 64 bytes
	input = h.Sum(nil)
	// Attention
	// hash result first byte  ^1
	// specialized processing
	// Prevent compatibility of existing ASIC
	input[0] ^= 1
	// third round
	h1 := qitmeerSha3.NewQitmeerKeccak256()
	// result length 32 bytes
	h1.Write(input)
	r := h.Sum(nil)
	hashR := [32]byte{}
	copy(hashR[:32], r[:32])
	return hashR
}
