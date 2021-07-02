package blockdag

import (
	"bytes"
	"container/list"
	"fmt"
	"github.com/Qitmeer/qitmeer/common/hash"
	"github.com/Qitmeer/qitmeer/common/roughtime"
	"github.com/Qitmeer/qitmeer/core/blockdag/anticone"
	"github.com/Qitmeer/qitmeer/core/dbnamespace"
	"github.com/Qitmeer/qitmeer/core/merkle"
	s "github.com/Qitmeer/qitmeer/core/serialization"
	"github.com/Qitmeer/qitmeer/database"
	"io"
	"math"
	"sort"
	"sync"
	"time"
)

// Some available DAG algorithm types
const (
	// A Scalable BlockDAG protocol
	phantom = "phantom"

	// Phantom protocol V2
	phantom_v2 = "phantom_v2"

	// The order of all transactions is solely determined by the Tree Graph (TG)
	conflux = "conflux"

	// Confirming Transactions via Recursive Elections
	spectre = "spectre"
)

// Maximum number of the DAG tip
const MaxTips = 100

// Maximum order of the DAG block
const MaxBlockOrder = uint(^uint32(0))

// Maximum id of the DAG block
const MaxId = uint(math.MaxUint32)

// Genesis id of the DAG block
const GenesisId = uint(0)

// MaxTipLayerGap
const MaxTipLayerGap = 10

// StableConfirmations
const StableConfirmations = 10

// It will create different BlockDAG instances
func NewBlockDAG(dagType string) IBlockDAG {
	switch dagType {
	case phantom:
		return &Phantom{}
	case phantom_v2:
		return &Phantom_v2{}
	case conflux:
		return &Conflux{}
	case spectre:
		return &Spectre{}
	}
	return nil
}

func GetDAGTypeIndex(dagType string) byte {
	switch dagType {
	case phantom:
		return 0
	case phantom_v2:
		return 1
	case conflux:
		return 2
	case spectre:
		return 3
	}
	return 0
}

func GetDAGTypeByIndex(dagType byte) string {
	switch dagType {
	case 0:
		return phantom
	case 1:
		return phantom_v2
	case 2:
		return conflux
	case 3:
		return spectre
	}
	return phantom
}

// The abstract inferface is used to build and manager DAG
type IBlockDAG interface {
	// Return the name
	GetName() string

	// This instance is initialized and will be executed first.
	Init(bd *BlockDAG) bool

	// Add a block
	AddBlock(ib IBlock) (*list.List, *list.List)

	// Build self block
	CreateBlock(b *Block) IBlock

	// If the successor return nil, the underlying layer will use the default tips list.
	GetTipsList() []IBlock

	// Query whether a given block is on the main chain.
	IsOnMainChain(ib IBlock) bool

	// return the tip of main chain
	GetMainChainTip() IBlock

	// return the tip of main chain id
	GetMainChainTipId() uint

	// return the main parent in the parents
	GetMainParent(parents *IdSet) IBlock

	// encode
	Encode(w io.Writer) error

	// decode
	Decode(r io.Reader) error

	// load
	Load(dbTx database.Tx) error

	// IsDAG
	IsDAG(parents []IBlock) bool

	// The main parent concurrency of block
	GetMainParentConcurrency(b IBlock) int

	// GetBlues
	GetBlues(parents *IdSet) uint

	// IsBlue
	IsBlue(id uint) bool

	// getMaxParents
	getMaxParents() int
}

// CalcWeight
type CalcWeight func(int64, *hash.Hash, BlockStatus) int64

type GetBlockData func(*hash.Hash) IBlockData

// The general foundation framework of DAG
type BlockDAG struct {
	// The genesis of block dag
	genesis hash.Hash

	// Use block hash to save all blocks with mapping
	blocks map[uint]IBlock

	// The total number blocks that this dag currently owned
	blockTotal uint

	// The terminal block is in block dag,this block have not any connecting at present.
	tips *IdSet

	// This is time when the last block have added
	lastTime time.Time

	// All orders relate to new block will be committed that need to be consensus
	commitOrder map[uint]uint

	commitBlock *IdSet

	// Current dag instance used. Different algorithms work according to
	// different dag types config.
	instance IBlockDAG

	// state lock
	stateLock sync.RWMutex

	//
	calcWeight CalcWeight

	getBlockData GetBlockData

	// blocks per second
	blockRate float64

	db database.DB
}

// Acquire the name of DAG instance
func (bd *BlockDAG) GetName() string {
	return bd.instance.GetName()
}

// GetInstance
func (bd *BlockDAG) GetInstance() IBlockDAG {
	return bd.instance
}

// Initialize self, the function to be invoked at the beginning
func (bd *BlockDAG) Init(dagType string, calcWeight CalcWeight, blockRate float64, db database.DB, getBlockData GetBlockData) IBlockDAG {
	bd.lastTime = time.Unix(roughtime.Now().Unix(), 0)
	bd.commitOrder = map[uint]uint{}
	bd.calcWeight = calcWeight
	bd.getBlockData = getBlockData
	bd.db = db
	bd.commitBlock = NewIdSet()
	bd.blockRate = blockRate
	if bd.blockRate < 0 {
		bd.blockRate = anticone.DefaultBlockRate
	}
	bd.instance = NewBlockDAG(dagType)
	bd.instance.Init(bd)

	err := db.Update(func(dbTx database.Tx) error {
		meta := dbTx.Metadata()
		serializedData := meta.Get(dbnamespace.DagInfoBucketName)
		if serializedData == nil {
			DBPutDAGInfo(dbTx, bd)
			// Create the bucket that houses the block index data.
			_, err := meta.CreateBucket(dbnamespace.BlockIndexBucketName)
			if err != nil {
				return err
			}

			// Create the bucket that houses the chain block order to hash
			// index.
			_, err = meta.CreateBucket(dbnamespace.OrderIdBucketName)
			if err != nil {
				return err
			}

			_, err = meta.CreateBucket(dbnamespace.DagMainChainBucketName)
			if err != nil {
				return err
			}

			_, err = meta.CreateBucket(dbnamespace.BlockIdBucketName)
			if err != nil {
				return err
			}
		}
		return nil
	})
	if err != nil {
		log.Error(err.Error())
		return nil
	}
	return bd.instance
}

