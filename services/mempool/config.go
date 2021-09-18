// Copyright (c) 2017-2018 The qitmeer developers
// Copyright (c) 2013-2016 The btcsuite developers
// Copyright (c) 2017-2018 The Decred developers
// Use of this source code is governed by an ISC
// license that can be found in the LICENSE file.
package mempool

import (
	"github.com/Qitmeer/qitmeer/common/hash"
	"github.com/Qitmeer/qitmeer/core/blockchain"
	"github.com/Qitmeer/qitmeer/core/blockdag"
	"github.com/Qitmeer/qitmeer/core/event"
	"github.com/Qitmeer/qitmeer/core/types"
	"github.com/Qitmeer/qitmeer/engine/txscript"
	"github.com/Qitmeer/qitmeer/params"
	"github.com/Qitmeer/qitmeer/services/index"
	"time"
)

const (

	//TODO, refactor config item
	DefaultMaxOrphanTxSize = 5000
)

// Config is a descriptor containing the memory pool configuration.
type Config struct {
	// Policy defines the various mempool configuration options related
	// to policy.
	Policy Policy

	// ChainParams identifies which chain parameters the txpool is
	// associated with.
	ChainParams *params.Params

	// FetchUtxoView defines the function to use to fetch unspent
	// transaction output information.
	FetchUtxoView func(*types.Tx) (*blockchain.UtxoViewpoint, error)

	// BlockByHash defines the function use to fetch the block identified
	// by the given hash.
	BlockByHash func(*hash.Hash) (*types.SerializedBlock, error)

	// BestHash defines the function to use to access the block hash of
	// the current best chain.
	BestHash func() *hash.Hash

	// BestHeight defines the function to use to access the block height of
	// the current best chain.
	BestHeight func() uint64

	// PastMedianTime defines the function to use in order to access the
	// median time calculated from the point-of-view of the current chain
	// tip within the best chain.
	PastMedianTime func() time.Time

	// CalcSequenceLock defines the function to use in order to generate
	// the current sequence lock for the given transaction using the passed
	// utxo view.
	CalcSequenceLock func(*types.Tx, *blockchain.UtxoViewpoint) (*blockchain.SequenceLock, error)

	// SubsidyCache defines a subsidy cache to use.
	SubsidyCache *blockchain.SubsidyCache

	// SigCache defines a signature cache to use.
	SigCache *txscript.SigCache

	// AddrIndex defines the optional address index instance to use for
	// indexing the unconfirmed transactions in the memory pool.
	// This can be nil if the address index is not enabled.
	AddrIndex *index.AddrIndex

	// ExistsAddrIndex defines the optional exists address index instance
	// to use for indexing the unconfirmed transactions in the memory pool.
	// This can be nil if the address index is not enabled.
	ExistsAddrIndex *index.ExistsAddrIndex

	// block dag
	BD *blockdag.BlockDAG

	// block chain
	BC *blockchain.BlockChain

	// Data Directory
	DataDir string

	// mempool expiry
	Expiry time.Duration

	// persist mempool
	Persist bool

	//  no mempool bar
	NoMempoolBar bool

	Events *event.Feed
}
