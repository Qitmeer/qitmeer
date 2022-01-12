package blockchain

import (
	"fmt"
	"github.com/Qitmeer/qitmeer/common/roughtime"
	"github.com/Qitmeer/qitmeer/core/dbnamespace"
	"github.com/Qitmeer/qitmeer/core/serialization"
	"github.com/Qitmeer/qitmeer/database"
)

// update db to new version
func (b *BlockChain) upgradeDB() error {
	if b.dbInfo.version == currentDatabaseVersion {
		return nil
	}
	log.Info(fmt.Sprintf("Update cur db to new version: version(%d) -> version(%d) ...", b.dbInfo.version, currentDatabaseVersion))
	err := b.db.Update(func(dbTx database.Tx) error {
		bidxStart := roughtime.Now()

		meta := dbTx.Metadata()
		serializedData := meta.Get(dbnamespace.ChainStateKeyName)
		if serializedData == nil {
			return nil
		}
		state, err := DeserializeBestChainState(serializedData)
		if err != nil {
			return err
		}

		err = b.bd.UpgradeDB(dbTx,&state.hash,state.total,b.params.GenesisHash)
		if err != nil {
			return err
		}

		// save
		b.dbInfo = &databaseInfo{
			version: currentDatabaseVersion,
			compVer: serialization.CurrentCompressionVersion,
			bidxVer: currentBlockIndexVersion,
			created: roughtime.Now(),
		}
		err = dbPutDatabaseInfo(dbTx, b.dbInfo)
		if err != nil {
			return err
		}

		log.Info(fmt.Sprintf("Finish update db version:time=%v", roughtime.Since(bidxStart)))
		return nil
	})
	if err != nil {
		return fmt.Errorf("You can cleanup your block data base by '--cleanup'.Your data is too old (%d -> %d). %s\n", b.dbInfo.version, currentDatabaseVersion,err)
	}
	return nil
}