// This is an entry for update the block dag,you need pass in a block parameter,
// If add block have failure,it will return false.
func (bd *BlockDAG) AddBlock(b IBlockData) (*list.List, *list.List, IBlock, bool) {
	bd.stateLock.Lock()
	defer bd.stateLock.Unlock()

	if b == nil {
		return nil, nil, nil, false
	}
	// Must keep no block in outside.
	/*	if bd.hasBlock(b.GetHash()) {
		return nil
	}*/
	parents := []IBlock{}
	if bd.blockTotal > 0 {
		parentsIds := b.GetParents()
		if len(parentsIds) == 0 {
			return nil, nil, nil, false
		}
		for _, v := range parentsIds {
			pib := bd.getBlock(v)
			if pib == nil {
				return nil, nil, nil, false
			}
			parents = append(parents, pib)
		}

		if !bd.isDAG(parents) {
			return nil, nil, nil, false
		}
	}
	lastMT := bd.instance.GetMainChainTipId()
	//
	block := Block{id: bd.blockTotal, hash: *b.GetHash(), layer: 0, status: StatusNone, mainParent: MaxId, data: b}

	if bd.blocks == nil {
		bd.blocks = map[uint]IBlock{}
	}
	ib := bd.instance.CreateBlock(&block)
	bd.blocks[block.id] = ib

	// db
	bd.commitBlock.AddPair(ib.GetID(), ib)

	if bd.db.Update(func(dbTx database.Tx) error {
		return DBPutDAGBlockIdByHash(dbTx, ib)
	}) != nil {
		return nil, nil, nil, false
	}

	//
	if bd.blockTotal == 0 {
		bd.genesis = *block.GetHash()
	}
	//
	bd.blockTotal++

	if len(parents) > 0 {
		block.parents = NewIdSet()
		var maxLayer uint = 0
		for _, v := range parents {
			parent := v.(IBlock)
			block.parents.AddPair(parent.GetID(), parent)
			parent.AddChild(ib)
			if block.mainParent > parent.GetID() {
				block.mainParent = parent.GetID()
			}

			if maxLayer == 0 || maxLayer < parent.GetLayer() {
				maxLayer = parent.GetLayer()
			}
		}
		block.SetLayer(maxLayer + 1)
	}

	//
	bd.updateTips(ib)
	//
	t := time.Unix(b.GetTimestamp(), 0)
	if bd.lastTime.Before(t) {
		bd.lastTime = t
	}
	//
	news, olds := bd.instance.AddBlock(ib)
	bd.optimizeReorganizeResult(news, olds)
	if news == nil {
		news = list.New()
	}
	if olds == nil {
		olds = list.New()
	}
	return news, olds, ib, lastMT != bd.instance.GetMainChainTipId()
}

// Acquire the genesis block of chain
func (bd *BlockDAG) getGenesis() IBlock {
	return bd.getBlockById(GenesisId)
}

// Acquire the genesis block hash of chain
func (bd *BlockDAG) GetGenesisHash() *hash.Hash {
	return &bd.genesis
}

// If the block is illegal dag,will return false.
// Exclude genesis block
func (bd *BlockDAG) isDAG(parents []IBlock) bool {
	return bd.checkLayerGap(parents) &&
		bd.checkLegality(parents) &&
		bd.instance.IsDAG(parents)
}

// Is there a block in DAG?
func (bd *BlockDAG) HasBlock(h *hash.Hash) bool {
	return bd.GetBlockId(h) != MaxId
}

// Is there a block in DAG?
func (bd *BlockDAG) hasBlockById(id uint) bool {
	return bd.getBlockById(id) != nil
}

// Is there a block in DAG?
func (bd *BlockDAG) HasBlockById(id uint) bool {
	bd.stateLock.Lock()
	defer bd.stateLock.Unlock()

	return bd.hasBlockById(id)
}

// Is there some block in DAG?
func (bd *BlockDAG) hasBlocks(ids []uint) bool {
	for _, id := range ids {
		if !bd.hasBlockById(id) {
			return false
		}
	}
	return true
}

// Acquire one block by hash
func (bd *BlockDAG) GetBlock(h *hash.Hash) IBlock {
	bd.stateLock.Lock()
	defer bd.stateLock.Unlock()

	return bd.getBlock(h)
}

// Acquire one block by hash
// Be careful, this is inefficient and cannot be called frequently
func (bd *BlockDAG) getBlock(h *hash.Hash) IBlock {
	return bd.getBlockById(bd.getBlockId(h))
}

func (bd *BlockDAG) GetBlockId(h *hash.Hash) uint {
	bd.stateLock.Lock()
	defer bd.stateLock.Unlock()

	return bd.getBlockId(h)
}

func (bd *BlockDAG) getBlockId(h *hash.Hash) uint {
	if h == nil {
		return MaxId
	}
	id := MaxId
	err := bd.db.View(func(dbTx database.Tx) error {
		bid, er := DBGetBlockIdByHash(dbTx, h)
		if er == nil {
			id = uint(bid)
		}
		return er
	})
	if err != nil {
		return MaxId
	}
	return id
}

// Acquire one block by hash
func (bd *BlockDAG) GetBlockById(id uint) IBlock {
	bd.stateLock.Lock()
	defer bd.stateLock.Unlock()

	return bd.getBlockById(id)
}

// Acquire one block by id
func (bd *BlockDAG) getBlockById(id uint) IBlock {
	if id == MaxId {
		return nil
	}
	block, ok := bd.blocks[id]
	if !ok {
		return nil
	}
	return block
}

