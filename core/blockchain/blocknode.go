// Copyright (c) 2017-2018 The nox developers
package blockchain

import (
	"github.com/HalalChain/qitmeer-lib/common/hash"
	"github.com/HalalChain/qitmeer-lib/common/util"
	"github.com/HalalChain/qitmeer-lib/core/dag"
	"github.com/HalalChain/qitmeer-lib/crypto/cuckoo"
	"github.com/HalalChain/qitmeer/core/merkle"
	"github.com/HalalChain/qitmeer-lib/core/types"
	"math/big"
	"sort"
	"time"
)

// blockStatus is a bit field representing the validation state of the block.
type blockStatus byte

// The following constants specify possible status bit flags for a block.
//
// NOTE: This section specifically does not use iota since the block status is
// serialized and must be stable for long-term storage.
const (
	// statusNone indicates that the block has no validation state flags set.
	statusNone blockStatus = 0

	// statusDataStored indicates that the block's payload is stored on disk.
	statusDataStored blockStatus = 1 << 0

	// statusValid indicates that the block has been fully validated.
	statusValid blockStatus = 1 << 1

	// statusValidateFailed indicates that the block has failed validation.
	statusValidateFailed blockStatus = 1 << 2

	// statusInvalidAncestor indicates that one of the ancestors of the block
	// has failed validation, thus the block is also invalid.
	statusInvalidAncestor = 1 << 3
)

// HaveData returns whether the full block data is stored in the database.  This
// will return false for a block node where only the header is downloaded or
// stored.
func (status blockStatus) HaveData() bool {
	return status&statusDataStored != 0
}

// KnownValid returns whether the block is known to be valid.  This will return
// false for a valid block that has not been fully validated yet.
func (status blockStatus) KnownValid() bool {
	return status&statusValid != 0
}

// KnownInvalid returns whether the block is known to be invalid.  This will
// return false for invalid blocks that have not been proven invalid yet.
func (status blockStatus) KnownInvalid() bool {
	return status&(statusValidateFailed|statusInvalidAncestor) != 0
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
	// children is all the children block for this node.
	children []*blockNode
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
	txRoot   	 hash.Hash
	stateRoot    hash.Hash
	nonce        uint64
	exNonce      uint64
	extraData    [32]byte

	// status is a bitfield representing the validation state of the block.
	// This field, unlike the other fields, may be changed after the block
	// node is created, so it must only be accessed or updated using the
	// concurrent-safe NodeStatus, SetStatusFlags, and UnsetStatusFlags
	// methods on blockIndex once the node has been added to the index.
	status blockStatus

	//order is in the position of whole block chain.(It is actually DAG order)
	order    uint64
	circleNonce [cuckoo.ProofSize]uint32
}

// newBlockNode returns a new block node for the given block header and parent
// node.  The workSum is calculated based on the parent, or, in the case no
// parent is provided, it will just be the work for the passed block.
func newBlockNode(blockHeader *types.BlockHeader, parents []*blockNode) *blockNode {
	var node blockNode
	initBlockNode(&node, blockHeader, parents)
	return &node
}

// initBlockNode initializes a block node from the given header, initialization
// vector for the ticket lottery, and parent node.  The workSum is calculated
// based on the parent, or, in the case no parent is provided, it will just be
// the work for the passed block.
//
// This function is NOT safe for concurrent access.  It must only be called when
// initially creating a node.
func initBlockNode(node *blockNode, blockHeader *types.BlockHeader, parents []*blockNode) {
	*node = blockNode{
		hash:         blockHeader.BlockHash(),
		workSum:      CalcWork(blockHeader.Difficulty),
		order:       0,
		blockVersion: blockHeader.Version,
		bits:         blockHeader.Difficulty,
		timestamp:    blockHeader.Timestamp.Unix(),
		txRoot:       blockHeader.TxRoot,
		nonce:        blockHeader.Nonce,
		exNonce:      blockHeader.ExNonce,
		stateRoot:    blockHeader.StateRoot,
		status:       statusNone,
		circleNonce:  blockHeader.CircleNonces,
	}
	if len(parents)>0 {
		node.parents = parents

		node.workSum = node.workSum.Add(node.GetParentsWorkSum(), node.workSum)
	}else {
		node.order=0
	}
}

