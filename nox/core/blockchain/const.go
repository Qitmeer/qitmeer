// Copyright (c) 2017-2018 The nox developers
package blockchain

import "github.com/noxproject/nox/core/types"

const (
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



