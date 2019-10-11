package blockdag

import (
	"bytes"
	"container/list"
	"fmt"
	"github.com/Qitmeer/qitmeer/common/hash"
	"github.com/Qitmeer/qitmeer/core/dbnamespace"
	"github.com/Qitmeer/qitmeer/core/merkle"
	"github.com/Qitmeer/qitmeer/database"
	"io"
	"math"
	"sort"
	"sync"
	"time"
	s "github.com/Qitmeer/qitmeer/core/serialization"
)

// Some available DAG algorithm types
const (
	// A Scalable BlockDAG protocol
	phantom="phantom"

	// Phantom protocol V2
	phantom_v2="phantom_v2"

	// The order of all transactions is solely determined by the Tree Graph (TG)
	conflux="conflux"

	// Confirming Transactions via Recursive Elections
	spectre="spectre"
)

// Maximum number of the DAG tip
const MaxTips=100

// Maximum order of the DAG block
const MaxBlockOrder=uint(^uint32(0))

// MaxTipLayerGap
const MaxTipLayerGap=10

// StableConfirmations
const StableConfirmations=10

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
	AddBlock(ib IBlock) *list.List

	// Build self block
	CreateBlock(b *Block) IBlock

	// If the successor return nil, the underlying layer will use the default tips list.
	GetTipsList() []IBlock

	// Find block hash by order, this is very fast.
	GetBlockByOrder(order uint) *hash.Hash

	// Query whether a given block is on the main chain.
	IsOnMainChain(ib IBlock) bool

	// return the tip of main chain
	GetMainChainTip() IBlock

	// return the main parent in the parents
	GetMainParent(parents *HashSet) IBlock

	// encode
	Encode(w io.Writer) error

	// decode
	Decode(r io.Reader) error

	// load
	Load(dbTx database.Tx) error
}