// Total number of blocks
func (bd *BlockDAG) GetBlockTotal() uint {
	bd.stateLock.Lock()
	defer bd.stateLock.Unlock()
	return bd.blockTotal
}

// return the terminal blocks, because there maybe more than one, so this is a set.
func (bd *BlockDAG) GetTips() *HashSet {
	bd.stateLock.Lock()
	defer bd.stateLock.Unlock()

	tips := NewHashSet()
	for k := range bd.tips.GetMap() {
		ib := bd.getBlockById(k)
		tips.AddPair(ib.GetHash(), ib)
	}
	return tips
}

// Acquire the tips array of DAG
func (bd *BlockDAG) GetTipsList() []IBlock {
	bd.stateLock.Lock()
	defer bd.stateLock.Unlock()

	result := bd.instance.GetTipsList()
	if result != nil {
		return result
	}
	result = []IBlock{}
	for k := range bd.tips.GetMap() {
		result = append(result, bd.getBlockById(k))
	}
	return result
}

// build merkle tree form current DAG tips
func (bd *BlockDAG) BuildMerkleTreeStoreFromTips() []*hash.Hash {
	parents := bd.GetTips().SortList(false)
	return merkle.BuildParentsMerkleTreeStore(parents)
}

// Refresh the dag tip with new block,it will cause changes in tips set.
func (bd *BlockDAG) updateTips(b IBlock) {
	if bd.tips == nil {
		bd.tips = NewIdSet()
		bd.tips.AddPair(b.GetID(), b)
		return
	}
	for k, v := range bd.tips.GetMap() {
		block := v.(IBlock)
		if block.HasChildren() {
			bd.tips.Remove(k)
		}
	}
	bd.tips.AddPair(b.GetID(), b)
}

// The last time is when add one block to DAG.
func (bd *BlockDAG) GetLastTime() *time.Time {
	bd.stateLock.Lock()
	defer bd.stateLock.Unlock()

	return &bd.lastTime
}

// Obtain block hash by global order
func (bd *BlockDAG) GetBlockHashByOrder(order uint) *hash.Hash {
	bd.stateLock.Lock()
	defer bd.stateLock.Unlock()

	ib := bd.getBlockByOrder(order)
	if ib != nil {
		return ib.GetHash()
	}
	return nil
}

func (bd *BlockDAG) GetBlockByOrder(order uint) IBlock {
	bd.stateLock.Lock()
	defer bd.stateLock.Unlock()

	return bd.getBlockByOrder(order)
}

func (bd *BlockDAG) GetBlockByOrderWithTx(dbTx database.Tx, order uint) *hash.Hash {
	bd.stateLock.Lock()
	defer bd.stateLock.Unlock()

	ib := bd.doGetBlockByOrder(dbTx, order)
	if ib != nil {
		return ib.GetHash()
	}
	return nil
}

func (bd *BlockDAG) getBlockByOrder(order uint) IBlock {
	return bd.doGetBlockByOrder(nil, order)
}

func (bd *BlockDAG) doGetBlockByOrder(dbTx database.Tx, order uint) IBlock {
	if order >= MaxBlockOrder {
		return nil
	}
	id, ok := bd.commitOrder[order]
	if ok {
		return bd.getBlockById(id)
	}

	bid := uint(MaxId)

	if dbTx == nil {
		err := bd.db.View(func(dbTx database.Tx) error {
			id, er := DBGetBlockIdByOrder(dbTx, order)
			if er == nil {
				bid = uint(id)
			}
			return er
		})
		if err != nil {
			log.Error(err.Error())
			return nil
		}
	} else {
		id, er := DBGetBlockIdByOrder(dbTx, order)
		if er == nil {
			bid = uint(id)
		} else {
			return nil
		}
	}

	return bd.getBlockById(bid)
}

// Return the last order block
func (bd *BlockDAG) GetLastBlock() IBlock {
	// TODO
	return bd.GetMainChainTip()
}

// This function need a stable sequence,so call it before sorting the DAG.
// If the h is invalid,the function will become a little inefficient.
func (bd *BlockDAG) GetPrevious(id uint) (uint, error) {
	bd.stateLock.Lock()
	defer bd.stateLock.Unlock()

	if id == 0 {
		return 0, fmt.Errorf("no pre")
	}
	b := bd.getBlockById(id)
	if b == nil {
		return 0, fmt.Errorf("no pre")
	}
	if b.GetOrder() == 0 {
		return 0, fmt.Errorf("no pre")
	}
	// TODO
	ib := bd.getBlockByOrder(b.GetOrder() - 1)
	if ib != nil {
		return ib.GetID(), nil
	}
	return 0, fmt.Errorf("no pre")
}

// Returns a future collection of block. This function is a recursively called function
// So we should consider its efficiency.
func (bd *BlockDAG) getFutureSet(fs *IdSet, b IBlock) {
	children := b.GetChildren()
	if children == nil || children.IsEmpty() {
		return
	}
	for k, v := range children.GetMap() {
		ib := v.(IBlock)
		if !fs.Has(k) {
			fs.AddPair(k, ib)
			bd.getFutureSet(fs, ib)
		}
	}
}

// Query whether a given block is on the main chain.
// Note that some DAG protocols may not support this feature.
func (bd *BlockDAG) IsOnMainChain(id uint) bool {
	bd.stateLock.Lock()
	defer bd.stateLock.Unlock()

	return bd.isOnMainChain(id)
}

// Query whether a given block is on the main chain.
// Note that some DAG protocols may not support this feature.
func (bd *BlockDAG) isOnMainChain(id uint) bool {
	return bd.instance.IsOnMainChain(bd.getBlockById(id))
}

// return the tip of main chain
func (bd *BlockDAG) GetMainChainTip() IBlock {
	bd.stateLock.Lock()
	defer bd.stateLock.Unlock()

	return bd.getMainChainTip()
}

