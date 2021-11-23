package peers

import (
	"fmt"
	"github.com/Qitmeer/qng-core/core/protocol"
	"github.com/prysmaticlabs/go-bitfield"
	"strings"
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

func ParseUserAgent(data string) (error, string, string, string) {
	name := ""
	version := ""
	network := ""
	if len(data) <= 0 {
		return fmt.Errorf("UserAgent is invalid"), name, version, network
	}
	formatArr := strings.Split(data, "|")
	if len(formatArr) <= 0 {
		return fmt.Errorf("UserAgent is invalid"), name, version, network
	}

	if len(formatArr) >= 1 {
		name = formatArr[0]
	}
	if len(formatArr) >= 2 {
		version = formatArr[1]
	}
	if len(formatArr) >= 3 {
		network = formatArr[2]
	}
	return nil, name, version, network
}
