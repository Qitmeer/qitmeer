// Copyright (c) 2017-2018 The qitmeer developers
package hash

import (
	qitmeerSha3 "github.com/Qitmeer/crypto/sha3"
)

// Qitmeer Keccak256 calculates hash(b) and returns the resulting bytes as a Hash.
func HashQitmeerKeccak256(b []byte) Hash {
	h := qitmeerSha3.NewQitmeerKeccak256()
	h.Write(b)
	r := h.Sum(nil)
	hashR := [32]byte{}
	copy(hashR[:32], r[:32])
	return Hash(hashR)
}
