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
	"github.com/noxproject/nox/common/hash"
	"time"
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
	apis := nox.acctmanager.APIs()
	apis = append(apis,nox.cpuMiner.APIs()...)
	apis = append(apis,nox.blockManager.API())
	apis = append(apis,nox.API())
	return apis
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

	// mem-pool
	txC := mempool.Config{
		Policy: mempool.Policy{
			MaxTxVersion:         2,
			DisableRelayPriority: cfg.NoRelayPriority,
			AcceptNonStd:         cfg.AcceptNonStd,
			FreeTxRelayLimit:     cfg.FreeTxRelayLimit,
			MaxOrphanTxs:         cfg.MaxOrphanTxs,
			MaxOrphanTxSize:      mempool.DefaultMaxOrphanTxSize,
			MaxSigOpsPerTx:       blockchain.MaxSigOpsPerBlock / 5,
			MinRelayTxFee:        cfg.MinRelayTxFee,
			StandardVerifyFlags: func() (txscript.ScriptFlags, error) {
				return standardScriptVerifyFlags(bm.GetChain())
			},
		},
		ChainParams:      node.Params,
		FetchUtxoView:    bm.GetChain().FetchUtxoView,
		BlockByHash:      bm.GetChain().BlockByHash,
		BestHash:         func() *hash.Hash { return &bm.GetChain().BestSnapshot().Hash },
		BestHeight:       func() uint64 { return bm.GetChain().BestSnapshot().Height },
		SigCache:         nox.sigCache,
		PastMedianTime:   func() time.Time { return bm.GetChain().BestSnapshot().MedianTime },
	}
	nox.txMemPool = mempool.New(&txC)

	// Cpu Miner

	return &nox, nil
}

// standardScriptVerifyFlags returns the script flags that should be used when
// executing transaction scripts to enforce additional checks which are required
// for the script to be considered standard.  Note these flags are different
// than what is required for the consensus rules in that they are more strict.
func standardScriptVerifyFlags(chain *blockchain.BlockChain) (txscript.ScriptFlags, error) {
	scriptFlags := mempool.BaseStandardVerifyFlags
	return scriptFlags, nil
}


