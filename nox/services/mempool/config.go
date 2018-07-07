// Copyright (c) 2017-2018 The nox developers
// Copyright (c) 2013-2016 The btcsuite developers
// Copyright (c) 2017-2018 The Decred developers
// Use of this source code is governed by an ISC
// license that can be found in the LICENSE file.
package mempool

import (
	"time"
	"github.com/noxproject/nox/params"
	"github.com/noxproject/nox/core/types"
	"github.com/noxproject/nox/common/hash"
	"github.com/noxproject/nox/engine/txscript"
	"github.com/noxproject/nox/core/blockchain"
)

const (
	DefaultMaxOrphanTxSize       = 5000
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
	FetchUtxoView func(*types.Tx, bool) (*blockchain.UtxoViewpoint, error)

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

	// SigCache defines a signature cache to use.
	SigCache *txscript.SigCache

}


