package vm

import (
	"github.com/Qitmeer/qitmeer/common/hash"
	"github.com/Qitmeer/qitmeer/core/types"
)

type ChainVM interface {
	Version() (string, error)

	GetBlock(*hash.Hash) (*types.Block, error)

	BuildBlock() (*types.Block, error)

	ParseBlock([]byte) (*types.Block, error)

	Shutdown() error
}
