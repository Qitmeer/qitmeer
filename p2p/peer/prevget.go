package peer

import (
	"github.com/Qitmeer/qitmeer-lib/core/dag"
	"github.com/Qitmeer/qitmeer-lib/log"
	"fmt"
	"sync"
	"github.com/Qitmeer/qitmeer-lib/common/hash"
)

type PrevGet struct {
	BlocksMtx   sync.Mutex
	GS          *dag.GraphState
	//Locator     *dag.HashSet
	Blocks      *dag.HashSet

	HdrsMtx     sync.Mutex
	HdrsBegin   *hash.Hash
	HdrsStop    *hash.Hash
}

func (pg *PrevGet) Init(p *Peer) {
	pg.Blocks=dag.NewHashSet()
/*	pg.Locator=dag.NewHashSet()
	pg.Locator.AddPair(p.cfg.ChainParams.GenesisHash,0)*/
}

func (pg *PrevGet) Clean() {
	if pg.Blocks!=nil {
		pg.Blocks.Clean()
	}
/*	if pg.Locator!=nil {
		pg.Locator.Clean()
	}*/
}

func (pg *PrevGet) Check(p *Peer,gs *dag.GraphState,blocks []*hash.Hash) (bool,*dag.HashSet) {
	pg.BlocksMtx.Lock()
	defer pg.BlocksMtx.Unlock()

	bs:=dag.NewHashSet()
	// Filter duplicate getblocks requests.
	if len(blocks)>0 {
		for _,v:=range blocks{
			if !pg.Blocks.Has(v) {
				bs.Add(v)
			}
		}
		if bs.IsEmpty() {
			log.Trace(fmt.Sprintf("Filtering duplicate [getblocks]: blocks=%d",len(blocks)))
			return false,nil
		}
		//isDuplicate=false
	}else {
		if pg.GS!=nil {
			if gs.IsEqual(pg.GS) {
				log.Trace(fmt.Sprintf("Filtering duplicate [getblocks]: gs=%s",gs.String()))
				return false,nil
			}
		}
	}
	return true,bs
}

func (pg *PrevGet) Update(gs *dag.GraphState,blocks []*hash.Hash) {
	pg.BlocksMtx.Lock()
	defer pg.BlocksMtx.Unlock()

	// Filter duplicate getblocks requests.
	if len(blocks)>0 {
		pg.Blocks.AddList(blocks)
	}else {
		pg.GS=gs
	}
}

/*func (pg *PrevGet) GetLocatorHeight() uint64 {
	for _,h:=range pg.Locator.GetMap() {
		return h.(uint64)
	}
	return 0
}*/
