package peers

import "github.com/prysmaticlabs/go-bitfield"

func retrieveIndicesFromBitfield(bitV bitfield.Bitvector64) []uint64 {
	committeeIdxs := []uint64{}
	for i := uint64(0); i < 64; i++ {
		if bitV.BitAt(i) {
			committeeIdxs = append(committeeIdxs, i)
		}
	}
	return committeeIdxs
}
