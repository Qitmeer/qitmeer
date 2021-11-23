// Copyright (c) 2017-2018 The qitmeer developers
package node

import (
	"fmt"
	"github.com/Qitmeer/qitmeer/core/blockchain"
	"github.com/Qitmeer/qitmeer/core/coinbase"
	"github.com/Qitmeer/qitmeer/database"
	"github.com/Qitmeer/qng-core/engine/txscript"
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

	// address service
	addressApi *address.AddressApi

	// clock time service
	timeSource blockchain.MedianTimeSource
	// signature cache
	sigCache *txscript.SigCache
}

func (qm *QitmeerFull) APIs() []api.API {
	apis := qm.Service.APIs()
	apis = append(apis, qm.addressApi.APIs()...)
	apis = append(apis, qm.apis()...)
	return apis
}

func (qm *QitmeerFull) RegisterP2PService() error {
	peerServer, err := p2p.NewService(qm.node.Config, &qm.node.events, qm.node.Params)
	if err != nil {
		return err
	}
	return qm.Services().RegisterService(peerServer)
}

func (qm *QitmeerFull) RegisterRpcService() error {
	if qm.node.Config.DisableRPC {
		return nil
	}
	rpcServer, err := rpc.NewRPCServer(qm.node.Config, &qm.node.events)
	if err != nil {
		return err
	}
	qm.Services().RegisterService(rpcServer)

	go func() {
		<-rpcServer.RequestedProcessShutdown()
		qm.node.shutdownRequestChannel <- struct{}{}
	}()
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
			if err := rpcServer.RegisterService(api.NameSpace, api.Service); err != nil {
				return err
			}
			log.Debug(fmt.Sprintf("RPC Service API registered. NameSpace:%s     %s", api.NameSpace, reflect.TypeOf(api.Service)))
		}
	}
	return nil
}

func (qm *QitmeerFull) RegisterBlkMgrService(indexManager blockchain.IndexManager) error {

	// block-manager
	bm, err := blkmgr.NewBlockManager(qm.nfManager, indexManager, qm.node.DB, qm.timeSource, qm.sigCache, qm.node.Config, qm.node.Params,
		qm.node.quit, &qm.node.events, qm.GetPeerServer())
	if err != nil {
		return err
	}
	qm.Services().RegisterService(bm)
	return nil
}

func (qm *QitmeerFull) RegisterTxManagerService(txIndex *index.TxIndex, addrIndex *index.AddrIndex) error {
	// txmanager
	tm, err := tx.NewTxManager(qm.GetBlockManager(), txIndex, addrIndex, qm.node.Config, qm.nfManager, qm.sigCache, qm.node.DB, &qm.node.events)
	if err != nil {
		return err
	}
	qm.Services().RegisterService(tm)
	return nil
}

func (qm *QitmeerFull) RegisterMinerService() error {
	cfg := qm.node.Config
	txManager := qm.GetTxManager()
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
		CoinbaseGenerator: coinbase.NewCoinbaseGenerator(qm.node.Params, qm.GetPeerServer().PeerID().String()),
	}
	miner := miner.NewMiner(cfg, &policy, qm.sigCache,
		txManager.MemPool().(*mempool.TxPool), qm.timeSource, qm.GetBlockManager(), &qm.node.events)
	qm.Services().RegisterService(miner)
	return nil
}

func (qm *QitmeerFull) RegisterAccountService() error {
	// account manager
	acctmgr, err := acct.New()
	if err != nil {
		return err
	}
	qm.Services().RegisterService(acctmgr)
	return nil
}

// return block manager
func (qm *QitmeerFull) GetBlockManager() *blkmgr.BlockManager {
	var service *blkmgr.BlockManager
	if err := qm.Services().FetchService(&service); err != nil {
		log.Error(err.Error())
		return nil
	}
	return service
}

// return address api
func (qm *QitmeerFull) GetAddressApi() *address.AddressApi {
	return qm.addressApi
}

// return peer server
func (qm *QitmeerFull) GetPeerServer() *p2p.Service {
	var service *p2p.Service
	if err := qm.Services().FetchService(&service); err != nil {
		log.Error(err.Error())
		return nil
	}
	return service
}

func (qm *QitmeerFull) GetRpcServer() *rpc.RpcServer {
	var service *rpc.RpcServer
	if err := qm.Services().FetchService(&service); err != nil {
		log.Error(err.Error())
		return nil
	}
	return service
}

func (qm *QitmeerFull) GetTxManager() *tx.TxManager {
	var service *tx.TxManager
	if err := qm.Services().FetchService(&service); err != nil {
		log.Error(err.Error())
		return nil
	}
	return service
}

func newQitmeerFullNode(node *Node) (*QitmeerFull, error) {
	qm := QitmeerFull{
		node:       node,
		db:         node.DB,
		timeSource: blockchain.NewMedianTime(),
		sigCache:   txscript.NewSigCache(node.Config.SigCacheMaxSize),
	}
	qm.Service.InitServices()

	cfg := node.Config

	if err := qm.RegisterAccountService(); err != nil {
		return nil, err
	}

	if err := qm.RegisterP2PService(); err != nil {
		return nil, err
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

	nfManager := &notifymgr.NotifyMgr{Server: qm.GetPeerServer()}
	qm.nfManager = nfManager

	if err := qm.RegisterBlkMgrService(indexManager); err != nil {
		return nil, err
	}
	bm := qm.GetBlockManager()

	if err := qm.RegisterTxManagerService(txIndex, addrIndex); err != nil {
		return nil, err
	}

	txManager := qm.GetTxManager()
	bm.SetTxManager(txManager)
	// prepare peerServer
	qm.GetPeerServer().SetBlockChain(bm.GetChain())
	qm.GetPeerServer().SetTimeSource(qm.timeSource)
	qm.GetPeerServer().SetTxMemPool(txManager.MemPool().(*mempool.TxPool))
	qm.GetPeerServer().SetNotify(qm.nfManager)

	//
	if err := qm.RegisterMinerService(); err != nil {
		return nil, err
	}
	// init address api
	qm.addressApi = address.NewAddressApi(cfg, node.Params)

	if err := qm.RegisterRpcService(); err != nil {
		return nil, err
	}
	if qm.GetRpcServer() != nil {
		qm.GetRpcServer().BC = bm.GetChain()
		qm.GetRpcServer().TxIndex = txIndex
		qm.GetRpcServer().ChainParams = bm.ChainParams()

		nfManager.RpcServer = qm.GetRpcServer()
	}
	return &qm, nil
}
