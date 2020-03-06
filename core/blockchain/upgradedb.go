package blockchain

import (
	"fmt"
	"github.com/Qitmeer/qitmeer/core/dbnamespace"
	"github.com/Qitmeer/qitmeer/database"
	"time"
)

// update db to new version
func (b *BlockChain) upgradeDB() error {
	if b.dbInfo.version == currentDatabaseVersion {
		return nil
	}
	log.Info(fmt.Sprintf("Update cur db to new version: version(%d) -> version(%d)", b.dbInfo.version, currentDatabaseVersion))
	err := b.db.View(func(dbTx database.Tx) error {
		meta := dbTx.Metadata()
		bidxStart := time.Now()

		serializedData := meta.Get(dbnamespace.ChainStateKeyName)
		if serializedData == nil {
			return nil
		}

		state, err := deserializeBestChainState(serializedData)
		if err != nil {
			return err
		}

		// Create the bucket that houses the block hash data.
		_, err = meta.CreateBucket(dbnamespace.BlockHashBucketName)
		if err != nil {
			return err
		}

		err = b.bd.UpgradeDB(dbTx, uint(state.total))
		if err != nil {
			return err
		}
		b.dbInfo = &databaseInfo{
			version: currentDatabaseVersion,
			compVer: currentCompressionVersion,
			bidxVer: currentBlockIndexVersion,
			created: time.Now(),
		}
		err = dbPutDatabaseInfo(dbTx, b.dbInfo)
		if err != nil {
			return err
		}

		log.Info(fmt.Sprintf("Update db version:time=%v", time.Since(bidxStart)))
		return nil
	})
	return err
}
