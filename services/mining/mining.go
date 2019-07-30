// Copyright (c) 2017-2018 The qitmeer developers
// Copyright (c) 2014-2016 The btcsuite developers
// Copyright (c) 2016-2018 The Decred developers
// Use of this source code is governed by an ISC
// license that can be found in the LICENSE file.

package mining

import (
	"github.com/HalalChain/qitmeer-lib/common/hash"
	"github.com/HalalChain/qitmeer/core/blockchain"
	s "github.com/HalalChain/qitmeer-lib/core/serialization"
	"github.com/HalalChain/qitmeer-lib/core/types"
	"time"
)

const (

	// kilobyte is the size of a kilobyte.
	// TODO, refactor the location of kilobyte const
	kilobyte = 1000

	// blockHeaderOverhead is the max number of bytes it takes to serialize
	// a block header and max possible transaction count.
	// TODO, refactor the location of blockHeaderOverhead const
	blockHeaderOverhead = types.MaxBlockHeaderPayload + s.MaxVarIntPayload

	// coinbaseFlags is some extra data appended to the coinbase script
	// sig.
	// TODO, refactor the location of coinbaseFlags const
	coinbaseFlags = "/qitmeer/"

	// generatedBlockVersion is the version of the block being generated for
	// the main network.  It is defined as a constant here rather than using
	// the wire.BlockVersion constant since a change in the block version
	// will require changes to the generated block.  Using the wire constant
	// for generated block version could allow creation of invalid blocks
	// for the updated version.
	// TODO, refactor the location of generatedBlockVersion const
	generatedBlockVersion = 0

	// generatedBlockVersionTest is the version of the block being generated
	// for networks other than the main network.
	// TODO, refactor the location of generatedBlockVersionTest const
	generatedBlockVersionTest = 2

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

}

// Allowed timestamp for a block building on the end of the provided best chain.
func MinimumMedianTime(chainState *blockchain.BestState) time.Time {
	return chainState.MedianTime.Add(time.Second)
}
