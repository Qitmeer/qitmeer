package blockdag

import (
	"container/list"
	"fmt"
	"qitmeer/common/anticone"
	"qitmeer/common/hash"
)

const (
	BlockDelay=15
	BlockRate=0.02
	SecurityLevel=0.01
)

type Phantom struct {
	// The general foundation framework of DAG
	bd *BlockDAG

	// The block anticone size is all in the DAG which did not reference it and
	// were not referenced by it.
	anticoneSize int

	blocks map[hash.Hash]*PhantomBlock

	mainChain *MainChain

	diffAnticone *HashSet

	virtualBlock *PhantomBlock
}

func (ph *Phantom) GetName() string {
	return phantom
}

func (ph *Phantom) Init(bd *BlockDAG) bool {
	ph.bd=bd

	ph.anticoneSize = anticone.GetSize(BlockDelay,BlockRate,SecurityLevel)

	if log!=nil {
		log.Info(fmt.Sprintf("anticone size:%d",ph.anticoneSize))
	}

	ph.bd.order = map[uint]*hash.Hash{}

	ph.mainChain=&MainChain{NewHashSet(),nil,nil}
	ph.diffAnticone=NewHashSet()

	//vb
	vb:= &Block{hash: hash.ZeroHash, weight: 1, layer:0}
	ph.virtualBlock=&PhantomBlock{vb,0,nil,NewHashSet(),NewHashSet()}

	return true
}

// Add a block
func (ph *Phantom) AddBlock(b *Block) *list.List {
	if ph.blocks == nil {
		ph.blocks = map[hash.Hash]*PhantomBlock{}
	}
	pb:=&PhantomBlock{b,0,nil,NewHashSet(),NewHashSet()}
	pb.SetOrder(MaxBlockOrder)
	ph.blocks[*pb.GetHash()]=pb

	ph.updateBlockColor(pb)
	ph.updateBlockOrder(pb)
	ph.updateMainChain(ph.getBluest(ph.bd.GetTips()),pb)
	ph.updateVirtualBlockOrder()

	return ph.getOrderChangeList(pb)
}

func (ph *Phantom) updateBlockColor(pb *PhantomBlock) {

	if pb.HasParents() {
		tp:=ph.getBluest(pb.GetParents())
		pb.mainParent=tp.GetHash()
		pb.blueNum=tp.blueNum

		pbAnticone:=ph.GetAnticone(pb.Block,nil)
		tpAnticone:=ph.GetAnticone(tp.Block,nil)
		diffAnticone:=tpAnticone.Clone()
		diffAnticone.RemoveSet(pbAnticone)

		ph.calculateBlueSet(pb,diffAnticone)
	}else{
		//It is genesis
		if !pb.GetHash().IsEqual(ph.bd.GetGenesisHash()) {
			log.Error("Error genesis")
		}
	}


}

func (ph *Phantom) getBluest(bs *HashSet) *PhantomBlock {
	return ph.getExtremeBlue(bs,true)
}

func (ph *Phantom) getExtremeBlue(bs *HashSet,bluest bool) *PhantomBlock {
	if bs.IsEmpty() {
		return nil
	}
	var result *PhantomBlock
	for k:=range bs.GetMap() {
		pb:=ph.blocks[k]
		if result==nil {
			result=pb
		}else {
			if bluest && pb.IsBluer(result) {
				result=pb
			}else if !bluest && result.IsBluer(pb) {
				result=pb
			}
		}
	}
	return result
}

func isVirtualTip(b *Block, futureSet *HashSet, anticone *HashSet, children *HashSet) bool {
	for k:= range children.GetMap() {
		if k.IsEqual(b.GetHash()) {
			return false
		}
		if !futureSet.Has(&k) && !anticone.Has(&k) {
			return false
		}
	}
	return true
}

// This function is used to GetAnticone recursion
func (ph *Phantom) recAnticone(b *Block, futureSet *HashSet, anticone *HashSet, h *hash.Hash) {
	if h.IsEqual(b.GetHash()) {
		return
	}
	node:=ph.bd.GetBlock(h)
	children := node.GetChildren()
	needRecursion := false
	if children == nil || children.Len() == 0 {
		needRecursion = true
	} else {
		needRecursion = isVirtualTip(b, futureSet, anticone, children)
	}
	if needRecursion {
		if !futureSet.Has(h) {
			anticone.Add(h)
		}
		parents := node.GetParents()

		//Because parents can not be empty, so there is no need to judge.
		for k:= range parents.GetMap() {
			ph.recAnticone(b, futureSet, anticone, &k)
		}
	}
}