// return the tip of main chain
func (bd *BlockDAG) getMainChainTip() IBlock {
	return bd.instance.GetMainChainTip()
}

// return the main parent in the parents
func (bd *BlockDAG) GetMainParent(parents *IdSet) IBlock {
	bd.stateLock.Lock()
	defer bd.stateLock.Unlock()

	return bd.instance.GetMainParent(parents)
}

// return the main parent in the parents
func (bd *BlockDAG) GetMainParentByHashs(parents []*hash.Hash) IBlock {
	bd.stateLock.Lock()
	defer bd.stateLock.Unlock()

	parentsSet := NewIdSet()
	for _, p := range parents {
		ib := bd.getBlock(p)
		if ib == nil {
			return nil
		}
		parentsSet.AddPair(ib.GetID(), ib)
	}
	return bd.instance.GetMainParent(parentsSet)
}

// Return the layer of block,it is stable.
// You can imagine that this is the main chain.
func (bd *BlockDAG) GetLayer(id uint) uint {
	return bd.GetBlockById(id).GetLayer()
}

// Return current general description of the whole state of DAG
func (bd *BlockDAG) GetGraphState() *GraphState {
	bd.stateLock.Lock()
	defer bd.stateLock.Unlock()
	return bd.getGraphState()
}

// Return current general description of the whole state of DAG
func (bd *BlockDAG) getGraphState() *GraphState {
	gs := NewGraphState()
	gs.SetLayer(0)

	tips := bd.getValidTips(false)
	for i := 0; i < len(tips); i++ {
		tip := tips[i]
		if i == 0 {
			gs.GetTips().AddPair(tip.GetHash(), true)
		} else {
			gs.GetTips().Add(tip.GetHash())
		}
		if tip.GetLayer() > gs.GetLayer() {
			gs.SetLayer(tip.GetLayer())
		}
	}
	gs.SetTotal(bd.blockTotal)
	gs.SetMainHeight(bd.getMainChainTip().GetHeight())
	gs.SetMainOrder(bd.getMainChainTip().GetOrder())
	return gs
}

// Locate all eligible block by current graph state.
func (bd *BlockDAG) locateBlocks(gs *GraphState, maxHashes uint) []*hash.Hash {
	if gs.IsExcellent(bd.getGraphState()) {
		return nil
	}
	queue := []IBlock{}
	fs := NewHashSet()
	tips := bd.getValidTips(false)
	queue = append(queue, tips...)
	for len(queue) > 0 {
		cur := queue[0]
		queue = queue[1:]

		if fs.Has(cur.GetHash()) {
			continue
		}
		if gs.GetTips().Has(cur.GetHash()) || cur.GetID() == 0 {
			continue
		}
		needRec := true
		if cur.HasChildren() {
			for _, v := range cur.GetChildren().GetMap() {
				ib := v.(IBlock)
				if gs.GetTips().Has(ib.GetHash()) || !fs.Has(ib.GetHash()) && ib.IsOrdered() {
					needRec = false
					break
				}
			}
		}
		if needRec {
			fs.AddPair(cur.GetHash(), cur)
			if cur.HasParents() {
				for _, v := range cur.GetParents().GetMap() {
					value := v.(IBlock)
					ib := value
					if fs.Has(ib.GetHash()) {
						continue
					}
					queue = append(queue, ib)

				}
			}
		}
	}

	fsSlice := BlockSlice{}
	for _, v := range fs.GetMap() {
		value := v.(IBlock)
		ib := value
		if gs.GetTips().Has(ib.GetHash()) {
			continue
		}
		if ib.HasChildren() {
			need := true
			for _, v := range ib.GetChildren().GetMap() {
				ib := v.(IBlock)
				if gs.GetTips().Has(ib.GetHash()) {
					need = false
					break
				}
			}
			if !need {
				continue
			}
		}
		fsSlice = append(fsSlice, ib)
	}

	result := []*hash.Hash{}
	if len(fsSlice) >= 2 {
		sort.Sort(fsSlice)
	}
	for i := 0; i < len(fsSlice); i++ {
		if maxHashes > 0 && i >= int(maxHashes) {
			break
		}
		result = append(result, fsSlice[i].GetHash())
	}
	return result
}

// Judging whether block is the virtual tip that it have not future set.
func isVirtualTip(bs *IdSet, futureSet *IdSet, anticone *IdSet, children *IdSet) bool {
	for k := range children.GetMap() {
		if bs.Has(k) {
			return false
		}
		if !futureSet.Has(k) && !anticone.Has(k) {
			return false
		}
	}
	return true
}

// This function is used to GetAnticone recursion
func (bd *BlockDAG) recAnticone(bs *IdSet, futureSet *IdSet, anticone *IdSet, ib IBlock) {
	if bs.Has(ib.GetID()) || anticone.Has(ib.GetID()) {
		return
	}
	children := ib.GetChildren()
	needRecursion := false
	if children == nil || children.Size() == 0 {
		needRecursion = true
	} else {
		needRecursion = isVirtualTip(bs, futureSet, anticone, children)
	}
	if needRecursion {
		if !futureSet.Has(ib.GetID()) {
			anticone.AddPair(ib.GetID(), ib)
		}
		parents := ib.GetParents()

		//Because parents can not be empty, so there is no need to judge.
		for _, v := range parents.GetMap() {
			pib := v.(IBlock)
			bd.recAnticone(bs, futureSet, anticone, pib)
		}
	}
}

// This function can get anticone set for an block that you offered in the block dag,If
// the exclude set is not empty,the final result will exclude set that you passed in.
func (bd *BlockDAG) getAnticone(b IBlock, exclude *IdSet) *IdSet {
	futureSet := NewIdSet()
	bd.getFutureSet(futureSet, b)
	anticone := NewIdSet()
	bs := NewIdSet()
	bs.AddPair(b.GetID(), b)
	for _, v := range bd.tips.GetMap() {
		ib := v.(IBlock)
		bd.recAnticone(bs, futureSet, anticone, ib)
	}
	if exclude != nil {
		anticone.Exclude(exclude)
	}
	return anticone
}

