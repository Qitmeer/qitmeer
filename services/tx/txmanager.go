package tx

import (
	"github.com/Qitmeer/qitmeer/common/hash"
	"github.com/Qitmeer/qitmeer/config"
	"github.com/Qitmeer/qitmeer/core/blockchain"
	"github.com/Qitmeer/qitmeer/core/blockdag"
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
	return nil
}

func (tm *TxManager) Stop() error {
	log.Info("Stopping tx manager")
	return nil
}

func (tm *TxManager) MemPool() blockchain.TxPool {
	return tm.txMemPool
}

func (tm *TxManager) IsInvalidTx(txh *hash.Hash) bool {
	_, ok := tm.invalidTx[*txh]
	return ok
}

func (tm *TxManager) GetInvalidTxFromBlock(bh *hash.Hash) []*hash.Hash {
	result := []*hash.Hash{}
	for k, v := range tm.invalidTx {
		if v.Has(bh) {
			txHash := k
			result = append(result, &txHash)
		}
	}
	return result
}

func (tm *TxManager) AddInvalidTx(txh *hash.Hash, bh *hash.Hash) {
	if tm.IsInvalidTx(txh) {
		tm.invalidTx[*txh].Add(bh)
	} else {
		set := blockdag.NewHashSet()
		set.Add(bh)
		tm.invalidTx[*txh] = set
	}
}

func (tm *TxManager) AddInvalidTxArray(txha []*hash.Hash, bh *hash.Hash) {
	if len(txha) == 0 {
		return
	}
	for _, v := range txha {
		tm.AddInvalidTx(v, bh)
	}
}

func (tm *TxManager) RemoveInvalidTx(bh *hash.Hash) {
	for k, v := range tm.invalidTx {
		if v.Has(bh) {
			v.Remove(bh)
			if v.IsEmpty() {
				delete(tm.invalidTx, k)
			}
		}
	}
}

func NewTxManager(bm *blkmgr.BlockManager, txIndex *index.TxIndex,
	addrIndex *index.AddrIndex, cfg *config.Config, ntmgr notify.Notify,
	sigCache *txscript.SigCache, db database.DB) (*TxManager, error) {
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
	}
	txMemPool := mempool.New(&txC)
	invalidTx := make(map[hash.Hash]*blockdag.HashSet)
	return &TxManager{bm, txIndex, addrIndex, txMemPool, ntmgr, db, invalidTx}, nil
}
