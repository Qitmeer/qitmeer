package blockchain

import (
	"fmt"
	"github.com/HalalChain/qitmeer/core/dbnamespace"
	"github.com/HalalChain/qitmeer-lib/core/types"
	"github.com/HalalChain/qitmeer/database"
	"github.com/HalalChain/qitmeer-lib/common/hash"
)

// -----------------------------------------------------------------------------
// The transaction spend journal consists of an entry for each block connected
// to the main chain which contains the transaction outputs the block spends
// serialized such that the order is the reverse of the order they were spent.
//
// This is required because reorganizing the chain necessarily entails
// disconnecting blocks to get back to the point of the fork which implies
// unspending all of the transaction outputs that each block previously spent.
// Since the utxo set, by definition, only contains unspent transaction outputs,
// the spent transaction outputs must be resurrected from somewhere.  There is
// more than one way this could be done, however this is the most straight
// forward method that does not require having a transaction index and unpruned
// blockchain.
//
// NOTE: This format is NOT self describing.  The additional details such as
// the number of entries (transaction inputs) are expected to come from the
// block itself and the utxo set.  The rationale in doing this is to save a
// significant amount of space.  This is also the reason the spent outputs are
// serialized in the reverse order they are spent because later transactions
// are allowed to spend outputs from earlier ones in the same block.
//
// The serialized format is:
//
//   [<flags><script version><compressed pk script>],...
//   OPTIONAL: [<txVersion>]
//
//   Field                Type           Size
//   flags                VLQ            byte
//   scriptVersion        uint16         2 bytes
//   pkScript             VLQ+[]byte     variable
//
//   OPTIONAL
//     txVersion          VLQ            variable
//     stakeExtra         []byte         variable
//
// The serialized flags code format is:
//   bit  0   - containing transaction is a coinbase
//   bit  1   - containing transaction has an expiry
//   bits 2-3 - transaction type
//   bit  4   - is fully spent
//
// The stake extra field contains minimally encoded outputs for all
// consensus-related outputs in the stake transaction. It is only
// encoded for tickets.
//
//   NOTE: The transaction version and flags are only encoded when the spent
//   txout was the final unspent output of the containing transaction.
//   Otherwise, the header code will be 0 and the version is not serialized at
//   all. This is  done because that information is only needed when the utxo
//   set no longer has it.
//
// Example:
//   TODO
// -----------------------------------------------------------------------------

// SpentTxOut contains a spent transaction output and potentially additional
// contextual information such as whether or not it was contained in a coinbase
// transaction, the txVersion of the transaction it was contained in, and which
// block height the containing transaction was included in.  As described in
// the comments above, the additional contextual information will only be valid
// when this spent txout is spending the last unspent output of the containing
// transaction.
//
// The struct is aligned for memory efficiency.
type SpentTxOut struct {
	Amount     uint64       // The amount of the output.
	PkScript   []byte // The public key script for the output.
	BlockHash  hash.Hash
	IsCoinBase   bool // Whether creating tx is a coinbase.

	stakeExtra []byte // Extra information for the staking system.
	txType        types.TxType // The tx type of the transaction.
	order         uint32       // order of the the block containing the tx.
	txIndex       uint32     // txIndex in the block of the transaction.
	inIndex       uint32       // Index in the txIn
	txVersion     uint32       // The version of creating tx.
	txFullySpent bool // Whether or not the transaction is fully spent.

	hasExpiry    bool // The expiry of the creating tx.
}

// SpentTxOutSerializeSize returns the number of bytes it would take to
// serialize the passed stxo according to the format described above.
// The amount is never encoded into spent transaction outputs because
// they're already encoded into the transactions, so skip them when
// determining the serialization size.
func spentTxOutSerializeSize(stxo *SpentTxOut) int {
	flags := encodeFlags(stxo.IsCoinBase, stxo.hasExpiry, stxo.txType,
		stxo.txFullySpent)
	size := serializeSizeVLQ(uint64(flags))

	// false below indicates that the txOut does not specify an amount.
	size += compressedTxOutSize(uint64(stxo.Amount), stxo.PkScript)

	// The transaction was fully spent, so we need to store some extra
	// data for UTX resurrection.
	if stxo.txFullySpent {
		size += serializeSizeVLQ(uint64(stxo.txVersion))
	}
	size+=8
	return size
}

