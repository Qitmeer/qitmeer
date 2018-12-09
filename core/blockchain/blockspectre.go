package blockchain

import (
	"github.com/noxproject/nox/common/hash"
)

type BlockSpectre struct {
	blocks map[hash.Hash]IBlock
	genesis hash.Hash
	tips    *BlockSet
	totalBlocks uint
}

func (bs *BlockSpectre) HasBlock(h *hash.Hash) bool {
	_, ok := bs.blocks[*h]
	return ok
}

func (bs *BlockSpectre) GetBlock(h *hash.Hash) IBlock {
	return bs.blocks[*h]
}

func (bs *BlockSpectre) GetFutureSet(fs *BlockSet, b IBlock) {
	if b.GetChildren() == nil || b.GetChildren().IsEmpty() {
		return
	}
	for h, _ := range b.GetChildren().GetMap() {
		if !fs.Has(&h) {
			fs.Add(&h)
			bs.GetFutureSet(fs, bs.GetBlock(&h))
		}
	}
}

func (bs *BlockSpectre) GetTips() *BlockSet {
	return bs.tips
}

func (bs *BlockSpectre) GetBlockCount() uint {
	return bs.totalBlocks
}

func (bs *BlockSpectre) AddBlock(b IBlock) bool {

	if bs.HasBlock(b.GetHash()) {
		return false
	}
	if bs.blocks==nil {
		bs.blocks= map[hash.Hash]IBlock{}
	}
	bs.blocks[*b.GetHash()]=b
	bs.totalBlocks++
	bs.updateTips(b)
	return true
}

func (bs *BlockSpectre) updateTips(b IBlock) {
	if bs.tips == nil {
		bs.tips = NewBlockSet()
		bs.tips.Add(b.GetHash())
		bs.genesis=*b.GetHash()
		return
	}
	isBelong:=bs.tips.Has(b.GetHash())

	for k, _ := range bs.tips.GetMap() {
		node:=bs.GetBlock(&k)
		if node==nil {
			continue
		}
		children:=node.GetChildren()
		if children !=nil &&!children.IsEmpty() {
			bs.tips.Remove(&k)
		}
	}
	if !isBelong {
		bs.tips.Add(b.GetHash())
	}
}

func (bs *BlockSpectre) GetGenesis() IBlock {
	return bs.GetBlock(&bs.genesis)
}

type SpectreBlock struct {
	Votes1, Votes2 int // votes in future set, -1 means not voted yet
	hash hash.Hash
	parents *BlockSet
	children *BlockSet
}

func (sb *SpectreBlock) GetHash() *hash.Hash {
	return &sb.hash
}

func (sb *SpectreBlock) GetParents() *BlockSet {
	return sb.parents
}

func (sb *SpectreBlock) GetChildren() *BlockSet {
	return sb.children
}

func (sb *SpectreBlock) GetTimestamp() int64 {
	return 0
}

func (sb *SpectreBlock) SetPastSetNum(num uint64) {

}

func (sb *SpectreBlock) GetPastSetNum() uint64 {
	return 0
}

func (sb *SpectreBlock) GetHeight() uint64 {
	return 0
}

func (sb *SpectreBlock) SetHeight(h uint64) {

}
func NewSpectreBlock(h *hash.Hash) IBlock {
	sb := &SpectreBlock{}
	sb.Votes1, sb.Votes2 = -1, -1
	sb.hash=*h
	sb.parents=NewBlockSet()
	sb.children=NewBlockSet()
	return sb
}
