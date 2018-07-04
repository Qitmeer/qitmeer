// Copyright (c) 2017-2018 The nox developers
package node

import (
	"github.com/noxproject/nox/services/acct"
	"github.com/noxproject/nox/services/blkmgr"
	"github.com/noxproject/nox/services/miner"
	"github.com/noxproject/nox/services/mempool"
	"github.com/noxproject/nox/database"
	"github.com/noxproject/nox/p2p"
	"github.com/noxproject/nox/log"
	"github.com/noxproject/nox/rpc"
	"github.com/noxproject/nox/services/index"
	"github.com/noxproject/nox/config"
)

// NoxFull implements the nox full node service.
type NoxFull struct {
	// database
	db                   *database.DB
	// account/wallet service
	acctmanager          *acct.AccountManager
	// block manager handles all incoming blocks.
	blockManager         *blkmgr.BlockManager
	// mempool hold tx that need to be mined into blocks and relayed to other peers.
	txMemPool            *mempool.TxPool
	// miner service
	cpuMiner             *miner.CPUMiner
	// index
	txIndex              *index.TxIndex

}

func (nox *NoxFull) Start(server *p2p.PeerServer) error {
	log.Debug("Starting Nox full node service")
	return nil
}

func (nox *NoxFull) Stop() error {
	log.Debug("Stopping Nox full node service")
	return nil
}

func (nox *NoxFull)	APIs() []rpc.API {
	return nox.acctmanager.APIs()
}
func newNoxFull(cfg *config.Config,db database.DB) (*NoxFull, error){

	// Create account manager
	acctmgr, err := acct.New()
	if err != nil{
		return nil,err
	}
	nox := NoxFull{
		acctmanager: acctmgr,
	}
	// Create the transaction and address indexes if needed.
	var indexes []index.Indexer
	if cfg.TxIndex {
		log.Info("Transaction index is enabled")
		nox.txIndex = index.NewTxIndex(db)
		indexes = append(indexes, nox.txIndex)
	}
	return &nox, nil
}

// register NoxFull service to node
func registerNoxFull(n *Node) error{
	// register acctmgr
	err := n.register(NewServiceConstructor("Nox",
		func(ctx *ServiceContext) (Service, error) {
		noxfull, err := newNoxFull(n.config,n.db)
		return noxfull, err
	}))
	return err
}
