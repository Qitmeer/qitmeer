// Copyright (c) 2017-2018 The qitmeer developers
package blockchain

import (
	"github.com/Qitmeer/qng-core/common/hash"
	"github.com/Qitmeer/qng-core/meerdag"
	"github.com/Qitmeer/qng-core/core/types"
	"github.com/Qitmeer/qng-core/database"
)

// IndexManager provides a generic interface that the is called when blocks are
// connected and disconnected to and from the tip of the main chain for the
// purpose of supporting optional indexes.
type IndexManager interface {
	// Init is invoked during chain initialize in order to allow the index
	// manager to initialize itself and any indexes it is managing.  The
	// channel parameter specifies a channel the caller can close to signal
	// that the process should be interrupted.  It can be nil if that
	// behavior is not desired.
	Init(*BlockChain, <-chan struct{}) error

	// ConnectBlock is invoked when a new block has been connected to the
	// main chain.
	ConnectBlock(tx database.Tx, block *types.SerializedBlock, stxos []SpentTxOut) error

	// DisconnectBlock is invoked when a block has been disconnected from
	// the main chain.
	DisconnectBlock(tx database.Tx, block *types.SerializedBlock, stxos []SpentTxOut) error

	// IsDuplicateTx
	IsDuplicateTx(tx database.Tx, txid *hash.Hash, blockHash *hash.Hash) bool
}

// LookupNode returns the block node identified by the provided hash.  It will
// return nil if there is no entry for the hash.
func (b *BlockChain) LookupNode(hash *hash.Hash) *BlockNode {
	ib := b.GetBlock(hash)
	if ib == nil {
		return nil
	}
	if ib.GetData() == nil {
		return nil
	}
	return ib.GetData().(*BlockNode)
}

func (b *BlockChain) LookupNodeById(id uint) *BlockNode {
	ib := b.bd.GetBlockById(id)
	if ib == nil {
		return nil
	}
	if ib.GetData() == nil {
		return nil
	}
	return ib.GetData().(*BlockNode)
}

func (b *BlockChain) GetBlockNode(ib meerdag.IBlock) *BlockNode {
	if ib == nil {
		return nil
	}
	if ib.GetData() == nil {
		return nil
	}
	return ib.GetData().(*BlockNode)
}

func (b *BlockChain) GetBlock(h *hash.Hash) meerdag.IBlock {
	return b.bd.GetBlock(h)
}
