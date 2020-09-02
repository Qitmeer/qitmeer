package blockdag

import (
	"bytes"
	"fmt"
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
	return bucket.Put(key, buff.Bytes())
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

func DBHasMainChainBlock(dbTx database.Tx, id uint) bool {
	bucket := dbTx.Metadata().Bucket(dbnamespace.DagMainChainBucketName)
	var serializedID [4]byte
	dbnamespace.ByteOrder.PutUint32(serializedID[:], uint32(id))

	data := bucket.Get(serializedID[:])
	return data != nil
}

func DBPutMainChainBlock(dbTx database.Tx, id uint) error {
	bucket := dbTx.Metadata().Bucket(dbnamespace.DagMainChainBucketName)
	var serializedID [4]byte
	dbnamespace.ByteOrder.PutUint32(serializedID[:], uint32(id))

	key := serializedID[:]
	return bucket.Put(key, []byte{0})
}

func DBRemoveMainChainBlock(dbTx database.Tx, id uint) error {
	bucket := dbTx.Metadata().Bucket(dbnamespace.DagMainChainBucketName)
	var serializedID [4]byte
	dbnamespace.ByteOrder.PutUint32(serializedID[:], uint32(id))

	key := serializedID[:]
	return bucket.Delete(key)
}
