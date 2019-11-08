// Copyright (c) 2013-2017 The btcsuite developers
// Use of this source code is governed by an ISC
// license that can be found in the LICENSE file.

package types

const (
	// MaxBlockWeight defines the maximum block weight
	MaxBlockWeight = 4000000

	// MaxBlockSigOpsCost is the maximum number of signature operations
	// allowed for a block. It is calculated via a weighted algorithm which
	// weights segregated witness sig ops lower than regular sig ops.
	MaxBlockSigOpsCost = 80000
)

// GetBlockWeight computes the value of the weight metric for a given block.
func GetBlockWeight(blk *Block) int {
	return blk.SerializeSize()
}
