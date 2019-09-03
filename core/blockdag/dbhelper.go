package blockdag

import (
	"fmt"
	"github.com/Qitmeer/qitmeer-lib/common/hash"
	"github.com/Qitmeer/qitmeer/core/dbnamespace"
	"github.com/Qitmeer/qitmeer/database"
)

// DBPutDAGBlock stores the information needed to reconstruct the provided
// block in the block index according to the format described above.
func DBPutDAGBlock(dbTx database.Tx, block IBlock,status byte) error {
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

	blockHash:=*block.GetHash()
	value:=blockHash.Bytes()
	value=append(value,status)
	if len(value) != hash.HashSize+1 {
		return fmt.Errorf("len is error")
	}
	return bucket.Put(key,value)
}

// DBGetDAGBlock get dag block data by resouce ID
func DBGetDAGBlock(dbTx database.Tx,id uint) (*hash.Hash,byte,error) {
	bucket := dbTx.Metadata().Bucket(dbnamespace.BlockIndexBucketName)
	var serializedID [4]byte
	dbnamespace.ByteOrder.PutUint32(serializedID[:], uint32(id))

	data:=bucket.Get(serializedID[:])
	var h hash.Hash
	err:=h.SetBytes(data[:hash.HashSize])
	if err!=nil {
		return nil,0,err
	}
	status:=data[hash.HashSize:]

	return &h,status[0],nil
}

func GetOrderLogStr(order uint) string {
	if order == MaxBlockOrder {
		return "uncertainty"
	}
	return fmt.Sprintf("%d",order)
}