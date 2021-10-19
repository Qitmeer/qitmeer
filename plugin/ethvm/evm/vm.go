package evm

import (
	"github.com/Qitmeer/qitmeer/common/hash"
	"github.com/Qitmeer/qitmeer/core/types"
)

type VM struct {
}

func (vm *VM) Version() (string, error) {
	return "", nil
}

func (vm *VM) GetBlock(*hash.Hash) (*types.Block, error) {
	return nil, nil
}

func (vm *VM) BuildBlock() (*types.Block, error) {
	return nil, nil
}

func (vm *VM) ParseBlock([]byte) (*types.Block, error) {
	return nil, nil
}

func (vm *VM) Shutdown() error {

	return nil
}
