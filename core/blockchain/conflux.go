package blockchain

import (
	"container/list"
	"github.com/noxproject/nox/common/hash"
)

//The abstract inferface is used to dag block
type IBlockData interface {
	// Get hash of block
	GetHash() *hash.Hash

	// Get all parents set,the dag block has more than one parent
	GetParents() []*hash.Hash

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

type Epoch struct {
	main    *Block
	depends []*Block
}

func (e *Epoch) GetSequence() []*Block {
	result := []*Block{}
	if e.depends != nil && len(e.depends) > 0 {
		for _, b := range e.depends {
			result = append(result, b)
		}
	}
	result = append(result, e.main)
	return result
}

func (e *Epoch) HasBlock(h *hash.Hash) bool {
	if e.main.GetHash().IsEqual(h) {
		return true
	}
	if e.depends != nil && len(e.depends) > 0 {
		for _, b := range e.depends {
			if b.GetHash().IsEqual(h) {
				return true
			}
		}
	}
	return false
}

func (e *Epoch) HasDepends() bool {
	if e.depends == nil {
		return false
	}
	if len(e.depends) == 0 {
		return false
	}
	return true
}

type Conflux struct {
	// The genesis of block dag
	genesis hash.Hash

	// Use block hash to save all blocks with mapping
	blocks map[hash.Hash]*Block

	// The terminal block is in block dag,this block have not any connecting at present.
	tips *BlockSet

	privotTip *Block

	// The total of block
	blockTotal uint

	// The full sequence of conflux
	order []*hash.Hash
}

func (con *Conflux) AddBlock(b IBlockData) bool {
	if b == nil {
		return false
	}
	if con.HasBlock(b.GetHash()) {
		return false
	}
	var parents []*hash.Hash
	if con.GetBlockTotal() > 0 {
		parents = b.GetParents()
		if parents == nil || len(parents) == 0 {
			return false
		}
		if !con.HasBlocks(parents) {
			return false
		}
	}

	//
	block := Block{hash: *b.GetHash(), weight: 1}
	if parents != nil {
		block.parents = NewBlockSet()
		for k, h := range parents {
			block.parents.Add(h)
			parent := con.GetBlock(h)
			parent.AddChild(block.GetHash())
			if k == 0 {
				block.privot = parent
			}
		}
	}
	if con.blocks == nil {
		con.blocks = map[hash.Hash]*Block{}
	}
	con.blocks[block.hash] = &block
	if con.GetBlockTotal() == 0 {
		con.genesis = *block.GetHash()
	}
	con.blockTotal++
	//
	con.updatePrivot(&block)
	con.updateTips(&block.hash)
	con.order = []*hash.Hash{}
	con.updateMainChain(con.GetBlock(&con.genesis), nil, nil)
	return true
}

func (con *Conflux) HasBlock(h *hash.Hash) bool {
	return con.GetBlock(h) != nil
}

func (con *Conflux) HasBlocks(hs []*hash.Hash) bool {
	for _, h := range hs {
		if !con.HasBlock(h) {
			return false
		}
	}
	return true
}

func (con *Conflux) GetBlock(h *hash.Hash) *Block {
	block, ok := con.blocks[*h]
	if !ok {
		return nil
	}
	return block
}

func (con *Conflux) GetTips() *BlockSet {
	return con.tips
}

func (con *Conflux) GetTipsList() []*hash.Hash {
	if con.tips.IsEmpty() || con.privotTip == nil {
		return nil
	}
	if con.tips.HasOnly(con.privotTip.GetHash()) {
		return []*hash.Hash{con.privotTip.GetHash()}
	}
	if !con.tips.Has(con.privotTip.GetHash()) {
		return nil
	}
	tips := con.tips.Clone()
	tips.Remove(con.privotTip.GetHash())
	tipsList := tips.List()
	result := []*hash.Hash{con.privotTip.GetHash()}
	for _, h := range tipsList {
		result = append(result, h)
	}
	return result
}

func (con *Conflux) GetBlockTotal() uint {
	return con.blockTotal
}

func (con *Conflux) GetGenesis() *Block {
	return con.GetBlock(&con.genesis)
}

func (con *Conflux) updatePrivot(b *Block) {
	if b.privot == nil {
		return
	}
	parent := b.privot
	var newWeight uint = 0
	for h := range parent.GetChildren().GetMap() {
		block := con.GetBlock(&h)
		if block.privot.GetHash().IsEqual(parent.GetHash()) {
			newWeight += block.GetWeight()
		}

	}
	parent.SetWeight(newWeight + 1)
	if parent.privot != nil {
		con.updatePrivot(parent)
	}
}

func (con *Conflux) updateTips(h *hash.Hash) {
	if con.tips == nil {
		con.tips = NewBlockSet()
		con.tips.Add(h)
		return
	}
	for k := range con.tips.GetMap() {
		block := con.GetBlock(&k)
		if block.HasChildren() {
			con.tips.Remove(&k)
		}
	}
	con.tips.Add(h)
}

func (con *Conflux) updateMainChain(b *Block, preEpoch *Epoch, main *BlockSet) {

	if main == nil {
		main = NewBlockSet()
	}
	main.Add(b.GetHash())

	curEpoch := con.updateOrder(b, preEpoch, main)
	if con.isVirtualBlock(b) {
		return
	}
	if !b.HasChildren() {
		con.privotTip = b
		if con.GetTips().Len() > 1 {
			virtualBlock := Block{hash: hash.Hash{}, weight: 1}
			virtualBlock.parents = NewBlockSet()
			virtualBlock.parents.AddSet(con.GetTips())
			con.updateMainChain(&virtualBlock, curEpoch, main)
		}
		return
	}
	children := b.GetChildren().List()
	if len(children) == 1 {
		con.updateMainChain(con.GetBlock(children[0]), curEpoch, main)
		return
	}
	var nextMain *Block = nil
	for _, h := range children {
		child := con.GetBlock(h)

		if nextMain == nil {
			nextMain = child
		} else {
			if child.GetWeight() > nextMain.GetWeight() {
				nextMain = child
			} else if child.GetWeight() == nextMain.GetWeight() {
				if child.GetHash().String() < nextMain.GetHash().String() {
					nextMain = child
				}
			}
		}

	}
	if nextMain != nil {
		con.updateMainChain(nextMain, curEpoch, main)
	}
}

func (con *Conflux) GetMainChain() []*hash.Hash {
	result := []*hash.Hash{}
	for p := con.privotTip; p != nil; p = p.privot {
		result = append(result, p.GetHash())
	}
	return result
}

func (con *Conflux) updateOrder(b *Block, preEpoch *Epoch, main *BlockSet) *Epoch {
	var result *Epoch
	if preEpoch == nil {
		b.order = 0
		result = &Epoch{main: b}
	} else {
		result = con.getEpoch(b, preEpoch, main)
		var dependsNum uint = 0
		if result.HasDepends() {
			dependsNum = uint(len(result.depends))
			if dependsNum == 1 {
				result.depends[0].order = preEpoch.main.order + 1
			} else {
				es := NewBlockSet()
				for _, dep := range result.depends {
					es.Add(dep.GetHash())
				}
				result.depends = []*Block{}
				order := 0
				for {
					if es.IsEmpty() {
						break
					}
					fbs := con.getForwardBlocks(es)
					for _, fb := range fbs {
						order++
						fb.order = preEpoch.main.order + uint(order)
						es.Remove(fb.GetHash())
					}
					result.depends = append(result.depends, fbs...)
				}
			}
		}
		b.order = preEpoch.main.order + 1 + dependsNum

	}
	//update list
	sequence := result.GetSequence()
	startOrder := len(con.order)
	for i, block := range sequence {
		if block.order != uint(startOrder+i) {
			panic("epoch order error")
		}
		if !con.isVirtualBlock(block) {
			con.order = append(con.order, block.GetHash())
		}
	}

	return result
}

func (con *Conflux) getEpoch(b *Block, preEpoch *Epoch, main *BlockSet) *Epoch {

	result := Epoch{main: b}
	var dependsS *BlockSet

	chain := list.New()
	chain.PushBack(b)
	for {
		if chain.Len() == 0 {
			break
		}
		ele := chain.Back()
		block := ele.Value.(*Block)
		chain.Remove(ele)
		//
		if block.HasParents() {
			for h := range block.GetParents().GetMap() {
				if main.Has(&h) || preEpoch.HasBlock(&h) {
					continue
				}
				if result.depends == nil {
					result.depends = []*Block{}
					dependsS = NewBlockSet()
				}
				if dependsS.Has(&h) {
					continue
				}
				parent := con.GetBlock(&h)
				result.depends = append(result.depends, parent)
				chain.PushBack(parent)
				dependsS.Add(&h)
			}
		}
	}
	return &result
}

func (con *Conflux) getForwardBlocks(bs *BlockSet) []*Block {
	result := []*Block{}
	rs := NewBlockSet()
	for h := range bs.GetMap() {
		block := con.GetBlock(&h)

		isParentsExit := false
		if block.HasParents() {
			for h := range block.GetParents().GetMap() {
				if bs.Has(&h) {
					isParentsExit = true
					break
				}
			}
		}
		if !isParentsExit {
			rs.Add(&h)
		}
	}
	if rs.Len() == 1 {
		result = append(result, con.GetBlock(rs.List()[0]))
	} else if rs.Len() > 1 {
		for {
			if rs.IsEmpty() {
				break
			}
			var minHash *hash.Hash
			for h := range rs.GetMap() {
				if minHash == nil {
					hv := h
					minHash = &hv
					continue
				}
				if minHash.String() > h.String() {
					minHash = &h
				}
			}
			result = append(result, con.GetBlock(minHash))
			rs.Remove(minHash)
		}
	}

	return result
}

func (con *Conflux) GetOrder() []*hash.Hash {
	return con.order
}

func (con *Conflux) isVirtualBlock(b *Block) bool {
	return b.GetHash().IsEqual(&hash.Hash{})
}
