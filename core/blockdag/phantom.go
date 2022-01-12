package blockdag

import (
	"container/list"
	"fmt"
	"github.com/Qitmeer/qitmeer/common/hash"
	"github.com/Qitmeer/qitmeer/common/math"
	"github.com/Qitmeer/qitmeer/core/blockdag/anticone"
	"github.com/Qitmeer/qitmeer/core/dbnamespace"
	s "github.com/Qitmeer/qitmeer/core/serialization"
	"github.com/Qitmeer/qitmeer/core/types"
	"github.com/Qitmeer/qitmeer/database"
	"io"
)

var (
	BlockRate = anticone.DefaultBlockRate
)

type Phantom struct {
	// The general foundation framework of DAG
	bd *BlockDAG

	// The block anticone size is all in the DAG which did not reference it and
	// were not referenced by it.
	anticoneSize int

	mainChain *MainChain

	diffAnticone *IdSet

	virtualBlock *PhantomBlock
}

func (ph *Phantom) GetName() string {
	return phantom
}

func (ph *Phantom) Init(bd *BlockDAG) bool {
	ph.bd = bd
	ph.anticoneSize = anticone.GetSize(anticone.BlockDelay, bd.blockRate, anticone.SecurityLevel)
	if log != nil {
		log.Info(fmt.Sprintf("anticone size:%d", ph.anticoneSize))
	}
	// main chain
	ph.mainChain = &MainChain{bd, MaxId, 0, NewIdSet()}
	ph.diffAnticone = NewIdSet()

	//vb
	vb := &Block{hash: hash.ZeroHash, layer: 0, mainParent: MaxId}
	ph.virtualBlock = &PhantomBlock{vb, 0, NewIdSet(), NewIdSet()}

	return true
}

// Add a block
func (ph *Phantom) AddBlock(ib IBlock) (*list.List, *list.List) {
	pb := ib.(*PhantomBlock)
	pb.SetOrder(MaxBlockOrder)

	ph.updateBlockColor(pb)
	ph.updateBlockOrder(pb)

	changeBlock, oldOrders := ph.updateMainChain(ph.getBluest(ph.bd.tips), pb)
	ph.preUpdateVirtualBlock()
	return ph.getOrderChangeList(changeBlock), oldOrders
}

// Build self block
func (ph *Phantom) CreateBlock(b *Block) IBlock {
	return &PhantomBlock{b, 0, NewIdSet(), NewIdSet()}
}

func (ph *Phantom) updateBlockColor(pb *PhantomBlock) {

	if pb.HasParents() {
		tp := ph.getBluest(pb.GetParents())
		pb.mainParent = tp.GetID()
		pb.blueNum = tp.blueNum + 1
		pb.height = tp.height + 1

		diffAnticone := ph.bd.getDiffAnticone(pb, true)
		if diffAnticone == nil {
			diffAnticone = NewIdSet()
		}
		ph.calculateBlueSet(pb, diffAnticone)
	} else {
		//It is genesis
		if !pb.GetHash().IsEqual(ph.bd.GetGenesisHash()) {
			log.Error("Error genesis")
		}
	}

}

func (ph *Phantom) getBluest(bs *IdSet) *PhantomBlock {
	return ph.getExtremeBlue(bs, true)
}

func (ph *Phantom) getExtremeBlue(bs *IdSet, bluest bool) *PhantomBlock {
	if bs.IsEmpty() {
		return nil
	}
	var result *PhantomBlock
	for k := range bs.GetMap() {
		pb := ph.getBlock(k)
		if result == nil {
			result = pb
		} else {
			if bluest && pb.IsBluer(result) {
				result = pb
			} else if !bluest && result.IsBluer(pb) {
				result = pb
			}
		}
	}
	return result
}

