package hash

import (
	h "hash"
)

// Calculate the hash of hasher over buf.
func CalcHash(buf []byte, hasher h.Hash) []byte {
	defer hasher.Reset()
	hasher.Write(buf)
	return hasher.Sum(nil)
}

// Hash160 calculates the hash ripemd160(hash256(b)).
func Hash160(buf []byte) []byte {
	return CalcHash(HashB(buf), GetHasher(Ripemd160))
}


