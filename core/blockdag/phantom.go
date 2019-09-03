package blockdag

import (
	"container/list"
	"fmt"
	"github.com/Qitmeer/qitmeer-lib/core/dag"
	"github.com/Qitmeer/qitmeer/core/blockdag/anticone"
	"github.com/Qitmeer/qitmeer-lib/common/hash"
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

	mainChain *MainChain

	diffAnticone *dag.HashSet

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

	ph.mainChain=&MainChain{dag.NewHashSet(),nil,nil}
	ph.diffAnticone=dag.NewHashSet()

	//vb
	vb:= &Block{hash: hash.ZeroHash, weight: 1, layer:0}
	ph.virtualBlock=&PhantomBlock{vb,0,dag.NewHashSet(),dag.NewHashSet()}

	return true
}

// Add a block
func (ph *Phantom) AddBlock(ib IBlock) *list.List {
	pb:=ib.(*PhantomBlock)
	pb.SetOrder(MaxBlockOrder)

	ph.updateBlockColor(pb)
	ph.updateBlockOrder(pb)

	changeBlock:=ph.updateMainChain(ph.getBluest(ph.bd.GetTips()),pb)
	ph.preUpdateVirtualBlock()
	return ph.getOrderChangeList(changeBlock)
}

// Build self block
func (ph *Phantom) CreateBlock(b *Block) IBlock {
	return &PhantomBlock{b,0,dag.NewHashSet(),dag.NewHashSet()}
}