func (ph *Phantom) calculateBlueSet(pb *PhantomBlock, diffAnticone *IdSet) {
	kc := ph.getKChain(pb)
	for _, v := range diffAnticone.GetMap() {
		cur, ok := v.(*PhantomBlock)
		if !ok {
			panic("phantom block type is error.")
		}
		ph.colorBlock(kc, cur, pb.blueDiffAnticone, pb.redDiffAnticone)
	}
	if diffAnticone.Size() != pb.blueDiffAnticone.Size()+pb.redDiffAnticone.Size() {
		log.Error(fmt.Sprintf("error blue set"))
	}
	pb.blueNum += uint(pb.blueDiffAnticone.Size())
}

func (ph *Phantom) getKChain(pb *PhantomBlock) *KChain {
	var blueCount int = 0
	result := &KChain{NewIdSet(), 0}
	curPb := pb
	deep := 0
	for {
		result.blocks.AddPair(curPb.GetID(), curPb)
		blueCount++
		deep++
		result.miniLayer = curPb.GetLayer()
		blueCount += curPb.blueDiffAnticone.Size()
		if (blueCount > ph.anticoneSize && deep > MaxTipLayerGap*2) || curPb.mainParent == MaxId {
			break
		}
		curPb = ph.getBlock(curPb.mainParent)
	}
	return result
}

func (ph *Phantom) colorBlock(kc *KChain, pb *PhantomBlock, blueOrder *IdSet, redOrder *IdSet) {
	if ph.coloringRule(kc, pb) {
		blueOrder.Add(pb.GetID())
	} else {
		redOrder.Add(pb.GetID())
	}
}

func (ph *Phantom) coloringRule(kc *KChain, pb *PhantomBlock) bool {
	curPb := pb
	for {
		if curPb.GetLayer() < kc.miniLayer {
			return false
		}
		if kc.blocks.Has(curPb.GetID()) {
			return true
		}
		if curPb.mainParent == MaxId {
			break
		}
		curPb = ph.getBlock(curPb.mainParent)
	}
	return false
}

func (ph *Phantom) updateBlockOrder(pb *PhantomBlock) {
	if !pb.HasParents() {
		return
	}
	order := ph.getDiffAnticoneOrder(pb)
	l := len(order)
	if l != pb.blueDiffAnticone.Size()+pb.redDiffAnticone.Size() {
		log.Error(fmt.Sprintf("error block order"))
	}
	for i := 0; i < l; i++ {
		index := i + 1
		if pb.blueDiffAnticone.Has(order[i]) {
			pb.blueDiffAnticone.AddPair(order[i], uint(index))
		} else if pb.redDiffAnticone.Has(order[i]) {
			pb.redDiffAnticone.AddPair(order[i], uint(index))
		} else {
			log.Error(fmt.Sprintf("error block order index"))
		}
	}
}

func (ph *Phantom) getDiffAnticoneOrder(pb *PhantomBlock) []uint {
	blueDiffAnticone := ph.buildSortDiffAnticone(pb.blueDiffAnticone)
	redDiffAnticone := ph.buildSortDiffAnticone(pb.redDiffAnticone)
	diffAnticone := blueDiffAnticone.Clone()
	diffAnticone.AddSet(redDiffAnticone)
	toOrder := ph.sortBlocks(pb.mainParent, blueDiffAnticone, pb.GetParents(), diffAnticone)
	ordered := IdSlice{}
	orderedSet := NewIdSet()

	for len(toOrder) > 0 {
		cur := toOrder[0]
		toOrder = toOrder[1:]

		if ordered.Has(cur) {
			continue
		}
		curBlock := ph.getBlock(cur)
		if curBlock.HasParents() {
			curParents := curBlock.GetParents().Intersection(diffAnticone)
			if !curParents.IsEmpty() && !orderedSet.Contain(curParents) {
				curParents.RemoveSet(orderedSet)
				toOrderP := ph.sortBlocks(curBlock.mainParent, blueDiffAnticone, curParents, diffAnticone)
				toOrder = append([]uint{cur}, toOrder...)
				toOrder = append(toOrderP, toOrder...)
				continue
			}
		}
		ordered = append(ordered, cur)
		orderedSet.Add(cur)
	}
	return ordered
}

