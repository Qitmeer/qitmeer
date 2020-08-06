package blockchain

import (
	"fmt"
	"github.com/Qitmeer/qitmeer/common/hash"
	"github.com/Qitmeer/qitmeer/core/dbnamespace"
	"github.com/Qitmeer/qitmeer/core/types"
	"github.com/Qitmeer/qitmeer/database"
	"time"
)

// update db to new version
func (b *BlockChain) upgradeDB() error {
	if b.dbInfo.version == currentDatabaseVersion {
		return nil
	}
	log.Info(fmt.Sprintf("Update cur db to new version: version(%d) -> version(%d) ...", b.dbInfo.version, currentDatabaseVersion))
	err := b.db.Update(func(dbTx database.Tx) error {
		meta := dbTx.Metadata()
		bidxStart := time.Now()
		spendBucket := meta.Bucket(dbnamespace.SpendJournalBucketName)
		if spendBucket == nil {
			return nil
		}
		err := checkSpendJournal(dbTx, spendBucket)
		if err != nil {
			return err
		}
		// save
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
	if err != nil {
		return fmt.Errorf("You can cleanup your block data base by '--cleanup'.The data is corrupted and cannot be upgraded (%s). ", err)
	}
	return err
}

func checkSpendJournal(dbTx database.Tx, spendBucket database.Bucket) (err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("%v", r)
		}
	}()
	var blockHash *hash.Hash
	var block *types.SerializedBlock
	cursor := spendBucket.Cursor()
	for ok := cursor.First(); ok; ok = cursor.Next() {
		serialized := spendBucket.Get(cursor.Key())
		if len(serialized) <= 0 {
			continue
		}

		blockHash, err = hash.NewHash(cursor.Key())
		if err != nil {
			return err
		}
		block, err = dbFetchBlockByHash(dbTx, blockHash)
		if err != nil {
			return err
		}
		_, err = dbFetchSpendJournalEntry(dbTx, block)
		if err != nil {
			return err
		}
	}
	return nil
}
