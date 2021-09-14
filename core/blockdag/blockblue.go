package blockdag

import (
	"github.com/Qitmeer/qitmeer/common/hash"
	"time"
)

// GetBlues
func (bd *BlockDAG) GetBlues(parents *IdSet) uint {
	bd.stateLock.Lock()
	defer bd.stateLock.Unlock()

	return bd.instance.GetBlues(parents)
}

func (bd *BlockDAG) GetBluesByHash(h *hash.Hash) uint {
	bd.stateLock.Lock()
	defer bd.stateLock.Unlock()

	return bd.getBluesByBlock(bd.getBlockById(bd.getBlockId(h)))
}

func (bd *BlockDAG) GetBluesByBlock(ib IBlock) uint {
	bd.stateLock.Lock()
	defer bd.stateLock.Unlock()

	return bd.getBluesByBlock(ib)
}

func (bd *BlockDAG) getBluesByBlock(ib IBlock) uint {
	if ib == nil {
		return 0
	}
	pb, ok := ib.(*PhantomBlock)
	if !ok {
		return 0
	}
	return pb.blueNum
}

// IsBlue
func (bd *BlockDAG) IsBlue(id uint) bool {
	bd.stateLock.Lock()
	defer bd.stateLock.Unlock()

	return bd.instance.IsBlue(id)
}

func (bd *BlockDAG) GetBlueInfoByHash(h *hash.Hash) *BlueInfo {
	bd.stateLock.Lock()
	defer bd.stateLock.Unlock()

	return bd.getBlueInfo(bd.getBlockById(bd.getBlockId(h)))
}

func (bd *BlockDAG) GetBlueInfo(ib IBlock) *BlueInfo {
	bd.stateLock.Lock()
	defer bd.stateLock.Unlock()

	return bd.getBlueInfo(ib)
}

func (bd *BlockDAG) getBlueInfo(ib IBlock) *BlueInfo {
	if ib == nil {
		return nil
	}
	if ib.GetID() == 0 {
		return nil
	}
	if !ib.HasParents() {
		return nil
	}
	mainIB, ok := ib.GetParents().Get(ib.GetMainParent()).(IBlock)
	if !ok {
		return nil
	}
	mt := ib.GetData().GetTimestamp() - mainIB.GetData().GetTimestamp()
	if mt <= 0 {
		mt = 1
	}
	mt *= int64(time.Second)

	pb, ok := ib.(*PhantomBlock)
	if !ok {
		return nil
	}
	blues := 1
	if pb.blueDiffAnticone != nil && !pb.blueDiffAnticone.IsEmpty() {
		blues += pb.blueDiffAnticone.Size()
	}
	return NewBlueInfo(pb.blueNum+1, mt/int64(blues), int64(mainIB.GetWeight()))
}
