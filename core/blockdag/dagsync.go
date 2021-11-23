package blockdag

import (
	"github.com/Qitmeer/qng-core/common/hash"
	"sync"
)

// This parameter can be set according to the size of TCP package(1500) to ensure the transmission stability of the network
const MaxMainLocatorNum = 32

// Synchronization mode
type SyncMode byte

const (
	DirectMode SyncMode = 0
	SubDAGMode SyncMode = 1
)

type DAGSync struct {
	bd *BlockDAG

	// The following fields are used to track the graph state being synced to from
	// peers.
	gsMtx sync.Mutex
	gs    *GraphState
}

// CalcSyncBlocks
func (ds *DAGSync) CalcSyncBlocks(gs *GraphState, locator []*hash.Hash, mode SyncMode, maxHashes uint) ([]*hash.Hash, *hash.Hash) {
	ds.bd.stateLock.Lock()
	defer ds.bd.stateLock.Unlock()

	if mode == DirectMode {
		result := []*hash.Hash{}
		if len(locator) == 0 {
			return result, nil
		}
		return ds.bd.sortBlock(locator), nil
	}

	var point IBlock
	for i := len(locator) - 1; i >= 0; i-- {
		mainBlock := ds.bd.getBlock(locator[i])
		if mainBlock == nil {
			continue
		}
		if !ds.bd.isOnMainChain(mainBlock.GetID()) {
			continue
		}
		point = mainBlock
		break
	}

	if point == nil && len(locator) > 0 {
		point = ds.bd.getBlock(locator[0])
		if point != nil {
			for !ds.bd.isOnMainChain(point.GetID()) {
				if point.GetMainParent() == MaxId {
					break
				}
				point = ds.bd.getBlockById(point.GetMainParent())
				if point == nil {
					break
				}
			}
		}

	}

	if point == nil {
		point = ds.bd.getGenesis()
	}
	//
	isSubDAG := false
	for k := range gs.tips.GetMap() {
		gst := ds.bd.getBlock(&k)
		if gst == nil || !gst.IsOrdered() {
			continue
		}
		isSubDAG = true
		break
	}
	if isSubDAG {
		return ds.bd.locateBlocks(gs, maxHashes), point.GetHash()
	}
	return ds.getBlockChainFromMain(point, maxHashes), point.GetHash()
}

// GetMainLocator
func (ds *DAGSync) GetMainLocator(point *hash.Hash) []*hash.Hash {
	ds.bd.stateLock.Lock()
	defer ds.bd.stateLock.Unlock()

	var endBlock IBlock
	if point != nil {
		endBlock = ds.bd.getBlock(point)
	}
	if endBlock != nil {
		for !ds.bd.isOnMainChain(endBlock.GetID()) {
			if endBlock.GetMainParent() == MaxId {
				break
			}
			endBlock = ds.bd.getBlockById(endBlock.GetMainParent())
			if endBlock == nil {
				break
			}
		}
	}
	if endBlock == nil {
		endBlock = ds.bd.getGenesis()
	}
	startBlock := ds.bd.getMainChainTip()
	dist := startBlock.GetHeight() - endBlock.GetHeight()
	locator := []*hash.Hash{}
	cur := startBlock
	if dist <= MaxMainLocatorNum {
		for cur.GetID() != endBlock.GetID() {
			if cur.GetID() == 0 {
				break
			}
			locator = append(locator, cur.GetHash())
			if cur.GetMainParent() == MaxId {
				break
			}
			cur = ds.bd.getBlockById(cur.GetMainParent())
			if cur == nil {
				break
			}
		}
	} else {
		const DefaultMainLocatorNum = 10
		deep := uint(1)
		for len(locator) < MaxMainLocatorNum {
			if cur.GetID() == 0 {
				break
			}
			if len(locator) < DefaultMainLocatorNum {
				locator = append(locator, cur.GetHash())
			} else {
				height := uint(0)
				if startBlock.GetHeight()-DefaultMainLocatorNum >= deep {
					height = startBlock.GetHeight() - DefaultMainLocatorNum - deep
				}
				if cur.GetHeight() <= height {
					locator = append(locator, cur.GetHash())
					deep *= 2
				}
			}

			if cur.GetMainParent() == MaxId {
				break
			}

			next := ds.bd.getBlockById(cur.GetMainParent())
			if next.GetID() == endBlock.GetID() {
				break
			}
			cur = next
			if cur == nil {
				break
			}
		}
	}
	locator = append(locator, endBlock.GetHash())
	if len(locator) >= 2 {
		tempL := locator
		locator = []*hash.Hash{}
		for i := len(tempL) - 1; i >= 0; i-- {
			if len(locator) >= MaxMainLocatorNum {
				break
			}
			locator = append(locator, tempL[i])
		}
	}

	return locator
}

func (ds *DAGSync) getBlockChainFromMain(point IBlock, maxHashes uint) []*hash.Hash {
	mainTip := ds.bd.getMainChainTip()
	result := []*hash.Hash{}
	for i := point.GetOrder() + 1; i <= mainTip.GetOrder(); i++ {
		block := ds.bd.getBlockByOrder(i)
		if block == nil {
			continue
		}
		result = append(result, block.GetHash())
		if uint(len(result)) >= maxHashes {
			break
		}
	}
	return result
}

func (ds *DAGSync) SetGraphState(gs *GraphState) {
	ds.gsMtx.Lock()
	defer ds.gsMtx.Unlock()

	ds.gs = gs
}

// NewDAGSync
func NewDAGSync(bd *BlockDAG) *DAGSync {
	return &DAGSync{bd: bd}
}