// getParentsAnticone
func (bd *BlockDAG) getParentsAnticone(parents *IdSet) *IdSet {
	anticone := NewIdSet()
	for _, v := range bd.tips.GetMap() {
		ib := v.(IBlock)
		bd.recAnticone(parents, NewIdSet(), anticone, ib)
	}
	return anticone
}

// getTreeTips
func getTreeTips(root IBlock, mainsubdag *IdSet, genealogy *IdSet) *IdSet {
	allmainsubdag := mainsubdag.Clone()
	queue := []IBlock{}
	for _, v := range root.GetParents().GetMap() {
		ib := v.(IBlock)
		queue = append(queue, ib)
		if genealogy != nil {
			genealogy.Add(ib.GetID())
		}
	}
	startQueue := queue
	for len(queue) > 0 {
		cur := queue[0]
		queue = queue[1:]

		if allmainsubdag.Has(cur.GetID()) {
			allmainsubdag.AddSet(cur.GetParents())
		}
		if !cur.HasParents() {
			continue
		}
		for _, v := range cur.GetParents().GetMap() {
			ib := v.(IBlock)
			queue = append(queue, ib)
		}
	}

	queue = startQueue
	tips := NewIdSet()
	for len(queue) > 0 {
		cur := queue[0]
		queue = queue[1:]

		if !allmainsubdag.Has(cur.GetID()) {
			if !cur.HasParents() {
				tips.AddPair(cur.GetID(), cur)
			}
			if genealogy != nil {
				genealogy.Add(cur.GetID())
			}
		}
		if !cur.HasParents() {
			continue
		}
		for _, v := range cur.GetParents().GetMap() {
			ib := v.(IBlock)
			queue = append(queue, ib)
		}
	}
	return tips
}

// getDiffAnticone
func (bd *BlockDAG) getDiffAnticone(b IBlock, verbose bool) *IdSet {
	if b.GetMainParent() == MaxId {
		return nil
	}
	parents := b.GetParents()
	if parents == nil || parents.Size() <= 1 {
		return nil
	}
	rootBlock := &Block{id: b.GetID(), hash: *b.GetHash(), parents: NewIdSet(), mainParent: MaxId}
	// find anticone
	anticone := NewIdSet()
	mainsubdag := NewIdSet()
	mainsubdag.Add(0)
	mainsubdagTips := NewIdSet()

	for _, v := range parents.GetMap() {
		ib := v.(IBlock)
		cur := &Block{id: ib.GetID(), hash: *ib.GetHash(), parents: NewIdSet(), mainParent: MaxId}
		if ib.GetID() == b.GetMainParent() {
			mainsubdag.Add(ib.GetID())
			mainsubdagTips.AddPair(ib.GetID(), ib)
		} else {
			rootBlock.parents.AddPair(cur.GetID(), cur)
			anticone.AddPair(cur.GetID(), cur)
		}
	}

	anticoneTips := getTreeTips(rootBlock, mainsubdag, nil)
	newmainsubdagTips := NewIdSet()

	for i := 0; i <= MaxTipLayerGap+1; i++ {

		for _, v := range mainsubdagTips.GetMap() {
			ib := v.(IBlock)
			if ib.HasParents() {
				for _, pv := range ib.GetParents().GetMap() {
					pib := pv.(IBlock)
					if mainsubdag.Has(pib.GetID()) {
						continue
					}
					mainsubdag.Add(pib.GetID())
					newmainsubdagTips.AddPair(pib.GetID(), pib)
				}
			}
			mainsubdagTips.Remove(ib.GetID())
		}
		mainsubdagTips.AddSet(newmainsubdagTips)

		if mainsubdagTips.Size() == 0 {
			break
		}
	}

	for anticoneTips.Size() > 0 {

		for _, v := range mainsubdagTips.GetMap() {
			ib := v.(IBlock)
			if ib.HasParents() {
				for _, pv := range ib.GetParents().GetMap() {
					pib := pv.(IBlock)
					if mainsubdag.Has(pib.GetID()) {
						continue
					}
					mainsubdag.Add(pib.GetID())
					newmainsubdagTips.AddPair(pib.GetID(), pib)
				}
			}
			mainsubdagTips.Remove(ib.GetID())
		}
		mainsubdagTips.AddSet(newmainsubdagTips)

		anticoneTips = getTreeTips(rootBlock, mainsubdag, nil)
		//
		for _, v := range anticoneTips.GetMap() {
			tb := v.(*Block)
			realib := bd.getBlockById(tb.GetID())
			if realib.HasParents() {
				for _, pv := range realib.GetParents().GetMap() {
					pib := pv.(IBlock)
					var cur *Block
					if anticone.Has(pib.GetID()) {
						cur = anticone.Get(pib.GetID()).(*Block)
					} else {
						cur = &Block{id: pib.GetID(), hash: *pib.GetHash(), parents: NewIdSet(), mainParent: MaxId}
						anticone.AddPair(cur.GetID(), cur)
					}
					tb.parents.AddPair(cur.GetID(), cur)
				}
			}
		}
		anticoneTips = getTreeTips(rootBlock, mainsubdag, nil)
	}
	result := NewIdSet()
	getTreeTips(rootBlock, mainsubdag, result)

	//
	if verbose && !result.IsEmpty() {
		optimizeDiffAnt := NewIdSet()
		for k := range result.GetMap() {
			optimizeDiffAnt.AddPair(k, bd.getBlockById(k))
		}
		return optimizeDiffAnt
	}
	return result
}

