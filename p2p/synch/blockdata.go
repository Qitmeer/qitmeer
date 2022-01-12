/*
 * Copyright (c) 2017-2020 The qitmeer developers
 */

package synch

import (
	"github.com/Qitmeer/qitmeer/common/hash"
	"github.com/Qitmeer/qitmeer/core/types"
)

type BlockData struct {
	Hash *hash.Hash
	Block *types.SerializedBlock
}