// This function can get anticone set for an block that you offered in the block dag,If
// the exclude set is not empty,the final result will exclude set that you passed in.
func (ph *Phantom) GetAnticone(b *Block, exclude *HashSet) *HashSet {
	futureSet := NewHashSet()
	ph.bd.GetFutureSet(futureSet, b)
	anticone := NewHashSet()
	for k:= range ph.bd.tips.GetMap() {
		ph.recAnticone(b, futureSet, anticone, &k)
	}
	if exclude != nil {
		anticone.Exclude(exclude)
	}
	return anticone
}

func (ph *Phantom) calculateBlueSet(pb *PhantomBlock,diffAnticone *HashSet) {
	kc:=ph.getKChain(pb)
	for k,_:=range diffAnticone.GetMap(){
		cur:=ph.blocks[k]
		ph.colorBlock(kc,cur,pb.blueDiffAnticone,pb.redDiffAnticone)
	}
	if diffAnticone.Len()!=pb.blueDiffAnticone.Len()+pb.redDiffAnticone.Len() {
		log.Error(fmt.Sprintf("error blue set"))
	}
	pb.blueNum+=uint(pb.blueDiffAnticone.Len())
}

func (ph *Phantom) getKChain(pb *PhantomBlock) *KChain {
	var blueCount int=0
	result:=&KChain{NewHashSet(),0}
	curPb:=pb
	for  {
		result.blocks.AddPair(curPb.GetHash(),curPb)
		result.minimalHeight=curPb.GetLayer()
		blueCount+=curPb.blueDiffAnticone.Len()
		if blueCount > ph.anticoneSize || curPb.mainParent==nil {
			break
		}
		curPb=ph.blocks[*curPb.mainParent]
	}
	return result
}

func (ph *Phantom) colorBlock(kc *KChain,pb *PhantomBlock,blueOrder *HashSet,redOrder *HashSet) {
	if ph.coloringRule(kc,pb) {
		blueOrder.Add(pb.GetHash())
	}else{
		redOrder.Add(pb.GetHash())
	}
}