// Sort block by id
func (bd *BlockDAG) sortBlock(src []*hash.Hash) []*hash.Hash {

	if len(src) <= 1 {
		return src
	}
	srcBlockS := BlockSlice{}
	for i := 0; i < len(src); i++ {
		ib := bd.getBlock(src[i])
		if ib != nil {
			srcBlockS = append(srcBlockS, ib)
		}
	}
	if len(srcBlockS) >= 2 {
		sort.Sort(srcBlockS)
	}
	result := []*hash.Hash{}
	for i := 0; i < len(srcBlockS); i++ {
		result = append(result, srcBlockS[i].GetHash())
	}
	return result
}

// Sort block by id
func (bd *BlockDAG) SortBlock(src []*hash.Hash) []*hash.Hash {
	bd.stateLock.Lock()
	defer bd.stateLock.Unlock()

	return bd.sortBlock(src)
}

// GetConfirmations
func (bd *BlockDAG) GetConfirmations(id uint) uint {
	bd.stateLock.Lock()
	defer bd.stateLock.Unlock()

	block := bd.getBlockById(id)
	if block == nil {
		return 0
	}
	if block.GetOrder() > bd.getMainChainTip().GetOrder() {
		return 0
	}
	mainTip := bd.getMainChainTip()
	if bd.isOnMainChain(id) {
		return mainTip.GetHeight() - block.GetHeight()
	}
	if !block.HasChildren() {
		return 0
	}
	//
	queue := []IBlock{}
	queue = append(queue, block)

	for len(queue) > 0 {
		cur := queue[0]
		queue = queue[1:]

		if bd.isOnMainChain(cur.GetID()) {
			return 1 + mainTip.GetHeight() - cur.GetHeight()
		}
		if !cur.HasChildren() {
			continue
		} else {
			childList := cur.GetChildren().SortHashList(false)
			for _, v := range childList {
				ib := cur.GetChildren().Get(v).(IBlock)
				queue = append(queue, ib)
			}
		}
	}
	return 0
}

func (bd *BlockDAG) GetBlockHash(id uint) *hash.Hash {
	bd.stateLock.Lock()
	defer bd.stateLock.Unlock()

	ib := bd.getBlockById(id)
	if ib != nil {
		return ib.GetHash()
	}
	return nil
}

func (bd *BlockDAG) GetValidTips() []*hash.Hash {
	bd.stateLock.Lock()
	defer bd.stateLock.Unlock()
	tips := bd.getValidTips(true)

	result := []*hash.Hash{}
	for _, v := range tips {
		result = append(result, v.GetHash())
	}
	return result
}

func (bd *BlockDAG) getValidTips(limit bool) []IBlock {
	temp := bd.tips.Clone()
	mainParent := bd.getMainChainTip()
	temp.Remove(mainParent.GetID())
	var parents []uint
	if temp.Size() > 1 {
		parents = temp.SortHashList(false)
	} else {
		parents = temp.List()
	}

	tips := []IBlock{mainParent}
	for i := 0; i < len(parents); i++ {
		if mainParent.GetID() == parents[i] {
			continue
		}
		block := bd.getBlockById(parents[i])
		if math.Abs(float64(block.GetLayer())-float64(mainParent.GetLayer())) > MaxTipLayerGap {
			continue
		}
		tips = append(tips, block)
		if limit && len(tips) >= bd.getMaxParents() {
			break
		}
	}
	return tips
}

// Checking the layer grap of block
func (bd *BlockDAG) checkLayerGap(parentsNode []IBlock) bool {
	if len(parentsNode) == 0 {
		return false
	}

	pLen := len(parentsNode)
	if pLen == 0 {
		return false
	}
	var gap float64
	if pLen == 1 {
		return true
	} else if pLen == 2 {
		gap = math.Abs(float64(parentsNode[0].GetLayer()) - float64(parentsNode[1].GetLayer()))
	} else {
		var minLayer int64 = -1
		var maxLayer int64 = -1
		for i := 0; i < pLen; i++ {
			parentLayer := int64(parentsNode[i].GetLayer())
			if maxLayer == -1 || parentLayer > maxLayer {
				maxLayer = parentLayer
			}
			if minLayer == -1 || parentLayer < minLayer {
				minLayer = parentLayer
			}
		}
		gap = math.Abs(float64(maxLayer) - float64(minLayer))
	}
	if gap > MaxTipLayerGap {
		log.Error(fmt.Sprintf("Parents gap is %f which is more than %d", gap, MaxTipLayerGap))
		return false
	}

	return true
}

// Checking the sub main chain for the parents of tip
func (bd *BlockDAG) CheckSubMainChainTip(parents []uint) (uint, bool) {
	bd.stateLock.Lock()
	defer bd.stateLock.Unlock()

	if len(parents) == 0 {
		return 0, false
	}
	for _, v := range parents {
		ib := bd.getBlockById(v)
		if ib == nil {
			return 0, false
		}
	}

	parentsSet := NewIdSet()
	parentsSet.AddList(parents)
	mainParent := bd.instance.GetMainParent(parentsSet)
	virtualHeight := mainParent.GetHeight() + 1

	if virtualHeight >= bd.getMainChainTip().GetHeight() {
		return virtualHeight, true
	}
	return 0, false
}

// Checking the parents of block legitimacy
func (bd *BlockDAG) checkLegality(parentsNode []IBlock) bool {
	if len(parentsNode) == 0 {
		return false
	}

	pLen := len(parentsNode)
	if pLen == 0 {
		return false
	} else if pLen == 1 {
		return true
	} else {
		parentsSet := NewIdSet()
		for _, v := range parentsNode {
			parentsSet.Add(v.GetID())
		}

		// Belonging to close relatives
		for _, p := range parentsNode {
			if p.HasParents() {
				inSet := p.GetParents().Intersection(parentsSet)
				if !inSet.IsEmpty() {
					return false
				}
			}
			if p.HasChildren() {
				inSet := p.GetChildren().Intersection(parentsSet)
				if !inSet.IsEmpty() {
					return false
				}
			}
		}
	}

	return true
}

