package hash

import "golang.org/x/crypto/blake2b"

// BlakeHashB using blake2b calculates hash(b) and returns the resulting bytes.
func BlakeHashB(b []byte) []byte {
	hash := blake2b.Sum256(b)
	return hash[:]
}

// HashH calculates hash(b) and returns the resulting bytes as a Hash.
func BlakeHashH(b []byte) Hash {
	return Hash(blake2b.Sum256(b))
}

// DoubleHashB calculates hash(hash(b)) and returns the resulting bytes.
func BlakeDoubleHashB(b []byte) []byte {
	first := blake2b.Sum256(b)
	second := blake2b.Sum256(first[:])
	return second[:]
}

// DoubleHashH calculates hash(hash(b)) and returns the resulting bytes as a
// Hash.
func BlakeDoubleHashH(b []byte) Hash {
	first := blake2b.Sum256(b)
	return Hash(blake2b.Sum256(first[:]))
}

