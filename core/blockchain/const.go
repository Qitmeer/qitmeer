// Copyright (c) 2017-2018 The nox developers
package blockchain

import (
	"github.com/HalalChain/qitmeer/common/hash"
	"github.com/HalalChain/qitmeer/core/types"
)

const (
	// MaxSigOpsPerBlock is the maximum number of signature operations
	// allowed for a block.  This really should be based upon the max
	// allowed block size for a network and any votes that might change it,
	// however, since it was not updated to be based upon it before
	// release, it will require a hard fork and associated vote agenda to
	// change it.  The original max block size for the protocol was 1MiB,
	// so that is what this is based on.
	MaxSigOpsPerBlock = 1000000 / 200

	// currentBlockIndexVersion indicates what the current block index
	// database version.
	currentBlockIndexVersion= 1

	// currentDatabaseVersion indicates what the current database
	// version is.
	currentDatabaseVersion = 1

	// blockHdrSize is the size of a block header.  This is simply the
	// constant from wire and is only provided here for convenience since
	// wire.MaxBlockHeaderPayload is quite long.
	blockHdrSize = types.MaxBlockHeaderPayload

	// medianTimeBlocks is the number of previous blocks which should be
	// used to calculate the median time used to validate block timestamps.
	medianTimeBlocks = 11

)

const (

	// MaxTimeOffsetSeconds is the maximum number of seconds a block time
	// is allowed to be ahead of the current time.  This is currently 2
	// hours.
	MaxTimeOffsetSeconds = 2 * 60 * 60

	// MinCoinbaseScriptLen is the minimum length a coinbase script can be.
	MinCoinbaseScriptLen = 2

	// MaxCoinbaseScriptLen is the maximum length a coinbase script can be.
	MaxCoinbaseScriptLen = 100

)

var (
	// zeroHash is the zero value for a hash.Hash and is defined as a
	// package level variable to avoid the need to create a new instance
	// every time a check is needed.
	zeroHash = &hash.ZeroHash
)



