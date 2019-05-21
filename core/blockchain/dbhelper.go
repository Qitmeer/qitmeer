// Copyright (c) 2017-2018 The nox developers
package blockchain

import (
	"qitmeer/core/dbnamespace"
	"qitmeer/database"
	"qitmeer/common/hash"
	"qitmeer/core/types"
	"fmt"
	"time"
	"encoding/binary"
	"math/big"
	"bytes"
	"sort"
)


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

// errDeserialize signifies that a problem was encountered when deserializing
// data.
type errDeserialize string

// Error implements the error interface.
func (e errDeserialize) Error() string {
	return string(e)
}

// isDeserializeErr returns whether or not the passed error is an errDeserialize
// error.
func isDeserializeErr(err error) bool {
	_, ok := err.(errDeserialize)
	return ok
}

// -----------------------------------------------------------------------------
// The database information contains information about the version and date
// of the blockchain database.
//
// It consists of a separate key for each individual piece of information:
//
//   Key        Value    Size      Description
//   version    uint32   4 bytes   The version of the database
//   compver    uint32   4 bytes   The script compression version of the database
//   bidxver    uint32   4 bytes   The block index version of the database
//   created    uint64   8 bytes   The date of the creation of the database
// -----------------------------------------------------------------------------

// databaseInfo is the structure for a database.
type databaseInfo struct {
	version uint32
	compVer uint32
	bidxVer uint32
	created time.Time
}

// blockIndexEntry represents a block index database entry.
type blockIndexEntry struct {
	header         types.BlockHeader
	status         blockStatus
}

// blockIndexKey generates the binary key for an entry in the block index
// bucket.  The key is composed of the block height encoded as a big-endian
// 32-bit unsigned int followed by the 32 byte block hash.  Big endian is used
// here so the entries can easily be iterated by height.
func blockIndexKey(blockHash *hash.Hash, blockHeight uint32) []byte {
	indexKey := make([]byte, hash.HashSize+4)
	binary.BigEndian.PutUint32(indexKey[0:4], blockHeight)
	copy(indexKey[4:hash.HashSize+4], blockHash[:])
	return indexKey
}

// -----------------------------------------------------------------------------
// The best chain state consists of the best block hash and height, the total
// number of transactions up to and including those in the best block, the
// total coin supply, the subsidy at the current block, the subsidy of the
// block prior (for rollbacks), and the accumulated work sum up to and
// including the best block.
//
// The serialized format is:
//
//   <block hash><block height><total txns><total subsidy><work sum length><work sum>
//
//   Field             Type             Size
//   block hash        chainhash.Hash   chainhash.HashSize
//   block height      uint32           4 bytes
//   total txns        uint64           8 bytes
//   total subsidy     int64            8 bytes
//   work sum length   uint32           4 bytes
//   work sum          big.Int          work sum length
// -----------------------------------------------------------------------------