func (ph *Phantom) coloringRule(kc *KChain,pb *PhantomBlock) bool {
	curPb:=pb
	for  {
		if curPb.GetLayer() < kc.minimalHeight {
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

func (ph *Phantom) updateBlockOrder(pb *PhantomBlock) {
	if !pb.HasParents() {
		return
	}
	order:=ph.getDiffAnticoneOrder(pb)
	l:=len(order)
	if l!=pb.blueDiffAnticone.Len()+pb.redDiffAnticone.Len() {
		log.Error(fmt.Sprintf("error block order"))
	}
	for i:=0;i<l ;i++  {
		index:=i+1
		if pb.blueDiffAnticone.Has(order[i]) {
			pb.blueDiffAnticone.AddPair(order[i],uint(index))
		}else if pb.redDiffAnticone.Has(order[i])  {
			pb.redDiffAnticone.AddPair(order[i],uint(index))
		}else {
			log.Error(fmt.Sprintf("error block order index"))
		}
	}
}

func (ph *Phantom) getDiffAnticoneOrder(pb *PhantomBlock) []*hash.Hash {
	diffAnticone:=pb.blueDiffAnticone.Clone()
	diffAnticone.AddSet(pb.redDiffAnticone)
	toOrder:=ph.sortBlocks(pb.mainParent,pb.blueDiffAnticone,pb.GetParents(),diffAnticone)
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
			curParents:=curBlock.GetParents().Intersection(diffAnticone)
			if !curParents.IsEmpty()&&!orderedSet.Contain(curParents) {
				curParents.RemoveSet(orderedSet)
				toOrderP:=ph.sortBlocks(curBlock.mainParent,pb.blueDiffAnticone,curParents,diffAnticone)
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

func(ph *Phantom) sortBlocks(lastBlock *hash.Hash,blueDiffAnticone *HashSet,toSort *HashSet,diffAnticone *HashSet) []*hash.Hash {
	if toSort==nil || toSort.IsEmpty() {
		return []*hash.Hash{}
	}
	remaining:=NewHashSet()
	remaining.AddSet(toSort)
	remaining.Remove(lastBlock)
	remaining=remaining.Intersection(diffAnticone)

	blueSet:=remaining.Intersection(blueDiffAnticone)
	blueList:=blueSet.SortList(false)

	redSet:=remaining.Clone()
	redSet.RemoveSet(blueSet)
	redList:=redSet.SortList(false)

	result:=[]*hash.Hash{}
	if lastBlock!=nil && diffAnticone.Has(lastBlock) && toSort.Has(lastBlock){
		result=append(result,lastBlock)
	}
	result=append(result,blueList...)
	result=append(result,redList...)

	return result
}

func (ph *Phantom) updateMainChain(buestTip *PhantomBlock,pb *PhantomBlock) {
	ph.virtualBlock.SetOrder(MaxBlockOrder)
	if !ph.isMaxMainTip(buestTip) {
		ph.diffAnticone.Add(pb.GetHash())
		return
	}
	if ph.mainChain.tip==nil {
		ph.mainChain.tip=buestTip.GetHash()
		ph.mainChain.genesis=buestTip.GetHash()
		ph.mainChain.blocks.Add(buestTip.GetHash())
		ph.diffAnticone.Clean()
		buestTip.SetOrder(0)
		ph.bd.order[0]=buestTip.GetHash()
		return
	}

	intersection,path:=ph.getIntersectionPathWithMainChain(buestTip)
	if intersection==nil {
		log.Error("DAG can't find intersection!")
	}
	ph.rollBackMainChain(intersection)

	ph.updateMainOrder(path,intersection)
	ph.mainChain.tip=buestTip.GetHash()

	ph.diffAnticone=ph.GetAnticone(ph.bd.GetBlock(ph.mainChain.tip),nil)
}

func (ph *Phantom) isMaxMainTip(pb *PhantomBlock) bool {
	if ph.mainChain.tip==nil {
		return true
	}
	if ph.mainChain.tip.IsEqual(pb.GetHash()) {
		return false
	}
	return pb.IsBluer(ph.blocks[*ph.mainChain.tip])
}

func (ph *Phantom) getIntersectionPathWithMainChain(pb *PhantomBlock) (*hash.Hash,[]*hash.Hash) {
	result:=[]*hash.Hash{}
	var intersection *hash.Hash
	curPb:=pb
	for  {

		if ph.mainChain.blocks.Has(curPb.GetHash()) {
			intersection=curPb.GetHash()
			break
		}
		result=append(result,curPb.GetHash())
		if curPb.mainParent==nil {
			break
		}
		curPb=ph.blocks[*curPb.mainParent]
	}
	return intersection,result
}

func (ph *Phantom) rollBackMainChain(intersection *hash.Hash) {
	curPb:=ph.blocks[*ph.mainChain.tip]
	for  {

		if curPb.GetHash().IsEqual(intersection) {
			break
		}
		ph.mainChain.blocks.Remove(curPb.GetHash())

		if curPb.mainParent==nil {
			break
		}
		curPb=ph.blocks[*curPb.mainParent]
	}
}

func (ph *Phantom) updateMainOrder(path []*hash.Hash,intersection *hash.Hash) {
	startOrder:=ph.blocks[*intersection].GetOrder()
	l:=len(path)
	for i:=l-1;i>=0 ;i--  {
		curBlock:=ph.blocks[*path[i]]
		curBlock.SetOrder(startOrder+uint(curBlock.blueDiffAnticone.Len()+curBlock.redDiffAnticone.Len()+1))
		ph.bd.order[curBlock.GetOrder()]=curBlock.GetHash()
		ph.mainChain.blocks.Add(curBlock.GetHash())
		for k,v:=range curBlock.blueDiffAnticone.GetMap() {
			dab:=ph.blocks[k]
			dab.SetOrder(startOrder+v.(uint))
			ph.bd.order[dab.GetOrder()]=dab.GetHash()
		}
		for k,v:=range curBlock.redDiffAnticone.GetMap() {
			dab:=ph.blocks[k]
			dab.SetOrder(startOrder+v.(uint))
			ph.bd.order[dab.GetOrder()]=dab.GetHash()
		}
		startOrder=curBlock.GetOrder()
	}
	//
}

func (ph *Phantom) updateVirtualBlockOrder() {
	if ph.diffAnticone.IsEmpty() ||
		ph.virtualBlock.GetOrder()!=MaxBlockOrder{
		return
	}
	ph.virtualBlock.parents = NewHashSet()
	var maxLayer uint=0
	for k, _ := range ph.bd.GetTips().GetMap() {
		parent := ph.bd.GetBlock(&k)
		ph.virtualBlock.parents.AddPair(&k,parent)

		if maxLayer==0 || maxLayer < parent.GetLayer() {
			maxLayer=parent.GetLayer()
		}
	}
	ph.virtualBlock.SetLayer(maxLayer+1)

	tp:=ph.blocks[*ph.mainChain.tip]
	ph.virtualBlock.mainParent=ph.mainChain.tip
	ph.virtualBlock.blueNum=tp.blueNum

	ph.virtualBlock.blueDiffAnticone.Clean()
	ph.virtualBlock.redDiffAnticone.Clean()

	ph.calculateBlueSet(ph.virtualBlock,ph.diffAnticone)
	ph.updateBlockOrder(ph.virtualBlock)

	startOrder:=ph.blocks[*ph.mainChain.tip].GetOrder()
	for k,v:=range ph.virtualBlock.blueDiffAnticone.GetMap(){
		dab:=ph.blocks[k]
		dab.SetOrder(startOrder+v.(uint))
		ph.bd.order[dab.GetOrder()]=dab.GetHash()
	}
	for k,v:=range ph.virtualBlock.redDiffAnticone.GetMap(){
		dab:=ph.blocks[k]
		dab.SetOrder(startOrder+v.(uint))
		ph.bd.order[dab.GetOrder()]=dab.GetHash()
	}

	ph.virtualBlock.SetOrder(ph.bd.GetBlockTotal()+1)
}

func (ph *Phantom) GetDiffBlueSet() *HashSet {
	if ph.mainChain.tip==nil {
		return nil
	}
	ph.updateVirtualBlockOrder()
	result:=NewHashSet()
	curPb:=ph.blocks[*ph.mainChain.tip]
	for  {
		result.AddSet(curPb.blueDiffAnticone)
		if curPb.mainParent==nil {
			break
		}
		curPb=ph.blocks[*curPb.mainParent]
	}

	if ph.virtualBlock.GetOrder()!=MaxBlockOrder {
		result.AddSet(ph.virtualBlock.blueDiffAnticone)
	}
	return result
}

// If the successor return nil, the underlying layer will use the default tips list.
func (ph *Phantom) GetTipsList() []*Block {
	return nil
}

// Find block hash by order, this is very fast.
func (ph *Phantom) GetBlockByOrder(order uint) *hash.Hash {
	ph.updateVirtualBlockOrder()
	if order>=uint(len(ph.bd.order)) {
		return nil
	}
	return ph.bd.order[order]
}

// Query whether a given block is on the main chain.
func (ph *Phantom) IsOnMainChain(b *Block) bool {
	for cur:=ph.blocks[*ph.mainChain.tip]; cur != nil; cur = ph.blocks[*cur.mainParent] {
		if cur.GetHash().IsEqual(b.GetHash()) {
			return true
		}
		if cur.GetLayer() < b.GetLayer() {
			break
		}
		if cur.mainParent==nil {
			break
		}
	}
	return false
}

func (ph *Phantom) getOrderChangeList(pb *PhantomBlock) *list.List {
	refNodes:=list.New()
	if ph.bd.GetBlockTotal() == 1 {
		refNodes.PushBack(*ph.bd.GetGenesisHash())
		return refNodes
	}
	tips:=ph.bd.GetTips()
	if tips.HasOnly(pb.GetHash()) || pb.GetOrder()==ph.bd.GetBlockTotal()-1 {
		refNodes.PushBack(pb.GetHash())
		return refNodes
	}
	////
	for i:=ph.bd.GetBlockTotal()-1;i>=0;i-- {
		refNodes.PushFront(ph.bd.order[i])
		if ph.bd.order[i].IsEqual(pb.GetHash()) {
			break
		}
	}
	return refNodes
}
// The main chain of DAG is support incremental expansion
type MainChain struct {
	blocks *HashSet
	tip *hash.Hash
	genesis *hash.Hash
}