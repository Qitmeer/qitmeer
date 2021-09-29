package tx

import (
	"fmt"
	"github.com/Qitmeer/qitmeer/common/hash"
	"github.com/Qitmeer/qitmeer/config"
	"github.com/Qitmeer/qitmeer/core/blockchain"
	"github.com/Qitmeer/qitmeer/core/blockdag"
	"github.com/Qitmeer/qitmeer/core/event"
	"github.com/Qitmeer/qitmeer/core/types"
	"github.com/Qitmeer/qitmeer/database"
	"github.com/Qitmeer/qitmeer/engine/txscript"
	"github.com/Qitmeer/qitmeer/node/notify"
	"github.com/Qitmeer/qitmeer/services/blkmgr"
	"github.com/Qitmeer/qitmeer/services/common"
	"github.com/Qitmeer/qitmeer/services/index"
	"github.com/Qitmeer/qitmeer/services/mempool"
	"time"
)

type TxManager struct {
	bm *blkmgr.BlockManager
	// tx index
	txIndex *index.TxIndex

	// addr index
	addrIndex *index.AddrIndex
	// mempool hold tx that need to be mined into blocks and relayed to other peers.
	txMemPool *mempool.TxPool

	// notify
	ntmgr notify.Notify

	// db
	db database.DB

	//invalidTx hash->block hash
	invalidTx map[hash.Hash]*blockdag.HashSet
}

func (tm *TxManager) Start() error {
	log.Info("Starting tx manager")
	err := tm.txMemPool.Load()
	if err != nil {
		log.Error(err.Error())
	}
	return nil
}

func (tm *TxManager) Stop() error {
	log.Info("Stopping tx manager")

	if tm.txMemPool.IsPersist() {
		num, err := tm.txMemPool.Save()
		if err != nil {
			log.Error(err.Error())
		} else {
			log.Info(fmt.Sprintf("Mempool persist:%d transactions", num))
		}
	}

	return nil
}

func (tm *TxManager) MemPool() blkmgr.TxPool {
	return tm.txMemPool
}

func NewTxManager(bm *blkmgr.BlockManager, txIndex *index.TxIndex,
	addrIndex *index.AddrIndex, cfg *config.Config, ntmgr notify.Notify,
	sigCache *txscript.SigCache, db database.DB, events *event.Feed) (*TxManager, error) {
	// mem-pool
	amt, _ := types.NewMeer(uint64(cfg.MinTxFee))
	txC := mempool.Config{
		Policy: mempool.Policy{
			MaxTxVersion:         2,
			DisableRelayPriority: cfg.NoRelayPriority,
			AcceptNonStd:         cfg.AcceptNonStd,
			FreeTxRelayLimit:     cfg.FreeTxRelayLimit,
			MaxOrphanTxs:         cfg.MaxOrphanTxs,
			MaxOrphanTxSize:      mempool.DefaultMaxOrphanTxSize,
			MaxSigOpsPerTx:       blockchain.MaxSigOpsPerBlock / 5,
			MinRelayTxFee:        *amt,
			StandardVerifyFlags: func() (txscript.ScriptFlags, error) {
				return common.StandardScriptVerifyFlags()
			},
		},
		ChainParams:      bm.ChainParams(),
		FetchUtxoView:    bm.GetChain().FetchUtxoView, //TODO, duplicated dependence of miner
		BlockByHash:      bm.GetChain().FetchBlockByHash,
		BestHash:         func() *hash.Hash { return &bm.GetChain().BestSnapshot().Hash },
		BestHeight:       func() uint64 { return uint64(bm.GetChain().BestSnapshot().GraphState.GetMainHeight()) },
		CalcSequenceLock: bm.GetChain().CalcSequenceLock,
		SubsidyCache:     bm.GetChain().FetchSubsidyCache(),
		SigCache:         sigCache,
		PastMedianTime:   func() time.Time { return bm.GetChain().BestSnapshot().MedianTime },
		AddrIndex:        addrIndex,
		BD:               bm.GetChain().BlockDAG(),
		BC:               bm.GetChain(),
		DataDir:          cfg.DataDir,
		Expiry:           time.Duration(cfg.MempoolExpiry),
		Persist:          cfg.Persistmempool,
		NoMempoolBar:     cfg.NoMempoolBar,
		Events:           events,
	}
	txMemPool := mempool.New(&txC)
	invalidTx := make(map[hash.Hash]*blockdag.HashSet)
	return &TxManager{bm, txIndex, addrIndex, txMemPool, ntmgr, db, invalidTx}, nil
}
