// Copyright (c) 2017-2018 The nox developers
package blockchain

import (
	"github.com/noxproject/nox/common/hash"
	"github.com/noxproject/nox/core/types"
	"github.com/noxproject/nox/engine/txscript"
	"github.com/noxproject/nox/database"
	"fmt"
	"github.com/noxproject/nox/core/dbnamespace"
)


// UtxoEntry contains contextual information about an unspent transaction such
// as whether or not it is a coinbase transaction, which block it was found in,
// and the spent status of its outputs.
//
// The struct is aligned for memory efficiency.
type UtxoEntry struct {
	sparseOutputs map[uint32]*utxoOutput // Sparse map of unspent outputs.

	txType    types.TxType // The stake type of the transaction.
	height    uint32       // Height of block containing tx.
	index     uint32       // Index of containing tx in block.
	txVersion uint32       // The tx version of this tx.

	isCoinBase bool // Whether entry is a coinbase tx.
	hasExpiry  bool // Whether entry has an expiry.
	modified   bool // Entry changed since load.
}

// utxoOutput houses details about an individual unspent transaction output such
// as whether or not it is spent, its public key script, and how much it pays.
//
// Standard public key scripts are stored in the database using a compressed
// format. Since the vast majority of scripts are of the standard form, a fairly
// significant savings is achieved by discarding the portions of the standard
// scripts that can be reconstructed.
//
// Also, since it is common for only a specific output in a given utxo entry to
// be referenced from a redeeming transaction, the script and amount for a given
// output is not uncompressed until the first time it is accessed.  This
// provides a mechanism to avoid the overhead of needlessly uncompressing all
// outputs for a given utxo entry at the time of load.
//
// The struct is aligned for memory efficiency.
type utxoOutput struct {
	pkScript      []byte  // The public key script for the output.
	amount        uint64  // The amount of the output.
	compressed    bool    // The public key script is compressed.
	spent         bool    // Output is spent.
}

// UtxoViewpoint represents a view into the set of unspent transaction outputs
// from a specific point of view in the chain.  For example, it could be for
// the end of the main chain, some point in the history of the main chain, or
// down a side chain.
//
// The unspent outputs are needed by other transactions for things such as
// script validation and double spend prevention.
type UtxoViewpoint struct {
	entries   map[hash.Hash]*UtxoEntry
	bestHash  hash.Hash
}

// NewUtxoViewpoint returns a new empty unspent transaction output view.
func NewUtxoViewpoint() *UtxoViewpoint {
	return &UtxoViewpoint{
		entries: make(map[hash.Hash]*UtxoEntry),
	}
}

// AddTxOuts adds all outputs in the passed transaction which are not provably
// unspendable to the view.  When the view already has entries for any of the
// outputs, they are simply marked unspent.  All fields will be updated for
// existing entries since it's possible it has changed during a reorg.
func (view *UtxoViewpoint) AddTxOuts(theTx *types.Tx, blockHeight int64, blockIndex uint32) {
	tx := theTx.Transaction()
	// When there are not already any utxos associated with the transaction,
	// add a new entry for it to the view.
	entry := view.LookupEntry(theTx.Hash())
	if entry == nil {
		txType := types.DetermineTxType(tx)
		entry = newUtxoEntry(tx.Version, uint32(blockHeight),
			blockIndex, IsCoinBaseTx(tx), tx.Expire != 0, txType)
		view.entries[*theTx.Hash()] = entry
	} else {
		entry.height = uint32(blockHeight)
		entry.index = blockIndex
	}
	entry.modified = true

	// Loop all of the transaction outputs and add those which are not
	// provably unspendable.
	for txOutIdx, txOut := range theTx.Transaction().TxOut {
		// TODO allow pruning of stake utxs after all other outputs are spent
		if txscript.IsUnspendable(txOut.Amount, txOut.PkScript) {
			continue
		}

		// Update existing entries.  All fields are updated because it's
		// possible (although extremely unlikely) that the existing
		// entry is being replaced by a different transaction with the
		// same hash.  This is allowed so long as the previous
		// transaction is fully spent.
		if output, ok := entry.sparseOutputs[uint32(txOutIdx)]; ok {
			output.spent = false
			output.amount = txOut.Amount
			output.pkScript = txOut.PkScript
			output.compressed = false
			continue
		}

		// Add the unspent transaction output.
		entry.sparseOutputs[uint32(txOutIdx)] = &utxoOutput{
			spent:         false,
			amount:        txOut.Amount,
			pkScript:      txOut.PkScript,
			compressed:    false,
		}
	}
}

// FetchUtxoView loads utxo details about the input transactions referenced by
// the passed transaction from the point of view of the end of the main chain.
// It also attempts to fetch the utxo details for the transaction itself so the
// returned view can be examined for duplicate unspent transaction outputs.
//
// This function is safe for concurrent access however the returned view is NOT.
func (b *BlockChain) FetchUtxoView(tx *types.Tx, treeValid bool) (*UtxoViewpoint, error) {
	b.chainLock.RLock()
	defer b.chainLock.RUnlock()

	// The genesis block does not have any spendable transactions, so there
	// can't possibly be any details about it.  This is also necessary
	// because the code below requires the parent block and the genesis
	// block doesn't have one.
	tip := b.bestNode
	view := NewUtxoViewpoint()
	if tip.height == 0 {
		view.SetBestHash(&tip.hash)
		return view, nil
	}

	view.SetBestHash(&tip.hash)

	// Create a set of needed transactions based on those referenced by the
	// inputs of the passed transaction.  Also, add the passed transaction
	// itself as a way for the caller to detect duplicates that are not
	// fully spent.
	txNeededSet := make(map[hash.Hash]struct{})
	txNeededSet[*tx.Hash()] = struct{}{}

	err := view.fetchUtxosMain(b.db, txNeededSet)

	return view, err
}

