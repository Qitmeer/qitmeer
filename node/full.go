// Copyright (c) 2017-2018 The qitmeer developers
package node

import (
	"context"
	"fmt"
	"github.com/Qitmeer/qitmeer/core/blockchain"
	"github.com/Qitmeer/qitmeer/core/coinbase"
	"github.com/Qitmeer/qitmeer/database"
	"github.com/Qitmeer/qitmeer/engine/txscript"
	"github.com/Qitmeer/qitmeer/node/notify"
	"github.com/Qitmeer/qitmeer/node/service"
	"github.com/Qitmeer/qitmeer/p2p"
	"github.com/Qitmeer/qitmeer/rpc"
	"github.com/Qitmeer/qitmeer/rpc/api"
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
	"reflect"
)

// QitmeerFull implements the qitmeer full node service.
type QitmeerFull struct {
	service.Service
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

	// network server
	peerServer *p2p.Service

	// api server
	rpcServer *rpc.RpcServer
}

func (qm *QitmeerFull) Start(ctx context.Context) error {
	log.Debug("Starting Qitmeer full node service")
	if err := qm.Service.Start(ctx); err != nil {
		return err
	}
	qm.blockManager.Start()
	qm.txManager.Start()
	qm.miner.Start()

	// start p2p server
	if err := qm.peerServer.Start(); err != nil {
		return err
	}
	// start RPC by service
	if !qm.node.Config.DisableRPC {
		if err := qm.startRPC(); err != nil {
			return err
		}
	}
	return nil
}

func (qm *QitmeerFull) Stop() error {
	log.Debug("Stopping Qitmeer full node service")
	if err := qm.Service.Stop(); err != nil {
		return err
	}

	log.Info("try stop miner")
	qm.miner.Stop()

	log.Info("try stop bm")
	qm.blockManager.Stop()
	qm.blockManager.WaitForStop()

	qm.txManager.Stop()

	// stop rpc server
	if qm.rpcServer != nil {
		qm.rpcServer.Stop()
	}

	// stop p2p server
	qm.peerServer.Stop()

	return nil
}

// startRPC is a helper method to start all the various RPC endpoint during node
// startup. It's not meant to be called at any time afterwards as it makes certain
// assumptions about the state of the node.
func (qm *QitmeerFull) startRPC() error {
	// Gather all the possible APIs to surface
	apis := qm.APIs()

	// Generate the whitelist based on the allowed modules
	whitelist := make(map[string]bool)
	for _, module := range qm.node.Config.Modules {
		whitelist[module] = true
	}

	// Register all the APIs exposed by the services
	for _, api := range apis {
		if whitelist[api.NameSpace] || (len(whitelist) == 0 && api.Public) {
			if err := qm.rpcServer.RegisterService(api.NameSpace, api.Service); err != nil {
				return err
			}
			log.Debug(fmt.Sprintf("RPC Service API registered. NameSpace:%s     %s", api.NameSpace, reflect.TypeOf(api.Service)))
		}
	}
	if err := qm.rpcServer.Start(); err != nil {
		return err
	}
	return nil
}

func (qm *QitmeerFull) APIs() []api.API {
	apis := qm.acctmanager.APIs()
	apis = append(apis, qm.addressApi.APIs()...)
	apis = append(apis, qm.miner.APIs()...)
	apis = append(apis, qm.blockManager.API())
	apis = append(apis, qm.txManager.APIs()...)
	apis = append(apis, qm.apis()...)
	return apis
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
	return qm.peerServer
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

	cfg := node.Config
	server, err := p2p.NewService(node.Config, &node.events, node.Params)
	if err != nil {
		return nil, err
	}
	qm.peerServer = server
	//
	if !cfg.DisableRPC {
		qm.rpcServer, err = rpc.NewRPCServer(cfg, &qm.node.events)
		if err != nil {
			return nil, err
		}
		go func() {
			<-qm.rpcServer.RequestedProcessShutdown()
			qm.node.shutdownRequestChannel <- struct{}{}
		}()
	}

	// Create the transaction and address indexes if needed.
	var indexes []index.Indexer

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

	qm.nfManager = &notifymgr.NotifyMgr{Server: qm.peerServer, RpcServer: qm.rpcServer}

	// block-manager
	bm, err := blkmgr.NewBlockManager(qm.nfManager, indexManager, node.DB, qm.timeSource, qm.sigCache, node.Config, node.Params,
		node.quit, &node.events, qm.peerServer)
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
	qm.peerServer.SetBlockChain(bm.GetChain())
	qm.peerServer.SetTimeSource(qm.timeSource)
	qm.peerServer.SetTxMemPool(qm.txManager.MemPool().(*mempool.TxPool))
	qm.peerServer.SetNotify(qm.nfManager)

	if qm.rpcServer != nil {
		qm.rpcServer.BC = bm.GetChain()
		qm.rpcServer.TxIndex = txIndex
		qm.rpcServer.ChainParams = bm.ChainParams()
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
		CoinbaseGenerator: coinbase.NewCoinbaseGenerator(node.Params, qm.peerServer.PeerID().String()),
	}
	qm.miner = miner.NewMiner(cfg, &policy, qm.sigCache,
		qm.txManager.MemPool().(*mempool.TxPool), qm.timeSource, qm.blockManager, &node.events)

	// init address api
	qm.addressApi = address.NewAddressApi(cfg, node.Params)
	return &qm, nil
}
