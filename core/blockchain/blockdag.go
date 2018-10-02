package blockchain

import (
	"sync"
	"sort"
	"container/list"
	"time"
	"github.com/noxproject/nox/common/hash"
	"fmt"
)

type BlockDAG struct {
	bc *BlockChain
	genesis hash.Hash

	mtx   sync.Mutex

	totalBlocks      uint

	publicBlueSet    *BlockSet
	tempBlueSet      *BlockSet

	lastPublicBlocks *BlockSet

	publicOrder      []*hash.Hash
	tempOrder        []*hash.Hash
	tips             *BlockSet
	//
	hourglassBlocks *BlockSet

	lastTime time.Time
}
func (bd *BlockDAG) Init(bch *BlockChain){
	bd.bc=bch
	bd.totalBlocks=0
	//bd.genesis=&bd.Genesis().hash
	bd.lastTime=time.Unix(time.Now().Unix(), 0)
}
func (bd *BlockDAG) Genesis() *blockNode {
	if bd.bc.params!=nil {
		return bd.bc.index.LookupNode(bd.bc.params.GenesisHash)
	}
	return nil
}
func (bd *BlockDAG) GetTips() *BlockSet {
	return bd.tips
}
func (bd *BlockDAG) SetTips(bs *BlockSet){
	bd.tips=bs
}
func (bd *BlockDAG) GetNodeTips() []*blockNode {
	result:=[]*blockNode{}
	for k,_:=range bd.tips.GetMap(){
		result=append(result,bd.bc.index.LookupNode(&k))
	}
	return result
}
func (bd *BlockDAG) AddBlock(b *blockNode) *list.List {
	if b == nil {
		return nil
	}
	bd.mtx.Lock()
	defer bd.mtx.Unlock()

	bd.bc.index.AddNode(b)
	bd.totalBlocks++
	bd.tempBlueSet=nil

	log.Trace(fmt.Sprintf("Add block:%v",b.hash.String()))

	t:=time.Unix(b.timestamp, 0)
	if bd.lastTime.Before(t) {
		bd.lastTime=t
	}

	bd.updateTips(b)
	bd.calculatePastBlockSetNum(b)
	//
	//obs:=NewBlockSet()
	bd.updatePublicBlueSet(&b.hash)
	bd.updateHourglass()

	return	bd.updateOrder(b)
}
func (bd *BlockDAG) updateTips(b *blockNode) {
	if bd.tips == nil {
		bd.tips = NewBlockSet()
		bd.tips.Add(&b.hash)
		bd.genesis=b.hash
		return
	}
	for k, _ := range bd.tips.GetMap() {
		node:=bd.bc.index.LookupNode(&k)
		if node==nil {
			continue
		}
		if node.children!=nil&&len(node.children)>0 {
			bd.tips.Remove(&k)
		}
	}
	bd.tips.Add(&b.hash)
}
func (bd *BlockDAG) addPastSetNum(b *blockNode, num uint64) {
	b.pastSetNum=num
}
func (bd *BlockDAG) GetPastSetNum(b *blockNode) uint64 {
	return b.pastSetNum
}
func isVirtualTip(b *blockNode, futureSet *BlockSet, anticone *BlockSet, children *BlockSet) bool {
	for k, _ := range children.GetMap() {
		if k.IsEqual(&b.hash) {
			return false
		}
		if !futureSet.Has(&k) && !anticone.Has(&k) {
			return false
		}
	}
	return true
}
func (bd *BlockDAG) recAnticone(b *blockNode, futureSet *BlockSet, anticone *BlockSet, h *hash.Hash) {
	if h.IsEqual(&b.hash) {
		return
	}
	node:=bd.bc.index.LookupNode(h)
	children := node.GetChildrenSet()
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
		parents := node.parents
		//因为parents不可能为空  所以不用判断了
		for _, v := range parents {
			bd.recAnticone(b, futureSet, anticone, &v.hash)
		}
	}
}
func (bd *BlockDAG) GetAnticone(b *blockNode, exclude *BlockSet) *BlockSet {
	futureSet := NewBlockSet()
	bd.GetFutureSet(futureSet, b)
	anticone := NewBlockSet()
	for k, _ := range bd.tips.GetMap() {
		bd.recAnticone(b, futureSet, anticone, &k)
	}
	if exclude != nil {
		anticone.Exclude(exclude)
	}
	return anticone
}
func (bd *BlockDAG) GetFutureSet(fs *BlockSet, b *blockNode) {
	children := b.children
	if children == nil || len(children) == 0 {
		return
	}
	for _, v := range children {
		if !fs.Has(&v.hash) {
			fs.Add(&v.hash)
			bd.GetFutureSet(fs, v)
		}
	}
}
func (bd *BlockDAG) calculatePastBlockSetNum(b *blockNode) {

	if b.hash.IsEqual(&bd.genesis) {
		bd.addPastSetNum(b, 0)
		return
	}
	parentsList := b.parents
	if parentsList == nil || len(parentsList) == 0 {
		return
	}
	if len(parentsList) == 1 {
		bd.addPastSetNum(b, bd.GetPastSetNum(parentsList[0])+1)
		return
	}
	anticone := bd.GetAnticone(b, nil)

	anOther := bd.GetAnticone(parentsList[0], anticone)

	bd.addPastSetNum(b, bd.GetPastSetNum(parentsList[0])+uint64(anOther.Len())+1)
}
func (bd *BlockDAG) sortBlockSet(set *BlockSet, bs *BlockSet) SortBlocks {
	sb0 := SortBlocks{}
	sb1 := SortBlocks{}

	for k, _ := range set.GetMap() {
		node:=bd.bc.index.LookupNode(&k)
		kv:=k
		if bs != nil && bs.Has(&k) {
			sb0 = append(sb0, SortBlock{&kv, bd.GetPastSetNum(node)})
		} else {
			sb1 = append(sb1, SortBlock{&kv, bd.GetPastSetNum(node)})
		}

	}
	sort.Sort(sb0)
	sort.Sort(sb1)
	sb0 = append(sb0, sb1...)
	return sb0
}
func (bd *BlockDAG) getPastSetByOrder(pastSet *BlockSet, exclude *BlockSet, h *hash.Hash) {
	if exclude.Has(h) || pastSet.Has(h) {
		return
	}

	if h.IsEqual(&bd.genesis) {
		return
	}

	parents := bd.bc.index.LookupNode(h).GetParentsSet()
	parentsList := parents.List()
	if parents == nil || len(parentsList) == 0 {
		return
	}
	for _, v := range parentsList {

		pastSet.Add(v)
		bd.getPastSetByOrder(pastSet, exclude, v)
	}
}
func (bd *BlockDAG) GetTempOrder(tempOrder *[]*hash.Hash, tempOrderM *BlockSet, bs *BlockSet, h *hash.Hash, exclude *BlockSet) {

	if exclude != nil && exclude.Has(h) {
		return
	}
	node:=bd.bc.index.LookupNode(h)
	parents := node.GetParentsSet()
	if parents != nil && parents.Len() > 0 {
		for k, _ := range parents.GetMap() {
			if exclude != nil && exclude.Has(&k) {
				continue
			}
			if !tempOrderM.Has(&k) {
				return
			}
		}
	}
	var anticone *BlockSet

	//
	if !tempOrderM.Has(h) {
		if !bd.genesis.IsEqual(h) && !bd.lastPublicBlocks.Has(h) {
			anticone = bd.GetAnticone(node, exclude)
			//
			if !anticone.IsEmpty() {
				ansb := bd.sortBlockSet(anticone, bs)
				if bs.Has(h) {
					for _, av := range ansb {
						avNode:=bd.bc.index.LookupNode(av.h)
						if bs.Has(av.h) && bd.GetPastSetNum(avNode) < bd.GetPastSetNum(node) && !tempOrderM.Has(av.h) {
							bd.GetTempOrder(tempOrder, tempOrderM, bs, av.h, exclude)
						}
					}
				} else {
					for _, av := range ansb {
						if bs.Has(av.h) && !tempOrderM.Has(av.h) {
							bd.GetTempOrder(tempOrder, tempOrderM, bs, av.h, exclude)
						}
					}
				}

			}
		}

	}
	if !tempOrderM.Has(h) {
		(*tempOrder) = append(*tempOrder, h)
		tempOrderM.Add(h)
	}
	//
	childrenSrc := node.GetChildrenSet()
	children := childrenSrc.Clone()
	if exclude != nil {
		children.Exclude(exclude)
	}
	if children == nil || children.Len() == 0 {
		return
	}
	pastSet := NewBlockSet()
	redSet := NewBlockSet()
	sb := bd.sortBlockSet(children, bs)

	for _, v := range sb {

		if bs.Has(v.h) {
			if !tempOrderM.Has(v.h) {
				pastSet.Clear()
				redSet.Clear()
				var excludeT *BlockSet
				if exclude != nil {
					excludeT = tempOrderM.Clone()
					excludeT.AddSet(exclude)
				} else {
					excludeT = tempOrderM
				}

				bd.getPastSetByOrder(pastSet, excludeT, v.h)

				inbs := pastSet.Intersection(anticone)
				if inbs != nil && inbs.Len() > 0 {
					insb := bd.sortBlockSet(inbs, bs)

					for _, v0 := range insb {
						if bs.Has(v0.h) {
							if !tempOrderM.Has(v0.h) {
								bd.GetTempOrder(tempOrder, tempOrderM, bs, v0.h, exclude)
							}
						} else {
							redSet.Add(v0.h)
						}
					}
					if !redSet.IsEmpty() {
						pastSet.Exclude(redSet)
						isAllOrder := true
						for k, _ := range pastSet.GetMap() {
							if !tempOrderM.Has(&k) {
								isAllOrder = false
								break
							}
						}
						if isAllOrder {
							redsb := bd.sortBlockSet(redSet, bs)
							for _, v1 := range redsb {
								bd.GetTempOrder(tempOrder, tempOrderM, bs, v1.h, exclude)
							}
						}
					}

				}
			}
			bd.GetTempOrder(tempOrder, tempOrderM, bs, v.h, exclude)
		}
	}
	for _, v := range sb {
		if !bs.Has(v.h) {
			bd.GetTempOrder(tempOrder, tempOrderM, bs, v.h, exclude)
		}
	}
}
func (bd *BlockDAG) updatePublicOrder(tip *hash.Hash, blueSet *BlockSet, isRollBack bool, exclude *BlockSet, curLastPublicBS *BlockSet, startIndex int) {

	if tip.IsEqual(&bd.genesis) {
		bd.publicOrder = []*hash.Hash{}
		return
	}
	node:=bd.bc.index.LookupNode(tip)
	parents := node.GetParentsSet()

	if parents.HasOnly(&bd.genesis) {
		if len(bd.publicOrder) == 0 {
			bd.publicOrder = append(bd.publicOrder, &bd.genesis)
		}
	}

	if !isRollBack {
		if blueSet == nil {
			return
		}
		tempOrder := []*hash.Hash{}
		tempOrderM := NewBlockSet()

		lpsb := bd.sortBlockSet(bd.lastPublicBlocks, blueSet)

		for _, v := range lpsb {
			bd.GetTempOrder(&tempOrder, tempOrderM, blueSet, v.h, exclude)
		}
		toLen := len(tempOrder)
		var poLen int = 0
		for i := 0; i < toLen; i++ {
			if bd.lastPublicBlocks.Has(tempOrder[i]) {
				continue
			}
			index := startIndex + i
			poLen = len(bd.publicOrder)
			if index < poLen {
				bd.publicOrder[index] = tempOrder[i]
			} else {
				bd.publicOrder = append(bd.publicOrder, tempOrder[i])
			}
		}
		poLen = len(bd.publicOrder)
		for i := poLen - 1; i >= 0; i-- {
			if bd.publicOrder[i]!=nil {
				if !curLastPublicBS.Has(bd.publicOrder[i]) {
					log.Error("order errer:end block is not new public block")
				}
				break
			}

		}

	} else {
		poLen := len(bd.publicOrder)
		rNum := 0
		for i := poLen - 1; i >= 0; i-- {
			if curLastPublicBS.Has(bd.publicOrder[i]) {
				break
			}
			bd.publicOrder[i] = nil
			rNum++
		}
		if (poLen - rNum) != startIndex {
			log.Error("order errer:number")
		}
	}
}
func (bd *BlockDAG) recPastBlockSet(genealogy *BlockSet, tipsAncestors *map[hash.Hash]*BlockSet, tipsGenealogy *map[hash.Hash]*BlockSet) {

	var maxPastHash *hash.Hash = nil
	var maxPastNum uint64 = 0
	var tipsHash *hash.Hash = nil

	for tk, v := range *tipsAncestors {
		tkv:=tk
		if v.Len() == 1 && v.Has(&bd.genesis) {
			continue
		}

		for k, _ := range v.GetMap() {
			kv:=k
			node:=bd.bc.index.LookupNode(&kv)
			pastNum := bd.GetPastSetNum(node)
			if maxPastHash == nil || maxPastNum < pastNum {
				maxPastHash = &kv
				maxPastNum = pastNum
				tipsHash = &tkv
			}
		}

	}
	if maxPastHash == nil {
		return
	}
	parents := bd.bc.index.LookupNode(maxPastHash).GetParentsSet()
	if parents == nil || parents.Len() == 0 {
		return
	}
	(*tipsAncestors)[*tipsHash].Remove(maxPastHash)
	for k, _ := range parents.GetMap() {

		if !(*tipsGenealogy)[*tipsHash].Has(&k) {
			(*tipsAncestors)[*tipsHash].Add(&k)
			(*tipsGenealogy)[*tipsHash].Add(&k)
			if genealogy != nil {
				genealogy.Add(&k)
			}
		}

	}
}
func (bd *BlockDAG) calLastPublicBlocks(tip *hash.Hash) *BlockSet {
	tips := bd.GetTips()
	if tips == nil {
		return nil
	}
	tipsList := tips.List()
	if len(tipsList) <= 1 {
		return nil
	}
	tipsGenealogy:=make(map[hash.Hash]*BlockSet)
	tipsAncestors := make(map[hash.Hash]*BlockSet)
	for _, v := range tipsList {
		tipsAncestors[*v] = NewBlockSet()
		tipsAncestors[*v].Add(v)

		tipsGenealogy[*v]=NewBlockSet()
		tipsGenealogy[*v].Add(v)
	}

	//
	for {
		hasDifferent := false
		for k, v := range tipsAncestors {
			if k.IsEqual(tip) {
				continue
			}
			if !tipsAncestors[*tip].IsEqual(v) {
				hasDifferent = true
				break
			}
		}
		if !hasDifferent {
			break
		}
		bd.recPastBlockSet(nil, &tipsAncestors, &tipsGenealogy)
	}
	return tipsAncestors[*tip]
}
func (bd *BlockDAG) calLastPublicBlocksPBS(pastBlueSet *map[hash.Hash]*BlockSet) {
	/////
	lastPFuture := NewBlockSet()
	for k, _ := range bd.lastPublicBlocks.GetMap() {
		bd.GetFutureSet(lastPFuture, bd.bc.index.LookupNode(&k))
	}

	if bd.lastPublicBlocks.Len() == 1 {
		lpbHash := bd.lastPublicBlocks.List()[0]
		if pastBlueSet != nil {
			(*pastBlueSet)[*lpbHash] = NewBlockSet()
		}

		//pastBlueSet[lpbHash].Add(lpbHash)

	} else {
		lastTempBlueSet := NewBlockSet()
		lpbAnti := make(map[hash.Hash]*BlockSet)

		for k, _ := range bd.lastPublicBlocks.GetMap() {
			lpbAnti[k] = bd.GetAnticone(bd.bc.index.LookupNode(&k), lastPFuture)
			lastTempBlueSet.AddSet(lpbAnti[k])
		}
		if pastBlueSet != nil {
			for k, _ := range lastTempBlueSet.GetMap() {
				if !bd.publicBlueSet.Has(&k) {
					lastTempBlueSet.Remove(&k)
				}
			}
			for k, _ := range bd.lastPublicBlocks.GetMap() {
				(*pastBlueSet)[k] = lastTempBlueSet.Clone()
				(*pastBlueSet)[k].Exclude(lpbAnti[k])
				(*pastBlueSet)[k].Remove(&k)
			}
		}

	}
}
func (bd *BlockDAG) calculateBlueSet(parents *BlockSet, exclude *BlockSet, pastBlueSet *map[hash.Hash]*BlockSet, usePublic bool) *BlockSet {

	parentsPBSS := make(map[hash.Hash]*BlockSet)
	for k, _ := range parents.GetMap() {
		if _, ok := (*pastBlueSet)[k]; ok {
			parentsPBSS[k] = (*pastBlueSet)[k]
		} else {
			parentsPBSS[k] = NewBlockSet()
		}

	}

	maxBluePBSHash := GetMaxLenBlockSet(parentsPBSS)
	if maxBluePBSHash == nil {
		return nil
	}
	//
	result := NewBlockSet()
	result.AddSet(parentsPBSS[*maxBluePBSHash])
	result.Add(maxBluePBSHash)

	if parents.Len() == 1 {
		return result
	}

	maxBlueAnBS := bd.GetAnticone(bd.bc.index.LookupNode(maxBluePBSHash), exclude)

	//

	if maxBlueAnBS != nil && maxBlueAnBS.Len() > 0 {

		for k, _ := range maxBlueAnBS.GetMap() {
			bAnBS := bd.GetAnticone(bd.bc.index.LookupNode(&k), exclude)
			if bAnBS == nil || bAnBS.Len() == 0 {
				continue
			}
			inBS := result.Intersection(bAnBS)
			if usePublic {
				inPBS := bd.publicBlueSet.Intersection(bAnBS)
				inBS.AddSet(inPBS)
			}

			if inBS == nil || uint32(inBS.Len()) <= bd.bc.params.AnticoneSize {
				result.Add(&k)
			}
		}
	}
	return result
}
func (bd *BlockDAG) calculatePastBlueSet(h *hash.Hash, pastBlueSet *map[hash.Hash]*BlockSet, usePublic bool) {

	_, ok := (*pastBlueSet)[*h]
	if ok {
		return
	}
	if h.IsEqual(&bd.genesis) {
		(*pastBlueSet)[*h] = NewBlockSet()
		return
	}
	//
	parents := bd.bc.index.LookupNode(h).GetParentsSet()
	if parents == nil || parents.IsEmpty() {
		return
	} else if parents.HasOnly(&bd.genesis) {
		(*pastBlueSet)[*h] = NewBlockSet()
		(*pastBlueSet)[*h].Add(&bd.genesis)
		return
	}

	for k, _ := range parents.GetMap() {
		bd.calculatePastBlueSet(&k, pastBlueSet, usePublic)
	}
	//
	anticone := bd.GetAnticone(bd.bc.index.LookupNode(h), nil)
	(*pastBlueSet)[*h] = bd.calculateBlueSet(parents, anticone, pastBlueSet, usePublic)
}
func (bd *BlockDAG) updatePublicBlueSet(tip *hash.Hash){

	if tip.IsEqual(&bd.genesis) {
		//needOrderBS.Add(tip)
		bd.publicBlueSet = NewBlockSet()
		bd.lastPublicBlocks = NewBlockSet()
		bd.updatePublicOrder(tip, nil, false, nil, nil, 0)

		return
	}
	parents := bd.bc.index.LookupNode(tip).GetParentsSet()

	if parents.HasOnly(&bd.genesis) {
		//needOrderBS.AddList(bd.tempOrder)
		bd.publicBlueSet.Clear()
		bd.publicBlueSet.Add(&bd.genesis)
		bd.lastPublicBlocks.Clear()
		bd.lastPublicBlocks.Add(&bd.genesis)
		bd.updatePublicOrder(tip, nil, false, nil, nil, 0)

	} else {
		tips := bd.GetTips()
		if tips.Len() <= 1 {
			//needOrderBS.Add(tip)
			return
		}
		curLastPublicBS := bd.calLastPublicBlocks(tip)
		if curLastPublicBS.IsEqual(bd.lastPublicBlocks) {
			return
		}
		curLPFuture := NewBlockSet()
		for k, _ := range curLastPublicBS.GetMap() {
			bd.GetFutureSet(curLPFuture, bd.bc.index.LookupNode(&k))
		}

		lastPFuture := NewBlockSet()
		for k, _ := range bd.lastPublicBlocks.GetMap() {
			bd.GetFutureSet(lastPFuture, bd.bc.index.LookupNode(&k))
		}
		//
		pastBlueSet := make(map[hash.Hash]*BlockSet)

		if lastPFuture.Contain(curLPFuture) {
			//needOrderBS.AddSet(lastPFuture)
			//
			oExclude := NewBlockSet()
			oExclude.AddSet(curLPFuture)
			for k, _ := range bd.lastPublicBlocks.GetMap() {
				oExclude.AddSet(bd.bc.index.LookupNode(&k).GetParentsSet())
			}

			bd.calLastPublicBlocksPBS(&pastBlueSet)

			for k, _ := range curLastPublicBS.GetMap() {
				bd.calculatePastBlueSet(&k, &pastBlueSet, false)
			}
			publicBlueSet := bd.calculateBlueSet(curLastPublicBS, curLPFuture, &pastBlueSet, false)
			//
			bd.updatePublicOrder(tip, publicBlueSet, false, oExclude, curLastPublicBS, int(bd.totalBlocks)-lastPFuture.Len())
			//
			bd.publicBlueSet.AddSet(publicBlueSet)
			bd.lastPublicBlocks = curLastPublicBS
		} else if curLPFuture.Contain(lastPFuture) {
			//needOrderBS.AddSet(curLPFuture)

			bd.updatePublicOrder(tip, nil, true, nil, curLastPublicBS, int(bd.totalBlocks)-curLPFuture.Len())
			bd.publicBlueSet.Exclude(curLPFuture)
			bd.lastPublicBlocks = curLastPublicBS
		} else {
			log.Error("error:public set")
		}

	}

}
func (bd *BlockDAG) GetTempBlueSet() *BlockSet {
	//
	tips := bd.GetTips()
	//

	result := NewBlockSet()
	if tips.HasOnly(&bd.genesis) {
		result = NewBlockSet()
		result.Add(&bd.genesis)
	} else {
		pastBlueSet := make(map[hash.Hash]*BlockSet)

		bd.calLastPublicBlocksPBS(&pastBlueSet)

		for k, _ := range tips.GetMap() {
			bd.calculatePastBlueSet(&k, &pastBlueSet, false)
		}

		result = bd.calculateBlueSet(tips, nil, &pastBlueSet, false)
	}
	return result
}
func (bd *BlockDAG) getTempBS() *BlockSet{
	if bd.tempBlueSet==nil {
		bd.tempBlueSet=bd.GetTempBlueSet()
	}
	return bd.tempBlueSet
}
func (bd *BlockDAG) recCalHourglass(genealogy *BlockSet, ancestors *BlockSet) {

	var maxPastHash *hash.Hash = nil
	var maxPastNum uint64 = 0

	for k, _ := range ancestors.GetMap() {
		pastNum := bd.GetPastSetNum(bd.bc.index.LookupNode(&k))
		if maxPastHash == nil || maxPastNum < pastNum {
			maxPastHash = &k
			maxPastNum = pastNum
		}
	}

	if maxPastHash == nil {
		return
	}
	parents := bd.bc.index.LookupNode(maxPastHash).GetParentsSet()
	if parents == nil || parents.Len() == 0 {
		return
	}
	ancestors.Remove(maxPastHash)
	for k, _ := range parents.GetMap() {
		if !genealogy.Has(&k) {
			ancestors.Add(&k)
			genealogy.Add(&k)
		}
	}

}
func (bd *BlockDAG) updateHourglass(){
	tips := bd.GetTips()
	if tips == nil||tips.Len()==0 {
		return
	}
	if bd.hourglassBlocks==nil {
		bd.hourglassBlocks=NewBlockSet()
	}
	if tips.HasOnly(&bd.genesis){

		bd.hourglassBlocks.Add(&bd.genesis)
		return
	}
	tempNum:=0
	for k,_:=range tips.GetMap(){
		parents:=bd.bc.index.LookupNode(&k).GetParentsSet()
		if parents!=nil&&parents.HasOnly(&bd.genesis) {
			tempNum++
		}
	}
	if tempNum==tips.Len() {
		return
	}
	//
	genealogy:=NewBlockSet()
	ancestors:=NewBlockSet()

	for k,_:=range tips.GetMap(){
		genealogy.Add(&k)
		ancestors.Add(&k)
	}
	tempBs:=bd.getTempBS()

	for  {
		bd.recCalHourglass(genealogy,ancestors)

		ne0:=tempBs.Intersection(ancestors)
		ne1:=bd.publicBlueSet.Intersection(ancestors)
		ne0.AddSet(ne1)

		ancestors=ne0


		//
		if ancestors.IsEmpty()||ancestors.HasOnly(&bd.genesis) {
			bd.hourglassBlocks.Clear()
			bd.hourglassBlocks.Add(&bd.genesis)
			return
		}

		sb := bd.sortBlockSet(ancestors,nil)
		for _,v:=range sb{
			anti:=bd.GetAnticone(bd.bc.index.LookupNode(v.h),nil)
			if anti.Len()==0 {
				bd.hourglassBlocks.Exclude(genealogy)
				bd.hourglassBlocks.Add(v.h)
				return
			}else{
				banti0:=tempBs.Intersection(anti)
				banti1:=bd.publicBlueSet.Intersection(anti)
				banti0.AddSet(banti1)

				if banti0.Len()==0 {
					bd.hourglassBlocks.Exclude(genealogy)
					bd.hourglassBlocks.Add(v.h)
					return
				}
			}
		}
	}
}
func (bd *BlockDAG) updateOrder(b *blockNode) *list.List{
	bd.tempOrder=[]*hash.Hash{}
	refNodes:=list.New()
	if bd.totalBlocks == 1 {
		bd.tempOrder=append(bd.tempOrder, &bd.genesis)
		refNodes.PushBack(bd.genesis)
		b.height=0
		return refNodes
	}
	tempOrder := []*hash.Hash{}
	tempOrderM := NewBlockSet()
	//
	blueSet := bd.getTempBS()
	lpsb := bd.sortBlockSet(bd.lastPublicBlocks, nil)
	exclude := NewBlockSet()
	for k, _ := range bd.lastPublicBlocks.GetMap() {
		exclude.AddSet(bd.bc.index.LookupNode(&k).GetParentsSet())
	}
	for _, v := range lpsb {
		bd.GetTempOrder(&tempOrder, tempOrderM, blueSet, v.h, exclude)
	}
	tLen := len(tempOrder)
	//
	pNum:=bd.GetPublicOrderNum()
	tIndex:=0
	for i := 0; i < tLen; i++ {
		if !bd.lastPublicBlocks.Has(tempOrder[i]) {
			bd.tempOrder = append(bd.tempOrder, tempOrder[i])
			//
			node:=bd.bc.index.LookupNode(tempOrder[i])

			node.height=uint64(pNum+tIndex)
			tIndex++
			if node.height==-1 {
				log.Error(fmt.Sprintf("Order error:%v",node.hash))
			}
		}
	}
	checkOrder:=bd.GetPublicOrderNum()+len(bd.tempOrder)
	if uint(checkOrder)!=bd.totalBlocks {
		log.Error(fmt.Sprintf("Order error:The number is a problem"))
	}
	//////
	tips:=bd.GetTips()
	if tips.HasOnly(&b.hash)||bd.tempOrder[len(bd.tempOrder)-1].IsEqual(&b.hash) {
		b.height=uint64(bd.totalBlocks-1)
		refNodes.PushBack(&b.hash)
		return refNodes
	}
	////
	tLen = len(bd.tempOrder)
	for i:=tLen-1;i>=0;i-- {
		refNodes.PushFront(bd.tempOrder[i])
		if bd.tempOrder[i].IsEqual(&b.hash) {
			break
		}
	}
	return refNodes
}
func (bd *BlockDAG) GetLastBlock() *blockNode{
	if bd.tempOrder==nil {
		return nil
	}
	tLen:=len(bd.tempOrder)
	if tLen>0 {
		return bd.bc.index.LookupNode(bd.tempOrder[tLen-1])
	}
	pLen:=len(bd.publicOrder)
	if pLen>0 {
		for i:=pLen-1;i>=0 ;i--  {
			if bd.publicOrder[i]!=nil {
				return bd.bc.index.LookupNode(bd.publicOrder[i])
			}
		}
	}
	return nil
}
func (bd *BlockDAG) GetPublicOrderNum() int{
	pLen:=len(bd.publicOrder)

	if pLen>0 {
		var i int
		for i=pLen-1;i>=0 ;i--  {
			if bd.publicOrder[i]!=nil {
				break
			}
		}
		return i+1
	}
	return 0
}
func (bd *BlockDAG) GetBlockOrder(h *hash.Hash) int32{
	var result int32=-1
	if bd.tempOrder==nil {
		return result
	}
	result=int32(bd.totalBlocks)
	tLen:=len(bd.tempOrder)
	if tLen>0 {
		for i:=tLen-1;i>=0 ;i--  {
			if bd.tempOrder[i]!=nil {
				result--
				if h.IsEqual(bd.tempOrder[i]) {
					return result
				}
			}
		}
	}
	pLen:=len(bd.publicOrder)
	if pLen>0 {
		for i:=pLen-1;i>=0 ;i--  {
			if bd.publicOrder[i]!=nil {
				result--
				if h.IsEqual(bd.publicOrder[i]) {
					return result
				}
			}
		}
	}

	return -1
}
func (bd *BlockDAG) GetPrevious(h *hash.Hash) *hash.Hash{
	if bd.tempOrder==nil {
		return nil
	}
	isEnd:=false
	tLen:=len(bd.tempOrder)
	if tLen>0 {
		for i:=tLen-1;i>=0 ;i--  {
			if bd.tempOrder[i]!=nil {
				if h.IsEqual(bd.tempOrder[i]) {
					if i>0 {
						return bd.tempOrder[i-1]
					}else{
						isEnd=true
					}
				}
			}
		}
	}
	pLen:=len(bd.publicOrder)
	if pLen>0 {
		for i:=pLen-1;i>=0 ;i--  {
			if bd.publicOrder[i]!=nil {
				if isEnd {
					return bd.publicOrder[i]
				}
				if h.IsEqual(bd.publicOrder[i]) {
					if i>0 {
						return bd.publicOrder[i-1]
					}
				}
			}
		}
	}

	return nil
}
func (bd *BlockDAG) NodeByOrder(order int) *hash.Hash{
	if bd.tempOrder==nil||order<0 {
		return nil
	}
	pNum:=bd.GetPublicOrderNum()
	if order<pNum {
		return bd.publicOrder[order]
	}
	rIndex:=order-pNum
	tLen:=len(bd.tempOrder)
	if rIndex<tLen {
		return bd.tempOrder[rIndex]
	}
	return nil
}
func (bd *BlockDAG) GetLastTime() *time.Time{
	return &bd.lastTime
}
///////
type SortBlock struct {
	h          *hash.Hash
	pastSetNum uint64
}
type SortBlocks []SortBlock

func (a SortBlocks) Len() int {
	return len(a)
}
func (a SortBlocks) Less(i, j int) bool {
	if a[i].pastSetNum == a[j].pastSetNum {
		return a[i].h.String() < a[j].h.String()
	}
	return a[i].pastSetNum < a[j].pastSetNum
}
func (a SortBlocks) Swap(i, j int) {
	a[i], a[j] = a[j], a[i]
}
/////////
