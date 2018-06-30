package hash

import "github.com/dchest/blake256"



func HashFunc(data []byte) [blake256.Size]byte {
	var outB [blake256.Size]byte
	a := blake256.New()
	a.Write(data)
	out := a.Sum(nil)
	for i, el := range out {
		outB[i] = el
	}

	return outB
}
// Blake256HashB using blake256 to calculates hash(b) and returns the resulting bytes.
func Blake256HashB(b []byte) []byte {
	a := blake256.New()
	a.Write(b)
	out := a.Sum(nil)
	return out
}

// Blake256HashH using blake256 to calculates hash(b) and returns the resulting bytes as a Hash.
func Blake256HashH(b []byte) Hash {
	var outB [blake256.Size]byte
	a := blake256.New()
	a.Write(b)
	out := a.Sum(nil)
	for i, el := range out {
		outB[i] = el
	}

	return Hash(outB)
}

// HashBlockSize is the block size of the hash algorithm in bytes.
const HashBlockSize = blake256.BlockSize
