package blockdag

import (
	"container/list"
	"fmt"
	"github.com/Qitmeer/qitmeer/common/hash"
	"github.com/Qitmeer/qitmeer/core/blockdag/anticone"
	"github.com/Qitmeer/qitmeer/database"
	"io"
)

type Phantom_v2 struct {
	// The general foundation framework of DAG
	bd *BlockDAG

	// The block anticone size is all in the DAG which did not reference it and
	// were not referenced by it.
	anticoneSize int

	blocks map[hash.Hash]*PhantomBlock

	coloringTip *hash.Hash

	blueAntipastOrder *HashSet

	redAntipastOrder *HashSet

	uncoloredUnorderedAntipast *HashSet

	coloringChain *HashSet

	pastOrder *HashSet

	kChain *KChain

	incrementGenesis *hash.Hash
}

func (ph *Phantom_v2) GetName() string {
	return phantom_v2
}

func (ph *Phantom_v2) Init(bd *BlockDAG) bool {
	ph.bd=bd

	ph.anticoneSize = anticone.GetSize(BlockDelay,BlockRate,SecurityLevel)

	if log!=nil {
		log.Info(fmt.Sprintf("anticone size:%d",ph.anticoneSize))
	}

	ph.blueAntipastOrder=NewHashSet()
	ph.redAntipastOrder=NewHashSet()
	ph.uncoloredUnorderedAntipast=NewHashSet()
	ph.coloringChain=NewHashSet()
	ph.pastOrder=NewHashSet()
	return true
}

// Add a block
func (ph *Phantom_v2) AddBlock(b IBlock) *list.List {
	if ph.blocks == nil {
		ph.blocks = map[hash.Hash]*PhantomBlock{}
	}
	pb:=b.(*PhantomBlock)
	pb.SetOrder(MaxBlockOrder)
	ph.blocks[*pb.GetHash()]=pb

	ph.updateColoringIncrementally(pb)
	ph.updateTopologicalOrderIncrementally(pb)

	result:=list.New()
	result.PushBack(pb)

	return result
}

// Build self block
func (ph *Phantom_v2) CreateBlock(b *Block) IBlock {
	return &PhantomBlock{b,0,nil,nil}
}

// If the successor return nil, the underlying layer will use the default tips list.
func (ph *Phantom_v2) GetTipsList() []IBlock {
	return nil
}

// Find block hash by order, this is very fast.
func (ph *Phantom_v2) GetBlockByOrder(order uint) *hash.Hash {
	ph.updateAntipastColoring()
	if order>=ph.bd.blockTotal {
		return nil
	}
	return ph.bd.order[order]
}

// Query whether a given block is on the main chain.
func (ph *Phantom_v2) IsOnMainChain(b IBlock) bool {
	return false
}

func (ph *Phantom_v2) updateColoringIncrementally(pb *PhantomBlock) {
	if pb.HasParents() {
		cp:=ph.getBluest(pb.GetParents())
		pb.mainParent=cp.GetHash()
		pb.blueNum=cp.blueNum
	}else{
		//It is genesis
		if !pb.GetHash().IsEqual(ph.bd.GetGenesisHash()) {
			log.Error("Error genesis")
		}
	}
	pb.blueDiffAnticone=NewHashSet()
	pb.redDiffAnticone=NewHashSet()

	ph.uncoloredUnorderedAntipast.AddPair(pb.GetHash(),pb)

	ph.updateDiffColoringOfBlock(pb)
	ph.updateMaxColoring(pb)
}

func (ph *Phantom_v2) updateDiffColoringOfBlock(pb *PhantomBlock) {
	blueDiffPastOrder:=NewHashSet()
	redDiffPastOrder:=NewHashSet()

	if pb.HasParents() {
		kchain:=ph.getKChain(pb)
		parentAntipast:=ph.getAntipast(pb.mainParent)
		diffPastQueue := []*PhantomBlock{}
		for k:=range pb.GetParents().GetMap(){
			diffPastQueue=append(diffPastQueue,ph.blocks[k])
		}
		filter:=NewHashSet()
		filter.Add(pb.GetHash())

		for len(diffPastQueue) > 0 {
			cur := diffPastQueue[0]
			diffPastQueue = diffPastQueue[1:]
			//
			if blueDiffPastOrder.Has(cur.GetHash()) ||
				redDiffPastOrder.Has(cur.GetHash()) ||
				parentAntipast.Has(cur.GetHash()) ||
				filter.Has(cur.GetHash()) {
				continue
			}
			ph.colorBlock(kchain,cur,blueDiffPastOrder,redDiffPastOrder)
			filter.Add(cur.GetHash())

			for k:= range cur.GetParents().GetMap() {
				diffPastQueue = append(diffPastQueue,ph.blocks[k])
				filter.Add(&k)
			}
		}
	}
	pb.blueDiffAnticone=blueDiffPastOrder
	pb.redDiffAnticone=redDiffPastOrder
	pb.blueNum+=uint(pb.blueDiffAnticone.Size())

}