func (ph *Phantom) sortBlocks(lastBlock uint, blueDiffAnticone *IdSet, toSort *IdSet, diffAnticone *IdSet) []uint {
	if toSort == nil || toSort.IsEmpty() {
		return []uint{}
	}
	remaining := NewIdSet()
	remaining.AddSet(toSort)
	remaining.Remove(lastBlock)
	remaining = remaining.Intersection(diffAnticone)

	blueSet := remaining.Intersection(blueDiffAnticone)
	blueList := blueSet.SortPriorityList(true)

	redSet := remaining.Clone()
	redSet.RemoveSet(blueSet)
	redList := redSet.SortPriorityList(true)

	result := []uint{}
	if lastBlock != MaxId && diffAnticone.Has(lastBlock) && toSort.Has(lastBlock) {
		result = append(result, lastBlock)
	}
	result = append(result, blueList...)
	result = append(result, redList...)

	return result
}

func (ph *Phantom) buildSortDiffAnticone(diffAn *IdSet) *IdSet {
	result := NewIdSet()
	for k := range diffAn.GetMap() {
		ib := ph.getBlock(k)
		if ib != nil {
			result.AddPair(k, ib)
		}
	}
	return result
}

func (ph *Phantom) updateMainChain(buestTip *PhantomBlock, pb *PhantomBlock) (*PhantomBlock, *list.List) {
	if ph.bd.lastSnapshot.IsValid() {
		ph.bd.lastSnapshot.diffAnticone = ph.diffAnticone.Clone()
		ph.bd.lastSnapshot.mainChainTip = ph.mainChain.tip
		ph.bd.lastSnapshot.mainChainGenesis = ph.mainChain.genesis
	}

	ph.virtualBlock.SetOrder(MaxBlockOrder)
	if !ph.isMaxMainTip(buestTip) {
		ph.diffAnticone.AddPair(pb.GetID(), pb)
		return nil, nil
	}
	if ph.mainChain.tip == MaxId {
		ph.mainChain.tip = buestTip.GetID()
		ph.mainChain.genesis = buestTip.GetID()
		ph.mainChain.commitBlocks.AddPair(buestTip.GetID(), true)
		ph.diffAnticone.Clean()
		buestTip.SetOrder(0)
		ph.bd.commitOrder[0] = buestTip.GetID()
		return buestTip, nil
	}

	intersection, path := ph.getIntersectionPathWithMainChain(buestTip)
	intersectionBlock := ph.bd.getBlockById(intersection)
	if intersectionBlock == nil {
		panic("DAG can't find intersection!")
	}

	// old orders
	oldOrders := list.New()
	for i := intersectionBlock.GetOrder() + 1; i <= ph.GetMainChainTip().GetOrder(); i++ {
		ib := ph.bd.getBlockByOrder(i)
		if ib == nil {
			panic(fmt.Errorf("DAG can't find block in order(%d)\n", i))
		}
		oldOrders.PushBack(&BlockOrderHelp{OldOrder: i, Block: ib})
	}

	//
	ph.rollBackMainChain(intersection)

	ph.updateMainOrder(path, intersection)
	ph.mainChain.tip = buestTip.GetID()

	ph.diffAnticone = ph.bd.getAnticone(ph.bd.getBlockById(ph.mainChain.tip), nil)

	changeOrder := intersectionBlock.GetOrder() + 1

	coPB, ok := ph.bd.getBlockByOrder(changeOrder).(*PhantomBlock)
	if !ok {
		return nil, nil
	}
	return coPB, oldOrders
}

