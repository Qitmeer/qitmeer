package mempool

import (
	"sync"
	"github.com/noxproject/nox/core/types"
	"github.com/noxproject/nox/common/hash"
)

// TxPool is used as a source of transactions that need to be mined into blocks
// and relayed to other peers.  It is safe for concurrent access from multiple
// peers.
type TxPool struct {
	// The following variables must only be used atomically.
	lastUpdated int64 // last time pool was updated.

	mtx           sync.RWMutex
	cfg           Config
	pool          map[hash.Hash]*TxDesc
	orphans       map[hash.Hash]*types.Tx
	orphansByPrev map[hash.Hash]map[hash.Hash]*types.Tx
	outpoints     map[types.TxOutPoint]*types.Tx

	pennyTotal    float64 // exponentially decaying total for penny spends.
	lastPennyUnix int64   // unix time of last ``penny spend''
}

// TxDesc is a descriptor containing a transaction in the mempool along with
// additional metadata.
type TxDesc struct {
	types.TxDesc

	// StartingPriority is the priority of the transaction when it was added
	// to the pool.
	StartingPriority float64
}