// Load from database
func (bd *BlockDAG) Load(dbTx database.Tx, blockTotal uint, genesis *hash.Hash) error {
	meta := dbTx.Metadata()
	serializedData := meta.Get(dbnamespace.DagInfoBucketName)
	if serializedData == nil {
		return fmt.Errorf("dag load error")
	}

	err := bd.Decode(bytes.NewReader(serializedData))
	if err != nil {
		return err
	}
	bd.genesis = *genesis
	bd.blockTotal = blockTotal
	bd.blocks = map[uint]IBlock{}
	bd.tips = NewIdSet()
	return bd.instance.Load(dbTx)
}

func (bd *BlockDAG) Encode(w io.Writer) error {
	dagTypeIndex := GetDAGTypeIndex(bd.instance.GetName())
	err := s.WriteElements(w, dagTypeIndex)
	if err != nil {
		return err
	}
	return bd.instance.Encode(w)
}

// decode
func (bd *BlockDAG) Decode(r io.Reader) error {
	var dagTypeIndex byte
	err := s.ReadElements(r, &dagTypeIndex)
	if err != nil {
		return err
	}
	if GetDAGTypeIndex(bd.instance.GetName()) != dagTypeIndex {
		return fmt.Errorf("The dag type is %s, but read is %s", bd.instance.GetName(), GetDAGTypeByIndex(dagTypeIndex))
	}
	return bd.instance.Decode(r)
}

// GetBlues
func (bd *BlockDAG) GetBlues(parents *IdSet) uint {
	bd.stateLock.Lock()
	defer bd.stateLock.Unlock()

	return bd.instance.GetBlues(parents)
}

// IsBlue
func (bd *BlockDAG) IsBlue(id uint) bool {
	bd.stateLock.Lock()
	defer bd.stateLock.Unlock()

	return bd.instance.IsBlue(id)
}

func (bd *BlockDAG) IsHourglass(id uint) bool {
	bd.stateLock.Lock()
	defer bd.stateLock.Unlock()

	if !bd.hasBlockById(id) {
		return false
	}
	if !bd.isOnMainChain(id) {
		return false
	}
	block := bd.getBlockById(id)
	if block == nil {
		return false
	}
	if !block.IsOrdered() {
		return false
	}
	//
	queueSet := NewIdSet()
	queue := []IBlock{}
	for _, v := range bd.tips.GetMap() {
		ib := v.(IBlock)
		if !ib.IsOrdered() {
			continue
		}
		queue = append(queue, ib)
		queueSet.Add(ib.GetID())
	}

	num := 0
	for len(queue) > 0 {
		cur := queue[0]
		queue = queue[1:]
		if cur.GetID() == id {
			num++
			continue
		}
		if cur.GetLayer() <= block.GetLayer() {
			num++
			continue
		}
		if !cur.HasParents() {
			continue
		}
		for _, v := range cur.GetParents().GetMap() {
			ib := v.(IBlock)
			if queueSet.Has(ib.GetID()) || !ib.IsOrdered() {
				continue
			}
			queue = append(queue, ib)
			queueSet.Add(ib.GetID())
		}
	}
	return num == 1
}

func (bd *BlockDAG) GetParentsMaxLayer(parents *IdSet) (uint, bool) {
	bd.stateLock.Lock()
	defer bd.stateLock.Unlock()

	maxLayer := uint(0)
	for k := range parents.GetMap() {
		ib := bd.getBlockById(k)
		if ib == nil {
			return 0, false
		}
		if maxLayer == 0 || maxLayer < ib.GetLayer() {
			maxLayer = ib.GetLayer()
		}
	}
	return maxLayer, true
}

// GetMaturity
func (bd *BlockDAG) GetMaturity(target uint, views []uint) uint {
	bd.stateLock.Lock()
	defer bd.stateLock.Unlock()

	if target == MaxId {
		return 0
	}
	targetBlock := bd.getBlockById(target)
	if targetBlock == nil {
		return 0
	}

	//
	maxLayer := targetBlock.GetLayer()
	queueSet := NewIdSet()
	queue := []IBlock{}
	for _, v := range views {
		ib := bd.getBlockById(v)
		if ib != nil && ib.GetLayer() > targetBlock.GetLayer() {
			queue = append(queue, ib)
			queueSet.Add(ib.GetID())

			if maxLayer < ib.GetLayer() {
				maxLayer = ib.GetLayer()
			}
		}
	}

	connected := false
	for len(queue) > 0 {
		cur := queue[0]
		queue = queue[1:]
		if cur.GetID() == target {
			connected = true
			break
		}
		if !cur.HasParents() {
			continue
		}
		if cur.GetLayer() <= targetBlock.GetLayer() {
			continue
		}

		for _, v := range cur.GetParents().GetMap() {
			ib := v.(IBlock)
			if queueSet.Has(ib.GetID()) {
				continue
			}
			queue = append(queue, ib)
			queueSet.Add(ib.GetID())
		}
	}

	if connected {
		return maxLayer - targetBlock.GetLayer()
	}
	return 0
}

// Get path intersection from block to main chain.
func (bd *BlockDAG) getMainFork(ib IBlock, backward bool) IBlock {
	if bd.instance.IsOnMainChain(ib) {
		return ib
	}

	//
	queue := []IBlock{}
	queue = append(queue, ib)

	for len(queue) > 0 {
		cur := queue[0]
		queue = queue[1:]

		if bd.instance.IsOnMainChain(cur) {
			return cur
		}

		if backward {
			if !cur.HasChildren() {
				continue
			} else {
				childList := cur.GetChildren().SortHashList(false)
				for _, v := range childList {
					ib := cur.GetChildren().Get(v).(IBlock)
					queue = append(queue, ib)
				}
			}
		} else {
			if !cur.HasParents() {
				continue
			} else {
				parentsList := cur.GetParents().SortHashList(false)
				for _, v := range parentsList {
					ib := cur.GetParents().Get(v).(IBlock)
					queue = append(queue, ib)
				}
			}
		}
	}

	return nil
}

