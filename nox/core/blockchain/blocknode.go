package blockchain

import (
	"math/big"
	"github.com/noxproject/nox/common/hash"
)

// blockStatus is a bit field representing the validation state of the block.
type blockStatus byte

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

	// parent is the parent block for this node.
	parent *blockNode

	// hash is the hash of the block this node represents.
	hash hash.Hash

	// workSum is the total amount of work in the chain up to and including
	// this node.
	workSum *big.Int

	// Some fields from block headers to aid in best chain selection and
	// reconstructing headers from memory.  These must be treated as
	// immutable and are intentionally ordered to avoid padding on 64-bit
	// platforms.
	height       int64
	finalState   [6]byte
	blockVersion int32
	poolSize     uint32
	bits         uint32
	timestamp    int64
	txRoot   	 hash.Hash
	stateRoot    hash.Hash
	blockSize    uint32
	nonce        uint32
	extraData    [32]byte

	// status is a bitfield representing the validation state of the block.
	// This field, unlike the other fields, may be changed after the block
	// node is created, so it must only be accessed or updated using the
	// concurrent-safe NodeStatus, SetStatusFlags, and UnsetStatusFlags
	// methods on blockIndex once the node has been added to the index.
	status blockStatus

	// inMainChain denotes whether the block node is currently on the
	// the main chain or not.  This is used to help find the common
	// ancestor when switching chains.
	inMainChain bool

}
