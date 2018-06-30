package blockchain

import (
	"github.com/noxproject/nox/common/hash"
	"github.com/noxproject/nox/core/types"
	"github.com/noxproject/nox/engine/txscript"
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
