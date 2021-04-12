package blockdag

import (
	"github.com/Qitmeer/qitmeer/database"
)

// update db to new version
func (bd *BlockDAG) UpgradeDB(dbTx database.Tx) error {
	return nil
}
