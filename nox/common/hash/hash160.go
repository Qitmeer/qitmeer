package hash

import (
	h "hash"
)

// Calculate the hash of hasher over buf.
func calcHash(buf []byte, hasher h.Hash) []byte {
	hasher.Write(buf)
	return hasher.Sum(nil)
}

// Hash160 calculates the hash ripemd160(hash256(b)).
func Hash160(buf []byte) []byte {
	return calcHash(DoubleHashB(buf), GetHasher(ripemd160))
}