func (ph *Phantom) isMaxMainTip(pb *PhantomBlock) bool {
	if ph.mainChain.tip == MaxId {
		return true
	}
	if ph.mainChain.tip == pb.GetID() {
		return false
	}
	return pb.IsBluer(ph.getBlock(ph.mainChain.tip))
}

func (ph *Phantom) getIntersectionPathWithMainChain(pb *PhantomBlock) (uint, []uint) {
	result := []uint{}
	var intersection uint = MaxId
	curPb := pb
	for {

		if ph.mainChain.Has(curPb.GetID()) {
			intersection = curPb.GetID()
			break
		}
		result = append(result, curPb.GetID())
		if curPb.mainParent == MaxId {
			break
		}
		curPb = ph.getBlock(curPb.mainParent)
	}
	return intersection, result
}

func (ph *Phantom) rollBackMainChain(intersection uint) {
	curPb := ph.getBlock(ph.mainChain.tip)
	for {

		if curPb.GetID() == intersection {
			break
		}
		ph.mainChain.commitBlocks.AddPair(curPb.GetID(), false)

		if curPb.mainParent == MaxId {
			break
		}
		curPb = ph.getBlock(curPb.mainParent)
	}
}

func (ph *Phantom) updateMainOrder(path []uint, intersection uint) {
	startOrder := ph.getBlock(intersection).GetOrder()
	l := len(path)
	for i := l - 1; i >= 0; i-- {
		curBlock := ph.getBlock(path[i])
		ph.bd.lastSnapshot.AddOrder(curBlock)
		curBlock.SetOrder(startOrder + uint(curBlock.blueDiffAnticone.Size()+curBlock.redDiffAnticone.Size()+1))
		ph.bd.commitOrder[curBlock.GetOrder()] = curBlock.GetID()
		ph.bd.commitBlock.AddPair(curBlock.GetID(), curBlock)

		ph.mainChain.commitBlocks.AddPair(curBlock.GetID(), true)
		for k, v := range curBlock.blueDiffAnticone.GetMap() {
			dab := ph.getBlock(k)
			ph.bd.lastSnapshot.AddOrder(dab)
			dab.SetOrder(startOrder + v.(uint))
			ph.bd.commitOrder[dab.GetOrder()] = dab.GetID()
			ph.bd.commitBlock.AddPair(dab.GetID(), dab)
		}
		for k, v := range curBlock.redDiffAnticone.GetMap() {
			dab := ph.getBlock(k)
			ph.bd.lastSnapshot.AddOrder(dab)
			dab.SetOrder(startOrder + v.(uint))
			ph.bd.commitOrder[dab.GetOrder()] = dab.GetID()
			ph.bd.commitBlock.AddPair(dab.GetID(), dab)
		}
		startOrder = curBlock.GetOrder()
	}
	//
}

