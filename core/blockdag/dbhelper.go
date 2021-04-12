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

	bucket = dbTx.Metadata().Bucket(dbnamespace.BlockIdBucketName)
	key = block.GetHash()[:]
	return bucket.Put(key, serializedID[:])
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

// block order

// errNotInMainChain signifies that a block hash or height that is not in the
// main chain was requested.
type errNotInMainChain string

// Error implements the error interface.
func (e errNotInMainChain) Error() string {
	return string(e)
}

func DBPutBlockIdByOrder(dbTx database.Tx, order uint, id uint) error {
	// Serialize the order for use in the index entries.
	var serializedOrder [4]byte
	dbnamespace.ByteOrder.PutUint32(serializedOrder[:], uint32(order))

	var serializedID [4]byte
	dbnamespace.ByteOrder.PutUint32(serializedID[:], uint32(id))

	// Add the block order to id mapping to the index.
	bucket := dbTx.Metadata().Bucket(dbnamespace.OrderIdBucketName)
	return bucket.Put(serializedOrder[:], serializedID[:])
}

func DBGetBlockIdByOrder(dbTx database.Tx, order uint) (uint32, error) {
	var serializedOrder [4]byte
	dbnamespace.ByteOrder.PutUint32(serializedOrder[:], uint32(order))

	bucket := dbTx.Metadata().Bucket(dbnamespace.OrderIdBucketName)
	idBytes := bucket.Get(serializedOrder[:])
	if idBytes == nil {
		str := fmt.Sprintf("no block at order %d exists", order)
		return uint32(MaxId), errNotInMainChain(str)
	}
	return dbnamespace.ByteOrder.Uint32(idBytes), nil
}

func DBGetBlockIdByHash(dbTx database.Tx, h *hash.Hash) (uint32, error) {
	bucket := dbTx.Metadata().Bucket(dbnamespace.BlockIdBucketName)
	data := bucket.Get(h[:])
	if data == nil {
		return uint32(MaxId), fmt.Errorf("get dag block error")
	}
	return dbnamespace.ByteOrder.Uint32(data), nil
}