// Header constructs a block header from the node and returns it.
//
// This function is safe for concurrent access.
func (node *blockNode) Header() types.BlockHeader {
	// No lock is needed because all accessed fields are immutable.
	var parentRoot hash.Hash
	if node.parents!=nil {
		paMerkles :=merkle.BuildParentsMerkleTreeStore(node.GetParents())
		parentRoot=*paMerkles[len(paMerkles)-1]
	}
	return types.BlockHeader{
		Version:    node.blockVersion,
		ParentRoot: parentRoot,
		TxRoot:   	node.txRoot,
		StateRoot:  node.stateRoot,
		Difficulty: node.bits,
		ExNonce:    node.exNonce,
		Timestamp:  time.Unix(node.timestamp, 0),
		Nonce:      node.nonce,
		CircleNonces:      node.circleNonce,
	}
}

// CalcPastMedianTime calculates the median time of the previous few blocks
// prior to, and including, the block node.
//
// This function is safe for concurrent access.
func (node *blockNode) CalcPastMedianTime() time.Time {
	// Create a slice of the previous few block timestamps used to calculate
	// the median per the number defined by the constant medianTimeBlocks.
	timestamps := make([]int64, medianTimeBlocks)
	numNodes := 0
	iterNode := node
	for i := 0; i < medianTimeBlocks && iterNode != nil; i++ {
		timestamps[i] = iterNode.timestamp
		numNodes++

		iterNode = iterNode.GetForwardParent()
	}

	// Prune the slice to the actual number of available timestamps which
	// will be fewer than desired near the beginning of the block chain
	// and sort them.
	timestamps = timestamps[:numNodes]
	sort.Sort(util.TimeSorter(timestamps))

	// NOTE: The consensus rules incorrectly calculate the median for even
	// numbers of blocks.  A true median averages the middle two elements
	// for a set with an even number of elements in it.   Since the constant
	// for the previous number of blocks to be used is odd, this is only an
	// issue for a few blocks near the beginning of the chain.  I suspect
	// this is an optimization even though the result is slightly wrong for
	// a few of the first blocks since after the first few blocks, there
	// will always be an odd number of blocks in the set per the constant.
	//
	// This code follows suit to ensure the same rules are used, however, be
	// aware that should the medianTimeBlocks constant ever be changed to an
	// even number, this code will be wrong.
	medianTimestamp := timestamps[numNodes/2]
	return time.Unix(medianTimestamp, 0)
}

func (node *blockNode) GetParentsWorkSum() *big.Int {
	var workSum *big.Int=&big.Int{}
	for _,v:=range node.parents{
		workSum.Add(v.workSum, workSum)
	}
	return workSum
}

// Include all parents for set
func (node *blockNode) GetParents() []*hash.Hash{
	if node.parents==nil||len(node.parents)==0 {
		return nil
	}
	result:=[]*hash.Hash{}
	for _,v:=range node.parents{
		result=append(result,&v.hash)
	}
	return result
}

// node has children in DAG
func (node *blockNode) AddChild(child *blockNode){
	if node.HasChild(child) {
		return
	}
	if node.children==nil {
		node.children=[]*blockNode{}
	}
	node.children=append(node.children,child)
}

// check is there any child
func (node *blockNode) HasChild(child *blockNode) bool{
	if node.children==nil||len(node.children)==0 {
		return false
	}
	for _,v:=range node.children{
		if v==child {
			return true
		}
	}
	return false
}

// For the moment,In order to match the DAG
func (node *blockNode) GetChildren() *dag.HashSet{
	if node.children==nil||len(node.children)==0 {
		return nil
	}
	result:=dag.NewHashSet()
	for _,v:=range node.children{
		result.Add(&v.hash)
	}
	return result
}

func (node *blockNode) SetOrder(h uint64) {
	node.order=h
}

// return node height (Actually,it is order)
func (node *blockNode) GetOrder() uint64{
	return node.order
}

func (node *blockNode) Clone() *blockNode{
	header:=node.Header()
	newNode := newBlockNode(&header,node.parents)
	newNode.status = node.status
	newNode.children=node.children
	newNode.order=node.order
	return newNode
}

//return parent that position is rather forward
func (node *blockNode) GetForwardParent() *blockNode {
	if node.parents==nil || len(node.parents)<=0 {
		return nil
	}
	var result *blockNode=nil
	for _,v:=range node.parents{
		if result==nil || v.GetOrder()<result.GetOrder(){
			result=v
		}
	}
	return result
}

//return parent that position is rather back
func (node *blockNode) GetBackParent() *blockNode {
	if node.parents==nil || len(node.parents)<=0 {
		return nil
	}
	var result *blockNode=nil
	for _,v:=range node.parents{
		if result==nil || v.GetOrder()>result.GetOrder(){
			result=v
		}
	}
	return result
}

//return the block node hash.
func (node *blockNode) GetHash() *hash.Hash {
	return &node.hash
}

//return the timestamp of node
func (node *blockNode) GetTimestamp() int64 {
	return node.timestamp
}