func (ph *Phantom) UpdateVirtualBlockOrder() *PhantomBlock {
	if ph.diffAnticone.IsEmpty() ||
		ph.virtualBlock.GetOrder() != MaxBlockOrder {
		return nil
	}
	ph.virtualBlock.parents = NewIdSet()
	var maxLayer uint = 0
	for k := range ph.bd.tips.GetMap() {
		parent := ph.bd.getBlockById(k)
		ph.virtualBlock.parents.AddPair(k, parent)

		if maxLayer == 0 || maxLayer < parent.GetLayer() {
			maxLayer = parent.GetLayer()
		}
	}
	ph.virtualBlock.SetLayer(maxLayer + 1)

	tp := ph.getBlock(ph.mainChain.tip)
	ph.virtualBlock.mainParent = ph.mainChain.tip
	ph.virtualBlock.blueNum = tp.blueNum + 1
	ph.virtualBlock.height = tp.height + 1
	//ph.virtualBlock.weight = tp.GetWeight()

	ph.virtualBlock.blueDiffAnticone.Clean()
	ph.virtualBlock.redDiffAnticone.Clean()

	ph.calculateBlueSet(ph.virtualBlock, ph.diffAnticone)
	ph.updateBlockOrder(ph.virtualBlock)

	startOrder := ph.getBlock(ph.mainChain.tip).GetOrder()
	for k, v := range ph.virtualBlock.blueDiffAnticone.GetMap() {
		dab := ph.getBlock(k)
		ph.bd.lastSnapshot.AddOrder(dab)
		dab.SetOrder(startOrder + v.(uint))
		ph.bd.commitOrder[dab.GetOrder()] = dab.GetID()
		ph.bd.commitBlock.AddPair(dab.GetID(), dab)
	}
	for k, v := range ph.virtualBlock.redDiffAnticone.GetMap() {
		dab := ph.getBlock(k)
		ph.bd.lastSnapshot.AddOrder(dab)
		dab.SetOrder(startOrder + v.(uint))
		ph.bd.commitOrder[dab.GetOrder()] = dab.GetID()
		ph.bd.commitBlock.AddPair(dab.GetID(), dab)
	}
	ph.bd.lastSnapshot.AddOrder(ph.virtualBlock)
	ph.virtualBlock.SetOrder(ph.bd.blockTotal + 1)

	return ph.getBlock(ph.mainChain.tip)
}

func (ph *Phantom) preUpdateVirtualBlock() *PhantomBlock {
	if ph.diffAnticone.IsEmpty() ||
		ph.virtualBlock.GetOrder() != MaxBlockOrder {
		return nil
	}
	for k := range ph.diffAnticone.GetMap() {
		dab := ph.getBlock(k)
		ph.bd.lastSnapshot.AddOrder(dab)
		dab.SetOrder(MaxBlockOrder)
		ph.bd.commitBlock.AddPair(dab.GetID(), dab)
	}
	return nil
}

func (ph *Phantom) GetDiffBlueSet() *IdSet {
	if ph.mainChain.tip == MaxId {
		return nil
	}
	ph.UpdateVirtualBlockOrder()
	result := NewIdSet()
	curPb := ph.getBlock(ph.mainChain.tip)
	for {
		result.AddSet(curPb.blueDiffAnticone)
		if curPb.mainParent == MaxId {
			break
		}
		curPb = ph.getBlock(curPb.mainParent)
	}

	if ph.virtualBlock.GetOrder() != MaxBlockOrder {
		result.AddSet(ph.virtualBlock.blueDiffAnticone)
	}
	return result
}

// If the successor return nil, the underlying layer will use the default tips list.
func (ph *Phantom) GetTipsList() []IBlock {
	return nil
}

// Query whether a given block is on the main chain.
func (ph *Phantom) IsOnMainChain(b IBlock) bool {
	if ph.mainChain.Has(b.GetID()) {
		return true
	}
	return false
}

func (ph *Phantom) getOrderChangeList(pb *PhantomBlock) *list.List {
	refNodes := list.New()
	if ph.bd.blockTotal == 1 {
		refNodes.PushBack(ph.bd.getGenesis())
		return refNodes
	}
	if pb != nil {
		tips := ph.bd.tips
		if tips.HasOnly(pb.GetID()) {
			refNodes.PushBack(pb)
			return refNodes
		}
		if pb.GetID() == ph.mainChain.tip {
			refNodes.PushBack(pb)
		} else if pb.IsOrdered() && pb.GetOrder() <= ph.GetMainChainTip().GetOrder() {
			for i := ph.GetMainChainTip().GetOrder(); i >= 0; i-- {
				curOrderPb, ok := ph.bd.getBlockByOrder(i).(*PhantomBlock)
				if ok {
					refNodes.PushFront(curOrderPb)
					if curOrderPb.GetID() == pb.GetID() {
						break
					}
				}
			}
		}
	}
	if !ph.diffAnticone.IsEmpty() {
		for _, v := range ph.diffAnticone.GetMap() {
			ib := v.(IBlock)
			refNodes.PushBack(ib)
		}
	}
	return refNodes
}

