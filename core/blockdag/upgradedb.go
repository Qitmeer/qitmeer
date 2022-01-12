package blockdag

import (
	"fmt"
	"github.com/Qitmeer/qitmeer/common/hash"
	"github.com/Qitmeer/qitmeer/core/dbnamespace"
	"github.com/Qitmeer/qitmeer/database"
)

// update db to new version
func (bd *BlockDAG) UpgradeDB(dbTx database.Tx, mainTip *hash.Hash, total uint64, genesis *hash.Hash) error {
	bucket := dbTx.Metadata().Bucket(dbnamespace.DAGTipsBucketName)
	cursor := bucket.Cursor()
	if cursor.First() {
		return fmt.Errorf("Data format error: already exists tips")
	}
	log.Info(fmt.Sprintf("Start upgrade MeerDAG.tips🛠 (total=%d mainTip=%s)", total, mainTip.String()))

	blocks := map[uint]IBlock{}
	var tips *IdSet
	var mainTipBlock IBlock

	getBlockById := func(id uint) IBlock {
		if id == MaxId {
			return nil
		}
		block, ok := blocks[id]
		if !ok {
			return nil
		}
		return block
	}

	updateTips := func(b IBlock) {
		if tips == nil {
			tips = NewIdSet()
			tips.AddPair(b.GetID(), b)
			return
		}
		for k, v := range tips.GetMap() {
			block := v.(IBlock)
			if block.HasChildren() {
				tips.Remove(k)
			}
		}
		tips.AddPair(b.GetID(), b)
	}

	for i := uint(0); i < uint(total); i++ {
		block := Block{id: i}
		ib := &PhantomBlock{&block, 0, NewIdSet(), NewIdSet()}
		err := DBGetDAGBlock(dbTx, ib)
		if err != nil {
			if err.(*DAGError).IsEmpty() {
				continue
			}
			return err
		}
		if i == 0 && !ib.GetHash().IsEqual(genesis) {
			return fmt.Errorf("genesis data mismatch")
		}
		if ib.HasParents() {
			parentsSet := NewIdSet()
			for k := range ib.GetParents().GetMap() {
				parent := getBlockById(k)
				parentsSet.AddPair(k, parent)
				parent.AddChild(ib)
			}
			ib.GetParents().Clean()
			ib.GetParents().AddSet(parentsSet)
		}
		blocks[ib.GetID()] = ib

		updateTips(ib)

		if ib.GetHash().IsEqual(mainTip) {
			mainTipBlock = ib
		}
	}
	if mainTipBlock == nil || tips == nil || tips.IsEmpty() || !tips.Has(mainTipBlock.GetID()) {
		return fmt.Errorf("Main chain tip error")
	}

	for k := range tips.GetMap() {
		err := DBPutDAGTip(dbTx, k, k == mainTipBlock.GetID())
		if err != nil {
			return err
		}
	}
	log.Info(fmt.Sprintf("End upgrade MeerDAG.tips🛠:bridging tips num(%d)", tips.Size()))
	return nil
}

