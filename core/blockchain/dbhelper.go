// Copyright (c) 2017-2018 The qitmeer developers
package blockchain

import (
	"bytes"
	"fmt"
	"github.com/Qitmeer/qitmeer/common/hash"
	"github.com/Qitmeer/qitmeer/common/roughtime"
	"github.com/Qitmeer/qitmeer/core/blockchain/token"
	"github.com/Qitmeer/qitmeer/core/blockdag"
	"github.com/Qitmeer/qitmeer/core/dbnamespace"
	"github.com/Qitmeer/qitmeer/core/serialization"
	"github.com/Qitmeer/qitmeer/core/types"
	"github.com/Qitmeer/qitmeer/core/types/pow"
	"github.com/Qitmeer/qitmeer/database"
	"math/big"
	"time"
)

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

// -----------------------------------------------------------------------------
// The best chain state consists of the best block hash and order, the total
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
//   block order       uint32           4 bytes
//   total txns        uint64           8 bytes
//   total subsidy     int64            8 bytes
//   tokenTipHash      chainhash.Hash   chainhash.HashSize
//   work sum length   uint32           4 bytes
//   work sum          big.Int          work sum length
// -----------------------------------------------------------------------------

// bestChainState represents the data to be stored the database for the current
// best chain state.
type bestChainState struct {
	hash         hash.Hash
	total        uint64
	totalTxns    uint64
	tokenTipHash hash.Hash
	workSum      *big.Int
}

func (bcs *bestChainState) GetTotal() uint64 {
	return bcs.total
}

// dbFetchBlockByOrder uses an existing database transaction to retrieve the
// raw block for the provided order, deserialize it, and return a Block
// with the height set.
func (b *BlockChain) DBFetchBlockByOrder(dbTx database.Tx, order uint64) (*types.SerializedBlock, error) {
	// First find the hash associated with the provided order in the index.
	h := b.bd.GetBlockByOrderWithTx(dbTx, uint(order))
	if h == nil {
		return nil, fmt.Errorf("No block\n")
	}

	// Load the raw block bytes from the database.
	blockBytes, err := dbTx.FetchBlock(h)
	if err != nil {
		return nil, err
	}

	// Create the encapsulated block and set the order appropriately.
	block, err := types.NewBlockFromBytes(blockBytes)
	if err != nil {
		return nil, err
	}
	block.SetOrder(order)
	return block, nil
}

// BlockByHeight returns the block at the given height in the main chain.
//
// This function is safe for concurrent access.
func (b *BlockChain) BlockByOrder(blockOrder uint64) (*types.SerializedBlock, error) {
	var block *types.SerializedBlock
	err := b.db.View(func(dbTx database.Tx) error {
		var err error
		block, err = b.DBFetchBlockByOrder(dbTx, blockOrder)
		return err
	})
	return block, err
}

// BlockHashByOrder returns the hash of the block at the given order in the
// main chain.
//
// This function is safe for concurrent access.
func (b *BlockChain) BlockHashByOrder(blockOrder uint64) (*hash.Hash, error) {
	hash := b.bd.GetBlockHashByOrder(uint(blockOrder))
	if hash == nil {
		return nil, fmt.Errorf("Can't find block")
	}
	return hash, nil
}

// MainChainHasBlock returns whether or not the block with the given hash is in
// the main chain.
//
// This function is safe for concurrent access.
func (b *BlockChain) MainChainHasBlock(hash *hash.Hash) bool {
	return b.bd.IsOnMainChain(b.bd.GetBlockId(hash))
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
	node := NewBlockNode(genesisBlock, genesisBlock.Block().Parents)
	_, _, ib, _ := b.bd.AddBlock(node)
	//node.FlushToDB(b)
	// Initialize the state related to the best block.  Since it is the
	// genesis block, use its timestamp for the median time.
	numTxns := uint64(len(genesisBlock.Block().Transactions))
	blockSize := uint64(genesisBlock.Block().SerializeSize())
	b.stateSnapshot = newBestState(node.GetHash(), node.Difficulty(), blockSize, numTxns,
		time.Unix(node.GetTimestamp(), 0), numTxns, 0, b.bd.GetGraphState(), node.GetHash())
	b.TokenTipID = 0
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
			compVer: serialization.CurrentCompressionVersion,
			bidxVer: currentBlockIndexVersion,
			created: roughtime.Now(),
		}
		err = dbPutDatabaseInfo(dbTx, b.dbInfo)
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

		// Create the bucket which house the token state
		if _, err := meta.CreateBucket(dbnamespace.TokenBucketName); err != nil {
			return err
		}
		initTS := token.BuildGenesisTokenState()
		err = initTS.Commit()
		if err != nil {
			return err
		}
		err = token.DBPutTokenState(dbTx, uint32(ib.GetID()), initTS)
		if err != nil {
			return err
		}
		// Store the current best chain state into the database.
		err = dbPutBestState(dbTx, b.stateSnapshot, pow.CalcWork(header.Difficulty, header.Pow.GetPowType()))
		if err != nil {
			return err
		}

		// Add genesis utxo
		view := NewUtxoViewpoint()
		view.SetViewpoints([]*hash.Hash{genesisBlock.Hash()})
		for _, tx := range genesisBlock.Transactions() {
			view.AddTxOuts(tx, genesisBlock.Hash())
		}
		err = dbPutUtxoView(dbTx, view)
		if err != nil {
			return err
		}

		// Store the genesis block into the database.
		if err := dbTx.StoreBlock(genesisBlock); err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		return err
	}
	return b.bd.Commit()
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

