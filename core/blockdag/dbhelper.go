package blockdag

import (
	"bytes"
	"fmt"
	"github.com/Qitmeer/qitmeer/common/hash"
	"github.com/Qitmeer/qitmeer/core/dbnamespace"
	"github.com/Qitmeer/qitmeer/database"
)

const (
	DAGErrorEmpty = "empty"
)

type DAGError struct {
	err string
}

func (e *DAGError) Error() string {
	return e.err
}

func (e *DAGError) IsEmpty() bool {
	return e.Error() == DAGErrorEmpty
}

func NewDAGError(e error) error {
	if e == nil {
		return nil
	}
	return &DAGError{e.Error()}
}

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
		return &DAGError{DAGErrorEmpty}
	}

	return NewDAGError(block.Decode(bytes.NewReader(data)))
}

func DBDelDAGBlock(dbTx database.Tx, id uint) error {
	bucket := dbTx.Metadata().Bucket(dbnamespace.BlockIndexBucketName)
	var serializedID [4]byte
	dbnamespace.ByteOrder.PutUint32(serializedID[:], uint32(id))
	return bucket.Delete(serializedID[:])
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
		return uint32(MaxId), &DAGError{str}
	}
	return dbnamespace.ByteOrder.Uint32(idBytes), nil
}

func DBPutDAGBlockIdByHash(dbTx database.Tx, block IBlock) error {
	var serializedID [4]byte
	dbnamespace.ByteOrder.PutUint32(serializedID[:], uint32(block.GetID()))

	bucket := dbTx.Metadata().Bucket(dbnamespace.BlockIdBucketName)
	key := block.GetHash()[:]
	return bucket.Put(key, serializedID[:])
}

func DBGetBlockIdByHash(dbTx database.Tx, h *hash.Hash) (uint32, error) {
	bucket := dbTx.Metadata().Bucket(dbnamespace.BlockIdBucketName)
	data := bucket.Get(h[:])
	if data == nil {
		return uint32(MaxId), fmt.Errorf("get dag block error")
	}
	return dbnamespace.ByteOrder.Uint32(data), nil
}


func DBDelBlockIdByHash(dbTx database.Tx, h *hash.Hash) error {
	bucket := dbTx.Metadata().Bucket(dbnamespace.BlockIdBucketName)
	return bucket.Delete(h[:])
}

// tips
func DBPutDAGTip(dbTx database.Tx, id uint, isMain bool) error {
	var serializedID [4]byte
	dbnamespace.ByteOrder.PutUint32(serializedID[:], uint32(id))

	bucket := dbTx.Metadata().Bucket(dbnamespace.DAGTipsBucketName)
	main := byte(0)
	if isMain {
		main = byte(1)
	}
	return bucket.Put(serializedID[:], []byte{main})
}

func DBGetDAGTips(dbTx database.Tx) ([]uint, error) {
	bucket := dbTx.Metadata().Bucket(dbnamespace.DAGTipsBucketName)
	cursor := bucket.Cursor()
	mainTip := MaxId
	tips := []uint{}
	for cok := cursor.First(); cok; cok = cursor.Next() {
		id := uint(dbnamespace.ByteOrder.Uint32(cursor.Key()))
		main := cursor.Value()
		if len(main) > 0 {
			if main[0] > 0 {
				if mainTip != MaxId {
					return nil, fmt.Errorf("Too many main tip")
				}
				mainTip = id
				continue
			}
		}
		tips = append(tips, id)
	}
	if mainTip == MaxId {
		return nil, fmt.Errorf("Can't find main tip")
	}
	result := []uint{mainTip}
	if len(tips) > 0 {
		result = append(result, tips...)
	}
	return result, nil
}

func DBDelDAGTip(dbTx database.Tx, id uint) error {
	bucket := dbTx.Metadata().Bucket(dbnamespace.DAGTipsBucketName)
	var serializedID [4]byte
	dbnamespace.ByteOrder.PutUint32(serializedID[:], uint32(id))
	return bucket.Delete(serializedID[:])
}

