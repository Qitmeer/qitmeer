// Copyright (c) 2017-2018 The qitmeer developers
package blockchain

import (
	"github.com/Qitmeer/qng-core/common/hash"
	"github.com/Qitmeer/qng-core/core/types"
	"github.com/Qitmeer/qng-core/core/types/pow"
	"time"
)

type BlockNode struct {
	// hash is the hash of the block this node represents.
	hash hash.Hash

	parents []hash.Hash

	header types.BlockHeader

	txNum int
}

//return the block node hash.
func (node *BlockNode) GetHash() *hash.Hash {
	return &node.hash
}

// Include all parents for set
func (node *BlockNode) GetParents() []*hash.Hash {
	parents := []*hash.Hash{}
	for _, p := range node.parents {
		pa := p
		parents = append(parents, &pa)
	}
	return parents
}

//return the timestamp of node
func (node *BlockNode) GetTimestamp() int64 {
	return node.header.Timestamp.Unix()
}

func (node *BlockNode) GetHeader() *types.BlockHeader {
	return &node.header
}

func (node *BlockNode) Difficulty() uint32 {
	return node.GetHeader().Difficulty
}

func (node *BlockNode) Pow() pow.IPow {
	return node.GetHeader().Pow
}

func (node *BlockNode) GetPowType() pow.PowType {
	return node.Pow().GetPowType()
}

func (node *BlockNode) Timestamp() time.Time {
	return node.GetHeader().Timestamp
}

func (node *BlockNode) GetPriority() int {
	return node.txNum
}

func NewBlockNode(block *types.SerializedBlock, parents []*hash.Hash) *BlockNode {
	header := &block.Block().Header
	bn := BlockNode{
		hash:    header.BlockHash(),
		header:  *header,
		parents: []hash.Hash{},
		txNum:   len(block.Transactions()),
	}
	for _, p := range parents {
		pa := *p
		bn.parents = append(bn.parents, pa)
	}
	return &bn
}
