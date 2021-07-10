package blockdag

import "time"

type DAGSnapshot struct {
	block            IBlock
	tips             *IdSet
	lastTime         time.Time
	diffAnticone     *IdSet
	mainChainTip     uint
	mainChainGenesis uint
	orders           *IdSet
}

func (d *DAGSnapshot) Clean() {
	d.block = nil
	d.tips = nil
	d.diffAnticone = nil
	d.mainChainTip = MaxId
	d.mainChainGenesis = MaxId
	d.orders.Clean()
}

func (d *DAGSnapshot) AddOrder(ib IBlock) {
	if d.IsValid() {
		if d.orders.Has(ib.GetID()) {
			log.Error("DAG snapshot orders is already exit %s", ib.GetHash())
		}
		d.orders.AddPair(ib.GetID(), &BlockOrderHelp{OldOrder: ib.GetOrder(), Block: ib})
	}
}

func (d *DAGSnapshot) IsValid() bool {
	if d.block == nil {
		return false
	}
	return d.block.GetID() != 0
}

func NewDAGSnapshot() *DAGSnapshot {
	return &DAGSnapshot{orders: NewIdSet()}
}