// The general foundation framework of DAG
type BlockDAG struct {
	// The genesis of block dag
	genesis hash.Hash

	// Use block hash to save all blocks with mapping
	blocks map[hash.Hash]IBlock

	// The total number blocks that this dag currently owned
	blockTotal uint

	// The terminal block is in block dag,this block have not any connecting at present.
	tips *HashSet

	// This is time when the last block have added
	lastTime time.Time

	// The full sequence of dag, please note that the order starts at zero.
	order map[uint]*hash.Hash

	// Current dag instance used. Different algorithms work according to
	// different dag types config.
	instance IBlockDAG

	// Use block id to save all blocks with mapping
	blockids map[uint]*hash.Hash

	// state lock
	stateLock sync.RWMutex
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
func (bd *BlockDAG) Init(dagType string) IBlockDAG{
	bd.instance=NewBlockDAG(dagType)
	bd.instance.Init(bd)

	bd.lastTime=time.Unix(time.Now().Unix(), 0)

	return bd.instance
}

// This is an entry for update the block dag,you need pass in a block parameter,
// If add block have failure,it will return false.
func (bd *BlockDAG) AddBlock(b IBlockData) *list.List {
	bd.stateLock.Lock()
	defer bd.stateLock.Unlock()

	if b == nil {
		return nil
	}
	if bd.hasBlock(b.GetHash()) {
		return nil
	}
	var parents []*hash.Hash
	if bd.blockTotal > 0 {
		parents = b.GetParents()
		if len(parents) == 0 {
			return nil
		}
		if !bd.hasBlocks(parents) {
			return nil
		}
	}
	if !bd.isDAG(b) {
		return nil
	}
	//
	block := Block{id:bd.blockTotal,hash: *b.GetHash(), weight: 1, layer:0,status:StatusNone}
	if parents != nil {
		block.parents = NewHashSet()
		var maxLayer uint=0
		for k, h := range parents {
			parent := bd.getBlock(h)
			block.parents.AddPair(h,parent)
			parent.AddChild(&block)
			if k == 0 {
				block.mainParent = parent.GetHash()
			}

			if maxLayer==0 || maxLayer < parent.GetLayer() {
				maxLayer=parent.GetLayer()
			}
		}
		block.SetLayer(maxLayer+1)
	}

	if bd.blocks == nil {
		bd.blocks = map[hash.Hash]IBlock{}
	}
	ib:=bd.instance.CreateBlock(&block)
	bd.blocks[block.hash] = ib
	if bd.blockTotal == 0 {
		bd.genesis = *block.GetHash()
	}
	//
	if bd.blockids == nil {
		bd.blockids = map[uint]*hash.Hash{}
	}
	bd.blockids[block.GetID()]=block.GetHash()
	//
	bd.blockTotal++
	//
	bd.updateTips(&block)
	//
	t:=time.Unix(b.GetTimestamp(), 0)
	if bd.lastTime.Before(t) {
		bd.lastTime=t
	}
	//
	return bd.instance.AddBlock(ib)
}

// Acquire the genesis block of chain
func (bd *BlockDAG) getGenesis() IBlock {
	return bd.getBlock(&bd.genesis)
}

// Acquire the genesis block hash of chain
func (bd *BlockDAG) GetGenesisHash() *hash.Hash {
	return &bd.genesis
}

// If the block is illegal dag,will return false.
func (bd *BlockDAG) isDAG(b IBlockData) bool {
	err:=bd.checkLayerGap(b.GetParents())
	if err != nil {
		log.Warn(err.Error())
		return false
	}
	return true
}

// Is there a block in DAG?
func (bd *BlockDAG) hasBlock(h *hash.Hash) bool {
	return bd.getBlock(h) != nil
}

// Is there a block in DAG?
func (bd *BlockDAG) HasBlock(h *hash.Hash) bool {
	return bd.GetBlock(h) != nil
}

// Is there some block in DAG?
func (bd *BlockDAG) hasBlocks(hs []*hash.Hash) bool {
	for _, h := range hs {
		if !bd.hasBlock(h) {
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
func (bd *BlockDAG) getBlock(h *hash.Hash) IBlock {
	if h == nil {
		return nil
	}
	block, ok := bd.blocks[*h]
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

	return bd.tips
}

// Acquire the tips array of DAG
func (bd *BlockDAG) GetTipsList() []IBlock {
	bd.stateLock.Lock()
	defer bd.stateLock.Unlock()

	result:=bd.instance.GetTipsList()
	if result!=nil {
		return result
	}
	result=[]IBlock{}
	for k:=range bd.tips.GetMap(){
		result=append(result,bd.getBlock(&k))
	}
	return result
}

// build merkle tree form current DAG tips
func (bd *BlockDAG) BuildMerkleTreeStoreFromTips() []*hash.Hash {
	parents:=bd.GetTips().SortList(false)
	return merkle.BuildParentsMerkleTreeStore(parents)
}

// Refresh the dag tip whith new block,it will cause changes in tips set.
func (bd *BlockDAG) updateTips(b *Block) {
	if bd.tips == nil {
		bd.tips = NewHashSet()
		bd.tips.AddPair(b.GetHash(),b)
		return
	}
	for k := range bd.tips.GetMap() {
		block := bd.getBlock(&k)
		if block.HasChildren() {
			bd.tips.Remove(&k)
		}
	}
	bd.tips.AddPair(b.GetHash(),b)
}

// The last time is when add one block to DAG.
func (bd *BlockDAG) GetLastTime() *time.Time{
	bd.stateLock.Lock()
	defer bd.stateLock.Unlock()

	return &bd.lastTime
}

// Return the full sequence array.
func (bd *BlockDAG) GetOrder() map[uint]*hash.Hash {
	bd.stateLock.Lock()
	defer bd.stateLock.Unlock()

	return bd.order
}

// Obtain block hash by global order
func (bd *BlockDAG) GetBlockByOrder(order uint) *hash.Hash{
	bd.stateLock.Lock()
	defer bd.stateLock.Unlock()

	return bd.instance.GetBlockByOrder(order)
}

// Return the last order block
func (bd *BlockDAG) GetLastBlock() IBlock{
	// TODO
	return bd.GetMainChainTip()
}

// This function need a stable sequence,so call it before sorting the DAG.
// If the h is invalid,the function will become a little inefficient.
func (bd *BlockDAG) GetPrevious(h *hash.Hash) *hash.Hash{
	bd.stateLock.Lock()
	defer bd.stateLock.Unlock()

	if h==nil {
		return nil
	}
	if h.IsEqual(bd.GetGenesisHash()) {
		return nil
	}
	b:=bd.getBlock(h)
	if b==nil {
		return nil
	}
	if b.GetOrder()==0{
		return nil
	}
	// TODO
	return bd.instance.GetBlockByOrder(b.GetOrder()-1)
}

// Returns a future collection of block. This function is a recursively called function
// So we should consider its efficiency.
func (bd *BlockDAG) getFutureSet(fs *HashSet, b IBlock) {
	children := b.GetChildren()
	if children == nil || children.IsEmpty() {
		return
	}
	for k:= range children.GetMap() {
		if !fs.Has(&k) {
			fs.Add(&k)
			bd.getFutureSet(fs, bd.getBlock(&k))
		}
	}
}

// Query whether a given block is on the main chain.
// Note that some DAG protocols may not support this feature.
func (bd *BlockDAG) IsOnMainChain(h *hash.Hash) bool {
	bd.stateLock.Lock()
	defer bd.stateLock.Unlock()

	return bd.isOnMainChain(h)
}

// Query whether a given block is on the main chain.
// Note that some DAG protocols may not support this feature.
func (bd *BlockDAG) isOnMainChain(h *hash.Hash) bool {
	return bd.instance.IsOnMainChain(bd.getBlock(h))
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
func (bd *BlockDAG) GetMainParent(parents *HashSet) IBlock {
	bd.stateLock.Lock()
	defer bd.stateLock.Unlock()

	return bd.instance.GetMainParent(parents)
}

// Return the layer of block,it is stable.
// You can imagine that this is the main chain.
func (bd *BlockDAG) GetLayer(h *hash.Hash) uint {
	return bd.GetBlock(h).GetLayer()
}

// Return current general description of the whole state of DAG
func (bd *BlockDAG) GetGraphState() *GraphState {
	bd.stateLock.Lock()
	defer bd.stateLock.Unlock()
	return bd.getGraphState()
}

// Return current general description of the whole state of DAG
func (bd *BlockDAG) getGraphState() *GraphState {
	gs:=NewGraphState()
	if bd.tips!=nil && !bd.tips.IsEmpty() {
		gs.GetTips().AddList(bd.tips.List())

		gs.SetLayer(0)
		for _,v:=range bd.tips.GetMap() {
			tip:=v.(*Block)
			if tip.GetLayer()>gs.GetLayer(){
				gs.SetLayer(tip.GetLayer())
			}
		}
	}
	gs.SetTotal(bd.blockTotal)
	gs.SetMainHeight(bd.getMainChainTip().GetHeight())
	gs.SetMainOrder(bd.getMainChainTip().GetOrder())
	return gs
}

// Locate all eligible block by current graph state.
func (bd *BlockDAG) LocateBlocks(gs *GraphState, maxHashes uint) []*hash.Hash {
	bd.stateLock.Lock()
	defer bd.stateLock.Unlock()

	if gs.IsExcellent(bd.getGraphState()) {
		return nil
	}
	queue := []IBlock{}
	fs:=NewHashSet()
	tips:=bd.getValidTips()
	for _,v:=range tips {
		ib:=bd.getBlock(v)
		queue=append(queue,ib)
	}
	for len(queue) > 0 {
		cur := queue[0]
		queue = queue[1:]

		if fs.Has(cur.GetHash()) {
			continue
		}
		if gs.GetTips().Has(cur.GetHash()) || cur.GetHash().IsEqual(&bd.genesis) {
			continue
		}
		needRec:=true
		if cur.HasChildren() {
			for _,v := range cur.GetChildren().GetMap() {
				ib:=v.(IBlock)
				if gs.GetTips().Has(ib.GetHash()) || !fs.Has(ib.GetHash())&&ib.IsOrdered() {
					needRec=false
					break
				}
			}
		}
		if needRec {
			fs.AddPair(cur.GetHash(),cur)
			if cur.HasParents() {
				for _,v := range cur.GetParents().GetMap() {
					value:=v.(IBlock)
					ib:=value
					if fs.Has(ib.GetHash()) {
						continue
					}
					queue=append(queue,ib)

				}
			}
		}
	}

	fsSlice:=BlockSlice{}
	for _,v:=range fs.GetMap(){
		value:=v.(IBlock)
		ib:=value
		if gs.GetTips().Has(ib.GetHash()) {
			continue
		}
		if ib.HasChildren() {
			need:=true
			for _,v := range ib.GetChildren().GetMap() {
				ib:=v.(IBlock)
				if gs.GetTips().Has(ib.GetHash()) {
					need=false
					break
				}
			}
			if !need {
				continue
			}
		}
		fsSlice=append(fsSlice,ib)
	}

	result:=[]*hash.Hash{}
	if len(fsSlice)>=2 {
		sort.Sort(fsSlice)
	}
	for i:=0;i<len(fsSlice) ;i++  {
		if maxHashes>0 && i>=int(maxHashes) {
			break
		}
		result=append(result,fsSlice[i].GetHash())
	}
	return result
}

// Judging whether block is the virtual tip that it have not future set.
func isVirtualTip(b IBlock, futureSet *HashSet, anticone *HashSet, children *HashSet) bool {
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
func (bd *BlockDAG) recAnticone(b IBlock, futureSet *HashSet, anticone *HashSet, h *hash.Hash) {
	if h.IsEqual(b.GetHash()) {
		return
	}
	node:=bd.getBlock(h)
	children := node.GetChildren()
	needRecursion := false
	if children == nil || children.Size() == 0 {
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
			bd.recAnticone(b, futureSet, anticone, &k)
		}
	}
}

// This function can get anticone set for an block that you offered in the block dag,If
// the exclude set is not empty,the final result will exclude set that you passed in.
func (bd *BlockDAG) getAnticone(b IBlock, exclude *HashSet) *HashSet {
	futureSet := NewHashSet()
	bd.getFutureSet(futureSet, b)
	anticone := NewHashSet()
	for k:= range bd.tips.GetMap() {
		bd.recAnticone(b, futureSet, anticone, &k)
	}
	if exclude != nil {
		anticone.Exclude(exclude)
	}
	return anticone
}

// Sort block by id
func (bd *BlockDAG) SortBlock(src []*hash.Hash) []*hash.Hash {
	bd.stateLock.Lock()
	defer bd.stateLock.Unlock()

	if len(src)<=1 {
		return src
	}
	srcBlockS:=BlockSlice{}
	for i:=0;i<len(src) ;i++  {
		ib:=bd.getBlock(src[i])
		if ib!=nil {
			srcBlockS=append(srcBlockS,ib)
		}
	}
	if len(srcBlockS)>=2 {
		sort.Sort(srcBlockS)
	}
	result:=[]*hash.Hash{}
	for i:=0;i<len(srcBlockS) ;i++  {
		result=append(result,srcBlockS[i].GetHash())
	}
	return result
}

// GetConfirmations
func (bd *BlockDAG) GetConfirmations(h *hash.Hash) uint {
	bd.stateLock.Lock()
	defer bd.stateLock.Unlock()

	block:=bd.getBlock(h)
	if block == nil {
		return 0
	}
	if block.GetOrder() > bd.getMainChainTip().GetOrder() {
		return 0
	}
	mainTip:=bd.getMainChainTip()
	if bd.isOnMainChain(h) {
		return mainTip.GetHeight()-block.GetHeight()
	}
	if !block.HasChildren() {
		return 0
	}
	//
	queue := []IBlock{}
	queue=append(queue,block)

	for len(queue) > 0 {
		cur := queue[0]
		queue = queue[1:]

		if bd.isOnMainChain(cur.GetHash()) {
			return 1+mainTip.GetHeight()-cur.GetHeight()
		}
		if !cur.HasChildren() {
			return 0
		}else {
			for _,v := range cur.GetChildren().GetMap() {
				ib:=v.(IBlock)
				queue=append(queue,ib)
			}
		}
	}
	return 0
}

func (bd *BlockDAG) GetBlockHash(id uint) *hash.Hash {
	bd.stateLock.Lock()
	defer bd.stateLock.Unlock()

	return bd.blockids[id]
}

func (bd *BlockDAG) GetValidTips() []*hash.Hash {
	bd.stateLock.Lock()
	defer bd.stateLock.Unlock()
	return bd.getValidTips()
}

func (bd *BlockDAG) getValidTips() []*hash.Hash {
	parents:=bd.tips.SortList(false)
	mainParent:=bd.getMainChainTip()
	tips:=[]*hash.Hash{}
	for i:=0;i<len(parents);i++ {
		if mainParent.GetHash().IsEqual(parents[i]) {
			tips=append(tips,parents[i])
			continue
		}
		block:=bd.getBlock(parents[i])
		if math.Abs(float64(block.GetLayer())-float64(mainParent.GetLayer()))>MaxTipLayerGap {
			continue
		}
		tips=append(tips,block.GetHash())
	}
	return tips
}

func (bd *BlockDAG) CheckLayerGap(parents []*hash.Hash) error {
	bd.stateLock.Lock()
	defer bd.stateLock.Unlock()

	return bd.checkLayerGap(parents)
}

func (bd *BlockDAG) checkLayerGap(parents []*hash.Hash) error {
	if len(parents)==0 {
		return nil
	}
	parentsNode:=[]IBlock{}
	for _,v:=range parents{
		ib:=bd.getBlock(v)
		if ib == nil {
			continue
		}
		parentsNode=append(parentsNode,ib)
	}

	pLen:=len(parentsNode)
	if pLen==0 {
		return nil
	}
	var gap float64
	if pLen == 1 {
		return nil
	}else if pLen == 2 {
		gap=math.Abs(float64(parentsNode[0].GetLayer())-float64(parentsNode[1].GetLayer()))
	}else{
		var minLayer int64=-1
		var maxLayer int64=-1
		for i:=0;i<pLen ;i++  {
			parentLayer:=int64(parentsNode[i].GetLayer())
			if maxLayer ==-1 || parentLayer > maxLayer {
				maxLayer=parentLayer
			}
			if minLayer == -1 || parentLayer < minLayer {
				minLayer=parentLayer
			}
		}
		gap=math.Abs(float64(maxLayer)-float64(minLayer))
	}
	if gap > MaxTipLayerGap {
		return fmt.Errorf("Parents gap is %f which is more than %d",gap,MaxTipLayerGap)
	}

	return nil
}

// Load from database
func (bd *BlockDAG) Load(dbTx database.Tx,blockTotal uint,genesis *hash.Hash) error {
	meta := dbTx.Metadata()
	serializedData := meta.Get(dbnamespace.DagInfoBucketName)
	if serializedData == nil {
		return fmt.Errorf("dag load error")
	}

	err := bd.Decode(bytes.NewReader(serializedData))
	if err != nil {
		return err
	}
	bd.genesis=*genesis
	bd.blockTotal=blockTotal
	bd.blocks = map[hash.Hash]IBlock{}
	bd.blockids = map[uint]*hash.Hash{}
	bd.tips = NewHashSet()
	return bd.instance.Load(dbTx)
}

func (bd *BlockDAG) Encode(w io.Writer) error {
	dagTypeIndex:=GetDAGTypeIndex(bd.instance.GetName())
	err:=s.WriteElements(w,dagTypeIndex)
	if err != nil {
		return err
	}
	return bd.instance.Encode(w)
}

// decode
func (bd *BlockDAG) Decode(r io.Reader) error {
	var dagTypeIndex byte
	err:=s.ReadElements(r,&dagTypeIndex)
	if err != nil {
		return err
	}
	if GetDAGTypeIndex(bd.instance.GetName()) != dagTypeIndex {
		return fmt.Errorf("The dag type is %s, but read is %s",bd.instance.GetName(),GetDAGTypeByIndex(dagTypeIndex))
	}
	return bd.instance.Decode(r)
}