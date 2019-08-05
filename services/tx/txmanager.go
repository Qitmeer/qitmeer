package tx

import (
	"github.com/HalalChain/qitmeer-lib/common/hash"
	"github.com/HalalChain/qitmeer-lib/config"
	"github.com/HalalChain/qitmeer-lib/core/types"
	"github.com/HalalChain/qitmeer-lib/engine/txscript"
	"github.com/HalalChain/qitmeer/core/blockchain"
	"github.com/HalalChain/qitmeer/database"
	"github.com/HalalChain/qitmeer/node/notify"
	"github.com/HalalChain/qitmeer/services/blkmgr"
	"github.com/HalalChain/qitmeer/services/common"
	"github.com/HalalChain/qitmeer/services/index"
	"github.com/HalalChain/qitmeer/services/mempool"
	"time"
)

type TxManager struct {
	bm                   *blkmgr.BlockManager
	// tx index
	txIndex              *index.TxIndex

	// addr index
	addrIndex            *index.AddrIndex
	// mempool hold tx that need to be mined into blocks and relayed to other peers.
	txMemPool            *mempool.TxPool

	// notify
	ntmgr                notify.Notify

	// db
	db                   database.DB
}

func (tm *TxManager) Start() error {
	log.Info("Starting tx manager")
	return nil
}

func (tm *TxManager) Stop() error {
	log.Info("Stopping tx manager")
	return nil
}

func (tm *TxManager) MemPool() *mempool.TxPool {
	return tm.txMemPool
}

func NewTxManager(bm *blkmgr.BlockManager,txIndex *index.TxIndex,
	addrIndex *index.AddrIndex,cfg *config.Config,ntmgr notify.Notify,
	sigCache *txscript.SigCache,db database.DB) (*TxManager,error) {
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
				return common.StandardScriptVerifyFlags()
			},
		},
		ChainParams:      bm.ChainParams(),
		FetchUtxoView:    bm.GetChain().FetchUtxoView,  //TODO, duplicated dependence of miner
		BlockByHash:      bm.GetChain().FetchBlockByHash,
		BestHash:         func() *hash.Hash { return &bm.GetChain().BestSnapshot().Hash },
		BestHeight:       func() uint64 { return uint64(bm.GetChain().BestSnapshot().GraphState.GetMainHeight()) },
		BestOrder:       func() uint64 { return bm.GetChain().BestSnapshot().Order },
		CalcSequenceLock: bm.GetChain().CalcSequenceLock,
		SubsidyCache:     bm.GetChain().FetchSubsidyCache(),
		SigCache:         sigCache,
		PastMedianTime:   func() time.Time { return bm.GetChain().BestSnapshot().MedianTime },
		AddrIndex:        addrIndex,
	}
	txMemPool := mempool.New(&txC)

	return &TxManager{bm,txIndex,addrIndex,txMemPool,ntmgr,db},nil
}