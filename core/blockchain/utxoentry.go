package blockchain

import (
	"github.com/HalalChain/qitmeer-lib/core/types"
)

// UtxoEntry contains contextual information about an unspent transaction such
// as whether or not it is a coinbase transaction, which block it was found in,
// and the spent status of its outputs.
//
// The struct is aligned for memory efficiency.
type UtxoEntry struct {
	sparseOutputs map[uint32]*utxoOutput // Sparse map of unspent outputs.

	txType    types.TxType // The stake type of the transaction.
	order    uint32       // Order of block containing tx.
	index     uint32       // Index of containing tx in block.
	txVersion uint32       // The tx version of this tx.

	isCoinBase bool // Whether entry is a coinbase tx.
	hasExpiry  bool // Whether entry has an expiry.
	modified   bool // Entry changed since load.
}

// TxVersion returns the transaction version of the transaction the
// utxo represents.
func (entry *UtxoEntry) TxVersion() uint32 {
	return entry.txVersion
}

// HasExpiry returns the transaction expiry for the transaction that the utxo
// entry represents.
func (entry *UtxoEntry) HasExpiry() bool {
	return entry.hasExpiry
}

// IsCoinBase returns whether or not the transaction the utxo entry represents
// is a coinbase.
func (entry *UtxoEntry) IsCoinBase() bool {
	return entry.isCoinBase
}


// TxIndex returns the transaction index of the block containing the transaction the
// utxo entry represents.
func (entry *UtxoEntry) TxIndex() uint32 {
	return entry.index
}


// SpendOutput marks the output at the provided index as spent.  Specifying an
// output index that does not exist will not have any effect.
func (entry *UtxoEntry) SpendOutput(outputIndex uint32) {
	output, ok := entry.sparseOutputs[outputIndex]
	if !ok {
		return
	}

	// Nothing to do if the output is already spent.
	if output.spent {
		return
	}

	entry.modified = true
	output.spent = true
}


// TransactionType returns the transaction type of the transaction the utxo entry
// represents.
func (entry *UtxoEntry) TransactionType() types.TxType {
	return entry.txType
}

// IsOutputSpent returns whether or not the provided output index has been
// spent based upon the current state of the unspent transaction output view
// the entry was obtained from.
//
// Returns true if the output index references an output that does not exist
// either due to it being invalid or because the output is not part of the view
// due to previously being spent/pruned.
func (entry *UtxoEntry) IsOutputSpent(outputIndex uint32) bool {
	output, ok := entry.sparseOutputs[outputIndex]
	if !ok {
		return true
	}

	return output.spent
}

// BlockOrder returns the order of the block containing the transaction the
// utxo entry represents.
func (entry *UtxoEntry) BlockOrder() uint64 {
	return uint64(entry.order) //TODO, remove type conversion
}

// AmountByIndex returns the amount of the provided output index.
//
// Returns 0 if the output index references an output that does not exist
// either due to it being invalid or because the output is not part of the view
// due to previously being spent/pruned.
func (entry *UtxoEntry) AmountByIndex(outputIndex uint32) uint64 {
	output, ok := entry.sparseOutputs[outputIndex]
	if !ok {
		return 0
	}

	return output.amount
}


// ScriptVersionByIndex returns the public key script for the provided output
// index.
//
// Returns 0 if the output index references an output that does not exist
// either due to it being invalid or because the output is not part of the view
// due to previously being spent/pruned.
func (entry *UtxoEntry) ScriptVersionByIndex(outputIndex uint32) uint16 {
	output, ok := entry.sparseOutputs[outputIndex]
	if !ok {
		return 0
	}

	return output.scriptVersion
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

// IsFullySpent returns whether or not the transaction the utxo entry represents
// is fully spent.
func (entry *UtxoEntry) IsFullySpent() bool {
	// The entry is not fully spent if any of the outputs are unspent.
	for _, output := range entry.sparseOutputs {
		if !output.spent {
			return false
		}
	}

	return true
}

