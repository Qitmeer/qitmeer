// Copyright (c) 2017-2018 The qitmeer developers
package hash

import (
	qitmeerSha3 "github.com/Qitmeer/crypto/sha3"
)

// Meer X Keccak V1 calculates hash(b) and returns the resulting bytes as a Hash.
// 2 round of NewLegacyKeccak512
// 1 round of NewQitmeerKeccak256
func HashMeerXKeccakV1(input []byte) Hash {
	// input length 117 bytes
	h1 := qitmeerSha3.NewLegacyKeccak512()
	// first round
	h1.Write(input)
	// result length 64 bytes
	input = h1.Sum(nil)
	// second round
	h2 := qitmeerSha3.NewLegacyKeccak512()
	h2.Write(input)
	// result length 64 bytes
	input = h2.Sum(nil)
	// Attention
	// hash result first byte  ^1
	// specialized processing
	// Prevent compatibility of existing ASIC
	input[0] ^= 1
	// third round
	h3 := qitmeerSha3.NewQitmeerKeccak256()
	// result length 32 bytes
	h3.Write(input)
	r := h3.Sum(nil)
	hashR := [32]byte{}
	copy(hashR[:32], r[:32])
	return hashR
}