// dbPutBestState uses an existing database transaction to update the best chain
// state with the given parameters.
func dbPutBestState(dbTx database.Tx, snapshot *BestState, workSum *big.Int) error {
	// Serialize the current best chain state.
	tth := hash.ZeroHash
	if snapshot.TokenTipHash != nil {
		tth = *snapshot.TokenTipHash
	}
	serializedData := serializeBestChainState(bestChainState{
		hash:         snapshot.Hash,
		total:        uint64(snapshot.GraphState.GetTotal()),
		totalTxns:    snapshot.TotalTxns,
		workSum:      workSum,
		tokenTipHash: tth,
	})

	// Store the current best chain state into the database.
	return dbTx.Metadata().Put(dbnamespace.ChainStateKeyName, serializedData)
}

// serializeBestChainState returns the serialization of the passed block best
// chain state.  This is data to be stored in the chain state bucket.
func serializeBestChainState(state bestChainState) []byte {
	// Calculate the full size needed to serialize the chain state.
	workSumBytes := state.workSum.Bytes()
	workSumBytesLen := uint32(len(workSumBytes))
	serializedLen := hash.HashSize + 8 + 8 + hash.HashSize + 4 + workSumBytesLen

	// Serialize the chain state.
	serializedData := make([]byte, serializedLen)
	copy(serializedData[0:hash.HashSize], state.hash[:])
	offset := uint32(hash.HashSize)
	dbnamespace.ByteOrder.PutUint64(serializedData[offset:], state.total)
	offset += 8
	dbnamespace.ByteOrder.PutUint64(serializedData[offset:], state.totalTxns)
	offset += 8
	copy(serializedData[offset:offset+hash.HashSize], state.tokenTipHash[:])
	offset += hash.HashSize
	dbnamespace.ByteOrder.PutUint32(serializedData[offset:], workSumBytesLen)
	offset += 4
	copy(serializedData[offset:], workSumBytes)
	return serializedData[:]
}

// deserializeBestChainState deserializes the passed serialized best chain
// state.  This is data stored in the chain state bucket and is updated after
// every block is connected or disconnected form the main chain.
// block.
func DeserializeBestChainState(serializedData []byte) (bestChainState, error) {
	// Ensure the serialized data has enough bytes to properly deserialize
	// the hash, total, total transactions, total subsidy, current subsidy,
	// and work sum length.
	expectedMinLen := hash.HashSize + 8 + 8 + hash.HashSize + 4
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
	state.total = dbnamespace.ByteOrder.Uint64(serializedData[offset : offset+8])
	offset += 8
	state.totalTxns = dbnamespace.ByteOrder.Uint64(serializedData[offset : offset+8])
	offset += 8
	copy(state.tokenTipHash[:], serializedData[offset:offset+hash.HashSize])
	offset += hash.HashSize
	workSumBytesLen := dbnamespace.ByteOrder.Uint32(serializedData[offset : offset+4])
	offset += 4
	// Ensure the serialized data has enough bytes to deserialize the work
	// sum.
	if uint32(len(serializedData[offset:])) < workSumBytesLen {
		return bestChainState{}, database.Error{
			ErrorCode:   database.ErrCorruption,
			Description: "corrupt best chain state",
		}
	}
	workSumBytes := serializedData[offset : offset+workSumBytesLen]
	state.workSum = new(big.Int).SetBytes(workSumBytes)
	return state, nil
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

// BlockOrderByHash returns the order of the block with the given hash in the
// chain.
//
// This function is safe for concurrent access.
func (b *BlockChain) BlockOrderByHash(hash *hash.Hash) (uint64, error) {
	ib := b.bd.GetBlock(hash)
	if ib == nil {
		return uint64(blockdag.MaxBlockOrder), fmt.Errorf("No block\n")
	}
	return uint64(ib.GetOrder()), nil
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
