package blockdag

import (
	"github.com/Qitmeer/qitmeer/database"
)

// update db to new version
func (b *BlockDAG) UpgradeDB(dbTx database.Tx, blockTotal uint) error {

	for i := uint(0); i < blockTotal; i++ {
		block := Block{id: i}
		err := DBGetDAGBlock(b.dbTx, &block)
		if err != nil {
			return err
		}

		err = DBPutDAGBlockId(dbTx, &block)
		if err != nil {
			return err
		}
	}
	return nil
}
