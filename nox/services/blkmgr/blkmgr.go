package blkmgr

import (
	"sync"
	"container/list"
	"github.com/noxproject/nox/common/hash"
	"github.com/noxproject/nox/core/blockchain"
	"github.com/noxproject/nox/params"
	"time"
	"github.com/noxproject/nox/p2p"
)

// BlockManager provides a concurrency safe block manager for handling all
// incoming blocks.
type BlockManager struct {
	started             int32
	shutdown            int32
	chain               *blockchain.BlockChain
	rejectedTxns        map[hash.Hash]struct{}
	requestedTxns       map[hash.Hash]struct{}
	requestedEverTxns   map[hash.Hash]uint8
	requestedBlocks     map[hash.Hash]struct{}
	requestedEverBlocks map[hash.Hash]uint8

	syncPeer            *p2p.Peer
	msgChan             chan interface{}
	chainState          chainState
	wg                  sync.WaitGroup
	quit                chan struct{}

	// The following fields are used for headers-first mode.
	headersFirstMode bool
	headerList       *list.List
	startHeader      *list.Element
	nextCheckpoint   *params.Checkpoint

	AggressiveMining      bool
}

// chainState tracks the state of the best chain as blocks are inserted.  This
// is done because blockchain is currently not safe for concurrent access and the
// block manager is typically quite busy processing block and inventory.
// Therefore, requesting this information from chain through the block manager
// would not be anywhere near as efficient as simply updating it as each block
// is inserted and protecting it with a mutex.
type chainState struct {
	sync.Mutex
	newestHash          *hash.Hash
	newestHeight        int64
	nextFinalState      [6]byte
	nextPoolSize        uint32
	nextStakeDifficulty int64
	winningTickets      []hash.Hash
	missedTickets       []hash.Hash
	curPrevHash         hash.Hash
	pastMedianTime      time.Time
}