// bestChainState represents the data to be stored the database for the current
// best chain state.
type bestChainState struct {
	hash         hash.Hash
	height       uint64
	totalTxns    uint64
	totalSubsidy int64
	workSum      *big.Int
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
func DBFetchBlockByHeight(dbTx database.Tx, height uint64) (*types.SerializedBlock, error) {
	return dbFetchBlockByHeight(dbTx, height)
}
// dbFetchBlockByHeight uses an existing database transaction to retrieve the
// raw block for the provided height, deserialize it, and return a Block
// with the height set.
func dbFetchBlockByHeight(dbTx database.Tx, height uint64) (*types.SerializedBlock, error) {
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
func dbFetchHashByHeight(dbTx database.Tx, height uint64) (*hash.Hash, error) {
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


// BlockByHeight returns the block at the given height in the main chain.
//
// This function is safe for concurrent access.
func (b *BlockChain) BlockByOrder(blockOrder uint64) (*types.SerializedBlock, error) {
	var block *types.SerializedBlock
	err := b.db.View(func(dbTx database.Tx) error {
		var err error
		block, err = dbFetchBlockByHeight(dbTx, blockOrder)
		return err
	})
	return block, err
}

// BlockHashByHeight returns the hash of the block at the given height in the
// main chain.
//
// This function is safe for concurrent access.
func (b *BlockChain) BlockHashByHeight(blockHeight uint64) (*hash.Hash, error) {
	var hash *hash.Hash
	err := b.db.View(func(dbTx database.Tx) error {
		var err error
		hash, err = dbFetchHashByHeight(dbTx, blockHeight)
		return err
	})
	return hash, err
}

// MainChainHasBlock returns whether or not the block with the given hash is in
// the main chain.
//
// This function is safe for concurrent access.
func (b *BlockChain) MainChainHasBlock(hash *hash.Hash) (bool, error) {
	var exists bool
	err := b.db.View(func(dbTx database.Tx) error {
		exists = dbMainChainHasBlock(dbTx, hash)
		return nil
	})
	return exists, err
}

// dbFetchDatabaseInfo uses an existing database transaction to fetch the
// database versioning and creation information.
func dbFetchDatabaseInfo(dbTx database.Tx) (*databaseInfo, error) {
	meta := dbTx.Metadata()
	bucket := meta.Bucket(dbnamespace.BCDBInfoBucketName)

	// Uninitialized state.
	if bucket == nil {
		return nil, nil
	}

	// Load the database version.
	var version uint32
	versionBytes := bucket.Get(dbnamespace.BCDBInfoVersionKeyName)
	if versionBytes != nil {
		version = dbnamespace.ByteOrder.Uint32(versionBytes)
	}

	// Load the database compression version.
	var compVer uint32
	compVerBytes := bucket.Get(dbnamespace.BCDBInfoCompressionVersionKeyName)
	if compVerBytes != nil {
		compVer = dbnamespace.ByteOrder.Uint32(compVerBytes)
	}

	// Load the database block index version.
	var bidxVer uint32
	bidxVerBytes := bucket.Get(dbnamespace.BCDBInfoBlockIndexVersionKeyName)
	if bidxVerBytes != nil {
		bidxVer = dbnamespace.ByteOrder.Uint32(bidxVerBytes)
	}

	// Load the database creation date.
	var created time.Time
	createdBytes := bucket.Get(dbnamespace.BCDBInfoCreatedKeyName)
	if createdBytes != nil {
		ts := dbnamespace.ByteOrder.Uint64(createdBytes)
		created = time.Unix(int64(ts), 0)
	}

	return &databaseInfo{
		version: version,
		compVer: compVer,
		bidxVer: bidxVer,
		created: created,
	}, nil
}

// createChainState initializes both the database and the chain state to the
// genesis block.  This includes creating the necessary buckets and inserting
// the genesis block, so it must only be called on an uninitialized database.
func (b *BlockChain) createChainState() error {
	// Create a new node from the genesis block and set it as the best node.
	genesisBlock := types.NewBlock(b.params.GenesisBlock)
	genesisBlock.SetOrder(0)
	header := &genesisBlock.Block().Header
	node := newBlockNode(header, nil)
	node.status = statusDataStored | statusValid
	b.bd.AddBlock(node)
	node.SetOrder(0)
	b.index.addNode(node)
	// Initialize the state related to the best block.  Since it is the
	// genesis block, use its timestamp for the median time.
	numTxns := uint64(len(genesisBlock.Block().Transactions))
	blockSize := uint64(genesisBlock.Block().SerializeSize())
	b.stateSnapshot= newBestState(node, blockSize, numTxns,
		time.Unix(node.timestamp,0), numTxns,0,b.bd.GetGraphState())

	// Create the initial the database chain state including creating the
	// necessary index buckets and inserting the genesis block.
	err := b.db.Update(func(dbTx database.Tx) error {
		meta := dbTx.Metadata()

		// Create the bucket that houses information about the database's
		// creation and version.
		_, err := meta.CreateBucket(dbnamespace.BCDBInfoBucketName)
		if err != nil {
			return err
		}

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

		// Create the bucket that houses the block index data.
		_, err = meta.CreateBucket(dbnamespace.BlockIndexBucketName)
		if err != nil {
			return err
		}

		// Create the bucket that houses the chain block hash to height
		// index.
		_, err = meta.CreateBucket(dbnamespace.HashIndexBucketName)
		if err != nil {
			return err
		}

		// Create the bucket that houses the chain block height to hash
		// index.
		_, err = meta.CreateBucket(dbnamespace.HeightIndexBucketName)
		if err != nil {
			return err
		}

		// Create the bucket that houses the spend journal data.
		_, err = meta.CreateBucket(dbnamespace.SpendJournalBucketName)
		if err != nil {
			return err
		}

		// Create the bucket that houses the utxo set.  Note that the
		// genesis block coinbase transaction is intentionally not
		// inserted here since it is not spendable by consensus rules.
		_, err = meta.CreateBucket(dbnamespace.UtxoSetBucketName)
		if err != nil {
			return err
		}

		// Add the genesis block to the block index.
		err = dbPutBlockNode(dbTx, node)
		if err != nil {
			return err
		}

		// Add the genesis block hash to height and height to hash
		// mappings to the index.
		err = dbPutMainChainIndex(dbTx, &node.hash, node.order)
		if err != nil {
			return err
		}

		// Store the current best chain state into the database.
		err = dbPutBestState(dbTx, b.stateSnapshot, node.workSum)
		if err != nil {
			return err
		}

		// Store the genesis block into the database.
		return dbTx.StoreBlock(genesisBlock)
	})
	return err
}


// dbPutDatabaseInfo uses an existing database transaction to store the database
// information.
func dbPutDatabaseInfo(dbTx database.Tx, dbi *databaseInfo) error {
	// uint32Bytes is a helper function to convert a uint32 to a byte slice
	// using the byte order specified by the database namespace.
	uint32Bytes := func(ui32 uint32) []byte {
		var b [4]byte
		dbnamespace.ByteOrder.PutUint32(b[:], ui32)
		return b[:]
	}

	// uint64Bytes is a helper function to convert a uint64 to a byte slice
	// using the byte order specified by the database namespace.
	uint64Bytes := func(ui64 uint64) []byte {
		var b [8]byte
		dbnamespace.ByteOrder.PutUint64(b[:], ui64)
		return b[:]
	}

	// Store the database version.
	meta := dbTx.Metadata()
	bucket := meta.Bucket(dbnamespace.BCDBInfoBucketName)
	err := bucket.Put(dbnamespace.BCDBInfoVersionKeyName,
		uint32Bytes(dbi.version))
	if err != nil {
		return err
	}

	// Store the compression version.
	err = bucket.Put(dbnamespace.BCDBInfoCompressionVersionKeyName,
		uint32Bytes(dbi.compVer))
	if err != nil {
		return err
	}

	// Store the block index version.
	err = bucket.Put(dbnamespace.BCDBInfoBlockIndexVersionKeyName,
		uint32Bytes(dbi.bidxVer))
	if err != nil {
		return err
	}

	// Store the database creation date.
	return bucket.Put(dbnamespace.BCDBInfoCreatedKeyName,
		uint64Bytes(uint64(dbi.created.Unix())))
}

// dbPutBlockNode stores the information needed to reconstruct the provided
// block node in the block index according to the format described above.
func dbPutBlockNode(dbTx database.Tx, node *blockNode) error {
	serialized, err := serializeBlockIndexEntry(&blockIndexEntry{
		header:         node.Header(),
		status:         node.status,
	})
	if err != nil {
		return err
	}

	bucket := dbTx.Metadata().Bucket(dbnamespace.BlockIndexBucketName)
	key := blockIndexKey(&node.hash, uint32(node.order))
	return bucket.Put(key, serialized)
}

// serializeBlockIndexEntry serializes the passed block index entry into a
// single byte slice according to the format described in detail above.
func serializeBlockIndexEntry(entry *blockIndexEntry) ([]byte, error) {
	serialized := make([]byte, blockIndexEntrySerializeSize(entry))
	_, err := putBlockIndexEntry(serialized, entry)
	return serialized, err
}

// blockIndexEntrySerializeSize returns the number of bytes it would take to
// serialize the passed block index entry according to the format described
// above.
func blockIndexEntrySerializeSize(entry *blockIndexEntry) int {

	return blockHdrSize + 1
}

// -----------------------------------------------------------------------------
// The main chain index consists of two buckets with an entry for every block in
// the main chain.  One bucket is for the hash to height mapping and the other
// is for the height to hash mapping.
//
// The serialized format for values in the hash to height bucket is:
//   <height>
//
//   Field      Type     Size
//   height     uint32   4 bytes
//
// The serialized format for values in the height to hash bucket is:
//   <hash>
//
//   Field      Type             Size
//   hash       chainhash.Hash   chainhash.HashSize
// -----------------------------------------------------------------------------

// dbPutMainChainIndex uses an existing database transaction to update or add
// index entries for the hash to height and height to hash mappings for the
// provided values.
func dbPutMainChainIndex(dbTx database.Tx, hash *hash.Hash, height uint64) error {
	// Serialize the height for use in the index entries.
	var serializedHeight [4]byte
	dbnamespace.ByteOrder.PutUint32(serializedHeight[:], uint32(height))

	// Add the block hash to height mapping to the index.
	meta := dbTx.Metadata()
	hashIndex := meta.Bucket(dbnamespace.HashIndexBucketName)
	if err := hashIndex.Put(hash[:], serializedHeight[:]); err != nil {
		return err
	}

	// Add the block height to hash mapping to the index.
	heightIndex := meta.Bucket(dbnamespace.HeightIndexBucketName)
	return heightIndex.Put(serializedHeight[:], hash[:])
}

// dbPutBestState uses an existing database transaction to update the best chain
// state with the given parameters.
func dbPutBestState(dbTx database.Tx, snapshot *BestState, workSum *big.Int) error {
	// Serialize the current best chain state.
	serializedData := serializeBestChainState(bestChainState{
		hash:         snapshot.Hash,
		height:       snapshot.Order,
	})

	// Store the current best chain state into the database.
	return dbTx.Metadata().Put(dbnamespace.ChainStateKeyName, serializedData)
}

// serializeBestChainState returns the serialization of the passed block best
// chain state.  This is data to be stored in the chain state bucket.
func serializeBestChainState(state bestChainState) []byte {
	// Calculate the full size needed to serialize the chain state.
	serializedLen := hash.HashSize + 4 + 8

	// Serialize the chain state.
	serializedData := make([]byte, serializedLen)
	copy(serializedData[0:hash.HashSize], state.hash[:])
	offset := uint32(hash.HashSize)
	dbnamespace.ByteOrder.PutUint64(serializedData[offset:], state.height)
	offset += 8

	return serializedData[:]
}


// putBlockIndexEntry serializes the passed block index entry according to the
// format described above directly into the passed target byte slice.  The
// target byte slice must be at least large enough to handle the number of bytes
// returned by the blockIndexEntrySerializeSize function or it will panic.
func putBlockIndexEntry(target []byte, entry *blockIndexEntry) (int, error) {

	// Serialize the entire block header.
	w := bytes.NewBuffer(target[0:0])
	if err := entry.header.Serialize(w); err != nil {
		return 0, err
	}

	// Serialize the status.
	offset := blockHdrSize
	target[offset] = byte(entry.status)
	offset++

	return offset, nil
}

// deserializeBestChainState deserializes the passed serialized best chain
// state.  This is data stored in the chain state bucket and is updated after
// every block is connected or disconnected form the main chain.
// block.
func deserializeBestChainState(serializedData []byte) (bestChainState, error) {
	// Ensure the serialized data has enough bytes to properly deserialize
	// the hash, height, total transactions, total subsidy, current subsidy,
	// and work sum length.
	expectedMinLen := hash.HashSize + 8
	if len(serializedData) < expectedMinLen {
		return bestChainState{}, database.Error{
			ErrorCode: database.ErrCorruption,
			Description: fmt.Sprintf("corrupt best chain state size; min %v "+
				"got %v", expectedMinLen, len(serializedData)),
		}
	}

	state := bestChainState{}
	copy(state.hash[:], serializedData[0:hash.HashSize])
	offset := uint32(hash.HashSize)
	state.height = dbnamespace.ByteOrder.Uint64(serializedData[offset : offset+8])
	offset += 8

	return state, nil
}

// deserializeBlockIndexEntry decodes the passed serialized byte slice into a
// block index entry according to the format described above.
func deserializeBlockIndexEntry(serialized []byte) (*blockIndexEntry, error) {
	var entry blockIndexEntry
	if _, err := decodeBlockIndexEntry(serialized, &entry); err != nil {
		return nil, err
	}
	return &entry, nil
}

// decodeBlockIndexEntry decodes the passed serialized block index entry into
// the passed struct according to the format described above.  It returns the
// number of bytes read.
func decodeBlockIndexEntry(serialized []byte, entry *blockIndexEntry) (int, error) {
	// Ensure there are enough bytes to decode header.
	if len(serialized) < blockHdrSize {
		return 0, errDeserialize("unexpected end of data while " +
			"reading block header")
	}
	hB := serialized[0:blockHdrSize]

	// Deserialize the header.
	var header types.BlockHeader
	if err := header.Deserialize(bytes.NewReader(hB)); err != nil {
		return 0, err
	}
	offset := blockHdrSize

	// Deserialize the status.
	if offset+1 > len(serialized) {
		return offset, errDeserialize("unexpected end of data while " +
			"reading status")
	}
	status := blockStatus(serialized[offset])
	offset++

	entry.header = header
	entry.status = status
	return offset, nil
}

// dbFetchBlockByHash uses an existing database transaction to retrieve the raw
// block for the provided hash, deserialize it, retrieve the appropriate height
// from the index, and return a dcrutil.Block with the height set.
func dbFetchBlockByHash(dbTx database.Tx, hash *hash.Hash) (*types.SerializedBlock, error) {
	// Load the raw block bytes from the database.
	blockBytes, err := dbTx.FetchBlock(hash)
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


// deserializeUtxoEntry decodes a utxo entry from the passed serialized byte
// slice into a new UtxoEntry using a format that is suitable for long-term
// storage.  The format is described in detail above.
func deserializeUtxoEntry(serialized []byte) (*UtxoEntry, error) {
	// Deserialize the version.
	version, bytesRead := deserializeVLQ(serialized)
	offset := bytesRead
	if offset >= len(serialized) {
		return nil, errDeserialize("unexpected end of data after version")
	}

	// Deserialize the block height.
	blockHeight, bytesRead := deserializeVLQ(serialized[offset:])
	offset += bytesRead
	if offset >= len(serialized) {
		return nil, errDeserialize("unexpected end of data after height")
	}

	// Deserialize the block index.
	blockIndex, bytesRead := deserializeVLQ(serialized[offset:])
	offset += bytesRead
	if offset >= len(serialized) {
		return nil, errDeserialize("unexpected end of data after index")
	}

	// Deserialize the flags.
	flags, bytesRead := deserializeVLQ(serialized[offset:])
	offset += bytesRead
	if offset >= len(serialized) {
		return nil, errDeserialize("unexpected end of data after flags")
	}
	isCoinBase, hasExpiry, txType, _ := decodeFlags(byte(flags))

	// Deserialize the header code.
	code, bytesRead := deserializeVLQ(serialized[offset:])
	offset += bytesRead
	if offset >= len(serialized) {
		return nil, errDeserialize("unexpected end of data after header")
	}

	// Decode the header code.
	//
	// Bit 0 indicates output 0 is unspent.
	// Bit 1 indicates output 1 is unspent.
	// Bits 2-x encodes the number of non-zero unspentness bitmap bytes that
	// follow.  When both output 0 and 1 are spent, it encodes N-1.
	output0Unspent := code&0x01 != 0
	output1Unspent := code&0x02 != 0
	numBitmapBytes := code >> 2
	if !output0Unspent && !output1Unspent {
		numBitmapBytes++
	}

	// Ensure there are enough bytes left to deserialize the unspentness
	// bitmap.
	if uint64(len(serialized[offset:])) < numBitmapBytes {
		return nil, errDeserialize("unexpected end of data for " +
			"unspentness bitmap")
	}

	// Create a new utxo entry with the details deserialized above to house
	// all of the utxos.
	//TODO, remove type conversion
	entry := newUtxoEntry(uint32(version), uint32(blockHeight),
		uint32(blockIndex), isCoinBase, hasExpiry, txType)

	// Add sparse output for unspent outputs 0 and 1 as needed based on the
	// details provided by the header code.
	var outputIndexes []uint32
	if output0Unspent {
		outputIndexes = append(outputIndexes, 0)
	}
	if output1Unspent {
		outputIndexes = append(outputIndexes, 1)
	}

	// Decode the unspentness bitmap adding a sparse output for each unspent
	// output.
	for i := uint32(0); i < uint32(numBitmapBytes); i++ {
		unspentBits := serialized[offset]
		for j := uint32(0); j < 8; j++ {
			if unspentBits&0x01 != 0 {
				// The first 2 outputs are encoded via the
				// header code, so adjust the output number
				// accordingly.
				outputNum := 2 + i*8 + j
				outputIndexes = append(outputIndexes, outputNum)
			}
			unspentBits >>= 1
		}
		offset++
	}

	// Decode and add all of the utxos.
	for i, outputIndex := range outputIndexes {
		// Decode the next utxo.  The script and amount fields of the
		// utxo output are left compressed so decompression can be
		// avoided on those that are not accessed.  This is done since
		// it is quite common for a redeeming transaction to only
		// reference a single utxo from a referenced transaction.
		//
		// 'true' below instructs the method to deserialize a stored
		// amount.
		// TODO script version is omit
		amount, _, compScript, bytesRead, err :=
			decodeCompressedTxOut(serialized[offset:], currentCompressionVersion,
				true)
		if err != nil {
			return nil, errDeserialize(fmt.Sprintf("unable to "+
				"decode utxo at index %d: %v", i, err))
		}
		offset += bytesRead

		entry.sparseOutputs[outputIndex] = &utxoOutput{
			spent:         false,
			compressed:    true,
			pkScript:      compScript,
			amount:        uint64(amount),  //TODO, remove type conversion
		}
	}

	return entry, nil
}

// BlockHeightByHash returns the height of the block with the given hash in the
// main chain.
//
// This function is safe for concurrent access.
func (b *BlockChain) BlockHeightByHash(hash *hash.Hash) (uint64, error) {
	var height uint64
	err := b.db.View(func(dbTx database.Tx) error {
		var err error
		height, err = dbFetchHeightByHash(dbTx, hash)
		return err
	})
	return height, err
}

// dbFetchHeightByHash uses an existing database transaction to retrieve the
// height for the provided hash from the index.
func dbFetchHeightByHash(dbTx database.Tx, hash *hash.Hash) (uint64, error) {
	meta := dbTx.Metadata()
	hashIndex := meta.Bucket(dbnamespace.HashIndexBucketName)
	serializedHeight := hashIndex.Get(hash[:])
	if serializedHeight == nil {
		str := fmt.Sprintf("block %s is not in the main chain", hash)
		return 0, errNotInMainChain(str)
	}

	return uint64(dbnamespace.ByteOrder.Uint32(serializedHeight)), nil
}


// dbFetchHeaderByHash uses an existing database transaction to retrieve the
// block header for the provided hash.
func dbFetchHeaderByHash(dbTx database.Tx, hash *hash.Hash) (*types.BlockHeader, error) {
	headerBytes, err := dbTx.FetchBlockHeader(hash)
	if err != nil {
		return nil, err
	}

	var header types.BlockHeader
	err = header.Deserialize(bytes.NewReader(headerBytes))
	if err != nil {
		return nil, err
	}

	return &header, nil
}

// dbFetchHeaderByHeight uses an existing database transaction to retrieve the
// block header for the provided height.
func dbFetchHeaderByHeight(dbTx database.Tx, height uint64) (*types.BlockHeader, error) {
	h, err := dbFetchHashByHeight(dbTx, height)
	if err != nil {
		return nil, err
	}

	return dbFetchHeaderByHash(dbTx, h)
}

// dbMaybeStoreBlock stores the provided block in the database if it's not
// already there.
func dbMaybeStoreBlock(dbTx database.Tx, block *types.SerializedBlock) error {
	// Store the block in ffldb if not already done.
	hasBlock, err := dbTx.HasBlock(block.Hash())
	if err != nil {
		return err
	}
	if hasBlock {
		return nil
	}

	return dbTx.StoreBlock(block)
}
// dbPutUtxoView uses an existing database transaction to update the utxo set
// in the database based on the provided utxo view contents and state.  In
// particular, only the entries that have been marked as modified are written
// to the database.
func dbPutUtxoView(dbTx database.Tx, view *UtxoViewpoint) error {
	utxoBucket := dbTx.Metadata().Bucket(dbnamespace.UtxoSetBucketName)
	for txHashIter, entry := range view.entries {
		// No need to update the database if the entry was not modified.
		if entry == nil || !entry.modified {
			continue
		}

		// Serialize the utxo entry without any entries that have been
		// spent.
		//TODO, move serializeUtxoEntry to serialization package
		serialized, err := serializeUtxoEntry(entry)
		if err != nil {
			return err
		}

		// Make a copy of the hash because the iterator changes on each
		// loop iteration and thus slicing it directly would cause the
		// data to change out from under the put/delete funcs below.
		txHash := txHashIter

		// Remove the utxo entry if it is now fully spent.
		if serialized == nil {
			if err := utxoBucket.Delete(txHash[:]); err != nil {
				return err
			}

			continue
		}

		// At this point the utxo entry is not fully spent, so store its
		// serialization in the database.
		err = utxoBucket.Put(txHash[:], serialized)
		if err != nil {
			return err
		}
	}

	return nil
}


//TODO, refactor & move serializeUtxoEntry to serialization package
// -----------------------------------------------------------------------------
// The unspent transaction output (utxo) set consists of an entry for each
// transaction which contains a utxo serialized using a format that is highly
// optimized to reduce space using domain specific compression algorithms.  This
// format is a slightly modified version of the format used in Bitcoin Core.
//
// The serialized format is:
//
//   <version><height><header code><unspentness bitmap>[<compressed txouts>,...]
//
//   Field                 Type     Size
//   transaction version   VLQ      variable
//   block height          VLQ      variable
//   block index           VLQ      variable
//   flags                 VLQ      variable (currently 1 byte)
//   header code           VLQ      variable
//   unspentness bitmap    []byte   variable
//   compressed txouts
//     compressed amount   VLQ      variable
//     compressed version  VLQ      variable
//     compressed script   []byte   variable
//
// The serialized flags code format is:
//   bit  0   - containing transaction is a coinbase
//   bit  1   - containing transaction has an expiry
//   bits 2-3 - transaction type
//   bits 4-7 - unused
//
// The serialized header code format is:
//   bit 0 - output zero is unspent
//   bit 1 - output one is unspent
//   bits 2-x - number of bytes in unspentness bitmap.  When both bits 1 and 2
//     are unset, it encodes N-1 since there must be at least one unspent
//     output.
//
// The rationale for the header code scheme is as follows:
//   - Transactions which only pay to a single output and a change output are
//     extremely common, thus an extra byte for the unspentness bitmap can be
//     avoided for them by encoding those two outputs in the low order bits.
//   - Given it is encoded as a VLQ which can encode values up to 127 with a
//     single byte, that leaves 4 bits to represent the number of bytes in the
//     unspentness bitmap while still only consuming a single byte for the
//     header code.  In other words, an unspentness bitmap with up to 120
//     transaction outputs can be encoded with a single-byte header code.
//     This covers the vast majority of transactions.
//   - Encoding N-1 bytes when both bits 0 and 1 are unset allows an additional
//     8 outpoints to be encoded before causing the header code to require an
//     additional byte.
//
// -----------------------------------------------------------------------------

// utxoEntryHeaderCode returns the calculated header code to be used when
// serializing the provided utxo entry and the number of bytes needed to encode
// the unspentness bitmap.
func utxoEntryHeaderCode(entry *UtxoEntry, highestOutputIndex uint32) (uint64, int, error) {
	// The first two outputs are encoded separately, so offset the index
	// accordingly to calculate the correct number of bytes needed to encode
	// up to the highest unspent output index.
	numBitmapBytes := int((highestOutputIndex + 6) / 8)

	// As previously described, one less than the number of bytes is encoded
	// when both output 0 and 1 are spent because there must be at least one
	// unspent output.  Adjust the number of bytes to encode accordingly and
	// encode the value by shifting it over 2 bits.
	output0Unspent := !entry.IsOutputSpent(0)
	output1Unspent := !entry.IsOutputSpent(1)
	var numBitmapBytesAdjustment int
	if !output0Unspent && !output1Unspent {
		if numBitmapBytes == 0 {
			return 0, 0, AssertError("attempt to serialize utxo " +
				"header for fully spent transaction")
		}
		numBitmapBytesAdjustment = 1
	}
	headerCode := uint64(numBitmapBytes-numBitmapBytesAdjustment) << 2

	// Set the output 0 and output 1 bits in the header code
	// accordingly.
	if output0Unspent {
		headerCode |= 0x01 // bit 0
	}
	if output1Unspent {
		headerCode |= 0x02 // bit 1
	}

	return headerCode, numBitmapBytes, nil
}

//TODO, move serializeUtxoEntry to serialization package
// serializeUtxoEntry returns the entry serialized to a format that is suitable
// for long-term storage.  The format is described in detail above.
func serializeUtxoEntry(entry *UtxoEntry) ([]byte, error) {
	// Fully spent entries have no serialization.
	if entry.IsFullySpent() {
		return nil, nil
	}

	// Determine the output order by sorting the sparse output index keys.
	outputOrder := make([]int, 0, len(entry.sparseOutputs))
	for outputIndex := range entry.sparseOutputs {
		outputOrder = append(outputOrder, int(outputIndex))
	}
	sort.Ints(outputOrder)

	// Encode the header code and determine the number of bytes the
	// unspentness bitmap needs.
	highIndex := uint32(outputOrder[len(outputOrder)-1])
	headerCode, numBitmapBytes, err := utxoEntryHeaderCode(entry, highIndex)
	if err != nil {
		return nil, err
	}

	// Calculate the size needed to serialize the entry.
	flags := encodeFlags(entry.isCoinBase, entry.hasExpiry, entry.txType, false)
	size := serializeSizeVLQ(uint64(entry.txVersion)) +
		serializeSizeVLQ(uint64(entry.height)) +
		serializeSizeVLQ(uint64(entry.index)) +
		serializeSizeVLQ(uint64(flags)) +
		serializeSizeVLQ(headerCode) + numBitmapBytes
	for _, outputIndex := range outputOrder {
		out := entry.sparseOutputs[uint32(outputIndex)]
		if out.spent {
			continue
		}
		size += compressedTxOutSize(uint64(out.amount), out.scriptVersion,
			out.pkScript, currentCompressionVersion, out.compressed, true)
	}

	// Serialize the version, block height, block index, and flags of the
	// containing transaction, and "header code" which is a complex bitmap
	// of spentness.
	serialized := make([]byte, size)
	offset := putVLQ(serialized, uint64(entry.txVersion))
	offset += putVLQ(serialized[offset:], uint64(entry.height))
	offset += putVLQ(serialized[offset:], uint64(entry.index))
	offset += putVLQ(serialized[offset:], uint64(flags))
	offset += putVLQ(serialized[offset:], headerCode)

	// Serialize the unspentness bitmap.
	for i := uint32(0); i < uint32(numBitmapBytes); i++ {
		unspentBits := byte(0)
		for j := uint32(0); j < 8; j++ {
			// The first 2 outputs are encoded via the header code,
			// so adjust the output index accordingly.
			if !entry.IsOutputSpent(2 + i*8 + j) {
				unspentBits |= 1 << uint8(j)
			}
		}
		serialized[offset] = unspentBits
		offset++
	}

	// Serialize the compressed unspent transaction outputs.  Outputs that
	// are already compressed are serialized without modifications.
	for _, outputIndex := range outputOrder {
		out := entry.sparseOutputs[uint32(outputIndex)]
		if out.spent {
			continue
		}

		offset += putCompressedTxOut(serialized[offset:],
			uint64(out.amount), out.scriptVersion, out.pkScript,
			currentCompressionVersion, out.compressed, true)
	}
	return serialized, nil
}

// dbRemoveMainChainIndex uses an existing database transaction remove main
// chain index entries from the hash to height and height to hash mappings for
// the provided values.
func dbRemoveMainChainIndex(dbTx database.Tx, hash *hash.Hash, height int64) error {
	// Remove the block hash to height mapping.
	meta := dbTx.Metadata()
	hashIndex := meta.Bucket(dbnamespace.HashIndexBucketName)
	if err := hashIndex.Delete(hash[:]); err != nil {
		return err
	}

	// Remove the block height to hash mapping.
	var serializedHeight [4]byte
	dbnamespace.ByteOrder.PutUint32(serializedHeight[:], uint32(height))
	heightIndex := meta.Bucket(dbnamespace.HeightIndexBucketName)
	return heightIndex.Delete(serializedHeight[:])
}


