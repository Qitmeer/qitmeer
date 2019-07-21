// Copyright (c) 2017-2018 The qitmeer developers
package node

import (
	"github.com/HalalChain/qitmeer-lib/common/hash"
	"github.com/HalalChain/qitmeer/core/blockchain"
	"github.com/HalalChain/qitmeer-lib/core/types"
	"github.com/HalalChain/qitmeer/database"
	"github.com/HalalChain/qitmeer-lib/engine/txscript"
	"github.com/HalalChain/qitmeer/node/notify"
	"github.com/HalalChain/qitmeer/p2p/peerserver"
	"github.com/HalalChain/qitmeer-lib/params"
	"github.com/HalalChain/qitmeer-lib/rpc"
	"github.com/HalalChain/qitmeer/services/acct"
	"github.com/HalalChain/qitmeer/services/blkmgr"
	"github.com/HalalChain/qitmeer/services/index"
	"github.com/HalalChain/qitmeer/services/mempool"
	"github.com/HalalChain/qitmeer/services/miner"
	"github.com/HalalChain/qitmeer/services/mining"
	"github.com/HalalChain/qitmeer/services/notifymgr"
	"time"
)

// QitmeerFull implements the qitmeer full node service.
type QitmeerFull struct {
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

func (qm *QitmeerFull) Start(server *peerserver.PeerServer) error {
	log.Debug("Starting Qitmeer full node service")

	// Start the CPU miner if generation is enabled.
	if qm.node.Config.Generate {
		qm.cpuMiner.Start()
	}

	qm.blockManager.Start()
	return nil
}

func (qm *QitmeerFull) Stop() error {
	log.Debug("Stopping Qitmeer full node service")

	log.Info("try stop bm")

	qm.blockManager.Stop()
	qm.blockManager.WaitForStop()

	log.Info("try stop cpu miner")
	// Stop the CPU miner if needed.
	if qm.node.Config.Generate && qm.cpuMiner != nil {
		qm.cpuMiner.Stop()
	}

	return nil
}

func (qm *QitmeerFull)	APIs() []rpc.API {
	apis := qm.acctmanager.APIs()
	apis = append(apis,qm.cpuMiner.APIs()...)
	apis = append(apis,qm.blockManager.API())
	apis = append(apis,qm.txMemPool.API())
	apis = append(apis,qm.API())
	return apis
}
func newQitmeerFullNode(node *Node) (*QitmeerFull, error){

	// account manager
	acctmgr, err := acct.New()
	if err != nil{
		return nil,err
	}
	qm := QitmeerFull{
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
		qm.txIndex = index.NewTxIndex(qm.db)
		indexes = append(indexes, qm.txIndex)
	}
	// index-manager
	var indexManager blockchain.IndexManager
	if len(indexes) > 0 {
		indexManager = index.NewManager(qm.db,indexes,node.Params)
	}

	qm.nfManager = &notifymgr.NotifyMgr{Server:node.peerServer, RpcServer:node.rpcServer}

	// block-manager
	bm, err := blkmgr.NewBlockManager(qm.nfManager,indexManager,node.DB, qm.timeSource, qm.sigCache, node.Config, node.Params,
		node.quit)
	if err != nil {
		return nil, err
	}
	qm.blockManager = bm

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
		SigCache:         qm.sigCache,
		PastMedianTime:   func() time.Time { return bm.GetChain().BestSnapshot().MedianTime },
	}
	qm.txMemPool = mempool.New(&txC)

	// set mempool to bm
	bm.SetMemPool(qm.txMemPool)

	// prepare peerServer
	node.peerServer.BlockManager = bm
	node.peerServer.TimeSource = qm.timeSource
	node.peerServer.TxMemPool = qm.txMemPool

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

	qm.cpuMiner = miner.NewCPUMiner(cfg,node.Params,&policy,qm.sigCache,
		qm.txMemPool,qm.timeSource,qm.blockManager,defaultNumWorkers)

	return &qm, nil
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
func (qm *QitmeerFull) GetBlockManager() *blkmgr.BlockManager{
	return qm.blockManager
}

// return cpu miner
func (qm *QitmeerFull) GetCpuMiner() *miner.CPUMiner{
	return qm.cpuMiner
}

// return peer server
func (qm *QitmeerFull) GetPeerServer() *peerserver.PeerServer {
	return qm.node.peerServer
}