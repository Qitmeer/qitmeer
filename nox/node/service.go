package node

import (
	"github.com/noxproject/nox/services/acct"
	"github.com/noxproject/nox/services/blkmgr"
	"github.com/noxproject/nox/services/mempool"
	"github.com/noxproject/nox/services/miner"
)

type Service interface {

	// Start is called after all services have been constructed and the networking
	// layer was also initialized to spawn any goroutines required by the service.
	Start() error

	// Stop terminates all goroutines belonging to the service, blocking until they
	// are all terminated.
	Stop() error
}

type service struct {
	//   account/wallet service
	acctmanager           *acct.AccountManager
	//   block manager handles all incoming blocks.
	blockManager         *blkmgr.BlockManager
	//   mempool of transactions that need to be mined into blocks and relayed to other peers.
	txMemPool            *mempool.TxPool
	//   miner service
	cpuMiner             *miner.CPUMiner

}

