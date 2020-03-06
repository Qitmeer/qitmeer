package blockdag

import (
	"bytes"
	"fmt"
	"github.com/Qitmeer/qitmeer/common/hash"
	"github.com/Qitmeer/qitmeer/core/dbnamespace"
	"github.com/Qitmeer/qitmeer/database"
)

// DBPutDAGBlock stores the information needed to reconstruct the provided
// block in the block index according to the format described above.
func DBPutDAGBlock(dbTx database.Tx, block IBlock) error {
	bucket := dbTx.Metadata().Bucket(dbnamespace.BlockIndexBucketName)
	var serializedID [4]byte
	dbnamespace.ByteOrder.PutUint32(serializedID[:], uint32(block.GetID()))

	key := serializedID[:]

	var buff bytes.Buffer
	err := block.Encode(&buff)
	if err != nil {
		return err
	}
	err = bucket.Put(key, buff.Bytes())
	if err != nil {
		return err
	}
	// block hash -> block id
	return DBPutDAGBlockId(dbTx, block)
}

// DBPutDAGBlockId
func DBPutDAGBlockId(dbTx database.Tx, block IBlock) error {
	var serializedID [4]byte
	dbnamespace.ByteOrder.PutUint32(serializedID[:], uint32(block.GetID()))

	// block hash -> block id
	hashId := dbTx.Metadata().Bucket(dbnamespace.BlockHashBucketName)
	err := hashId.Put(block.GetHash()[:], serializedID[:])
	if err != nil {
		return err
	}
	return nil
}

// DBGetDAGBlock get dag block data by resouce ID
func DBGetDAGBlock(dbTx database.Tx, block IBlock) error {
	bucket := dbTx.Metadata().Bucket(dbnamespace.BlockIndexBucketName)
	var serializedID [4]byte
	dbnamespace.ByteOrder.PutUint32(serializedID[:], uint32(block.GetID()))

	data := bucket.Get(serializedID[:])
	if data == nil {
		return fmt.Errorf("get dag block error")
	}

	return block.Decode(bytes.NewReader(data))
}

// DBGetDAGBlockId get dag block id by block hash
func DBGetDAGBlockId(dbTx database.Tx, h *hash.Hash) (uint32, error) {
	hashId := dbTx.Metadata().Bucket(dbnamespace.BlockHashBucketName)
	serializedID := hashId.Get(h[:])
	if serializedID == nil {
		return 0, fmt.Errorf("get dag block error")
	}
	return dbnamespace.ByteOrder.Uint32(serializedID), nil
}

func GetOrderLogStr(order uint) string {
	if order == MaxBlockOrder {
		return "uncertainty"
	}
	return fmt.Sprintf("%d", order)
}

func DBPutDAGInfo(dbTx database.Tx, bd *BlockDAG) error {
	var buff bytes.Buffer
	err := bd.Encode(&buff)
	if err != nil {
		return err
	}
	return dbTx.Metadata().Put(dbnamespace.DagInfoBucketName, buff.Bytes())
}
