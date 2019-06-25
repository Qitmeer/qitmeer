package blockdag

import "github.com/HalalChain/qitmeer-lib/common/hash"

//The abstract inferface is used to dag block
type IBlockData interface {
	// Get hash of block
	GetHash() *hash.Hash

	// Get all parents set,the dag block has more than one parent
	GetParents() []*hash.Hash

	// Timestamp
	GetTimestamp() int64
}

//The interface of block
type IBlock interface {
	// Return the hash of block. It will be a pointer.
	GetHash() *hash.Hash

	// Acquire the layer of block
	GetLayer() uint

	// Setting the order of block
	SetOrder(o uint)

	// Acquire the order of block
	GetOrder() uint

	// Get all parents set,the dag block has more than one parent
	GetParents() *HashSet

	// Testing whether it has parents
	HasParents() bool

	// Add child nodes to block
	AddChild(child *Block)

    // Get all the children of block
    GetChildren() *HashSet

	// Detecting the presence of child nodes
	HasChildren() bool

	GetMainParent() *hash.Hash

	// Setting the weight of block
	SetWeight(weight uint)

    // Acquire the weight of block
    GetWeight() uint

	// Acquire the height of block in main chain
	GetHeight() uint
}

// It is the element of a DAG. It is the most basic data unit.
type Block struct {
	hash     hash.Hash
	parents  *HashSet
	children *HashSet

	mainParent *hash.Hash
	weight uint
	order  uint
	layer  uint
	height uint
}

// Return the hash of block. It will be a pointer.
func (b *Block) GetHash() *hash.Hash {
	return &b.hash
}

// Get all parents set,the dag block has more than one parent
func (b *Block) GetParents() *HashSet {
	return b.parents
}

func (b *Block) GetMainParent() *hash.Hash {
	return b.mainParent
}

// Testing whether it has parents
func (b *Block) HasParents() bool {
	if b.parents == nil {
		return false
	}
	if b.parents.IsEmpty() {
		return false
	}
	return true
}

// Parent with order in front.
func (b *Block) GetForwardParent() *Block {
	if b.parents==nil || b.parents.IsEmpty() {
		return nil
	}
	var result *Block=nil
	for _,v:=range b.parents.GetMap(){
		parent:=v.(*Block)
		if result==nil || parent.GetOrder()<result.GetOrder(){
			result=parent
		}
	}
	return result
}

// Parent with order in back.
func (b *Block) GetBackParent() *Block {
	if b==nil || b.parents==nil || b.parents.IsEmpty() {
		return nil
	}
	var result *Block=nil
	for _,v:=range b.parents.GetMap(){
		parent:=v.(*Block)
		if result==nil || parent.GetOrder()>result.GetOrder(){
			result=parent
		}
	}
	return result
}

// Add child nodes to block
func (b *Block) AddChild(child *Block) {
	if b.children == nil {
		b.children = NewHashSet()
	}
	b.children.AddPair(child.GetHash(),child)
}

// Get all the children of block
func (b *Block) GetChildren() *HashSet {
	return b.children
}

// Detecting the presence of child nodes
func (b *Block) HasChildren() bool {
	if b.children == nil {
		return false
	}
	if b.children.IsEmpty() {
		return false
	}
	return true
}

// Setting the weight of block
func (b *Block) SetWeight(weight uint) {
	b.weight = weight
}

// Acquire the weight of block
func (b *Block) GetWeight() uint {
	return b.weight
}

// Setting the layer of block
func (b *Block) SetLayer(layer uint) {
	b.layer=layer
}

// Acquire the layer of block
func (b *Block) GetLayer() uint {
	return b.layer
}

// Setting the order of block
func (b *Block) SetOrder(o uint) {
	b.order=o
}

// Acquire the order of block
func (b *Block) GetOrder() uint {
	return b.order
}

// Setting the height of block in main chain
func (b *Block) SetHeight(h uint) {
	b.height=h
}

// Acquire the height of block in main chain
func (b *Block) GetHeight() uint {
	return b.height
}