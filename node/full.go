// Copyright (c) 2017-2018 The nox developers
package node

import (
	"github.com/HalalChain/qitmeer-lib/common/hash"
	"github.com/HalalChain/qitmeer/core/blockchain"
	"github.com/HalalChain/qitmeer/core/types"
	"github.com/HalalChain/qitmeer/database"
	"github.com/HalalChain/qitmeer/engine/txscript"
	"github.com/HalalChain/qitmeer/node/notify"
	"github.com/HalalChain/qitmeer/p2p/peerserver"
	"github.com/HalalChain/qitmeer/params"
	"github.com/HalalChain/qitmeer/rpc"
	"github.com/HalalChain/qitmeer/services/acct"
	"github.com/HalalChain/qitmeer/services/blkmgr"
	"github.com/HalalChain/qitmeer/services/index"
	"github.com/HalalChain/qitmeer/services/mempool"
	"github.com/HalalChain/qitmeer/services/miner"
	"github.com/HalalChain/qitmeer/services/mining"
	"github.com/HalalChain/qitmeer/services/notifymgr"
	"time"
)

// NoxFull implements the nox full node service.
type NoxFull struct {
	// under node
	node                 *Node
	// msg notifier
	nfManager            notify.Notify
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

	// Start the CPU miner if generation is enabled.
	if nox.node.Config.Generate {
		nox.cpuMiner.Start()
	}

	nox.blockManager.Start()
	return nil
}

func (nox *NoxFull) Stop() error {
	log.Debug("Stopping Nox full node service")

	log.Info("try stop bm")

	go func() {
		nox.blockManager.Stop()
		nox.blockManager.WaitForStop()
	}()

	log.Info("try stop cpu miner")
	// Stop the CPU miner if needed.
	if nox.node.Config.Generate && nox.cpuMiner != nil {
		nox.cpuMiner.Stop()
	}

	return nil
}

func (nox *NoxFull)	APIs() []rpc.API {
	apis := nox.acctmanager.APIs()
	apis = append(apis,nox.cpuMiner.APIs()...)
	apis = append(apis,nox.blockManager.API())
	apis = append(apis,nox.txMemPool.API())
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

	nox.nfManager = &notifymgr.NotifyMgr{Server:node.peerServer, RpcServer:node.rpcServer}

	// block-manager
	bm, err := blkmgr.NewBlockManager(nox.nfManager,indexManager,node.DB, nox.timeSource, nox.sigCache, node.Config, node.Params,
		node.quit)
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
			MinRelayTxFee:        types.Amount(cfg.MinTxFee),
			StandardVerifyFlags: func() (txscript.ScriptFlags, error) {
				return standardScriptVerifyFlags(bm.GetChain())
			},
		},
		ChainParams:      node.Params,
		FetchUtxoView:    bm.GetChain().FetchUtxoView,  //TODO, duplicated dependence of miner
		BlockByHash:      bm.GetChain().FetchBlockByHash,
		BestHash:         func() *hash.Hash { return &bm.GetChain().BestSnapshot().Hash },
		BestHeight:       func() uint64 { return bm.GetChain().BestSnapshot().Order },
		CalcSequenceLock: bm.GetChain().CalcSequenceLock,
		SubsidyCache:     bm.GetChain().FetchSubsidyCache(),
		SigCache:         nox.sigCache,
		PastMedianTime:   func() time.Time { return bm.GetChain().BestSnapshot().MedianTime },
	}
	nox.txMemPool = mempool.New(&txC)

	// set mempool to bm
	bm.SetMemPool(nox.txMemPool)

	// prepare peerServer
	node.peerServer.BlockManager = bm
	node.peerServer.TimeSource = nox.timeSource
	node.peerServer.TxMemPool = nox.txMemPool

	// Cpu Miner
	// Create the mining policy based on the configuration options.
	// NOTE: The CPU miner relies on the mempool, so the mempool has to be
	// created before calling the function to create the CPU miner.
	policy := mining.Policy{
		BlockMinSize:      cfg.BlockMinSize,
		BlockMaxSize:      cfg.BlockMaxSize,
		BlockPrioritySize: cfg.BlockPrioritySize,
		TxMinFreeFee:      cfg.MinTxFee,    //TODO, duplicated config item with mem-pool
		StandardVerifyFlags: func() (txscript.ScriptFlags, error) {
				return standardScriptVerifyFlags(bm.GetChain())
		}, //TODO, duplicated config item with mem-pool
	}
	// defaultNumWorkers is the default number of workers to use for mining
	// and is based on the number of processor cores.  This helps ensure the
	// system stays reasonably responsive under heavy load.
	defaultNumWorkers := uint32(params.CPUMinerThreads) //TODO, move to config

	nox.cpuMiner = miner.NewCPUMiner(cfg,node.Params,&policy,nox.sigCache,
		nox.txMemPool,nox.timeSource,nox.blockManager,defaultNumWorkers)

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

// return block manager
func (nox *NoxFull) GetBlockManager() *blkmgr.BlockManager{
	return nox.blockManager
}

// return cpu miner
func (nox *NoxFull) GetCpuMiner() *miner.CPUMiner{
	return nox.cpuMiner
}
