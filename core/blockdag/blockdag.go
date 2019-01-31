package blockdag

import (
	"container/list"
	"time"
	"github.com/noxproject/nox/common/hash"
)

const (
	phantom="phantom"
	conflux="conflux"
	spectre="spectre"
)

func NewBlockDAG(dagType string) IBlockDAG {
	switch dagType {
	case phantom:
		return &Phantom{}
	case conflux:
		return &Conflux{}
	case spectre:
		return &Spectre{}
	}
	return nil
}

//The abstract inferface is used to build and manager DAG
type IBlockDAG interface {
	GetName() string
	Init(bd *BlockDAG) bool

	// Add a block
	AddBlock(b *Block) *list.List

	// If the successor return nil, the underlying layer will use the default tips list.
	GetTipsList() []*Block

	// Find block hash by order, this is very fast.
	GetBlockByOrder(order uint) *hash.Hash
}


//The abstract inferface is used to dag block
type IBlockData interface {
	// Get hash of block
	GetHash() *hash.Hash

	// Get all parents set,the dag block has more than one parent
	GetParents() []*hash.Hash

	// Timestamp
	GetTimestamp() int64
}

type Block struct {
	hash     hash.Hash
	parents  *BlockSet
	children *BlockSet

	privot *Block
	weight uint
	order  uint
}

func (b *Block) GetHash() *hash.Hash {
	return &b.hash
}

// Get all parents set,the dag block has more than one parent
func (b *Block) GetParents() *BlockSet {
	return b.parents
}

func (b *Block) HasParents() bool {
	if b.parents == nil {
		return false
	}
	if b.parents.IsEmpty() {
		return false
	}
	return true
}

func (b *Block) AddChild(child *hash.Hash) {
	if b.children == nil {
		b.children = NewBlockSet()
	}
	b.children.Add(child)
}

func (b *Block) GetChildren() *BlockSet {
	return b.children
}

func (b *Block) HasChildren() bool {
	if b.children == nil {
		return false
	}
	if b.children.IsEmpty() {
		return false
	}
	return true
}

func (b *Block) SetWeight(weight uint) {
	b.weight = weight
}

func (b *Block) GetWeight() uint {
	return b.weight
}

func (b *Block) GetOrder() uint {
	return b.order
}

// The general foundation framework of DAG
type BlockDAG struct {
	// The genesis of block dag
	genesis hash.Hash

	// Use block hash to save all blocks with mapping
	blocks map[hash.Hash]*Block

	// The total number blocks that this dag currently owned
	blockTotal uint

	// The terminal block is in block dag,this block have not any connecting at present.
	tips *BlockSet

	// This is time when the last block have added
	lastTime time.Time

	// The full sequence of dag, please note that the order starts at zero.
	order []*hash.Hash

	// Current dag instance used. Different algorithms work according to
	// different dag types config.
	instance IBlockDAG
}

func (bd *BlockDAG) GetName() string {
	return bd.instance.GetName()
}

func (bd *BlockDAG) Init(dagType string) IBlockDAG{
	bd.instance=NewBlockDAG(dagType)
	bd.instance.Init(bd)

	bd.lastTime=time.Unix(time.Now().Unix(), 0)

	return bd.instance
}

// This is an entry for update the block dag,you need pass in a block parameter,
// If add block have failure,it will return false.
func (bd *BlockDAG) AddBlock(b IBlockData) *list.List {
	if b == nil {
		return nil
	}
	if bd.HasBlock(b.GetHash()) {
		return nil
	}
	var parents []*hash.Hash
	if bd.GetBlockTotal() > 0 {
		parents = b.GetParents()
		if parents == nil || len(parents) == 0 {
			return nil
		}
		if !bd.HasBlocks(parents) {
			return nil
		}
	}
	if !bd.IsDAG(b) {
		return nil
	}
	//
	block := Block{hash: *b.GetHash(), weight: 1}
	if parents != nil {
		block.parents = NewBlockSet()
		for k, h := range parents {
			block.parents.Add(h)
			parent := bd.GetBlock(h)
			parent.AddChild(block.GetHash())
			if k == 0 {
				block.privot = parent
			}
		}
	}
	if bd.blocks == nil {
		bd.blocks = map[hash.Hash]*Block{}
	}
	bd.blocks[block.hash] = &block
	if bd.GetBlockTotal() == 0 {
		bd.genesis = *block.GetHash()
	}
	bd.blockTotal++
	//
	bd.updateTips(block.GetHash())
	//
	t:=time.Unix(b.GetTimestamp(), 0)
	if bd.lastTime.Before(t) {
		bd.lastTime=t
	}
	//
	return bd.instance.AddBlock(&block)
}

