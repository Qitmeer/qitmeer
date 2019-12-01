package blockdag

import (
	"github.com/Qitmeer/qitmeer/common/hash"
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
	GSMtx sync.Mutex
	GS    *GraphState
}

// CalcSyncBlocks
func (ds *DAGSync) CalcSyncBlocks(gs *GraphState, locator []*hash.Hash, mode SyncMode, maxHashes uint) []*hash.Hash {
	ds.bd.stateLock.Lock()
	defer ds.bd.stateLock.Unlock()

	result := []*hash.Hash{}
	if mode == DirectMode {
		if len(locator) == 0 {
			return result
		}
		for _, v := range locator {
			if ds.bd.hasBlock(v) {
				result = append(result, v)
			}
		}
		if len(result) >= 2 {
			result = ds.bd.sortBlock(result)
		}
	} else {
		//result = ds.bd.LocateBlocks(gs, message.MaxBlockLocatorsPerMsg)
	}

	return result
}

// GetMainLocator
func (ds *DAGSync) GetMainLocator(point *hash.Hash) []*hash.Hash {
	ds.bd.stateLock.Lock()
	defer ds.bd.stateLock.Unlock()

	var endBlock IBlock
	if point != nil {
		endBlock = ds.bd.getBlock(point)
	}
	if endBlock == nil {
		endBlock = ds.bd.getGenesis()
	} else {
		for !ds.bd.isOnMainChain(endBlock.GetHash()) {
			if endBlock.GetMainParent() == nil {
				break
			}
			endBlock = ds.bd.getBlock(endBlock.GetMainParent())
			if endBlock == nil {
				endBlock = ds.bd.getGenesis()
				break
			}
		}
	}
	startBlock := ds.bd.getMainChainTip()
	dist := startBlock.GetHeight() - endBlock.GetHeight()
	locator := []*hash.Hash{}
	cur := startBlock
	if dist <= MaxMainLocatorNum {
		for !cur.GetHash().IsEqual(endBlock.GetHash()) {
			if cur.GetHash().IsEqual(ds.bd.GetGenesisHash()) {
				break
			}
			locator = append(locator, cur.GetHash())
			if cur.GetMainParent() == nil {
				break
			}
			cur = ds.bd.getBlock(cur.GetMainParent())
			if cur == nil {
				break
			}
		}
	} else {
		const DefaultMainLocatorNum = 10
		deep := uint(1)
		for len(locator) < MaxMainLocatorNum {
			if cur.GetHash().IsEqual(ds.bd.GetGenesisHash()) {
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

			if cur.GetMainParent() == nil {
				break
			}

			next := ds.bd.getBlock(cur.GetMainParent())
			if next.GetHash().IsEqual(endBlock.GetHash()) {
				locator = append(locator, cur.GetHash())
				break
			}
			cur = next
			if cur == nil {
				break
			}
		}
	}

	if len(locator) >= 2 {
		tempL := locator
		locator = []*hash.Hash{}
		for i := len(tempL) - 1; i >= 0; i-- {
			locator = append(locator, tempL[i])
		}
	}

	return locator
}

// NewDAGSync
func NewDAGSync(bd *BlockDAG) *DAGSync {
	return &DAGSync{bd: bd}
}
