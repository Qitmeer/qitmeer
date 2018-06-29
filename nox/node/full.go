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
func newNoxFull() (*NoxFull, error){
	acctmgr, err := acct.New()
	if err != nil{
		return nil,err
	}
	nox := NoxFull{
		acctmanager: acctmgr,
	}
	return &nox, nil
}

// register NoxFull service to node
func RegisterNoxFull(n *Node) error{
	// register acctmgr
	err := n.Register(NewServiceConstructor("Nox",
		func(ctx *ServiceContext) (Service, error) {
		noxfull, err := newNoxFull()
		return noxfull, err
	}))
	return err
}

// register account manger service to node
func RegisterAcctMgr(n *Node) error{
	// register acctmgr
	err := n.Register(NewServiceConstructor("acctMgr",
		func(ctx *ServiceContext) (Service, error) {
		acctmgr, err := acct.New()
		return acctmgr, err
	}))
	return err
}

