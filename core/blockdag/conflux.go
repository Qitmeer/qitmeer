package blockdag

import (
	"container/list"
	"github.com/Qitmeer/qng-core/common/hash"
	"github.com/Qitmeer/qng-core/database"
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

	// The full sequence of dag, please note that the order starts at zero.
	order map[uint]uint
}

func (con *Conflux) GetName() string {
	return conflux
}

func (con *Conflux) Init(bd *BlockDAG) bool {
	con.bd = bd
	return true
}

func (con *Conflux) AddBlock(b IBlock) (*list.List, *list.List) {
	if b == nil {
		return nil, nil
	}
	//
	con.updatePrivot(b)
	oldOrder := con.order
	con.order = map[uint]uint{}
	con.updateMainChain(con.bd.getGenesis(), nil, nil)

	var result *list.List
	var i uint
	for i = 0; i < con.bd.blockTotal; i++ {
		if result == nil {
			oldOrderL := len(oldOrder)
			if oldOrderL == 0 ||
				i >= uint(oldOrderL) ||
				oldOrder[i] != con.order[i] {
				result = list.New()
				result.PushBack(con.order[i])
			}
		} else {
			result.PushBack(con.order[i])
		}

	}
	return result, nil
}

// Build self block
func (con *Conflux) CreateBlock(b *Block) IBlock {
	return b
}

func (con *Conflux) GetTipsList() []IBlock {
	if con.bd.tips.IsEmpty() || con.privotTip == nil {
		return nil
	}
	if con.bd.tips.HasOnly(con.privotTip.GetID()) {
		return []IBlock{con.privotTip}
	}
	if !con.bd.tips.Has(con.privotTip.GetID()) {
		return nil
	}
	tips := con.bd.tips.Clone()
	tips.Remove(con.privotTip.GetID())
	//tipsList := tips.List()
	result := []IBlock{con.privotTip}
	for _, v := range tips.GetMap() {
		ib := v.(IBlock)
		result = append(result, ib)
	}
	return result
}

func (con *Conflux) updatePrivot(b IBlock) {
	if b.GetMainParent() == MaxId {
		return
	}
	parent := con.bd.getBlockById(b.GetMainParent())
	var newWeight uint64 = 0
	for h := range parent.GetChildren().GetMap() {
		block := con.bd.getBlockById(h)
		if block.GetMainParent() == parent.GetID() {
			newWeight += block.GetWeight()
		}

	}
	parent.SetWeight(newWeight + 1)
	if parent.GetMainParent() != MaxId {
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
			virtualBlock.parents = NewIdSet()
			virtualBlock.parents.AddSet(con.bd.tips)
			con.updateMainChain(&virtualBlock, curEpoch, main)
		}
		return
	}
	children := b.GetChildren().SortList(false)
	if len(children) == 1 {
		con.updateMainChain(con.bd.getBlockById(children[0]), curEpoch, main)
		return
	}
	var nextMain IBlock = nil
	for _, h := range children {
		child := con.bd.getBlockById(h)

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

func (con *Conflux) GetMainChain() []uint {
	result := []uint{}
	for p := con.privotTip; p != nil; p = con.bd.getBlockById(p.GetMainParent()) {
		result = append(result, p.GetID())
	}
	return result
}

func (con *Conflux) updateOrder(b IBlock, preEpoch *Epoch, main *HashSet) *Epoch {

	var result *Epoch
	if preEpoch == nil {
		b.SetOrder(0)
		result = &Epoch{main: b}
	} else {
		result = con.getEpoch(b, preEpoch, main)
		var dependsNum uint = 0
		if result.HasDepends() {
			dependsNum = uint(len(result.depends))
			if dependsNum == 1 {
				result.depends[0].SetOrder(preEpoch.main.GetOrder() + 1)
			} else {
				es := NewIdSet()
				for _, dep := range result.depends {
					es.Add(dep.GetID())
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
						fb.SetOrder(preEpoch.main.GetOrder() + uint(order))
						es.Remove(fb.GetID())
					}
					result.depends = append(result.depends, fbs...)
				}
			}
		}
		b.SetOrder(preEpoch.main.GetOrder() + 1 + dependsNum)

	}
	//update list
	sequence := result.GetSequence()
	startOrder := len(con.order)
	for i, block := range sequence {
		if block.GetOrder() != uint(startOrder+i) {
			panic("epoch order error")
		}
		if !con.isVirtualBlock(block) {
			con.order[block.GetOrder()] = block.GetID()
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
			ids := block.GetParents().SortList(false)
			for _, id := range ids {
				h := con.bd.getBlockById(id).GetHash()
				if main.Has(h) || preEpoch.HasBlock(h) {
					continue
				}
				if result.depends == nil {
					result.depends = []IBlock{}
					dependsS = NewHashSet()
				}
				if dependsS.Has(h) {
					continue
				}
				parent := con.bd.getBlockById(id)
				result.depends = append(result.depends, parent)
				chain.PushBack(parent)
				dependsS.Add(h)
			}
		}
	}
	return &result
}

func (con *Conflux) getForwardBlocks(bs *IdSet) []IBlock {
	result := []IBlock{}
	rs := NewIdSet()
	for h := range bs.GetMap() {
		block := con.bd.getBlockById(h)

		isParentsExit := false
		if block.HasParents() {
			for id := range block.GetParents().GetMap() {
				if bs.Has(id) {
					isParentsExit = true
					break
				}
			}
		}
		if !isParentsExit {
			rs.Add(h)
		}
	}
	if rs.Size() == 1 {
		result = append(result, con.bd.getBlockById(rs.List()[0]))
	} else if rs.Size() > 1 {
		for {
			if rs.IsEmpty() {
				break
			}
			var minHash uint = MaxId
			for h := range rs.GetMap() {
				if minHash == MaxId {
					hv := h
					minHash = hv
					continue
				}
				if minHash > h {
					minHash = h
				}
			}
			result = append(result, con.bd.getBlockById(minHash))
			rs.Remove(minHash)
		}
	}

	return result
}

func (con *Conflux) isVirtualBlock(b IBlock) bool {
	return b.GetHash().IsEqual(&hash.Hash{})
}

// Query whether a given block is on the main chain.
func (con *Conflux) IsOnMainChain(b IBlock) bool {
	for p := con.privotTip; p != nil; p = con.bd.getBlockById(p.GetMainParent()) {
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

// return the tip of main chain id
func (con *Conflux) GetMainChainTipId() uint {
	return 0
}

// return the main parent in the parents
func (con *Conflux) GetMainParent(parents *IdSet) IBlock {
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

// IsDAG
func (con *Conflux) IsDAG(parents []IBlock) bool {
	return true
}

// GetBlues
func (con *Conflux) GetBlues(parents *IdSet) uint {
	return 0
}

// IsBlue
func (con *Conflux) IsBlue(id uint) bool {
	return false
}

// getMaxParents
func (con *Conflux) getMaxParents() int {
	return 0
}

// The main parent concurrency of block
func (con *Conflux) GetMainParentConcurrency(b IBlock) int {
	return 0
}
