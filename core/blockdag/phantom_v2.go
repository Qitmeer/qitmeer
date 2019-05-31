package blockdag

import (
	"container/list"
	"fmt"
	"qitmeer/common/anticone"
	"qitmeer/common/hash"
)

type PhantomBlock struct {
	*Block
	blueNum uint
	coloringParent *hash.Hash
	selfOrderIndex uint

	blueDiffPastOrder *HashSet
	redDiffPastOrder *HashSet
}

func (pb *PhantomBlock) IsBluer(other *PhantomBlock) bool {
	if pb.blueNum > other.blueNum ||
		(pb.blueNum == other.blueNum && pb.GetHash().String() < other.GetHash().String()) {
		return true
	}
	return false
}

type KChain struct {
	blocks *HashSet
	minimalHeight uint
}

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

	incrementGenesis hash.Hash

	coloringParent *hash.Hash
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
func (ph *Phantom_v2) AddBlock(b *Block) *list.List {
	if ph.blocks == nil {
		ph.blocks = map[hash.Hash]*PhantomBlock{}
	}
	pb:=&PhantomBlock{b,0,nil,0,nil,nil}
	ph.blocks[*pb.GetHash()]=pb

	ph.updateColoringIncrementally(pb)
	ph.updateTopologicalOrderIncrementally(pb)
	return nil
}

// If the successor return nil, the underlying layer will use the default tips list.
func (ph *Phantom_v2) GetTipsList() []*Block {
	return nil
}

// Find block hash by order, this is very fast.
func (ph *Phantom_v2) GetBlockByOrder(order uint) *hash.Hash {
	return nil
}

// Query whether a given block is on the main chain.
func (ph *Phantom_v2) IsOnMainChain(b *Block) bool {
	return false
}

func (ph *Phantom_v2) updateColoringIncrementally(pb *PhantomBlock) {
	if pb.HasParents() {
		cp:=ph.getBluest(pb.GetParents())
		pb.coloringParent=cp.GetHash()
		pb.blueNum=cp.blueNum
	}
	pb.blueDiffPastOrder=NewHashSet()
	pb.redDiffPastOrder=NewHashSet()

	ph.uncoloredUnorderedAntipast.AddPair(pb.GetHash(),pb)

	ph.updateDiffColoringOfBlock(pb)
	ph.updateMaxColoring(pb)
}

func (ph *Phantom_v2) updateDiffColoringOfBlock(pb *PhantomBlock) {
	blueDiffPastOrder:=NewHashSet()
	redDiffPastOrder:=NewHashSet()
	kchain:=ph.getKChain(pb)
	parentAntipast:=ph.getAntipast(pb.coloringParent)

	if pb.HasParents() {
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

			reference:=[]*hash.Hash{}
			if cur.HasParents() {
				reference=append(reference,cur.GetParents().List()...)
			}

			if cur.HasChildren() {
				reference=append(reference,cur.GetChildren().List()...)
			}

			for _,v := range reference {
				diffPastQueue = append(diffPastQueue,ph.blocks[*v])
				filter.Add(v)
			}
		}
	}
	pb.blueDiffPastOrder=blueDiffPastOrder
	pb.redDiffPastOrder=redDiffPastOrder
	pb.blueNum+=uint(pb.blueDiffPastOrder.Len())

}

func (ph *Phantom_v2) getBluest(bs *HashSet) *PhantomBlock {
	return ph.getExtremeBlue(bs,true)
}