// putSpentTxOut serializes the passed stxo according to the format described
// above directly into the passed target byte slice.  The target byte slice must
// be at least large enough to handle the number of bytes returned by the
// SpentTxOutSerializeSize function or it will panic.
func putSpentTxOut(target []byte, stxo *SpentTxOut) int {
	flags := encodeFlags(stxo.IsCoinBase, stxo.hasExpiry, stxo.txType,
		stxo.txFullySpent)
	offset := putVLQ(target, uint64(flags))

	// false below indicates that the txOut does not specify an amount.
	offset += putCompressedTxOut(target[offset:], stxo.Amount,stxo.PkScript)

	// The transaction was fully spent, so we need to store some extra
	// data for UTX resurrection.
	if stxo.txFullySpent {
		offset += putVLQ(target[offset:], uint64(stxo.txVersion))
	}
	serializedIndex:=[]byte{0,0,0,0}
	dbnamespace.ByteOrder.PutUint32(serializedIndex[:], uint32(stxo.txIndex))
	target[offset]=serializedIndex[0]
	target[offset+1]=serializedIndex[1]
	target[offset+2]=serializedIndex[2]
	target[offset+3]=serializedIndex[3]
	offset+=4

	serializedIndex=[]byte{0,0,0,0}
	dbnamespace.ByteOrder.PutUint32(serializedIndex[:], uint32(stxo.inIndex))
	target[offset]=serializedIndex[0]
	target[offset+1]=serializedIndex[1]
	target[offset+2]=serializedIndex[2]
	target[offset+3]=serializedIndex[3]
	offset+=4

	return offset
}

// decodeSpentTxOut decodes the passed serialized stxo entry, possibly followed
// by other data, into the passed stxo struct.  It returns the number of bytes
// read.
//
// Since the serialized stxo entry does not contain the height, version, or
// coinbase flag of the containing transaction when it still has utxos, the
// caller is responsible for passing in the containing transaction version in
// that case.  The provided version is ignore when it is serialized as a part of
// the stxo.
//
// An error will be returned if the version is not serialized as a part of the
// stxo and is also not provided to the function.
func decodeSpentTxOut(serialized []byte, stxo *SpentTxOut) (int, error) {
	// Ensure there are bytes to decode.
	if len(serialized) == 0 {
		return 0, errDeserialize("no serialized bytes")
	}

	// Deserialize the header code.
	flags, offset := deserializeVLQ(serialized)
	if offset >= len(serialized) {
		return offset, errDeserialize("unexpected end of data after " +
			"spent tx out flags")
	}

	// Decode the flags. If the flags are non-zero, it means that the
	// transaction was fully spent at this spend.
	isCoinBase, hasExpiry, txType, txFullySpent := decodeFlags(byte(flags))
	stxo.IsCoinBase = isCoinBase
	stxo.hasExpiry = hasExpiry
	stxo.txType = txType
	stxo.txFullySpent = txFullySpent

	// Decode the compressed txout. We pass false for the amount flag,
	// since we only need pkScript at most due to fraud proofs already
	// storing the decompressed amount.
	amount, script, bytesRead, err := decodeCompressedTxOut(serialized[offset:])
	offset += bytesRead
	if err != nil {
		return offset, errDeserialize(fmt.Sprintf("unable to decode "+
			"txout: %v", err))
	}
	stxo.Amount = uint64(amount)
	stxo.PkScript = script

	// Deserialize the containing transaction if the flags indicate that
	// the transaction has been fully spent.
	if txFullySpent {
		txVersion, bytesRead := deserializeVLQ(serialized[offset:])
		offset += bytesRead
		if offset == 0 || offset > len(serialized) {
			return offset, errDeserialize("unexpected end of data " +
				"after version")
		}

		stxo.txVersion = uint32(txVersion)

	}

	stxo.txIndex=dbnamespace.ByteOrder.Uint32(serialized[offset:offset+4])
	offset+=4
	stxo.inIndex=dbnamespace.ByteOrder.Uint32(serialized[offset:offset+4])
	offset+=4

	return offset, nil
}

