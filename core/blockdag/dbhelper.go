package blockdag

import (
	"github.com/HalalChain/qitmeer-lib/common/hash"
	"github.com/HalalChain/qitmeer/core/dbnamespace"
	"github.com/HalalChain/qitmeer/database"
)

// DBPutDAGBlock stores the information needed to reconstruct the provided
// block in the block index according to the format described above.
func DBPutDAGBlock(dbTx database.Tx, block IBlock) error {
	// TODO save DAG block to accelerating Reconfiguration
	/*var w bytes.Buffer
	err:=block.Encode(&w)
	if err != nil {
		return err
	}*/
	bucket := dbTx.Metadata().Bucket(dbnamespace.BlockIndexBucketName)
	var serializedID [4]byte
	dbnamespace.ByteOrder.PutUint32(serializedID[:], uint32(block.GetID()))

	key := serializedID[:]
	return bucket.Put(key,block.GetHash()[:])
}

// DBGetDAGBlock get dag block data by resouce ID
func DBGetDAGBlock(dbTx database.Tx,id uint) (*hash.Hash,error) {
	bucket := dbTx.Metadata().Bucket(dbnamespace.BlockIndexBucketName)
	var serializedID [4]byte
	dbnamespace.ByteOrder.PutUint32(serializedID[:], uint32(id))

	data:=bucket.Get(serializedID[:])
	var h hash.Hash
	err:=h.SetBytes(data)
	if err!=nil {
		return nil,err
	}
	return &h,nil
}
