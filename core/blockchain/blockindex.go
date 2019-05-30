// Copyright (c) 2017-2018 The nox developers
package blockchain

import (
	"sync"
	"qitmeer/params"
	"qitmeer/database"
	"qitmeer/common/hash"
	"qitmeer/core/types"
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
	ConnectBlock(tx database.Tx, block *types.SerializedBlock, utxoView *UtxoViewpoint) error

	// DisconnectBlock is invoked when a block has been disconnected from
	// the main chain.
	DisconnectBlock(tx database.Tx, block *types.SerializedBlock,  utxoView *UtxoViewpoint) error
}

// blockIndex provides facilities for keeping track of an in-memory index of the
// block chain.  Although the name block chain suggests a single chain of
// blocks, it is actually a tree-shaped structure where any node can have
// multiple children.  However, there can only be one active branch which does
// indeed form a chain from the tip all the way back to the genesis block.
type blockIndex struct {
	// The following fields are set when the instance is created and can't
	// be changed afterwards, so there is no need to protect them with a
	// separate mutex.
	db          database.DB
	params      *params.Params

	sync.RWMutex
	index     map[hash.Hash]*blockNode
}

// newBlockIndex returns a new empty instance of a block index.  The index will
// be dynamically populated as block nodes are loaded from the database and
// manually added.
func newBlockIndex(db database.DB, par *params.Params) *blockIndex {
	return &blockIndex{
		db:          db,
		params:      par,
		index:       make(map[hash.Hash]*blockNode),
	}
}

// lookupNode returns the block node identified by the provided hash.  It will
// return nil if there is no entry for the hash.
//
// This function MUST be called with the block index lock held (for reads).
func (bi *blockIndex) lookupNode(hash *hash.Hash) *blockNode {
	return bi.index[*hash]
}

// LookupNode returns the block node identified by the provided hash.  It will
// return nil if there is no entry for the hash.
//
// This function is safe for concurrent access.
func (bi *blockIndex) LookupNode(hash *hash.Hash) *blockNode {
	bi.RLock()
	node := bi.lookupNode(hash)
	bi.RUnlock()
	return node
}

// addNode adds the provided node to the block index.  Duplicate entries are not
// checked so it is up to caller to avoid adding them.
//
// This function MUST be called with the block index lock held (for writes).
func (bi *blockIndex) addNode(node *blockNode) {
	bi.index[node.hash] = node
	if node.parents!=nil&&len(node.parents)>0 {
		for _,v:=range node.parents{
			v.AddChild(node)
		}
	}
}
// AddNode adds the provided node to the block index.  Duplicate entries are not
// checked so it is up to caller to avoid adding them.
//
// This function is safe for concurrent access.
func (bi *blockIndex) AddNode(node *blockNode) {
	bi.Lock()
	bi.addNode(node)
	bi.Unlock()
}


// HaveBlock returns whether or not the block index contains the provided hash.
//
// This function is safe for concurrent access.
func (bi *blockIndex) HaveBlock(hash *hash.Hash) bool {
	bi.RLock()
	_, hasBlock := bi.index[*hash]
	bi.RUnlock()
	return hasBlock
}

// NodeStatus returns the status associated with the provided node.
//
// This function is safe for concurrent access.
func (bi *blockIndex) NodeStatus(node *blockNode) blockStatus {
	bi.RLock()
	status := node.status
	bi.RUnlock()
	return status
}

// SetStatusFlags sets the provided status flags for the given block node
// regardless of their previous state.  It does not unset any flags.
//
// This function is safe for concurrent access.
func (bi *blockIndex) SetStatusFlags(node *blockNode, flags blockStatus) {
	bi.Lock()
	node.status |= flags
	bi.Unlock()
}

// UnsetStatusFlags unsets the provided status flags for the given block node
// regardless of their previous state.
//
// This function is safe for concurrent access.
func (bi *blockIndex) UnsetStatusFlags(node *blockNode, flags blockStatus) {
	bi.Lock()
	node.status &^= flags
	bi.Unlock()
}

// This function can get backward block hash from list.
func (bi *blockIndex)GetMaxOrderFromList(list []*hash.Hash) *hash.Hash{
	var maxOrder uint64=0
	var maxHash *hash.Hash=nil
	for _,v:=range list{
		node:=bi.LookupNode(v)
		if node==nil {
			continue
		}
		if maxOrder==0||maxOrder<node.order {
			maxOrder=node.order
			maxHash=v
		}
	}
	return maxHash
}
