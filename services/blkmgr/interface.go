// Copyright (c) 2020 The qitmeer developers
// Use of this source code is governed by an ISC
// license that can be found in the LICENSE file.

package blkmgr

import (
	"github.com/Qitmeer/qitmeer/common/hash"
	"github.com/Qitmeer/qitmeer/core/blockchain"
	"github.com/Qitmeer/qitmeer/core/types"
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

	ProcessOrphans(hash *hash.Hash) []*types.TxDesc

	MaybeAcceptTransaction(tx *types.Tx, isNew, rateLimit bool) ([]*hash.Hash, error)

	HaveTransaction(hash *hash.Hash) bool

	PruneExpiredTx()

	ProcessTransaction(tx *types.Tx, allowOrphan, rateLimit, allowHighFees bool) ([]*types.TxDesc, error)
}
