package blockdag

import (
	"fmt"
	"github.com/Qitmeer/qitmeer/core/dbnamespace"
	"github.com/Qitmeer/qitmeer/database"
	"time"
)

// update db to new version
func (bd *BlockDAG) UpgradeDB(dbTx database.Tx) error {
	return bd.upgradeMainChain(dbTx)
}

func (bd *BlockDAG) upgradeMainChain(dbTx database.Tx) error {
	meta := dbTx.Metadata()
	mchBucket := meta.Bucket(dbnamespace.DagMainChainBucketName)
	if mchBucket != nil {
		return nil
	}
	// Need build
	mchBucket, err := meta.CreateBucket(dbnamespace.DagMainChainBucketName)
	if err != nil {
		return err
	}
	umcStart := time.Now()
	mcNum := int(0)

	for cur := bd.getMainChainTip(); cur != nil; cur = bd.getBlockById(cur.GetMainParent()) {
		err = DBPutMainChainBlock(dbTx, cur.GetID())
		mcNum++
		if err != nil {
			return err
		}
		if cur.GetMainParent() == MaxId {
			break
		}
	}
	log.Info(fmt.Sprintf("Build DAG main chain bucket:%v/%d", time.Since(umcStart), mcNum))
	return nil
}
