package blockchain

import (
	"encoding/binary"
	"fmt"
	"github.com/Qitmeer/qitmeer/common/hash"
	"github.com/Qitmeer/qitmeer/core/dbnamespace"
	"github.com/Qitmeer/qitmeer/core/serialization"
	"github.com/Qitmeer/qitmeer/core/types"
	"github.com/Qitmeer/qitmeer/database"
)

var byteOrder = binary.LittleEndian

// The bytes of TxIndex
const SpentTxOutTxIndexSize = 4

// The bytes of TxInIndex
const SpentTxOutTxInIndexSize = 4

// The bytes of Fees field
const SpentTxoutFeesCoinIDSize = 2
const SpentTxOutFeesValueSize = 8

// The bytes of Amount's CoinId
const SpentTxoutAmountCoinIDSize = 2

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
	Amount     types.Amount // The total amount of the output.
	PkScript   []byte       // The public key script for the output.
	BlockHash  hash.Hash
	IsCoinBase bool         // Whether creating tx is a coinbase.
	TxIndex    uint32       // The index of tx in block.
	TxInIndex  uint32       // The index of TxInput in the tx.
	Fees       types.Amount // The fees of block
}

func spentTxOutHeaderCode(stxo *SpentTxOut) uint64 {
	// As described in the serialization format comments, the header code
	// encodes the height shifted over one bit and the coinbase flag in the
	// lowest bit.
	headerCode := uint64(0)
	if stxo.IsCoinBase {
		headerCode |= 0x01
	}

	return headerCode
}

// SpentTxOutSerializeSize returns the number of bytes it would take to
// serialize the passed stxo according to the format described above.
// The amount is never encoded into spent transaction outputs because
// they're already encoded into the transactions, so skip them when
// determining the serialization size.
func spentTxOutSerializeSize(stxo *SpentTxOut) int {
	size := serialization.SerializeSizeVLQ(spentTxOutHeaderCode(stxo))
	size += hash.HashSize
	size += SpentTxOutTxIndexSize + SpentTxOutTxInIndexSize
	size += SpentTxoutFeesCoinIDSize + SpentTxOutFeesValueSize
	size += SpentTxoutAmountCoinIDSize
	return size + compressedTxOutSize(uint64(stxo.Amount.Value), stxo.PkScript)
}

// putSpentTxOut serializes the passed stxo according to the format described
// above directly into the passed target byte slice.  The target byte slice must
// be at least large enough to handle the number of bytes returned by the
// SpentTxOutSerializeSize function or it will panic.
func putSpentTxOut(target []byte, stxo *SpentTxOut) int {
	headerCode := spentTxOutHeaderCode(stxo)
	offset := serialization.PutVLQ(target, headerCode)
	copy(target[offset:], stxo.BlockHash.Bytes())
	offset += hash.HashSize

	byteOrder.PutUint32(target[offset:], uint32(stxo.TxIndex))
	offset += SpentTxOutTxIndexSize
	byteOrder.PutUint32(target[offset:], uint32(stxo.TxInIndex))
	offset += SpentTxOutTxInIndexSize
	// add Fees coinId
	byteOrder.PutUint16(target[offset:], uint16(stxo.Fees.Id))
	offset += SpentTxoutFeesCoinIDSize
	// add Fees Value
	byteOrder.PutUint64(target[offset:], uint64(stxo.Fees.Value))
	offset += SpentTxOutFeesValueSize
	// add Amount coinId
	byteOrder.PutUint16(target[offset:], uint16(stxo.Amount.Id))
	offset += SpentTxoutAmountCoinIDSize
	return offset + putCompressedTxOut(target[offset:], uint64(stxo.Amount.Value), stxo.PkScript)
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
	code, offset := serialization.DeserializeVLQ(serialized)
	if offset >= len(serialized) {
		return offset, errDeserialize("unexpected end of data after " +
			"header code")
	}

	// Decode the header code.
	//
	// Bit 0 indicates containing transaction is a coinbase.
	// Bits 1-x encode height of containing transaction.
	stxo.IsCoinBase = code&0x01 != 0

	stxo.BlockHash.SetBytes(serialized[offset : offset+hash.HashSize])
	offset += hash.HashSize

	stxo.TxIndex = byteOrder.Uint32(serialized[offset : offset+SpentTxOutTxIndexSize])
	offset += SpentTxOutTxIndexSize
	stxo.TxInIndex = byteOrder.Uint32(serialized[offset : offset+SpentTxOutTxInIndexSize])
	offset += SpentTxOutTxInIndexSize

	// Decode Fees
	feesCoinId := byteOrder.Uint16(serialized[offset : offset+SpentTxoutFeesCoinIDSize])
	offset += SpentTxoutFeesCoinIDSize
	feesValue := byteOrder.Uint64(serialized[offset : offset+SpentTxOutFeesValueSize])
	offset += SpentTxOutFeesValueSize
	stxo.Fees = types.Amount{Value: int64(feesValue), Id: types.CoinID(feesCoinId)}

	// Decode amount coinId
	amountCoinId := byteOrder.Uint16(serialized[offset : offset+SpentTxoutAmountCoinIDSize])
	offset += SpentTxoutAmountCoinIDSize

	// Decode the compressed txout.
	amount, pkScript, bytesRead, err := decodeCompressedTxOut(
		serialized[offset:])
	offset += bytesRead
	if err != nil {
		return offset, errDeserialize(fmt.Sprintf("unable to decode "+
			"txout: %v", err))
	}
	stxo.Amount = types.Amount{Value: int64(amount), Id: types.CoinID(amountCoinId)}
	stxo.PkScript = pkScript

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
	numStxos := int(byteOrder.Uint32(serialized[0:4]))
	// Loop backwards through all transactions so everything is read in
	// reverse order to match the serialization order.

	offset := 4
	stxos := make([]SpentTxOut, numStxos)
	for stxoIdx := numStxos - 1; stxoIdx > -1; stxoIdx-- {
		stxo := &stxos[stxoIdx]

		n, err := decodeSpentTxOut(serialized[offset:], stxo)
		offset += n
		if err != nil {
			return nil, errDeserialize(fmt.Sprintf("unable "+
				"to decode stxo for %v: %v",
				stxoIdx, err))
		}
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
	var size int = 4
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
	byteOrder.PutUint32(serialized[offset:], uint32(len(stxos)))
	offset += 4

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
	blockTxns := block.Block().Transactions[1:]

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
	if len(stxos) == 0 {
		return nil
	}
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

func GetStxo(txIndex uint32, txInIndex uint32, stxos []SpentTxOut) *SpentTxOut {
	for _, stxo := range stxos {
		if stxo.TxIndex == txIndex &&
			stxo.TxInIndex == txInIndex {
			return &stxo
		}
	}
	return nil
}
