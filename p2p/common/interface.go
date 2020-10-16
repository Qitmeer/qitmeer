package common

import "github.com/Qitmeer/qitmeer/common/hash"

type P2P interface {
	GetGenesisHash() *hash.Hash
}