func (ph *Phantom) updateBlockColor(pb *PhantomBlock) {

	if pb.HasParents() {
		tp:=ph.getBluest(pb.GetParents())
		pb.mainParent=tp.GetHash()
		pb.blueNum=tp.blueNum+1
		pb.height=tp.height+1

		pbAnticone:=ph.bd.GetAnticone(pb.Block,nil)
		tpAnticone:=ph.bd.GetAnticone(tp.Block,nil)
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

func (ph *Phantom) getBluest(bs *dag.HashSet) *PhantomBlock {
	return ph.getExtremeBlue(bs,true)
}

func (ph *Phantom) getExtremeBlue(bs *dag.HashSet,bluest bool) *PhantomBlock {
	if bs.IsEmpty() {
		return nil
	}
	var result *PhantomBlock
	for k:=range bs.GetMap() {
		pb:=ph.getBlock(&k)
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

func (ph *Phantom) calculateBlueSet(pb *PhantomBlock,diffAnticone *dag.HashSet) {
	kc:=ph.getKChain(pb)
	for k:=range diffAnticone.GetMap(){
		cur:=ph.getBlock(&k)
		ph.colorBlock(kc,cur,pb.blueDiffAnticone,pb.redDiffAnticone)
	}
	if diffAnticone.Size()!=pb.blueDiffAnticone.Size()+pb.redDiffAnticone.Size() {
		log.Error(fmt.Sprintf("error blue set"))
	}
	pb.blueNum+=uint(pb.blueDiffAnticone.Size())
}

func (ph *Phantom) getKChain(pb *PhantomBlock) *KChain {
	var blueCount int=0
	result:=&KChain{dag.NewHashSet(),0}
	curPb:=pb
	for  {
		result.blocks.AddPair(curPb.GetHash(),curPb)
		result.miniLayer=curPb.GetLayer()
		blueCount+=curPb.blueDiffAnticone.Size()
		if blueCount > ph.anticoneSize || curPb.mainParent==nil {
			break
		}
		curPb=ph.getBlock(curPb.mainParent)
	}
	return result
}

func (ph *Phantom) colorBlock(kc *KChain,pb *PhantomBlock,blueOrder *dag.HashSet,redOrder *dag.HashSet) {
	if ph.coloringRule(kc,pb) {
		blueOrder.Add(pb.GetHash())
	}else{
		redOrder.Add(pb.GetHash())
	}
}

func (ph *Phantom) coloringRule(kc *KChain,pb *PhantomBlock) bool {
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
		curPb=ph.getBlock(curPb.mainParent)
	}
	return false
}

func (ph *Phantom) updateBlockOrder(pb *PhantomBlock) {
	if !pb.HasParents() {
		return
	}
	order:=ph.getDiffAnticoneOrder(pb)
	l:=len(order)
	if l!=pb.blueDiffAnticone.Size()+pb.redDiffAnticone.Size() {
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
	ordered:=dag.HashList{}
	orderedSet:=dag.NewHashSet()

	for len(toOrder)>0 {
		cur:=toOrder[0]
		toOrder=toOrder[1:]

		if ordered.Has(cur) {
			continue
		}
		curBlock:=ph.getBlock(cur)
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

func(ph *Phantom) sortBlocks(lastBlock *hash.Hash,blueDiffAnticone *dag.HashSet,toSort *dag.HashSet,diffAnticone *dag.HashSet) []*hash.Hash {
	if toSort==nil || toSort.IsEmpty() {
		return []*hash.Hash{}
	}
	remaining:=dag.NewHashSet()
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

func (ph *Phantom) updateMainChain(buestTip *PhantomBlock,pb *PhantomBlock) *PhantomBlock {
	ph.virtualBlock.SetOrder(MaxBlockOrder)
	if !ph.isMaxMainTip(buestTip) {
		ph.diffAnticone.Add(pb.GetHash())
		return nil
	}
	if ph.mainChain.tip==nil {
		ph.mainChain.tip=buestTip.GetHash()
		ph.mainChain.genesis=buestTip.GetHash()
		ph.mainChain.blocks.Add(buestTip.GetHash())
		ph.diffAnticone.Clean()
		buestTip.SetOrder(0)
		ph.bd.order[0]=buestTip.GetHash()
		return buestTip
	}

	intersection,path:=ph.getIntersectionPathWithMainChain(buestTip)
	if intersection==nil {
		panic("DAG can't find intersection!")
	}
	ph.rollBackMainChain(intersection)

	ph.updateMainOrder(path,intersection)
	ph.mainChain.tip=buestTip.GetHash()

	ph.diffAnticone=ph.bd.GetAnticone(ph.bd.GetBlock(ph.mainChain.tip),nil)

	changeOrder:=ph.bd.GetBlock(intersection).GetOrder()+1
	return ph.getBlock(ph.bd.order[changeOrder])
}

func (ph *Phantom) isMaxMainTip(pb *PhantomBlock) bool {
	if ph.mainChain.tip==nil {
		return true
	}
	if ph.mainChain.tip.IsEqual(pb.GetHash()) {
		return false
	}
	return pb.IsBluer(ph.getBlock(ph.mainChain.tip))
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
		curPb=ph.getBlock(curPb.mainParent)
	}
	return intersection,result
}

func (ph *Phantom) rollBackMainChain(intersection *hash.Hash) {
	curPb:=ph.getBlock(ph.mainChain.tip)
	for  {

		if curPb.GetHash().IsEqual(intersection) {
			break
		}
		ph.mainChain.blocks.Remove(curPb.GetHash())

		if curPb.mainParent==nil {
			break
		}
		curPb=ph.getBlock(curPb.mainParent)
	}
}

func (ph *Phantom) updateMainOrder(path []*hash.Hash,intersection *hash.Hash) {
	startOrder:=ph.getBlock(intersection).GetOrder()
	l:=len(path)
	for i:=l-1;i>=0 ;i--  {
		curBlock:=ph.getBlock(path[i])
		curBlock.SetOrder(startOrder+uint(curBlock.blueDiffAnticone.Size()+curBlock.redDiffAnticone.Size()+1))
		ph.bd.order[curBlock.GetOrder()]=curBlock.GetHash()
		ph.mainChain.blocks.Add(curBlock.GetHash())
		for k,v:=range curBlock.blueDiffAnticone.GetMap() {
			dab:=ph.getBlock(&k)
			dab.SetOrder(startOrder+v.(uint))
			ph.bd.order[dab.GetOrder()]=dab.GetHash()
		}
		for k,v:=range curBlock.redDiffAnticone.GetMap() {
			dab:=ph.getBlock(&k)
			dab.SetOrder(startOrder+v.(uint))
			ph.bd.order[dab.GetOrder()]=dab.GetHash()
		}
		startOrder=curBlock.GetOrder()
	}
	//
}

func (ph *Phantom) updateVirtualBlockOrder() *PhantomBlock {
	if ph.diffAnticone.IsEmpty() ||
		ph.virtualBlock.GetOrder()!=MaxBlockOrder {
		return nil
	}
	ph.virtualBlock.parents = dag.NewHashSet()
	var maxLayer uint=0
	for k:= range ph.bd.GetTips().GetMap() {
		parent := ph.bd.GetBlock(&k)
		ph.virtualBlock.parents.AddPair(&k,parent)

		if maxLayer==0 || maxLayer < parent.GetLayer() {
			maxLayer=parent.GetLayer()
		}
	}
	ph.virtualBlock.SetLayer(maxLayer+1)

	tp:=ph.getBlock(ph.mainChain.tip)
	ph.virtualBlock.mainParent=ph.mainChain.tip
	ph.virtualBlock.blueNum=tp.blueNum+1

	ph.virtualBlock.blueDiffAnticone.Clean()
	ph.virtualBlock.redDiffAnticone.Clean()

	ph.calculateBlueSet(ph.virtualBlock,ph.diffAnticone)
	ph.updateBlockOrder(ph.virtualBlock)

	startOrder:=ph.getBlock(ph.mainChain.tip).GetOrder()
	for k,v:=range ph.virtualBlock.blueDiffAnticone.GetMap(){
		dab:=ph.getBlock(&k)
		dab.SetOrder(startOrder+v.(uint))
		ph.bd.order[dab.GetOrder()]=dab.GetHash()
	}
	for k,v:=range ph.virtualBlock.redDiffAnticone.GetMap(){
		dab:=ph.getBlock(&k)
		dab.SetOrder(startOrder+v.(uint))
		ph.bd.order[dab.GetOrder()]=dab.GetHash()
	}

	ph.virtualBlock.SetOrder(ph.bd.GetBlockTotal()+1)

	return ph.getBlock(ph.mainChain.tip)
}

func (ph *Phantom) preUpdateVirtualBlock() *PhantomBlock {
	if ph.diffAnticone.IsEmpty() ||
		ph.virtualBlock.GetOrder()!=MaxBlockOrder {
		return nil
	}
	for k:=range ph.diffAnticone.GetMap(){
		dab:=ph.getBlock(&k)
		dab.SetOrder(MaxBlockOrder)
	}
	return nil
}

func (ph *Phantom) GetDiffBlueSet() *dag.HashSet {
	if ph.mainChain.tip==nil {
		return nil
	}
	ph.updateVirtualBlockOrder()
	result:=dag.NewHashSet()
	curPb:=ph.getBlock(ph.mainChain.tip)
	for  {
		result.AddSet(curPb.blueDiffAnticone)
		if curPb.mainParent==nil {
			break
		}
		curPb=ph.getBlock(curPb.mainParent)
	}

	if ph.virtualBlock.GetOrder()!=MaxBlockOrder {
		result.AddSet(ph.virtualBlock.blueDiffAnticone)
	}
	return result
}

// If the successor return nil, the underlying layer will use the default tips list.
func (ph *Phantom) GetTipsList() []IBlock {
	return nil
}

// Find block hash by order, this is very fast.
func (ph *Phantom) GetBlockByOrder(order uint) *hash.Hash {
	if order>=ph.GetMainChainTip().GetOrder() {
		return nil
	}
	return ph.bd.order[order]
}

// Query whether a given block is on the main chain.
func (ph *Phantom) IsOnMainChain(b IBlock) bool {
	for cur:=ph.getBlock(ph.mainChain.tip); cur != nil; cur = ph.getBlock(cur.mainParent) {
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
		refNodes.PushBack(ph.bd.GetGenesisHash())
		return refNodes
	}
	if pb != nil {
		tips:=ph.bd.GetTips()
		if tips.HasOnly(pb.GetHash()) {
			refNodes.PushBack(pb.GetHash())
			return refNodes
		}
		if pb.GetHash().IsEqual(ph.GetMainChainTip().GetHash()) {
			refNodes.PushBack(pb.GetHash())
		}else if pb.IsOrdered() && pb.GetOrder() <=ph.bd.GetMainChainTip().GetOrder() {
			for i:=ph.bd.GetMainChainTip().GetOrder();i>=0;i-- {
				refNodes.PushFront(ph.bd.order[i])
				if ph.bd.order[i].IsEqual(pb.GetHash()) {
					break
				}
			}
		}
	}
	if !ph.diffAnticone.IsEmpty() {
		for k:=range ph.diffAnticone.GetMap(){
			dk:=k
			refNodes.PushBack(&dk)
		}
	}
	return refNodes
}

// return the tip of main chain
func (ph *Phantom) GetMainChainTip() IBlock {
	return ph.bd.GetBlock(ph.mainChain.tip)
}

// return the main parent in the parents
func (ph *Phantom) GetMainParent(parents *dag.HashSet) IBlock {
	if parents == nil || parents.IsEmpty() {
		return nil
	}
	if parents.Size() == 1 {
		return ph.getBlock(parents.List()[0])
	}
	return ph.getBluest(parents)
}

func (ph *Phantom) getBlock(h *hash.Hash) *PhantomBlock {
	return ph.bd.GetBlock(h).(*PhantomBlock)
}

func (ph *Phantom) GetDiffAnticone() *dag.HashSet {
	return ph.diffAnticone
}

// The main chain of DAG is support incremental expansion
type MainChain struct {
	blocks *dag.HashSet
	tip *hash.Hash
	genesis *hash.Hash
}

type KChain struct {
	blocks *dag.HashSet
	miniLayer uint
}