// deserializeSpendJournalEntry decodes the passed serialized byte slice into a
// slice of spent txouts according to the format described in detail above.
//
// Since the serialization format is not self describing, as noted in the
// format comments, this function also requires the transactions that spend the
// txouts and a utxo view that contains any remaining existing utxos in the
// transactions referenced by the inputs to the passed transasctions.
func deserializeSpendJournalEntry(serialized []byte, txns []*types.Transaction) ([]SpentTxOut, error) {
	// When a block has no spent txouts there is nothing to serialize.
	if len(serialized) == 0 {
		return nil, nil
	}

	// Loop backwards through all transactions so everything is read in
	// reverse order to match the serialization order.
	stxos := []SpentTxOut{}

	for offset:=0;offset<len(serialized); {
		stxo := SpentTxOut{}
		n, err := decodeSpentTxOut(serialized[offset:], &stxo)
		offset += n

		if n==0 || err != nil {
			return nil, errDeserialize(fmt.Sprintf("unable "+
				"to decode stxo for %v: %v",
				offset, err))
		}
		//
		tx := txns[stxo.txIndex]
		txIn := tx.TxIn[stxo.inIndex]

		stxo.order=txIn.BlockOrder
		//stxo.amount=txIn.AmountIn
		//
		//indexStr:=fmt.Sprintf("%d-%d",stxo.index,stxo.inIndex)
		stxos=append(stxos,stxo)
	}

	return stxos, nil
}

// serializeSpendJournalEntry serializes all of the passed spent txouts into a
// single byte slice according to the format described in detail above.
func serializeSpendJournalEntry(stxos []SpentTxOut) ([]byte, error) {
	if len(stxos) == 0 {
		return nil, nil
	}

	// Calculate the size needed to serialize the entire journal entry.
	var size int
	var sizes []int
	for i := range stxos {
		sz := spentTxOutSerializeSize(&stxos[i])
		sizes = append(sizes, sz)
		size += sz
	}
	serialized := make([]byte, size)

	// Serialize each individual stxo directly into the slice in reverse
	// order one after the other.
	var offset int
	for i := len(stxos) - 1; i > -1; i-- {
		oldOffset := offset
		offset += putSpentTxOut(serialized[offset:], &stxos[i])

		if offset-oldOffset != sizes[i] {
			return nil, AssertError(fmt.Sprintf("bad write; expect sz %v, "+
				"got sz %v (wrote %x)", sizes[i], offset-oldOffset,
				serialized[oldOffset:offset]))
		}
	}

	return serialized, nil
}

// dbFetchSpendJournalEntry fetches the spend journal entry for the passed
// block and deserializes it into a slice of spent txout entries.  The provided
// view MUST have the utxos referenced by all of the transactions available for
// the passed block since that information is required to reconstruct the spent
// txouts.
func dbFetchSpendJournalEntry(dbTx database.Tx, block *types.SerializedBlock) ([]SpentTxOut, error) {
	// Exclude the coinbase transaction since it can't spend anything.
	spendBucket := dbTx.Metadata().Bucket(dbnamespace.SpendJournalBucketName)
	serialized := spendBucket.Get(block.Hash()[:])
	blockTxns := block.Block().Transactions[:]

	stxos, err := deserializeSpendJournalEntry(serialized, blockTxns)
	if err != nil {
		// Ensure any deserialization errors are returned as database
		// corruption errors.
		if isDeserializeErr(err) {
			return nil, database.Error{
				ErrorCode: database.ErrCorruption,
				Description: fmt.Sprintf("corrupt spend "+
					"information for %v: %v", block.Hash(),
					err),
			}
		}

		return nil, err
	}

	return stxos, nil
}

// dbPutSpendJournalEntry uses an existing database transaction to update the
// spend journal entry for the given block hash using the provided slice of
// spent txouts.   The spent txouts slice must contain an entry for every txout
// the transactions in the block spend in the order they are spent.
func dbPutSpendJournalEntry(dbTx database.Tx, blockHash *hash.Hash, stxos []SpentTxOut) error {
	spendBucket := dbTx.Metadata().Bucket(dbnamespace.SpendJournalBucketName)
	serialized, err := serializeSpendJournalEntry(stxos)
	if err != nil {
		return err
	}
	return spendBucket.Put(blockHash[:], serialized)
}

// dbRemoveSpendJournalEntry uses an existing database transaction to remove the
// spend journal entry for the passed block hash.
func dbRemoveSpendJournalEntry(dbTx database.Tx, blockHash *hash.Hash) error {
	spendBucket := dbTx.Metadata().Bucket(dbnamespace.SpendJournalBucketName)
	return spendBucket.Delete(blockHash[:])
}