// return the tip of main chain
func (ph *Phantom) GetMainChainTip() IBlock {
	return ph.bd.getBlockById(ph.mainChain.tip)
}

func (ph *Phantom) GetMainChainTipId() uint {
	return ph.mainChain.tip
}

// return the main parent in the parents
func (ph *Phantom) GetMainParent(parents *IdSet) IBlock {
	if parents == nil || parents.IsEmpty() {
		return nil
	}
	if parents.Size() == 1 {
		ib := ph.getBlock(parents.List()[0])
		if ib == nil {
			return nil
		}
		return ib
	}
	return ph.getBluest(parents)
}

func (ph *Phantom) getBlock(id uint) *PhantomBlock {
	ib := ph.bd.getBlockById(id)
	if ib == nil {
		return nil
	}
	return ib.(*PhantomBlock)
}

func (ph *Phantom) GetDiffAnticone() *IdSet {
	return ph.diffAnticone
}

// encode
func (ph *Phantom) Encode(w io.Writer) error {
	err := s.WriteElements(w, uint32(ph.anticoneSize))
	if err != nil {
		return err
	}
	return nil
}

// decode
func (ph *Phantom) Decode(r io.Reader) error {
	var anticoneSize uint32
	err := s.ReadElements(r, &anticoneSize)
	if err != nil {
		return err
	}
	if anticoneSize != uint32(ph.anticoneSize) {
		return fmt.Errorf("The anticoneSize (%d) is not the same. (%d)", ph.anticoneSize, anticoneSize)
	}
	return nil
}

// load
func (ph *Phantom) Load(dbTx database.Tx) error {
	tips, err := DBGetDAGTips(dbTx)
	if err != nil {
		return err
	}

	ph.mainChain.genesis = 0
	ph.mainChain.tip = tips[0]

	for i := uint(0); i < ph.bd.blockTotal; i++ {
		block := Block{id: i}
		ib := ph.CreateBlock(&block)
		err := DBGetDAGBlock(dbTx, ib)
		if err != nil {
			if err.(*DAGError).IsEmpty() {
				continue
			}
			return err
		}
		if i == 0 && !ib.GetHash().IsEqual(ph.bd.GetGenesisHash()) {
			return fmt.Errorf("genesis data mismatch")
		}
		// Make up for missing
		if ib.HasParents() {
			parentsSet := NewIdSet()
			for k := range ib.GetParents().GetMap() {
				parent := ph.bd.getBlockById(k)
				parentsSet.AddPair(k, parent)
				parent.AddChild(ib)
			}
			ib.GetParents().Clean()
			ib.GetParents().AddSet(parentsSet)
		}
		ph.bd.blocks[ib.GetID()] = ib

		//
		if !ib.IsOrdered() {
			ph.diffAnticone.AddPair(ib.GetID(), ib)
		} else {
			// check order index
			id, err := DBGetBlockIdByOrder(dbTx, ib.GetOrder())
			if err != nil {
				return err
			}
			if uint(id) != ib.GetID() {
				return fmt.Errorf("The order(%d) of %s is inconsistent: Order Index (%d)\n", ib.GetOrder(), ib.GetHash(), id)
			}
		}
		block.data = ph.bd.getBlockData(ib.GetHash())
	}
	// load tips
	for _, v := range tips {
		tip := ph.getBlock(v)
		if tip == nil {
			return fmt.Errorf("Can't find tip:%d\n", v)
		}
		ph.bd.updateTips(tip)
	}
	ph.bd.optimizeTips(dbTx)

	return ph.CheckMainChainDB(dbTx)
}

