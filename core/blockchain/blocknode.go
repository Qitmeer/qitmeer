// Copyright (c) 2017-2018 The qitmeer developers
package blockchain

import (
	"github.com/Qitmeer/qitmeer/common/hash"
	"github.com/Qitmeer/qitmeer/core/blockdag"
	"github.com/Qitmeer/qitmeer/core/types"
	"github.com/Qitmeer/qitmeer/core/types/pow"
	"github.com/Qitmeer/qitmeer/database"
	"math/big"
	"time"
)

// BlockStatus is a bit field representing the validation state of the block.
type BlockStatus byte

// The following constants specify possible status bit flags for a block.
//
// NOTE: This section specifically does not use iota since the block status is
// serialized and must be stable for long-term storage.
const (
	// statusNone indicates that the block has no validation state flags set.
	statusNone BlockStatus = 0

	// statusDataStored indicates that the block's payload is stored on disk.
	statusDataStored BlockStatus = 1 << 0

	// statusValid indicates that the block has been fully validated.
	statusValid BlockStatus = 1 << 1

	// statusInvalid indicates that the block has failed validation.
	statusInvalid BlockStatus = 1 << 2
)

// KnownInvalid returns whether the block is known to be invalid.  This will
// return false for invalid blocks that have not been proven invalid yet.
func (status BlockStatus) KnownInvalid() bool {
	return status&statusInvalid != 0
}

// blockNode represents a block within the block chain and is primarily used to
// aid in selecting the best chain to be the main chain.  The main chain is
// stored into the block database.
type blockNode struct {
	// NOTE: Additions, deletions, or modifications to the order of the
	// definitions in this struct should not be changed without considering
	// how it affects alignment on 64-bit platforms.  The current order is
	// specifically crafted to result in minimal padding.  There will be
	// hundreds of thousands of these in memory, so a few extra bytes of
	// padding adds up.

	// parents is all the parents block for this node.
	parents []*blockNode
	// hash is the hash of the block this node represents.
	hash hash.Hash

	// workSum is the total amount of work in the chain up to and including
	// this node.
	workSum *big.Int

	// Some fields from block headers to aid in best chain selection and
	// reconstructing headers from memory.  These must be treated as
	// immutable and are intentionally ordered to avoid padding on 64-bit
	// platforms.
	blockVersion uint32
	bits         uint32
	timestamp    int64
	txRoot       hash.Hash
	stateRoot    hash.Hash
	extraData    [32]byte

	// status is a bitfield representing the validation state of the block.
	// This field, unlike the other fields, may be changed after the block
	// node is created, so it must only be accessed or updated using the
	// concurrent-safe NodeStatus, SetStatusFlags, and UnsetStatusFlags
	// methods on blockIndex once the node has been added to the index.
	status BlockStatus

	// order is in the position of whole block chain.(It is actually DAG order)
	order uint64

	// height
	height uint

	// layer
	layer uint

	// pow
	pow pow.IPow

	// dirty
	dirty bool

	// dag block id
	dagID uint
}

func (node *blockNode) Valid(b *BlockChain) {
	node.SetStatusFlags(statusValid)
	node.UnsetStatusFlags(statusInvalid)
	node.FlushToDB(b)
}

func (node *blockNode) Invalid(b *BlockChain) {
	node.SetStatusFlags(statusInvalid)
	node.UnsetStatusFlags(statusValid)
	node.FlushToDB(b)
}

func (node *blockNode) FlushToDB(b *BlockChain) error {
	if !node.dirty {
		return nil
	}
	block := b.bd.GetBlockById(node.dagID)
	block.SetStatus(blockdag.BlockStatus(node.status))

	err := b.db.Update(func(dbTx database.Tx) error {
		return blockdag.DBPutDAGBlock(dbTx, block)
	})
	// If write was successful, clear the dirty set.
	if err == nil {
		node.dirty = false
	}
	return b.bd.Commit()
}

type BlockNode struct {
	// hash is the hash of the block this node represents.
	hash hash.Hash

	parents []hash.Hash

	header types.BlockHeader
}

//return the block node hash.
func (node *BlockNode) GetHash() *hash.Hash {
	return &node.hash
}

// Include all parents for set
func (node *BlockNode) GetParents() []*hash.Hash {
	parents := []*hash.Hash{}
	for _, p := range node.parents {
		parents = append(parents, &p)
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

func NewBlockNode(header *types.BlockHeader, parents []*hash.Hash) *BlockNode {
	bn := BlockNode{
		hash:    header.BlockHash(),
		header:  *header,
		parents: []hash.Hash{},
	}
	for _, p := range parents {
		bn.parents = append(bn.parents, *p)
	}
	return &bn
}
