// Copyright (c) 2017-2018 The nox developers
package node

import (
	"github.com/noxproject/nox/services/acct"
	"github.com/noxproject/nox/services/blkmgr"
	"github.com/noxproject/nox/services/miner"
	"github.com/noxproject/nox/services/mempool"
	"github.com/noxproject/nox/database"
	"github.com/noxproject/nox/log"
	"github.com/noxproject/nox/rpc"
	"github.com/noxproject/nox/services/index"
	"github.com/noxproject/nox/core/blockchain"
	"github.com/noxproject/nox/engine/txscript"
	"github.com/noxproject/nox/p2p/peerserver"
)

// NoxFull implements the nox full node service.
type NoxFull struct {
	// under node
	node                 *Node
	// database
	db                   database.DB
	// account/wallet service
	acctmanager          *acct.AccountManager
	// block manager handles all incoming blocks.
	blockManager         *blkmgr.BlockManager
	// mempool hold tx that need to be mined into blocks and relayed to other peers.
	txMemPool            *mempool.TxPool
	// miner service
	cpuMiner             *miner.CPUMiner
	// tx index
	txIndex              *index.TxIndex
	// clock time service
	timeSource    		 blockchain.MedianTimeSource
	// signature cache
	sigCache             *txscript.SigCache
}

func (nox *NoxFull) Start(server *peerserver.PeerServer) error {
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
func newNoxFullNode(node *Node) (*NoxFull, error){

	// account manager
	acctmgr, err := acct.New()
	if err != nil{
		return nil,err
	}
	nox := NoxFull{
		node:         node,
		db:           node.DB,
		acctmanager:  acctmgr,
		timeSource:   blockchain.NewMedianTime(),
		sigCache:     txscript.NewSigCache(node.Config.SigCacheMaxSize),
	}
	// Create the transaction and address indexes if needed.
	var indexes []index.Indexer
	cfg := node.Config

	if cfg.TxIndex {
		log.Info("Transaction index is enabled")
		nox.txIndex = index.NewTxIndex(nox.db)
		indexes = append(indexes, nox.txIndex)
	}
	// index-manager
	var indexManager blockchain.IndexManager
	if len(indexes) > 0 {
		indexManager = index.NewManager(nox.db,indexes,node.Params)
	}

	// block-manager
	bm, err := blkmgr.NewBlockManager(indexManager,node.DB, nox.timeSource, nox.sigCache, node.Config, node.Params,
		node.peerServer, node.quit)
	if err != nil {
		return nil, err
	}
	nox.blockManager = bm


	return &nox, nil
}


