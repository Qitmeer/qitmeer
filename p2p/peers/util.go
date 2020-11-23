package peers

import (
	"github.com/Qitmeer/qitmeer/core/protocol"
	"github.com/prysmaticlabs/go-bitfield"
)

func retrieveIndicesFromBitfield(bitV bitfield.Bitvector64) []uint64 {
	committeeIdxs := []uint64{}
	for i := uint64(0); i < 64; i++ {
		if bitV.BitAt(i) {
			committeeIdxs = append(committeeIdxs, i)
		}
	}
	return committeeIdxs
}

func HasConsensusService(services protocol.ServiceFlag) bool {
	if protocol.HasServices(protocol.ServiceFlag(services), protocol.Full) ||
		protocol.HasServices(protocol.ServiceFlag(services), protocol.Light) {
		return true
	}
	return false
}
