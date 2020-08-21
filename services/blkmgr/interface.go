/*
 * Copyright (c) 2020.
 * Project:qitmeer
 * File:interface.go
 * Date:8/21/20 3:47 PM
 * Author:Jin
 * Email:lochjin@gmail.com
 */

package blkmgr

import (
	"github.com/Qitmeer/qitmeer/common/hash"
	"github.com/Qitmeer/qitmeer/core/blockchain"
	"github.com/Qitmeer/qitmeer/core/types"
	"github.com/Qitmeer/qitmeer/services/mempool"
)

type TxManager interface {
	MemPool() TxPool
}

type TxPool interface {
	AddTransaction(utxoView *blockchain.UtxoViewpoint,
		tx *types.Tx, height uint64, fee int64)

	RemoveTransaction(tx *types.Tx, removeRedeemers bool)

	RemoveDoubleSpends(tx *types.Tx)

	RemoveOrphan(txHash *hash.Hash)

	ProcessOrphans(hash *hash.Hash) []*mempool.TxDesc

	MaybeAcceptTransaction(tx *types.Tx, isNew, rateLimit bool) ([]*hash.Hash, error)

	HaveTransaction(hash *hash.Hash) bool

	PruneExpiredTx()

	ProcessTransaction(tx *types.Tx, allowOrphan, rateLimit, allowHighFees bool) ([]*mempool.TxDesc, error)
}