// BestHash returns the hash of the best block in the chain the view currently
// respresents.
func (view *UtxoViewpoint) BestHash() *hash.Hash {
	return &view.bestHash
}

// SetBestHash sets the hash of the best block in the chain the view currently
// respresents.
func (view *UtxoViewpoint) SetBestHash(hash *hash.Hash) {
	view.bestHash = *hash
}

// fetchUtxosMain fetches unspent transaction output data about the provided
// set of transactions from the point of view of the end of the main chain at
// the time of the call.
//
// Upon completion of this function, the view will contain an entry for each
// requested transaction.  Fully spent transactions, or those which otherwise
// don't exist, will result in a nil entry in the view.
func (view *UtxoViewpoint) fetchUtxosMain(db database.DB, txSet map[hash.Hash]struct{}) error {
	// Nothing to do if there are no requested hashes.
	if len(txSet) == 0 {
		return nil
	}

	// Load the unspent transaction output information for the requested set
	// of transactions from the point of view of the end of the main chain.
	//
	// NOTE: Missing entries are not considered an error here and instead
	// will result in nil entries in the view.  This is intentionally done
	// since other code uses the presence of an entry in the store as a way
	// to optimize spend and unspend updates to apply only to the specific
	// utxos that the caller needs access to.
	return db.View(func(dbTx database.Tx) error {
		for hash := range txSet {
			hashCopy := hash
			// If the UTX already exists in the view, skip adding it.
			if _, ok := view.entries[hashCopy]; ok {
				continue
			}
			entry, err := dbFetchUtxoEntry(dbTx, &hashCopy)
			if err != nil {
				return err
			}

			view.entries[hash] = entry
		}

		return nil
	})
}

// dbFetchUtxoEntry uses an existing database transaction to fetch all unspent
// outputs for the provided Bitcoin transaction hash from the utxo set.
//
// When there is no entry for the provided hash, nil will be returned for the
// both the entry and the error.
func dbFetchUtxoEntry(dbTx database.Tx, hash *hash.Hash) (*UtxoEntry, error) {
	// Fetch the unspent transaction output information for the passed
	// transaction hash.  Return now when there is no entry.
	utxoBucket := dbTx.Metadata().Bucket(dbnamespace.UtxoSetBucketName)
	serializedUtxo := utxoBucket.Get(hash[:])
	if serializedUtxo == nil {
		return nil, nil
	}

	// A non-nil zero-length entry means there is an entry in the database
	// for a fully spent transaction which should never be the case.
	if len(serializedUtxo) == 0 {
		return nil, AssertError(fmt.Sprintf("database contains entry "+
			"for fully spent tx %v", hash))
	}

	// Deserialize the utxo entry and return it.
	entry, err := deserializeUtxoEntry(serializedUtxo)
	if err != nil {
		// Ensure any deserialization errors are returned as database
		// corruption errors.
		if isDeserializeErr(err) {
			return nil, database.Error{
				ErrorCode: database.ErrCorruption,
				Description: fmt.Sprintf("corrupt utxo entry "+
					"for %v: %v", hash, err),
			}
		}

		return nil, err
	}

	return entry, nil
}

// newUtxoEntry returns a new unspent transaction output entry with the provided
// coinbase flag and block height ready to have unspent outputs added.
func newUtxoEntry(txVersion uint32, height uint32, index uint32, isCoinBase bool, hasExpiry bool, tt types.TxType) *UtxoEntry {
	return &UtxoEntry{
		sparseOutputs: make(map[uint32]*utxoOutput),
		txVersion:     txVersion,
		height:        height,
		index:         index,
		isCoinBase:    isCoinBase,
		hasExpiry:     hasExpiry,
		txType:        tt,
	}
}

// LookupEntry returns information about a given transaction according to the
// current state of the view.  It will return nil if the passed transaction
// hash does not exist in the view or is otherwise not available such as when
// it has been disconnected during a reorg.
func (view *UtxoViewpoint) LookupEntry(txHash *hash.Hash) *UtxoEntry {
	entry, ok := view.entries[*txHash]
	if !ok {
		return nil
	}

	return entry
}
// PkScriptByIndex returns the public key script for the provided output index.
//
// Returns nil if the output index references an output that does not exist
// either due to it being invalid or because the output is not part of the view
// due to previously being spent/pruned.
func (entry *UtxoEntry) PkScriptByIndex(outputIndex uint32) []byte {
	output, ok := entry.sparseOutputs[outputIndex]
	if !ok {
		return nil
	}

	// Ensure the output is decompressed before returning the script.
	output.maybeDecompress(currentCompressionVersion)
	return output.pkScript
}

// maybeDecompress decompresses the amount and public key script fields of the
// utxo and marks it decompressed if needed.
func (o *utxoOutput) maybeDecompress(compressionVersion uint32) {
	// Nothing to do if it's not compressed.
	if !o.compressed {
		return
	}

	//TODO: impl compressed/decompressScript
	// o.pkScript = decompressScript(o.pkScript, compressionVersion)
	o.compressed = false
}

// TransactionType returns the transaction type of the transaction the utxo entry
// represents.
func (entry *UtxoEntry) TransactionType() types.TxType {
	return entry.txType
}