func (ph *Phantom) GetBlues(parents *IdSet) uint {
	if parents == nil || parents.IsEmpty() {
		return 0
	}
	for k := range parents.GetMap() {
		if !ph.bd.hasBlockById(k) {
			return 0
		}
	}

	//vb
	vb := &Block{hash: hash.ZeroHash, layer: 0, mainParent: MaxId, parents: parents}
	pb := &PhantomBlock{vb, 0, NewIdSet(), NewIdSet()}

	tp := ph.GetMainParent(parents).(*PhantomBlock)
	pb.mainParent = tp.GetID()
	pb.blueNum = tp.blueNum + 1
	pb.height = tp.height + 1

	diffAnticone := ph.bd.getDiffAnticone(pb, true)
	if diffAnticone == nil {
		diffAnticone = NewIdSet()
	}

	ph.calculateBlueSet(pb, diffAnticone)

	return pb.blueNum
}

func (ph *Phantom) IsBlue(id uint) bool {
	b := ph.getBlock(id)
	if b == nil {
		return false
	}
	return ph.doIsBlue(b, nil)
}

// Functions that really handle isblue.
// fork: Path intersection from block to main chain.
func (ph *Phantom) doIsBlue(ib IBlock, fork IBlock) bool {
	b := ib.(*PhantomBlock)
	if ph.diffAnticone.Has(b.GetID()) {
		return false
	}
	var cur *PhantomBlock
	if fork == nil {
		cur = ph.bd.getMainFork(b, true).(*PhantomBlock)
		if cur == nil {
			cur = ph.getBlock(ph.mainChain.tip)
		}
	} else {
		cur = fork.(*PhantomBlock)
	}

	for ; cur != nil; cur = ph.getBlock(cur.mainParent) {
		if cur.GetHash().IsEqual(b.GetHash()) ||
			cur.blueDiffAnticone.Has(b.GetID()) {
			return true
		}
		if cur.GetLayer() < b.GetLayer() {
			break
		}
		if cur.mainParent == MaxId {
			break
		}
	}
	return false
}

// IsDAG
func (ph *Phantom) IsDAG(parents []IBlock) bool {
	if len(parents) == 0 {
		return false
	} else if len(parents) == 1 {
		return true
	}
	tp := parents[0]
	ops := NewIdSet()
	//
	minTipLayer := uint(math.MaxUint32)
	for _, v := range parents {
		ib := v.(IBlock)
		if ib == nil ||
			ib.GetID() == tp.GetID() {
			continue
		}
		if ib.GetLayer() < minTipLayer {
			minTipLayer = ib.GetLayer()
		}
		ops.Add(ib.GetID())
	}
	//
	mainsubdag := NewIdSet()
	mainsubdag.Add(0)

	mlg := MaxTipLayerGap

	for curMP := tp; curMP != nil && mlg > 0; curMP = ph.bd.getBlockById(curMP.GetMainParent()) {
		mainsubdag.Add(curMP.GetID())
		mainsubdag.AddSet(curMP.(*PhantomBlock).GetBlueDiffAnticone())
		mainsubdag.AddSet(curMP.(*PhantomBlock).GetRedDiffAnticone())

		if curMP.GetLayer() < minTipLayer {
			mlg--
		}
	}

	for k := range ops.GetMap() {
		if mainsubdag.Has(k) {
			return false
		}
	}
	return true
}

// The main parent concurrency of block
func (ph *Phantom) GetMainParentConcurrency(b IBlock) int {
	if !b.HasParents() {
		return 0
	}
	pblock := b.(*PhantomBlock)
	result := pblock.redDiffAnticone.Size() + pblock.blueDiffAnticone.Size() + 1
	return result
}

func (ph *Phantom) getMaxParents() int {
	dagMax := ph.anticoneSize + 1
	if dagMax < types.MaxParentsPerBlock {
		return dagMax
	}
	return types.MaxParentsPerBlock
}

