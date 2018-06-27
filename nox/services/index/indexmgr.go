package index

import (
	"github.com/noxproject/nox/params"
	"github.com/noxproject/nox/database"
	"github.com/noxproject/nox/core/blockchain"
	"github.com/noxproject/nox/core/types"
)

// Manager defines an index manager that manages multiple optional indexes and
// implements the blockchain.IndexManager interface so it can be seamlessly
// plugged into normal chain processing.
type Manager struct {
	params         *params.Params
	db             database.DB
	enabledIndexes []Indexer
}

// Indexer provides a generic interface for an indexer that is managed by an
// index manager such as the Manager type provided by this package.
type Indexer interface {
	// Key returns the key of the index as a byte slice.
	Key() []byte

	// Name returns the human-readable name of the index.
	Name() string

	// Create is invoked when the indexer manager determines the index needs
	// to be created for the first time.
	Create(dbTx database.Tx) error

	// Init is invoked when the index manager is first initializing the
	// index.  This differs from the Create method in that it is called on
	// every load, including the case the index was just created.
	Init() error

	// ConnectBlock is invoked when the index manager is notified that a new
	// block has been connected to the main chain.
	ConnectBlock(dbTx database.Tx, block, parent *types.Block, view *blockchain.UtxoViewpoint) error

	// DisconnectBlock is invoked when the index manager is notified that a
	// block has been disconnected from the main chain.
	DisconnectBlock(dbTx database.Tx, block, parent *types.Block, view *blockchain.UtxoViewpoint) error
}

// Ensure the Manager type implements the blockchain.IndexManager interface.
var _ blockchain.IndexManager = (*Manager)(nil)
