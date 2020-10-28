// Copyright (c) 2017-2018 The qitmeer developers
package hash

import (
	"github.com/Qitmeer/crypto/cryptonight"
)

// CryptoNight calculates hash(b) and returns the resulting bytes as a Hash.
func HashCryptoNight(b []byte) Hash {
	h := cryptonight.Sum(b, 2)
	hashR := [32]byte{}
	copy(hashR[:32], h[:32])
	return Hash(hashR)
}