func (ph *Phantom_v2) getExtremeBlue(bs *HashSet,bluest bool) *PhantomBlock {
	if bs.IsEmpty() {
		return nil
	}
	var result *PhantomBlock
	for k,_:=range bs.GetMap() {
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
		result.minimalHeight=curPb.GetLayer()
		blueCount+=curPb.blueDiffPastOrder.Len()
		if blueCount > ph.anticoneSize || curPb.coloringParent==nil {
			break
		}
		curPb=ph.blocks[*curPb.coloringParent]
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

		negative.AddSet(curPb.blueDiffPastOrder)
		negative.AddSet(curPb.redDiffPastOrder)

		if ph.coloringChain.Has(curPb.GetHash()) {
			intersection=curPb.GetHash()
			break
		}

		if curPb.coloringParent==nil {
			break
		}
		curPb=ph.blocks[*curPb.coloringParent]
	}

	curPb=ph.blocks[*ph.coloringTip]
	for !intersection.IsEqual(curPb.GetHash()) {

		positive.AddSet(curPb.blueDiffPastOrder)
		positive.AddSet(curPb.redDiffPastOrder)

		if curPb.coloringParent==nil {
			break
		}
		curPb=ph.blocks[*curPb.coloringParent]
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
		if curPb.GetLayer() < kc.minimalHeight {
			return false
		}
		if kc.blocks.Has(curPb.GetHash()) {
			return true
		}
		if curPb.coloringParent==nil {
			break
		}
		curPb=ph.blocks[*curPb.coloringParent]
	}
	return false
}

func (ph *Phantom_v2) updateMaxColoring(pb *PhantomBlock) {
	if ph.isMaxColoringTip(pb) {
		ph.updatePastColoringAccordingTo(pb)
		ph.kChain=ph.getKChain(pb)
		if !ph.coloringChain.Has(&ph.incrementGenesis) {
			ph.incrementGenesis=*ph.getExtremeBlue(ph.coloringChain,false).GetHash()
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

		if curPb.coloringParent==nil {
			break
		}
		curPb=ph.blocks[*curPb.coloringParent]
	}

	curPb=ph.blocks[*ph.coloringTip]
	for !intersection.IsEqual(curPb.GetHash()) {
		ph.coloringChain.Remove(curPb.GetHash())
		ph.uncoloredUnorderedAntipast.AddSet(curPb.blueDiffPastOrder)
		ph.uncoloredUnorderedAntipast.AddSet(curPb.redDiffPastOrder)
		ph.pastOrder.Remove(curPb.GetHash())

		if curPb.coloringParent==nil {
			break
		}
		curPb=ph.blocks[*curPb.coloringParent]
	}

	for _,v:=range diffPastOrderings{
		ph.pastOrder.AddPair(v.GetHash(),v)

		ph.uncoloredUnorderedAntipast.RemoveSet(v.blueDiffPastOrder)
		ph.uncoloredUnorderedAntipast.RemoveSet(v.redDiffPastOrder)
	}

	ph.coloringTip=pb.GetHash()
}

func (ph *Phantom_v2) updateTopologicalOrderIncrementally(pb *PhantomBlock) {
	ph.updateTopologicalOrderInDicts(pb)
	ph.updateOrder(pb)
}

func (ph *Phantom_v2) updateTopologicalOrderInDicts(pb *PhantomBlock) {
	var startOrder uint=0
	if pb.coloringParent!=nil && ph.blocks[*pb.coloringParent].GetOrder()!=0 {
		startOrder=ph.blocks[*pb.coloringParent].GetOrder()
	}
	ordered:=ph.calculateTopologicalOrder(pb)
	for i,v:=range ordered{
		newLOrder:=uint(i)+startOrder
		if pb.blueDiffPastOrder.Has(v) {
			pb.blueDiffPastOrder.AddPair(v,newLOrder)
		}else{
			pb.redDiffPastOrder.AddPair(v,newLOrder)
		}
	}
}

func (ph *Phantom_v2) calculateTopologicalOrder(pb *PhantomBlock) []*hash.Hash {
	unordered:=pb.blueDiffPastOrder.Clone()
	unordered.AddSet(pb.redDiffPastOrder)
	toOrder:=ph.sortBlocks(pb.coloringParent,pb.blueDiffPastOrder,pb.GetParents(),unordered)
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
				toOrderP:=ph.sortBlocks(curBlock.coloringParent,pb.blueDiffPastOrder,curParents,unordered)
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
	if toSort.IsEmpty() {
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
	pb.order=uint(pb.blueDiffPastOrder.Len()+pb.redDiffPastOrder.Len())
	ph.coloringParent=pb.coloringParent
	if ph.coloringParent!=nil &&ph.blocks[*ph.coloringParent].GetOrder()!=0{
		pb.order+=ph.blocks[*ph.coloringParent].GetOrder()
	}
}