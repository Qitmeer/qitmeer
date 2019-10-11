package blockdag

import (
	"container/list"
	"github.com/Qitmeer/qitmeer/common/hash"
	"github.com/Qitmeer/qitmeer/database"
	"io"
)


type Epoch struct {
	main    IBlock
	depends []IBlock
}

func (e *Epoch) GetSequence() []IBlock {
	result := []IBlock{}
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

	privotTip IBlock
}

func (con *Conflux) GetName() string {
	return conflux
}

func (con *Conflux) Init(bd *BlockDAG) bool {
	con.bd=bd
	return true
}

func (con *Conflux) AddBlock(b IBlock) *list.List {
	if b == nil {
		return nil
	}
	//
	con.updatePrivot(b)
	oldOrder:=con.bd.order
	con.bd.order = map[uint]*hash.Hash{}
	con.updateMainChain(con.bd.getGenesis(), nil, nil)

	var result *list.List
	var i uint
	for i=0;i<con.bd.blockTotal;i++ {
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

// Build self block
func (con *Conflux) CreateBlock(b *Block) IBlock {
	return b
}

func (con *Conflux) GetTipsList() []IBlock {
	if con.bd.tips.IsEmpty() || con.privotTip == nil {
		return nil
	}
	if con.bd.tips.HasOnly(con.privotTip.GetHash()) {
		return []IBlock{con.privotTip}
	}
	if !con.bd.tips.Has(con.privotTip.GetHash()) {
		return nil
	}
	tips := con.bd.tips.Clone()
	tips.Remove(con.privotTip.GetHash())
	tipsList := tips.List()
	result := []IBlock{con.privotTip}
	for _, h := range tipsList {
		result = append(result, con.bd.getBlock(h))
	}
	return result
}

func (con *Conflux) updatePrivot(b IBlock) {
	if b.GetMainParent() == nil {
		return
	}
	parent := con.bd.getBlock(b.GetMainParent())
	var newWeight uint = 0
	for h := range parent.GetChildren().GetMap() {
		block := con.bd.getBlock(&h)
		if block.GetMainParent().IsEqual(parent.GetHash()) {
			newWeight += block.GetWeight()
		}

	}
	parent.SetWeight(newWeight + 1)
	if parent.GetMainParent() != nil {
		con.updatePrivot(parent)
	}
}

func (con *Conflux) updateMainChain(b IBlock, preEpoch *Epoch, main *HashSet) {

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
		if con.bd.tips.Size() > 1 {
			virtualBlock := Block{hash: hash.Hash{}, weight: 1}
			virtualBlock.parents = NewHashSet()
			virtualBlock.parents.AddSet(con.bd.tips)
			con.updateMainChain(&virtualBlock, curEpoch, main)
		}
		return
	}
	children := b.GetChildren().List()
	if len(children) == 1 {
		con.updateMainChain(con.bd.getBlock(children[0]), curEpoch, main)
		return
	}
	var nextMain IBlock = nil
	for _, h := range children {
		child := con.bd.getBlock(h)

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
	for p := con.privotTip; p != nil; p = con.bd.getBlock(p.GetMainParent()) {
		result = append(result, p.GetHash())
	}
	return result
}

func (con *Conflux) updateOrder(b IBlock, preEpoch *Epoch, main *HashSet) *Epoch {
	var result *Epoch
	if preEpoch == nil {
		b.SetOrder( 0)
		result = &Epoch{main: b}
	} else {
		result = con.getEpoch(b, preEpoch, main)
		var dependsNum uint = 0
		if result.HasDepends() {
			dependsNum = uint(len(result.depends))
			if dependsNum == 1 {
				result.depends[0].SetOrder(preEpoch.main.GetOrder() + 1)
			} else {
				es := NewHashSet()
				for _, dep := range result.depends {
					es.Add(dep.GetHash())
				}
				result.depends = []IBlock{}
				order := 0
				for {
					if es.IsEmpty() {
						break
					}
					fbs := con.getForwardBlocks(es)
					for _, fb := range fbs {
						order++
						fb.SetOrder( preEpoch.main.GetOrder() + uint(order))
						es.Remove(fb.GetHash())
					}
					result.depends = append(result.depends, fbs...)
				}
			}
		}
		b.SetOrder(preEpoch.main.GetOrder() + 1 + dependsNum)

	}
	//update list
	sequence := result.GetSequence()
	startOrder := len(con.bd.order)
	for i, block := range sequence {
		if block.GetOrder() != uint(startOrder+i) {
			panic("epoch order error")
		}
		if !con.isVirtualBlock(block) {
			con.bd.order[block.GetOrder()]=block.GetHash()
		}
	}

	return result
}

func (con *Conflux) getEpoch(b IBlock, preEpoch *Epoch, main *HashSet) *Epoch {

	result := Epoch{main: b}
	var dependsS *HashSet

	chain := list.New()
	chain.PushBack(b)
	for {
		if chain.Len() == 0 {
			break
		}
		ele := chain.Back()
		block := ele.Value.(IBlock)
		chain.Remove(ele)
		//
		if block.HasParents() {
			for h := range block.GetParents().GetMap() {
				if main.Has(&h) || preEpoch.HasBlock(&h) {
					continue
				}
				if result.depends == nil {
					result.depends = []IBlock{}
					dependsS = NewHashSet()
				}
				if dependsS.Has(&h) {
					continue
				}
				parent := con.bd.getBlock(&h)
				result.depends = append(result.depends, parent)
				chain.PushBack(parent)
				dependsS.Add(&h)
			}
		}
	}
	return &result
}

func (con *Conflux) getForwardBlocks(bs *HashSet) []IBlock {
	result := []IBlock{}
	rs := NewHashSet()
	for h := range bs.GetMap() {
		block := con.bd.getBlock(&h)

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
	if rs.Size() == 1 {
		result = append(result, con.bd.getBlock(rs.List()[0]))
	} else if rs.Size() > 1 {
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
			result = append(result, con.bd.getBlock(minHash))
			rs.Remove(minHash)
		}
	}

	return result
}

func (con *Conflux) isVirtualBlock(b IBlock) bool {
	return b.GetHash().IsEqual(&hash.Hash{})
}

func (con *Conflux) GetBlockByOrder(order uint) *hash.Hash {
	if order>=con.bd.blockTotal {
		return nil
	}
	return con.bd.order[order]
}

// Query whether a given block is on the main chain.
func (con *Conflux) IsOnMainChain(b IBlock) bool {
	for p := con.privotTip; p != nil; p = con.bd.getBlock(p.GetMainParent()) {
		if p.GetHash().IsEqual(b.GetHash()) {
			return true
		}
		if p.GetLayer() < b.GetLayer() {
			break
		}
	}
	return false
}

// return the tip of main chain
func (con *Conflux) GetMainChainTip() IBlock {
	return nil
}

// return the main parent in the parents
func (con *Conflux) GetMainParent(parents *HashSet) IBlock {
	return nil
}

// encode
func (con *Conflux) Encode(w io.Writer) error {
	return nil
}

// decode
func (con *Conflux) Decode(r io.Reader) error {
	return nil
}

func (con *Conflux) Load(dbTx database.Tx) error {
	return nil
}