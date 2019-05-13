package blockdag

import (
	"container/list"
	"time"
	"qitmeer/common/hash"
)

// Some available DAG algorithm types
const (
	phantom="phantom"
	conflux="conflux"
	spectre="spectre"
)

// It will create different BlockDAG instances
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

// The abstract inferface is used to build and manager DAG
type IBlockDAG interface {
	GetName() string
	Init(bd *BlockDAG) bool

	// Add a block
	AddBlock(b *Block) *list.List

	// If the successor return nil, the underlying layer will use the default tips list.
	GetTipsList() []*Block

	// Find block hash by order, this is very fast.
	GetBlockByOrder(order uint) *hash.Hash

	// Query whether a given block is on the main chain.
	IsOnMainChain(b *Block) bool
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

// It is the element of a DAG. It is the most basic data unit.
type Block struct {
	hash     hash.Hash
	parents  *HashSet
	children *HashSet

	privot *Block
	weight uint
	order  uint
	layer  uint
}

// Return the hash of block. It will be a pointer.
func (b *Block) GetHash() *hash.Hash {
	return &b.hash
}

// Get all parents set,the dag block has more than one parent
func (b *Block) GetParents() *HashSet {
	return b.parents
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

// Acquire the order of block
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
	tips *HashSet

	// This is time when the last block have added
	lastTime time.Time

	// The full sequence of dag, please note that the order starts at zero.
	order []*hash.Hash

	// Current dag instance used. Different algorithms work according to
	// different dag types config.
	instance IBlockDAG
}

// Acquire the name of DAG instance
func (bd *BlockDAG) GetName() string {
	return bd.instance.GetName()
}

// Initialize self, the function to be invoked at the beginning
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
		if len(parents) == 0 {
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
	block := Block{hash: *b.GetHash(), weight: 1, layer:0}
	if parents != nil {
		block.parents = NewHashSet()
		var maxLayer uint=0
		for k, h := range parents {
			parent := bd.GetBlock(h)
			block.parents.AddPair(h,parent)
			parent.AddChild(&block)
			if k == 0 {
				block.privot = parent
			}

			if maxLayer==0 || maxLayer < parent.GetLayer() {
				maxLayer=parent.GetLayer()
			}
		}
		block.SetLayer(maxLayer+1)
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
	bd.updateTips(&block)
	//
	t:=time.Unix(b.GetTimestamp(), 0)
	if bd.lastTime.Before(t) {
		bd.lastTime=t
	}
	//
	return bd.instance.AddBlock(&block)
}

// Acquire the genesis block of chain
func (bd *BlockDAG) GetGenesis() *Block {
	return bd.GetBlock(&bd.genesis)
}

// Acquire the genesis block hash of chain
func (bd *BlockDAG) GetGenesisHash() *hash.Hash {
	return &bd.genesis
}

// If the block is illegal dag,will return false.
func (bd *BlockDAG) IsDAG(b IBlockData) bool {
	return true
}

// Is there a block in DAG?
func (bd *BlockDAG) HasBlock(h *hash.Hash) bool {
	return bd.GetBlock(h) != nil
}

// Is there some block in DAG?
func (bd *BlockDAG) HasBlocks(hs []*hash.Hash) bool {
	for _, h := range hs {
		if !bd.HasBlock(h) {
			return false
		}
	}
	return true
}

// Acquire one block by hash
func (bd *BlockDAG) GetBlock(h *hash.Hash) *Block {
	block, ok := bd.blocks[*h]
	if !ok {
		return nil
	}
	return block
}

// Total number of blocks
func (bd *BlockDAG) GetBlockTotal() uint {
	return bd.blockTotal
}

// return the terminal blocks, because there maybe more than one, so this is a set.
func (bd *BlockDAG) GetTips() *HashSet {
	return bd.tips
}

// Acquire the tips array of DAG
func (bd *BlockDAG) GetTipsList() []*Block {
	result:=bd.instance.GetTipsList()
	if result!=nil {
		return result
	}
	result=[]*Block{}
	for k:=range bd.tips.GetMap(){
		result=append(result,bd.GetBlock(&k))
	}
	return result
}

// Refresh the dag tip whith new block,it will cause changes in tips set.
func (bd *BlockDAG) updateTips(b *Block) {
	if bd.tips == nil {
		bd.tips = NewHashSet()
		bd.tips.AddPair(b.GetHash(),b)
		return
	}
	for k := range bd.tips.GetMap() {
		block := bd.GetBlock(&k)
		if block.HasChildren() {
			bd.tips.Remove(&k)
		}
	}
	bd.tips.AddPair(b.GetHash(),b)
}

// The last time is when add one block to DAG.
func (bd *BlockDAG) GetLastTime() *time.Time{
	return &bd.lastTime
}

// Return the full sequence array.
func (bd *BlockDAG) GetOrder() []*hash.Hash {
	return bd.order
}

// Obtain block hash by global order
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

// Returns a future collection of block. This function is a recursively called function
// So we should consider its efficiency.
func (bd *BlockDAG) GetFutureSet(fs *HashSet, b *Block) {
	children := b.GetChildren()
	if children == nil || children.IsEmpty() {
		return
	}
	for k:= range children.GetMap() {
		if !fs.Has(&k) {
			fs.Add(&k)
			bd.GetFutureSet(fs, bd.GetBlock(&k))
		}
	}
}

// Query whether a given block is on the main chain.
// Note that some DAG protocols may not support this feature.
func (bd *BlockDAG) IsOnMainChain(h *hash.Hash) bool {
	return bd.instance.IsOnMainChain(bd.GetBlock(h))
}

// Return the layer of block,it is stable.
// You can imagine that this is the main chain.
func (bd *BlockDAG) GetLayer(h *hash.Hash) uint{
	return bd.GetBlock(h).GetLayer()
}

