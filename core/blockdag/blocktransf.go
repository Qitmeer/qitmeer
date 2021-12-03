package blockdag

import (
	"fmt"
	"github.com/Qitmeer/qng-core/common/hash"
	"github.com/Qitmeer/qng-core/database"
	"sort"
)

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
	if bd.lastSnapshot.block != nil {
		if bd.lastSnapshot.block.GetHash().IsEqual(h) {
			return bd.lastSnapshot.block.GetID()
		}
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

func (bd *BlockDAG) GetBlockHash(id uint) *hash.Hash {
	bd.stateLock.Lock()
	defer bd.stateLock.Unlock()

	ib := bd.getBlockById(id)
	if ib != nil {
		return ib.GetHash()
	}
	return nil
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

// GetIdSet
func (bd *BlockDAG) GetIdSet(hs []*hash.Hash) *IdSet {
	result := NewIdSet()

	err := bd.db.View(func(dbTx database.Tx) error {
		for _, v := range hs {
			if bd.lastSnapshot.block != nil {
				if bd.lastSnapshot.block.GetHash().IsEqual(v) {
					result.Add(bd.lastSnapshot.block.GetID())
					continue
				}
			}
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

// Return the layer of block,it is stable.
// You can imagine that this is the main chain.
func (bd *BlockDAG) GetLayer(id uint) uint {
	return bd.GetBlockById(id).GetLayer()
}