func (ph *Phantom_v2) getBluest(bs *HashSet) *PhantomBlock {
	return ph.getExtremeBlue(bs,true)
}

func (ph *Phantom_v2) getExtremeBlue(bs *HashSet,bluest bool) *PhantomBlock {
	if bs.IsEmpty() {
		return nil
	}
	var result *PhantomBlock
	for k:=range bs.GetMap() {
		pb:=ph.blocks[k]
		if result==nil {
			result=pb
		}else {
			if bluest && result.blueNum < pb.blueNum {
				result=pb
			}else if !bluest && result.blueNum > pb.blueNum {
				result=pb
			}
		}
	}
	return result
}

func (ph *Phantom_v2) getKChain(pb *PhantomBlock) *KChain {
	var blueCount int=0
	result:=&KChain{NewHashSet(),0}
	curPb:=pb
	for  {
		result.blocks.AddPair(curPb.GetHash(),curPb)
		result.miniLayer=curPb.GetLayer()
		blueCount+=curPb.blueDiffAnticone.Size()
		if blueCount > ph.anticoneSize || curPb.mainParent==nil {
			break
		}
		curPb=ph.blocks[*curPb.mainParent]
	}
	return result
}

func (ph *Phantom_v2) getAntipast(h *hash.Hash) *HashSet {
	result:=NewHashSet()
	result.AddSet(ph.blueAntipastOrder)
	result.AddSet(ph.redAntipastOrder)
	result.AddSet(ph.uncoloredUnorderedAntipast)

	if h.IsEqual(ph.coloringTip) {
		return result
	}

	curPb:=ph.blocks[*h]
	var intersection *hash.Hash
	positive:=NewHashSet()
	negative:=NewHashSet()

	for  {

		negative.AddSet(curPb.blueDiffAnticone)
		negative.AddSet(curPb.redDiffAnticone)

		if ph.coloringChain.Has(curPb.GetHash()) {
			intersection=curPb.GetHash()
			break
		}

		if curPb.mainParent==nil {
			break
		}
		curPb=ph.blocks[*curPb.mainParent]
	}

	if ph.coloringTip!=nil {
		curPb=ph.blocks[*ph.coloringTip]
		for !curPb.GetHash().IsEqual(intersection) {

			positive.AddSet(curPb.blueDiffAnticone)
			positive.AddSet(curPb.redDiffAnticone)

			if curPb.mainParent==nil {
				break
			}
			curPb=ph.blocks[*curPb.mainParent]
		}
	}


	result.AddSet(positive)
	result.RemoveSet(negative)

	return result
}

func (ph *Phantom_v2) colorBlock(kc *KChain,pb *PhantomBlock,blueOrder *HashSet,redOrder *HashSet) {
	if ph.coloringRule2(kc,pb) {
		blueOrder.Add(pb.GetHash())
	}else{
		redOrder.Add(pb.GetHash())
	}
}

func (ph *Phantom_v2) coloringRule2(kc *KChain,pb *PhantomBlock) bool {
	curPb:=pb
	for  {
		if curPb.GetLayer() < kc.miniLayer {
			return false
		}
		if kc.blocks.Has(curPb.GetHash()) {
			return true
		}
		if curPb.mainParent==nil {
			break
		}
		curPb=ph.blocks[*curPb.mainParent]
	}
	return false
}

func (ph *Phantom_v2) updateMaxColoring(pb *PhantomBlock) {
	if ph.isMaxColoringTip(pb) {
		ph.updatePastColoringAccordingTo(pb)
		ph.kChain=ph.getKChain(pb)
		if ph.incrementGenesis==nil || !ph.coloringChain.Has(ph.incrementGenesis) {
			ph.incrementGenesis=ph.getExtremeBlue(ph.coloringChain,false).GetHash()
		}
	}
}

func (ph *Phantom_v2) isMaxColoringTip(pb *PhantomBlock) bool {
	if ph.coloringTip==nil {
		return true
	}
	return pb.IsBluer(ph.blocks[*ph.coloringTip])
}

func (ph *Phantom_v2) updatePastColoringAccordingTo(pb *PhantomBlock) {
	ph.uncoloredUnorderedAntipast.AddSet(ph.blueAntipastOrder)
	ph.uncoloredUnorderedAntipast.AddSet(ph.redAntipastOrder)
	ph.uncoloredUnorderedAntipast.Add(pb.GetHash())
	ph.blueAntipastOrder.Clean()
	ph.redAntipastOrder.Clean()

	var intersection *hash.Hash
	diffPastOrderings:=[]*PhantomBlock{}
	curPb:=pb

	for  {
		if ph.coloringChain.Has(curPb.GetHash()) {
			intersection=curPb.GetHash()
			break
		}

		ph.coloringChain.Add(curPb.GetHash())
		diffPastOrderings=append(diffPastOrderings,curPb)

		if curPb.mainParent==nil {
			break
		}
		curPb=ph.blocks[*curPb.mainParent]
	}

	if ph.coloringTip!=nil {
		curPb=ph.blocks[*ph.coloringTip]
		for !curPb.GetHash().IsEqual(intersection) {
			ph.coloringChain.Remove(curPb.GetHash())
			ph.uncoloredUnorderedAntipast.AddSet(curPb.blueDiffAnticone)
			ph.uncoloredUnorderedAntipast.AddSet(curPb.redDiffAnticone)
			ph.pastOrder.Remove(curPb.GetHash())

			if curPb.mainParent==nil {
				break
			}
			curPb=ph.blocks[*curPb.mainParent]
		}
	}

	for _,v:=range diffPastOrderings{
		ph.pastOrder.AddPair(v.GetHash(),v)

		ph.uncoloredUnorderedAntipast.RemoveSet(v.blueDiffAnticone)
		ph.uncoloredUnorderedAntipast.RemoveSet(v.redDiffAnticone)
	}

	ph.coloringTip=pb.GetHash()
}

