package blockchain

import (
	"sync"
	"github.com/noxproject/nox/params"
	"github.com/noxproject/nox/database"
	"github.com/noxproject/nox/common/hash"
	"github.com/noxproject/nox/core/types"
)

// IndexManager provides a generic interface that the is called when blocks are
// connected and disconnected to and from the tip of the main chain for the
// purpose of supporting optional indexes.
type IndexManager interface {
	// Init is invoked during chain initialize in order to allow the index
	// manager to initialize itself and any indexes it is managing.  The
	// channel parameter specifies a channel the caller can close to signal
	// that the process should be interrupted.  It can be nil if that
	// behavior is not desired.
	Init(*BlockChain, <-chan struct{}) error

	// ConnectBlock is invoked when a new block has been connected to the
	// main chain.
	ConnectBlock(tx database.Tx, block *types.SerializedBlock, parent *types.SerializedBlock, utxoView *UtxoViewpoint) error

	// DisconnectBlock is invoked when a block has been disconnected from
	// the main chain.
	DisconnectBlock(tx database.Tx, block *types.SerializedBlock, parent *types.SerializedBlock,  utxoView *UtxoViewpoint) error
}

// blockIndex provides facilities for keeping track of an in-memory index of the
// block chain.  Although the name block chain suggests a single chain of
// blocks, it is actually a tree-shaped structure where any node can have
// multiple children.  However, there can only be one active branch which does
// indeed form a chain from the tip all the way back to the genesis block.
type blockIndex struct {
	// The following fields are set when the instance is created and can't
	// be changed afterwards, so there is no need to protect them with a
	// separate mutex.
	db          database.DB
	chainParams *params.Params

	sync.RWMutex
	index     map[hash.Hash]*blockNode
	chainTips map[int64][]*blockNode
}
