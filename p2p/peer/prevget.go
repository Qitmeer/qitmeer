package peer

import (
	"fmt"
	"github.com/Qitmeer/qitmeer/common/hash"
	"github.com/Qitmeer/qitmeer/core/blockdag"
	"github.com/Qitmeer/qitmeer/log"
	"sync"
)

type PrevGet struct {
	sync.Mutex
	GS      *blockdag.GraphState
	Locator []*hash.Hash
	Point   *hash.Hash
	Blocks  *blockdag.HashSet
}

func (pg *PrevGet) Init(p *Peer) {
	pg.Blocks = blockdag.NewHashSet()
	/*	pg.Locator=blockdag.NewHashSet()
		pg.Locator.AddPair(p.cfg.ChainParams.GenesisHash,0)*/
}

func (pg *PrevGet) Clean() {
	pg.Lock()
	defer pg.Unlock()

	if pg.Blocks != nil {
		pg.Blocks.Clean()
	}
	/*	if pg.Locator!=nil {
		pg.Locator.Clean()
	}*/
}

func (pg *PrevGet) CheckBlocks(p *Peer, gs *blockdag.GraphState, blocks []*hash.Hash) (bool, *blockdag.HashSet) {
	pg.Lock()
	defer pg.Unlock()

	bs := blockdag.NewHashSet()
	// Filter duplicate getblocks requests.
	if len(blocks) > 0 {
		for _, v := range blocks {
			if !pg.Blocks.Has(v) {
				bs.Add(v)
			}
		}
		if bs.IsEmpty() {
			log.Trace(fmt.Sprintf("Filtering duplicate [getblocks]: blocks=%d", len(blocks)))
			return false, nil
		}
		//isDuplicate=false
	}
	return true, bs
}

func (pg *PrevGet) UpdateBlocks(blocks []*hash.Hash) {
	pg.Lock()
	defer pg.Unlock()

	// Filter duplicate getblocks requests.
	if len(blocks) > 0 {
		pg.Blocks.AddList(blocks)
	}
}

func (pg *PrevGet) UpdateGS(gs *blockdag.GraphState, locator []*hash.Hash) {
	pg.Lock()
	defer pg.Unlock()
	pg.GS = gs
	pg.Locator = locator
}

func (pg *PrevGet) UpdatePoint(point *hash.Hash) {
	pg.Lock()
	defer pg.Unlock()
	pg.Point = point
}
