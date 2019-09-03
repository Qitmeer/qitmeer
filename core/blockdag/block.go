package blockdag

import (
	"github.com/Qitmeer/qitmeer-lib/common/hash"
	"github.com/Qitmeer/qitmeer-lib/core/dag"
	s "github.com/Qitmeer/qitmeer-lib/core/serialization"
	"io"
)

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
	// Return block ID
	GetID() uint
	// Return the hash of block. It will be a pointer.
	GetHash() *hash.Hash

	// Acquire the layer of block
	GetLayer() uint

	// Setting the order of block
	SetOrder(o uint)

	// Acquire the order of block
	GetOrder() uint

	// IsOrdered
	IsOrdered() bool

	// Get all parents set,the dag block has more than one parent
	GetParents() *dag.HashSet

	// Testing whether it has parents
	HasParents() bool

	// Add child nodes to block
	AddChild(child *Block)

    // Get all the children of block
    GetChildren() *dag.HashSet

	// Detecting the presence of child nodes
	HasChildren() bool

	// GetMainParent
	GetMainParent() *hash.Hash

	// Setting the weight of block
	SetWeight(weight uint)

    // Acquire the weight of block
    GetWeight() uint

	// Acquire the height of block in main chain
	GetHeight() uint

	// encode
	Encode(w io.Writer) error

	// decode
	Decode(r io.Reader) error
}

// It is the element of a DAG. It is the most basic data unit.
type Block struct {
	id         uint
	hash       hash.Hash
	parents    *dag.HashSet
	children   *dag.HashSet

	mainParent *hash.Hash
	weight     uint
	order      uint
	layer      uint
	height     uint
}

// Return block ID
func (b *Block) GetID() uint {
	return b.id
}

// Return the hash of block. It will be a pointer.
func (b *Block) GetHash() *hash.Hash {
	return &b.hash
}

// Get all parents set,the dag block has more than one parent
func (b *Block) GetParents() *dag.HashSet {
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
		b.children = dag.NewHashSet()
	}
	b.children.AddPair(child.GetHash(),child)
}

// Get all the children of block
func (b *Block) GetChildren() *dag.HashSet {
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

// IsOrdered
func (b *Block) IsOrdered() bool {
	return b.GetOrder()!=MaxBlockOrder
}

// Setting the height of block in main chain
func (b *Block) SetHeight(h uint) {
	b.height=h
}

// Acquire the height of block in main chain
func (b *Block) GetHeight() uint {
	return b.height
}

// encode
func (b *Block) Encode(w io.Writer) error {
	err:=s.WriteElements(w,uint32(b.id))
	if err != nil {
		return err
	}
	err=s.WriteElements(w,&b.hash)
	if err != nil {
		return err
	}
	// parents
	parents:=[]*hash.Hash{}
	if b.parents!=nil && b.parents.Size()>0 {
		parents=b.parents.List()
	}
	parentsSize:=len(parents)
	err=s.WriteElements(w,uint32(parentsSize))
	if err != nil {
		return err
	}
	for i:=0;i<parentsSize ;i++  {
		err=s.WriteElements(w,parents[i])
		if err != nil {
			return err
		}
	}
	// children
	children:=[]*hash.Hash{}
	if b.children!=nil && b.children.Size()>0 {
		children=b.children.List()
	}
	childrenSize:=len(parents)
	err=s.WriteElements(w,uint32(childrenSize))
	if err != nil {
		return err
	}
	for i:=0;i<childrenSize ;i++  {
		err=s.WriteElements(w,children[i])
		if err != nil {
			return err
		}
	}
	// mainParent
	mainParent:=&hash.ZeroHash
	if b.mainParent!=nil {
		mainParent=b.mainParent
	}
	err=s.WriteElements(w,mainParent)
	if err != nil {
		return err
	}

	err=s.WriteElements(w,uint32(b.weight))
	if err != nil {
		return err
	}
	err=s.WriteElements(w,uint32(b.order))
	if err != nil {
		return err
	}
	err=s.WriteElements(w,uint32(b.layer))
	if err != nil {
		return err
	}
	err=s.WriteElements(w,uint32(b.height))
	if err != nil {
		return err
	}
	return nil
}

// decode
func (b *Block) Decode(r io.Reader) error {
	var id uint32
	err:=s.ReadElements(r,&id)
	if err != nil {
		return err
	}
	b.id=uint(id)

	err=s.ReadElements(r,&b.hash)
	if err != nil {
		return err
	}
	// parents
	var parentsSize uint32
	err=s.ReadElements(r,&parentsSize)
	if err != nil {
		return err
	}
	if parentsSize>0 {
		b.parents = dag.NewHashSet()
		for i:=uint32(0);i<parentsSize ;i++  {
			var parent hash.Hash
			err:=s.ReadElements(r,&parent)
			if err != nil {
				return err
			}
			b.parents.Add(&parent)
		}
	}
	// children
	var childrenSize uint32
	err=s.ReadElements(r,&childrenSize)
	if err != nil {
		return err
	}
	if childrenSize>0 {
		b.children = dag.NewHashSet()
		for i:=uint32(0);i<childrenSize ;i++  {
			var children hash.Hash
			err:=s.ReadElements(r,&children)
			if err != nil {
				return err
			}
			b.children.Add(&children)
		}
	}
	// mainParent
	var mainParent hash.Hash
	err=s.ReadElements(r,&mainParent)
	if err != nil {
		return err
	}
	if mainParent.IsEqual(&hash.ZeroHash) {
		b.mainParent=nil
	}else{
		b.mainParent=&mainParent
	}

	var weight uint32
	err=s.ReadElements(r,&weight)
	if err != nil {
		return err
	}
	b.weight=uint(weight)

	var order uint32
	err=s.ReadElements(r,&order)
	if err != nil {
		return err
	}
	b.order=uint(order)

	var layer uint32
	err=s.ReadElements(r,&layer)
	if err != nil {
		return err
	}
	b.layer=uint(layer)

	var height uint32
	err=s.ReadElements(r,&height)
	if err != nil {
		return err
	}
	b.height=uint(height)

	return nil
}