// MaxParentsPerBlock
func (bd *BlockDAG) getMaxParents() int {
	return bd.instance.getMaxParents()
}

// GetIdSet
func (bd *BlockDAG) GetIdSet(hs []*hash.Hash) *IdSet {
	result := NewIdSet()

	err := bd.db.View(func(dbTx database.Tx) error {
		for _, v := range hs {
			bid, er := DBGetBlockIdByHash(dbTx, v)
			if er == nil {
				result.Add(uint(bid))
			} else {
				return er
			}
		}
		return nil
	})
	if err != nil {
		return nil
	}

	return result
}

// The main parent concurrency of block
func (bd *BlockDAG) GetMainParentConcurrency(b IBlock) int {
	bd.stateLock.Lock()
	defer bd.stateLock.Unlock()
	return bd.instance.GetMainParentConcurrency(b)
}

// GetBlockConcurrency : Temporarily use blue set of the past blocks as the criterion
func (bd *BlockDAG) GetBlockConcurrency(h *hash.Hash) (uint, error) {
	bd.stateLock.Lock()
	defer bd.stateLock.Unlock()

	ib := bd.getBlock(h)
	if ib == nil {
		return 0, fmt.Errorf("No find block")
	}
	return ib.(*PhantomBlock).GetBlueNum(), nil
}

func (bd *BlockDAG) UpdateWeight(ib IBlock) {
	bd.instance.(*Phantom).UpdateWeight(ib)
}

// Commit the consensus content to the database for persistence
func (bd *BlockDAG) Commit() error {
	bd.stateLock.Lock()
	defer bd.stateLock.Unlock()
	return bd.commit()
}

// Commit the consensus content to the database for persistence
func (bd *BlockDAG) commit() error {
	if len(bd.commitOrder) > 0 {
		err := bd.db.Update(func(dbTx database.Tx) error {
			var e error
			for order, id := range bd.commitOrder {
				er := DBPutBlockIdByOrder(dbTx, order, id)
				if er != nil {
					log.Error(er.Error())
					e = er
				}
			}
			return e
		})
		bd.commitOrder = map[uint]uint{}
		if err != nil {
			return err
		}
	}

	if !bd.commitBlock.IsEmpty() {
		err := bd.db.Update(func(dbTx database.Tx) error {
			for _, v := range bd.commitBlock.GetMap() {
				block, ok := v.(IBlock)
				if !ok {
					return fmt.Errorf("Commit block error\n")
				}
				e := DBPutDAGBlock(dbTx, block)
				if e != nil {
					return e
				}
			}
			return nil
		})
		bd.commitBlock.Clean()
		if err != nil {
			return err
		}
	}
	return nil
}

// Just for custom Virtual block
func (bd *BlockDAG) CreateVirtualBlock(data IBlockData) IBlock {
	if _, ok := bd.instance.(*Phantom); !ok {
		return nil
	}
	parents := NewIdSet()
	var maxLayer uint = 0
	for _, p := range data.GetParents() {
		ib := bd.GetBlock(p)
		if ib == nil {
			return nil
		}
		parents.AddPair(ib.GetID(), ib)
		if maxLayer == 0 || maxLayer < ib.GetLayer() {
			maxLayer = ib.GetLayer()
		}
	}
	mainParent := bd.GetMainParent(parents)
	mainParentId := MaxId
	mainHeight := uint(0)
	if mainParent != nil {
		mainParentId = mainParent.GetID()
		mainHeight = mainParent.GetHeight()
	}
	block := Block{id: bd.GetBlockTotal(), hash: *data.GetHash(), parents: parents, layer: maxLayer + 1, status: StatusNone, mainParent: mainParentId, data: data, order: MaxBlockOrder, height: mainHeight + 1}
	return &PhantomBlock{&block, 0, NewIdSet(), NewIdSet()}
}

func (bd *BlockDAG) optimizeReorganizeResult(newOrders *list.List, oldOrders *list.List) {
	if newOrders == nil || oldOrders == nil {
		return
	}
	if newOrders.Len() <= 0 || oldOrders.Len() <= 0 {
		return
	}
	// optimization
	ne := newOrders.Front()
	oe := oldOrders.Front()
	for {
		if ne == nil || oe == nil {
			break
		}
		neNext := ne.Next()
		oeNext := oe.Next()

		neBlock := ne.Value.(IBlock)
		oeBlock := oe.Value.(*BlockOrderHelp)
		if neBlock.GetID() == oeBlock.Block.GetID() {
			newOrders.Remove(ne)
			oldOrders.Remove(oe)
		} else {
			break
		}

		ne = neNext
		oe = oeNext
	}
}

func (bd *BlockDAG) GetMainAncestor(block IBlock, height int64) IBlock {
	if height < 0 || height > int64(block.GetHeight()) {
		return nil
	}

	ib := block

	for ib != nil && int64(ib.GetHeight()) != height {
		if !ib.HasParents() {
			ib = nil
			break
		}
		ib = bd.GetBlockById(ib.GetMainParent())
	}
	return ib
}

func (bd *BlockDAG) RelativeMainAncestor(block IBlock, distance int64) IBlock {
	return bd.GetMainAncestor(block, int64(block.GetHeight())-distance)
}

func (bd *BlockDAG) ValidBlock(block IBlock) {
	block.Valid()
	bd.commitBlock.AddPair(block.GetID(), block)
}

func (bd *BlockDAG) InvalidBlock(block IBlock) {
	block.Invalid()
	bd.commitBlock.AddPair(block.GetID(), block)
}
