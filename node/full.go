// Copyright (c) 2017-2018 The qitmeer developers
package node

import (
	"github.com/Qitmeer/qitmeer/core/blockchain"
	"github.com/Qitmeer/qitmeer/core/coinbase"
	"github.com/Qitmeer/qitmeer/database"
	"github.com/Qitmeer/qitmeer/engine/txscript"
	"github.com/Qitmeer/qitmeer/node/notify"
	"github.com/Qitmeer/qitmeer/p2p"
	"github.com/Qitmeer/qitmeer/rpc"
	"github.com/Qitmeer/qitmeer/services/acct"
	"github.com/Qitmeer/qitmeer/services/address"
	"github.com/Qitmeer/qitmeer/services/blkmgr"
	"github.com/Qitmeer/qitmeer/services/common"
	"github.com/Qitmeer/qitmeer/services/index"
	"github.com/Qitmeer/qitmeer/services/mempool"
	"github.com/Qitmeer/qitmeer/services/miner"
	"github.com/Qitmeer/qitmeer/services/mining"
	"github.com/Qitmeer/qitmeer/services/notifymgr"
	"github.com/Qitmeer/qitmeer/services/tx"
)

// QitmeerFull implements the qitmeer full node service.
type QitmeerFull struct {
	// under node
	node *Node
	// msg notifier
	nfManager notify.Notify
	// database
	db database.DB
	// account/wallet service
	acctmanager *acct.AccountManager
	// block manager handles all incoming blocks.
	blockManager *blkmgr.BlockManager
	// tx manager
	txManager *tx.TxManager

	// miner service
	miner *miner.Miner

	// address service
	addressApi *address.AddressApi

	// clock time service
	timeSource blockchain.MedianTimeSource
	// signature cache
	sigCache *txscript.SigCache
}

func (qm *QitmeerFull) Start() error {
	log.Debug("Starting Qitmeer full node service")
	qm.blockManager.Start()
	qm.txManager.Start()
	qm.miner.Start()
	return nil
}

func (qm *QitmeerFull) Stop() error {
	log.Debug("Stopping Qitmeer full node service")
	log.Info("try stop miner")
	qm.miner.Stop()

	log.Info("try stop bm")
	qm.blockManager.Stop()
	qm.blockManager.WaitForStop()

	qm.txManager.Stop()
	return nil
}

func (qm *QitmeerFull) APIs() []rpc.API {
	apis := qm.acctmanager.APIs()
	apis = append(apis, qm.addressApi.APIs()...)
	apis = append(apis, qm.miner.APIs()...)
	apis = append(apis, qm.blockManager.API())
	apis = append(apis, qm.txManager.APIs()...)
	apis = append(apis, qm.apis()...)
	return apis
}
func newQitmeerFullNode(node *Node) (*QitmeerFull, error) {

	// account manager
	acctmgr, err := acct.New()
	if err != nil {
		return nil, err
	}
	qm := QitmeerFull{
		node:        node,
		db:          node.DB,
		acctmanager: acctmgr,
		timeSource:  blockchain.NewMedianTime(),
		sigCache:    txscript.NewSigCache(node.Config.SigCacheMaxSize),
	}
	// Create the transaction and address indexes if needed.
	var indexes []index.Indexer
	cfg := node.Config

	var txIndex *index.TxIndex
	var addrIndex *index.AddrIndex
	log.Info("Transaction index is enabled")
	txIndex = index.NewTxIndex(qm.db)
	indexes = append(indexes, txIndex)
	if cfg.AddrIndex {
		log.Info("Address index is enabled")
		addrIndex = index.NewAddrIndex(qm.db, node.Params)
		indexes = append(indexes, addrIndex)
	}
	// index-manager
	var indexManager blockchain.IndexManager
	if len(indexes) > 0 {
		indexManager = index.NewManager(qm.db, indexes, node.Params)
	}

	qm.nfManager = &notifymgr.NotifyMgr{Server: node.peerServer, RpcServer: node.rpcServer}

	// block-manager
	bm, err := blkmgr.NewBlockManager(qm.nfManager, indexManager, node.DB, qm.timeSource, qm.sigCache, node.Config, node.Params,
		node.quit, &node.events, node.peerServer)
	if err != nil {
		return nil, err
	}
	qm.blockManager = bm

	// txmanager
	tm, err := tx.NewTxManager(bm, txIndex, addrIndex, cfg, qm.nfManager, qm.sigCache, node.DB, &node.events)
	if err != nil {
		return nil, err
	}
	qm.txManager = tm
	bm.SetTxManager(tm)
	// prepare peerServer
	node.peerServer.SetBlockChain(bm.GetChain())
	node.peerServer.SetTimeSource(qm.timeSource)
	node.peerServer.SetTxMemPool(qm.txManager.MemPool().(*mempool.TxPool))
	node.peerServer.SetNotify(qm.nfManager)

	if node.rpcServer != nil {
		node.rpcServer.BC = bm.GetChain()
		node.rpcServer.TxIndex = txIndex
		node.rpcServer.ChainParams = bm.ChainParams()
	}

	// Cpu Miner
	// Create the mining policy based on the configuration options.
	// NOTE: The CPU miner relies on the mempool, so the mempool has to be
	// created before calling the function to create the CPU miner.
	policy := mining.Policy{
		BlockMinSize:      cfg.BlockMinSize,
		BlockMaxSize:      cfg.BlockMaxSize,
		BlockPrioritySize: cfg.BlockPrioritySize,
		TxMinFreeFee:      cfg.MinTxFee, //TODO, duplicated config item with mem-pool
		StandardVerifyFlags: func() (txscript.ScriptFlags, error) {
			return common.StandardScriptVerifyFlags()
		}, //TODO, duplicated config item with mem-pool
		CoinbaseGenerator: coinbase.NewCoinbaseGenerator(node.Params, qm.node.peerServer.PeerID().String()),
	}
	qm.miner = miner.NewMiner(cfg, &policy, qm.sigCache,
		qm.txManager.MemPool().(*mempool.TxPool), qm.timeSource, qm.blockManager, &node.events)

	// init address api
	qm.addressApi = address.NewAddressApi(cfg, node.Params)
	return &qm, nil
}

// return block manager
func (qm *QitmeerFull) GetBlockManager() *blkmgr.BlockManager {
	return qm.blockManager
}

// return address api
func (qm *QitmeerFull) GetAddressApi() *address.AddressApi {
	return qm.addressApi
}

// return peer server
func (qm *QitmeerFull) GetPeerServer() *p2p.Service {
	return qm.node.peerServer
}