func (bd *BlockDAG) GetGenesis() *Block {
	return bd.GetBlock(&bd.genesis)
}

func (bd *BlockDAG) GetGenesisHash() *hash.Hash {
	return &bd.genesis
}

// If the block is illegal dag,will return false.
func (bd *BlockDAG) IsDAG(b IBlockData) bool {
	return true
}

func (bd *BlockDAG) HasBlock(h *hash.Hash) bool {
	return bd.GetBlock(h) != nil
}

func (bd *BlockDAG) HasBlocks(hs []*hash.Hash) bool {
	for _, h := range hs {
		if !bd.HasBlock(h) {
			return false
		}
	}
	return true
}

func (bd *BlockDAG) GetBlock(h *hash.Hash) *Block {
	block, ok := bd.blocks[*h]
	if !ok {
		return nil
	}
	return block
}

func (bd *BlockDAG) GetBlockTotal() uint {
	return bd.blockTotal
}

// return the terminal blocks, because there maybe more than one, so this is a set.
func (bd *BlockDAG) GetTips() *BlockSet {
	return bd.tips
}

func (bd *BlockDAG) GetTipsList() []*Block {
	result:=bd.instance.GetTipsList()
	if result!=nil {
		return result
	}
	result=[]*Block{}
	for k,_:=range bd.tips.GetMap(){
		result=append(result,bd.GetBlock(&k))
	}
	return result
}

// Refresh the dag tip whith new block,it will cause changes in tips set.
func (bd *BlockDAG) updateTips(h *hash.Hash) {
	if bd.tips == nil {
		bd.tips = NewBlockSet()
		bd.tips.Add(h)
		return
	}
	for k := range bd.tips.GetMap() {
		block := bd.GetBlock(&k)
		if block.HasChildren() {
			bd.tips.Remove(&k)
		}
	}
	bd.tips.Add(h)
}

func (bd *BlockDAG) GetLastTime() *time.Time{
	return &bd.lastTime
}

func (bd *BlockDAG) GetOrder() []*hash.Hash {
	return bd.order
}

func (bd *BlockDAG) GetBlockByOrder(order uint) *hash.Hash{
	result:=bd.instance.GetBlockByOrder(order)
	if result!=nil {
		return result
	}
	if order>=uint(len(bd.order)) {
		return nil
	}
	return bd.order[order]
}

// Return the last order block
func (bd *BlockDAG) GetLastBlock() *Block{
	if bd.GetBlockTotal()==0 {
		return nil
	}
	result:=bd.GetBlockByOrder(bd.GetBlockTotal()-1)
	if result==nil {
		return nil
	}
	return bd.GetBlock(result)
}

// This function need a stable sequence,so call it before sorting the DAG.
// If the h is invalid,the function will become a little inefficient.
func (bd *BlockDAG) GetPrevious(h *hash.Hash) *hash.Hash{
	if h==nil {
		return nil
	}
	if h.IsEqual(bd.GetGenesisHash()) {
		return nil
	}
	b:=bd.GetBlock(h)
	if b==nil {
		return nil
	}
	if b.order==0{
		return nil
	}
	return bd.GetBlockByOrder(b.order-1)
}

func (bd *BlockDAG) GetFutureSet(fs *BlockSet, b *Block) {
	children := b.GetChildren()
	if children == nil || children.IsEmpty() {
		return
	}
	for k, _ := range children.GetMap() {
		if !fs.Has(&k) {
			fs.Add(&k)
			bd.GetFutureSet(fs, bd.GetBlock(&k))
		}
	}
}
