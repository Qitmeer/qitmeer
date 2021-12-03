package index

import (
	"fmt"
	"github.com/Qitmeer/qng-core/common/hash"
	"github.com/Qitmeer/qng-core/core/types"
	"github.com/Qitmeer/qng-core/database"
)

var (
	itxIndexKey             = []byte("invalid_txbyhashidx")
	itxidByTxhashBucketName = []byte("invalid_txidbytxhash")
)

func dbAddInvalidTxIndexEntries(dbTx database.Tx, block *types.SerializedBlock, blockID uint32) error {
	addEntries := func(txns []*types.Tx, txLocs []types.TxLoc, blockID uint32) error {
		offset := 0
		serializedValues := make([]byte, len(txns)*txEntrySize)
		for i, tx := range txns {
			putTxIndexEntry(serializedValues[offset:], blockID,
				txLocs[i])
			endOffset := offset + txEntrySize

			if !tx.IsDuplicate {
				if err := dbPutInvalidTxIndexEntry(dbTx, tx.Hash(),
					serializedValues[offset:endOffset:endOffset]); err != nil {
					return err
				}
				if err := dbPutInvalidTxIdByHash(dbTx, tx.Tx.TxHashFull(), tx.Hash()); err != nil {
					return err
				}
			}
			offset += txEntrySize
		}
		return nil
	}
	txLocs, err := block.TxLoc()
	if err != nil {
		return err
	}

	err = addEntries(block.Transactions(), txLocs, blockID)
	if err != nil {
		return err
	}

	return nil
}

func dbPutInvalidTxIndexEntry(dbTx database.Tx, txHash *hash.Hash, serializedData []byte) error {
	itxIndex := dbTx.Metadata().Bucket(itxIndexKey)
	return itxIndex.Put(txHash[:], serializedData)
}

func dbPutInvalidTxIdByHash(dbTx database.Tx, txHash hash.Hash, txId *hash.Hash) error {
	itxidByTxhash := dbTx.Metadata().Bucket(itxidByTxhashBucketName)
	return itxidByTxhash.Put(txHash[:], txId[:])
}

func dbRemoveInvalidTxIndexEntries(dbTx database.Tx, block *types.SerializedBlock) error {
	removeEntries := func(txns []*types.Tx) error {
		for _, tx := range txns {
			region, _ := dbFetchInvalidTxIndexEntry(dbTx, tx.Hash())
			if region != nil && !region.Hash.IsEqual(block.Hash()) {
				continue
			}
			if err := dbRemoveInvalidTxIndexEntry(dbTx, tx.Hash()); err != nil {
				return err
			}
			if err := dbRemoveInvalidTxIdByHash(dbTx, tx.Tx.TxHashFull()); err != nil {
				return err
			}
		}
		return nil
	}
	if err := removeEntries(block.Transactions()); err != nil {
		return err
	}
	return nil
}

func dbFetchInvalidTxIndexEntry(dbTx database.Tx, txid *hash.Hash) (*database.BlockRegion, error) {
	itxIndex := dbTx.Metadata().Bucket(itxIndexKey)
	serializedData := itxIndex.Get(txid[:])
	if len(serializedData) == 0 {
		return nil, nil
	}

	// Ensure the serialized data has enough bytes to properly deserialize.
	if len(serializedData) < 12 {
		return nil, database.Error{
			ErrorCode: database.ErrCorruption,
			Description: fmt.Sprintf("corrupt transaction index "+
				"entry for %s", txid),
		}
	}

	// Load the block hash associated with the block ID.
	h, err := dbFetchBlockHashBySerializedID(dbTx, serializedData[0:4])
	if err != nil {
		return nil, database.Error{
			ErrorCode: database.ErrCorruption,
			Description: fmt.Sprintf("corrupt transaction index "+
				"entry for %s: %v", txid, err),
		}
	}

	// Deserialize the final entry.
	region := database.BlockRegion{Hash: &hash.Hash{}}
	copy(region.Hash[:], h[:])
	region.Offset = byteOrder.Uint32(serializedData[4:8])
	region.Len = byteOrder.Uint32(serializedData[8:12])

	return &region, nil
}

func dbRemoveInvalidTxIndexEntry(dbTx database.Tx, txHash *hash.Hash) error {
	itxIndex := dbTx.Metadata().Bucket(itxIndexKey)
	serializedData := itxIndex.Get(txHash[:])
	if len(serializedData) == 0 {
		return nil
	}

	return itxIndex.Delete(txHash[:])
}

func dbRemoveInvalidTxIdByHash(dbTx database.Tx, txhash hash.Hash) error {
	itxidByTxhash := dbTx.Metadata().Bucket(itxidByTxhashBucketName)
	serializedData := itxidByTxhash.Get(txhash[:])
	if len(serializedData) == 0 {
		return nil
	}
	return itxidByTxhash.Delete(txhash[:])
}

func dbFetchInvalidTxIdByHash(dbTx database.Tx, txhash hash.Hash) (*hash.Hash, error) {
	itxidByTxhash := dbTx.Metadata().Bucket(itxidByTxhashBucketName)
	serializedData := itxidByTxhash.Get(txhash[:])
	if serializedData == nil {
		return nil, errNoTxHashEntry
	}
	txId := hash.Hash{}
	txId.SetBytes(serializedData[:])

	return &txId, nil
}

func (idx *TxIndex) InvalidTxBlockRegion(id hash.Hash) (*database.BlockRegion, error) {
	var region *database.BlockRegion
	err := idx.db.View(func(dbTx database.Tx) error {
		var err error
		region, err = dbFetchInvalidTxIndexEntry(dbTx, &id)
		return err
	})
	return region, err
}

func (idx *TxIndex) GetInvalidTxIdByHash(txhash hash.Hash) (*hash.Hash, error) {
	var txid *hash.Hash
	err := idx.db.View(func(dbTx database.Tx) error {
		var err error
		id, err := dbFetchInvalidTxIdByHash(dbTx, txhash)
		if err != nil {
			return err
		}
		txid = id
		return nil
	})
	return txid, err
}

// Because there was a problem with the previous version[1234f -> 4f610] of droptxindex,
// in order to be compatible with its data.
func (idx *TxIndex) compatibleOldData(dbTx database.Tx) {
	meta := dbTx.Metadata()
	if meta.Bucket(itxIndexKey) != nil {
		meta.DeleteBucket(itxIndexKey)
	}
	if meta.Bucket(itxidByTxhashBucketName) != nil {
		meta.DeleteBucket(itxidByTxhashBucketName)
	}
}
