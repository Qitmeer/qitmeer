package dcr

import (
	"github.com/noxproject/nox/common/hash"
	"github.com/dchest/blake256"
)


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
// HashB using blake256 to calculates hash(b) and returns the resulting bytes.
func HashB(b []byte) []byte {
	a := blake256.New()
	a.Write(b)
	out := a.Sum(nil)
	return out
}

// DoubleHashB calculates hash(hash(b)) and returns the resulting bytes.
func DoubleHashB(b []byte) (output [blake256.Size]byte) {
	h := blake256.New()
	h.Write(b)
	intermediateHash := h.Sum(nil)
	h.Reset()
	h.Write(intermediateHash)
	finalHash := h.Sum(nil)
	copy(output[:], finalHash[:])
	return
}

// HashH using blake256 to calculates hash(b) and returns the resulting bytes as a Hash.
func HashH(b []byte) hash.Hash {
	var outB [blake256.Size]byte
	a := blake256.New()
	a.Write(b)
	out := a.Sum(nil)
	for i, el := range out {
		outB[i] = el
	}

	return hash.Hash(outB)
}

func Hash160(buf []byte) []byte {
	return hash.CalcHash(HashB(buf), hash.GetHasher(hash.Ripemd160))
}
// HashBlockSize is the block size of the hash algorithm in bytes.
const HashBlockSize = blake256.BlockSize
