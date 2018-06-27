package miner

import (
	"time"
	"github.com/noxproject/nox/core/types"
	"github.com/noxproject/nox/common/hash"
)

// TxSource represents a source of transactions to consider for inclusion in
// new blocks.
//
// The interface contract requires that all of these methods are safe for
// concurrent access with respect to the source.
type TxSource interface {
	// LastUpdated returns the last time a transaction was added to or
	// removed from the source pool.
	LastUpdated() time.Time

	// MiningDescs returns a slice of mining descriptors for all the
	// transactions in the source pool.
	MiningDescs() []*types.TxDesc

	// HaveTransaction returns whether or not the passed transaction hash
	// exists in the source pool.
	HaveTransaction(hash *hash.Hash) bool

	// HaveAllTransactions returns whether or not all of the passed
	// transaction hashes exist in the source pool.
	HaveAllTransactions(hashes []hash.Hash) bool

	// IsTxTreeKnownInvalid returns whether or not the transaction tree of
	// the provided hash is known to be invalid according to the votes
	// currently in the memory pool.
	IsTxTreeKnownInvalid(hash *hash.Hash) bool
}
