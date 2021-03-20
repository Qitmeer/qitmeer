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

// block order

// errNotInMainChain signifies that a block hash or height that is not in the
// main chain was requested.
type errNotInMainChain string

// Error implements the error interface.
func (e errNotInMainChain) Error() string {
	return string(e)
}

// isNotInMainChainErr returns whether or not the passed error is an
// errNotInMainChain error.
func isNotInMainChainErr(err error) bool {
	_, ok := err.(errNotInMainChain)
	return ok
}

// -----------------------------------------------------------------------------
// The block index consists of two buckets with an entry for every block in
// the chain.  One bucket is for the hash to order mapping and the other
// is for the order to hash mapping.
//
// The serialized format for values in the hash to order bucket is:
//   <order>
//
//   Field      Type     Size
//   order     uint32   4 bytes
//
// The serialized format for values in the order to hash bucket is:
//   <hash>
//
//   Field      Type             Size
//   hash       chainhash.Hash   chainhash.HashSize
// -----------------------------------------------------------------------------

// dbPutBlockIndex uses an existing database transaction to update or add
// index entries for the hash to order and order to hash mappings for the
// provided values.
func DBPutBlockIndex(dbTx database.Tx, hash *hash.Hash, order uint64) error {
	// Serialize the order for use in the index entries.
	var serializedOrder [4]byte
	dbnamespace.ByteOrder.PutUint32(serializedOrder[:], uint32(order))

	// Add the block hash to order mapping to the index.
	meta := dbTx.Metadata()
	hashIndex := meta.Bucket(dbnamespace.HashIndexBucketName)
	if err := hashIndex.Put(hash[:], serializedOrder[:]); err != nil {
		return err
	}

	// Add the block order to hash mapping to the index.
	orderIndex := meta.Bucket(dbnamespace.OrderIndexBucketName)
	return orderIndex.Put(serializedOrder[:], hash[:])
}

// dbRemoveBlockIndex uses an existing database transaction remove block
// index entries from the hash to order and order to hash mappings for
// the provided values.
func DBRemoveBlockIndex(dbTx database.Tx, hash *hash.Hash, order int64) error {
	// Remove the block hash to height mapping.
	meta := dbTx.Metadata()
	hashIndex := meta.Bucket(dbnamespace.HashIndexBucketName)
	if err := hashIndex.Delete(hash[:]); err != nil {
		return err
	}

	// Remove the block height to hash mapping.
	var serializedOrdert [4]byte
	dbnamespace.ByteOrder.PutUint32(serializedOrdert[:], uint32(order))
	orderIndex := meta.Bucket(dbnamespace.OrderIndexBucketName)
	return orderIndex.Delete(serializedOrdert[:])
}

func DBFetchOrderByHash(dbTx database.Tx, hash *hash.Hash) (uint64, error) {
	meta := dbTx.Metadata()
	hashIndex := meta.Bucket(dbnamespace.HashIndexBucketName)
	serializedOrder := hashIndex.Get(hash[:])
	if serializedOrder == nil {
		str := fmt.Sprintf("block %s is not in the chain", hash)
		return 0, errNotInMainChain(str)
	}

	return uint64(dbnamespace.ByteOrder.Uint32(serializedOrder)), nil
}

// dbFetchHashByOrder uses an existing database transaction to retrieve the
// hash for the provided order from the index.
func DBFetchHashByOrder(dbTx database.Tx, order uint64) (*hash.Hash, error) {
	var serializedOrder [4]byte
	dbnamespace.ByteOrder.PutUint32(serializedOrder[:], uint32(order))

	meta := dbTx.Metadata()
	orderIndex := meta.Bucket(dbnamespace.OrderIndexBucketName)
	hashBytes := orderIndex.Get(serializedOrder[:])
	if hashBytes == nil {
		str := fmt.Sprintf("no block at order %d exists", order)
		return nil, errNotInMainChain(str)
	}

	var h hash.Hash
	copy(h[:], hashBytes)
	return &h, nil
}