func (ph *Phantom) UpdateWeight(ib IBlock) {
	if ib.GetID() != GenesisId {
		pb := ib.(*PhantomBlock)
		tp := ph.getBlock(pb.GetMainParent())
		pb.weight = tp.GetWeight()

		pb.weight += uint64(ph.bd.calcWeight(pb, ph.bd.getBlueInfo(pb)))
		for k := range pb.blueDiffAnticone.GetMap() {
			bdpb := ph.getBlock(k)
			pb.weight += uint64(ph.bd.calcWeight(bdpb, ph.bd.getBlueInfo(bdpb)))
		}
		ph.bd.commitBlock.AddPair(ib.GetID(), ib)
	}
}

func (ph *Phantom) CheckMainChainDB(dbTx database.Tx) error {
	mainChainM := map[uint]bool{}
	var cur *PhantomBlock
	for cur = ph.getBlock(ph.mainChain.tip); cur != nil; cur = ph.getBlock(cur.mainParent) {
		mainChainM[cur.id] = true
		if cur.mainParent == MaxId {
			break
		}
	}
	if cur.id != 0 {
		return fmt.Errorf("Main chain genesis error:%d %s\n", cur.id, cur.GetHash().String())
	}
	bucket := dbTx.Metadata().Bucket(dbnamespace.DagMainChainBucketName)
	cursor := bucket.Cursor()
	mcLen := int(0)
	for cok := cursor.First(); cok; cok = cursor.Next() {
		id := dbnamespace.ByteOrder.Uint32(cursor.Key())
		_, ok := mainChainM[uint(id)]
		if !ok {
			return fmt.Errorf("Main chain error:invalid %d\n", id)
		}
		mcLen++
	}
	if len(mainChainM) != mcLen {
		return fmt.Errorf("Inconsistent main chain length:%d != %d\n", len(mainChainM), mcLen)
	}
	return nil
}

// The main chain of DAG is support incremental expansion
type MainChain struct {
	bd      *BlockDAG
	tip     uint
	genesis uint

	commitBlocks *IdSet
}

func (mc *MainChain) Has(id uint) bool {
	result := false

	if mc.commitBlocks.Has(id) {
		opt := mc.commitBlocks.Get(id)
		add, ok := opt.(bool)
		if ok {
			if add {
				return true
			}
		}
	}

	mc.bd.db.View(func(dbTx database.Tx) error {
		meta := dbTx.Metadata()
		mchBucket := meta.Bucket(dbnamespace.DagMainChainBucketName)
		if mchBucket == nil {
			return nil
		}
		result = DBHasMainChainBlock(dbTx, id)
		return nil
	})
	return result
}

func (mc *MainChain) Add(id uint) error {
	return mc.bd.db.Update(func(dbTx database.Tx) error {
		meta := dbTx.Metadata()
		mchBucket := meta.Bucket(dbnamespace.DagMainChainBucketName)
		if mchBucket == nil {
			return fmt.Errorf("no %s", string(dbnamespace.DagMainChainBucketName))
		}
		return DBPutMainChainBlock(dbTx, id)
	})
}

func (mc *MainChain) Remove(id uint) error {
	if !mc.Has(id) {
		return nil
	}
	return mc.bd.db.Update(func(dbTx database.Tx) error {
		meta := dbTx.Metadata()
		mchBucket := meta.Bucket(dbnamespace.DagMainChainBucketName)
		if mchBucket == nil {
			return fmt.Errorf("no %s", string(dbnamespace.DagMainChainBucketName))
		}
		return DBRemoveMainChainBlock(dbTx, id)
	})
}

func (mc *MainChain) commit() error {
	if mc.commitBlocks.IsEmpty() {
		return nil
	}
	for k, v := range mc.commitBlocks.GetMap() {
		opt, ok := v.(bool)
		if !ok || !opt {
			mc.Remove(k)
		} else {
			mc.Add(k)
		}
	}
	mc.commitBlocks.Clean()

	return nil
}

type KChain struct {
	blocks    *IdSet
	miniLayer uint
}
