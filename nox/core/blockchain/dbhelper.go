// Copyright (c) 2017-2018 The nox developers
package blockchain

import (
	"github.com/noxproject/nox/core/dbnamespace"
	"github.com/noxproject/nox/database"
	"github.com/noxproject/nox/common/hash"
	"github.com/noxproject/nox/core/types"
	"fmt"
)

// errNotInMainChain signifies that a block hash or height that is not in the
// main chain was requested.
type errNotInMainChain string

// Error implements the error interface.
func (e errNotInMainChain) Error() string {
	return string(e)
}

// DBMainChainHasBlock is the exported version of dbMainChainHasBlock.
func DBMainChainHasBlock(dbTx database.Tx, hash *hash.Hash) bool {
	return dbMainChainHasBlock(dbTx, hash)
}

// dbMainChainHasBlock uses an existing database transaction to return whether
// or not the main chain contains the block identified by the provided hash.
func dbMainChainHasBlock(dbTx database.Tx, hash *hash.Hash) bool {
	hashIndex := dbTx.Metadata().Bucket(dbnamespace.HashIndexBucketName)
	return hashIndex.Get(hash[:]) != nil
}

// DBFetchBlockByHeight is the exported version of dbFetchBlockByHeight.
func DBFetchBlockByHeight(dbTx database.Tx, height int64) (*types.SerializedBlock, error) {
	return dbFetchBlockByHeight(dbTx, height)
}
// dbFetchBlockByHeight uses an existing database transaction to retrieve the
// raw block for the provided height, deserialize it, and return a Block
// with the height set.
func dbFetchBlockByHeight(dbTx database.Tx, height int64) (*types.SerializedBlock, error) {
	// First find the hash associated with the provided height in the index.
	h, err := dbFetchHashByHeight(dbTx, height)
	if err != nil {
		return nil, err
	}

	// Load the raw block bytes from the database.
	blockBytes, err := dbTx.FetchBlock(h)
	if err != nil {
		return nil, err
	}

	// Create the encapsulated block and set the height appropriately.
	block, err := types.NewBlockFromBytes(blockBytes)
	if err != nil {
		return nil, err
	}

	return block, nil
}

// dbFetchHashByHeight uses an existing database transaction to retrieve the
// hash for the provided height from the index.
func dbFetchHashByHeight(dbTx database.Tx, height int64) (*hash.Hash, error) {
	var serializedHeight [4]byte
	dbnamespace.ByteOrder.PutUint32(serializedHeight[:], uint32(height))

	meta := dbTx.Metadata()
	heightIndex := meta.Bucket(dbnamespace.HeightIndexBucketName)
	hashBytes := heightIndex.Get(serializedHeight[:])
	if hashBytes == nil {
		str := fmt.Sprintf("no block at height %d exists", height)
		return nil, errNotInMainChain(str)
	}

	var h hash.Hash
	copy(h[:], hashBytes)
	return &h, nil
}