func (ph *Phantom_v2) updateTopologicalOrderIncrementally(pb *PhantomBlock) {
	ph.updateTopologicalOrderInDicts(pb)
	ph.updateOrder(pb)
}

func (ph *Phantom_v2) updateTopologicalOrderInDicts(pb *PhantomBlock) {
	var startOrder uint=0
	if pb.mainParent!=nil && ph.blocks[*pb.mainParent].GetOrder()!=MaxBlockOrder {
		startOrder=ph.blocks[*pb.mainParent].GetOrder()
	}
	ordered:=ph.calculateTopologicalOrder(pb)
	for i,v:=range ordered{
		newLOrder:=uint(i)+startOrder
		if pb.blueDiffAnticone.Has(v) {
			pb.blueDiffAnticone.AddPair(v,newLOrder)
		}else{
			pb.redDiffAnticone.AddPair(v,newLOrder)
		}
	}
}

func (ph *Phantom_v2) calculateTopologicalOrder(pb *PhantomBlock) []*hash.Hash {
	unordered:=pb.blueDiffAnticone.Clone()
	unordered.AddSet(pb.redDiffAnticone)
	toOrder:=ph.sortBlocks(pb.mainParent,pb.blueDiffAnticone,pb.GetParents(),unordered)
	ordered:=HashList{}
	orderedSet:=NewHashSet()

	for len(toOrder)>0 {
		cur:=toOrder[0]
		toOrder=toOrder[1:]

		if ordered.Has(cur) {
			continue
		}
		curBlock:=ph.blocks[*cur]
		if curBlock.HasParents() {
			curParents:=curBlock.GetParents().Intersection(unordered)
			if !curParents.IsEmpty()&&!orderedSet.Contain(curParents) {
				toOrderP:=ph.sortBlocks(curBlock.mainParent,pb.blueDiffAnticone,curParents,unordered)
				toOrder=append([]*hash.Hash{cur},toOrder...)
				toOrder=append(toOrderP,toOrder...)
				continue
			}
		}
		ordered=append(ordered,cur)
		orderedSet.Add(cur)
	}
	return ordered
}

func(ph *Phantom_v2) sortBlocks(lastBlock *hash.Hash,laterBlocks *HashSet,toSort *HashSet,unsorted *HashSet) []*hash.Hash {
	if toSort==nil || toSort.IsEmpty() {
		return []*hash.Hash{}
	}
	remaining:=NewHashSet()
	remaining.AddSet(toSort)
	remaining.Remove(lastBlock)
	remaining=remaining.Intersection(unsorted)

	blueSet:=remaining.Intersection(laterBlocks)
	blueList:=blueSet.SortList(false)

	redSet:=remaining.Clone()
	redSet.RemoveSet(blueSet)
	redList:=redSet.SortList(false)

	result:=[]*hash.Hash{}
	if lastBlock!=nil {
		result=append(result,lastBlock)
	}
	result=append(result,blueList...)
	result=append(result,redList...)

	return result
}

func(ph *Phantom_v2) updateOrder(pb *PhantomBlock) {
	pb.order=uint(pb.blueDiffAnticone.Size()+pb.redDiffAnticone.Size())
	if pb.mainParent!=nil &&ph.blocks[*pb.mainParent].GetOrder()!=MaxBlockOrder{
		pb.order+=ph.blocks[*pb.mainParent].GetOrder()
	}
}

func(ph *Phantom_v2) updateAntipastColoring() {
	for k:=range ph.uncoloredUnorderedAntipast.GetMap() {
		ph.colorBlock(ph.kChain,ph.blocks[k],ph.blueAntipastOrder,ph.redAntipastOrder)
	}
	ph.uncoloredUnorderedAntipast.Clean()
}

// return the tip of main chain
func (ph *Phantom_v2) GetMainChainTip() IBlock {
	return nil
}

// return the main parent in the parents
func (ph *Phantom_v2) GetMainParent(parents *HashSet) IBlock {
	return nil
}

// encode
func (ph *Phantom_v2) Encode(w io.Writer) error {
	return nil
}

// decode
func (ph *Phantom_v2) Decode(r io.Reader) error {
	return nil
}

func (ph *Phantom_v2) Load(dbTx database.Tx) error {
	return nil
}