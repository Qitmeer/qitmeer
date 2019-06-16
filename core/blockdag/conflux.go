package blockdag

import (
	"container/list"
	"qitmeer/common/hash"
)


type Epoch struct {
	main    *Block
	depends []*Block
}

func (e *Epoch) GetSequence() []*Block {
	result := []*Block{}
	if e.depends != nil && len(e.depends) > 0 {
		result = append(result, e.depends...)
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
	// The general foundation framework of DAG
	bd *BlockDAG

	privotTip *Block
}

func (con *Conflux) GetName() string {
	return conflux
}

func (con *Conflux) Init(bd *BlockDAG) bool {
	con.bd=bd
	return true
}

func (con *Conflux) AddBlock(b *Block) *list.List {
	if b == nil {
		return nil
	}
	//
	con.updatePrivot(b)
	oldOrder:=con.bd.order
	con.bd.order = map[uint]*hash.Hash{}
	con.updateMainChain(con.bd.GetGenesis(), nil, nil)

	var result *list.List
	var i uint
	for i=0;i<con.bd.GetBlockTotal();i++ {
		if result==nil {
			if len(oldOrder)==0||
				i>=uint(len(oldOrder))||
				!oldOrder[i].IsEqual(con.bd.order[i]) {
				result=list.New()
				result.PushBack(con.bd.order[i])
			}
		}else{
			result.PushBack(con.bd.order[i])
		}

	}
	return result
}

func (con *Conflux) GetTipsList() []*Block {
	if con.bd.tips.IsEmpty() || con.privotTip == nil {
		return nil
	}
	if con.bd.tips.HasOnly(con.privotTip.GetHash()) {
		return []*Block{con.privotTip}
	}
	if !con.bd.tips.Has(con.privotTip.GetHash()) {
		return nil
	}
	tips := con.bd.tips.Clone()
	tips.Remove(con.privotTip.GetHash())
	tipsList := tips.List()
	result := []*Block{con.privotTip}
	for _, h := range tipsList {
		result = append(result, con.bd.GetBlock(h))
	}
	return result
}

func (con *Conflux) updatePrivot(b *Block) {
	if b.privot == nil {
		return
	}
	parent := b.privot
	var newWeight uint = 0
	for h := range parent.GetChildren().GetMap() {
		block := con.bd.GetBlock(&h)
		if block.privot.GetHash().IsEqual(parent.GetHash()) {
			newWeight += block.GetWeight()
		}

	}
	parent.SetWeight(newWeight + 1)
	if parent.privot != nil {
		con.updatePrivot(parent)
	}
}

func (con *Conflux) updateMainChain(b *Block, preEpoch *Epoch, main *HashSet) {

	if main == nil {
		main = NewHashSet()
	}
	main.Add(b.GetHash())

	curEpoch := con.updateOrder(b, preEpoch, main)
	if con.isVirtualBlock(b) {
		return
	}
	if !b.HasChildren() {
		con.privotTip = b
		if con.bd.GetTips().Len() > 1 {
			virtualBlock := Block{hash: hash.Hash{}, weight: 1}
			virtualBlock.parents = NewHashSet()
			virtualBlock.parents.AddSet(con.bd.GetTips())
			con.updateMainChain(&virtualBlock, curEpoch, main)
		}
		return
	}
	children := b.GetChildren().List()
	if len(children) == 1 {
		con.updateMainChain(con.bd.GetBlock(children[0]), curEpoch, main)
		return
	}
	var nextMain *Block = nil
	for _, h := range children {
		child := con.bd.GetBlock(h)

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

func (con *Conflux) updateOrder(b *Block, preEpoch *Epoch, main *HashSet) *Epoch {
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
				es := NewHashSet()
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
	startOrder := len(con.bd.order)
	for i, block := range sequence {
		if block.order != uint(startOrder+i) {
			panic("epoch order error")
		}
		if !con.isVirtualBlock(block) {
			con.bd.order[block.order]=block.GetHash()
		}
	}

	return result
}

func (con *Conflux) getEpoch(b *Block, preEpoch *Epoch, main *HashSet) *Epoch {

	result := Epoch{main: b}
	var dependsS *HashSet

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
					dependsS = NewHashSet()
				}
				if dependsS.Has(&h) {
					continue
				}
				parent := con.bd.GetBlock(&h)
				result.depends = append(result.depends, parent)
				chain.PushBack(parent)
				dependsS.Add(&h)
			}
		}
	}
	return &result
}

func (con *Conflux) getForwardBlocks(bs *HashSet) []*Block {
	result := []*Block{}
	rs := NewHashSet()
	for h := range bs.GetMap() {
		block := con.bd.GetBlock(&h)

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
		result = append(result, con.bd.GetBlock(rs.List()[0]))
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
			result = append(result, con.bd.GetBlock(minHash))
			rs.Remove(minHash)
		}
	}

	return result
}

func (con *Conflux) isVirtualBlock(b *Block) bool {
	return b.GetHash().IsEqual(&hash.Hash{})
}

func (con *Conflux) GetBlockByOrder(order uint) *hash.Hash {
	return nil
}

// Query whether a given block is on the main chain.
func (con *Conflux) IsOnMainChain(b *Block) bool {
	for p := con.privotTip; p != nil; p = p.privot {
		if p.GetHash().IsEqual(b.GetHash()) {
			return true
		}
		if p.GetLayer() < b.GetLayer() {
			break
		}
	}
	return false
}