// Copyright (c) 2017-2018 The qitmeer developers
// Copyright (c) 2013-2016 The btcsuite developers
// Copyright (c) 2015-2018 The Decred developers
// Use of this source code is governed by an ISC
// license that can be found in the LICENSE file.

package blockchain

import (
	"github.com/Qitmeer/qitmeer/core/types"
	"github.com/Qitmeer/qitmeer/core/types/pow"
	"math/big"
)

// find block node by pow type
func (b *BlockChain) GetDagDiff(curNode *blockNode, powInstance pow.IPow, curDiff *big.Int, changed bool) (uint32, error) {
	if !b.needAjustPowDifficulty(curNode, powInstance.GetPowType(), b.params.DagDiffAdjustmentConfig.WorkDiffWindowSize) {
		return pow.BigToCompact(curDiff), nil
	}
	oldDiffMax := big.NewInt(0).Add(big.NewInt(0), curDiff)
	oldDiffMin := big.NewInt(0).Add(big.NewInt(0), curDiff)
	curBlock := b.bd.GetBlock(curNode.GetHash())
	if curBlock == nil {
		return pow.BigToCompact(curDiff), nil
	}
	lastPastBlockCount := b.bd.GetMainParentConcurrency(curBlock)
	i := int64(0)
	allBlockSize := 0
	for {
		if curNode.parents == nil || i >= b.params.DagDiffAdjustmentConfig.WorkDiffWindowSize {
			break
		}
		sblock, err := b.FetchBlockByHash(curNode.GetHash())
		if err != nil {
			return pow.BigToCompact(curDiff), err
		}
		allBlockSize += sblock.Block().SerializeSize()
		oldBlock := b.bd.GetBlock(curNode.GetHash())

		oldMainParent := b.bd.GetBlock(oldBlock.GetMainParent())
		if oldMainParent != nil {
			curNode = b.index.LookupNode(oldMainParent.GetHash())
			curNode = b.getPowTypeNode(curNode, powInstance.GetPowType())
		}
		i++
	}
	if i < b.params.DagDiffAdjustmentConfig.WorkDiffWindowSize {
		return pow.BigToCompact(curDiff), nil
	}
	curBlock = b.bd.GetBlock(curNode.GetHash())
	if curBlock == nil {
		return pow.BigToCompact(curDiff), nil
	}
	firstPastBlockCount := b.bd.GetMainParentConcurrency(curBlock)
	allBlockDagCount := int64(lastPastBlockCount - firstPastBlockCount)
	targetAllowMaxBlockCount := b.params.DagDiffAdjustmentConfig.WorkDiffWindowSize * b.params.DagDiffAdjustmentConfig.MaxConcurrencyCount
	if allBlockDagCount == targetAllowMaxBlockCount {
		return pow.BigToCompact(curDiff), nil
	}
	baseTarget := powInstance.GetSafeDiff(0)
	// fault-tolerant
	allAllowMaxBLockSize := b.params.DagDiffAdjustmentConfig.MaxConcurrencyCount * (types.MaxBlockPayload - b.params.DagDiffAdjustmentConfig.FaultTolerantBlockSize)
	if allBlockDagCount < targetAllowMaxBlockCount {
		if int64(allBlockSize) < allAllowMaxBLockSize {
			return pow.BigToCompact(curDiff), nil
		}
		curDiff = curDiff.Mul(curDiff, big.NewInt(allBlockDagCount))
		curDiff = curDiff.Div(curDiff, big.NewInt(targetAllowMaxBlockCount))
		// Limit new value to the proof of work limit.
		if powInstance.CompareDiff(curDiff, baseTarget) {
			curDiff.Set(baseTarget)
		}

		return pow.BigToCompact(curDiff), nil
	}
	curDiff = curDiff.Mul(curDiff, big.NewInt(targetAllowMaxBlockCount))
	curDiff = curDiff.Div(curDiff, big.NewInt(allBlockDagCount))
	maxDiff := oldDiffMax.Mul(oldDiffMax, big.NewInt(b.params.DagDiffAdjustmentConfig.RetargetAdjustmentFactor))
	minDiff := oldDiffMin.Div(oldDiffMin, big.NewInt(b.params.DagDiffAdjustmentConfig.RetargetAdjustmentFactor))
	if changed {
		if curDiff.Cmp(maxDiff) > 0 {
			curDiff.Set(maxDiff)
		}
		if curDiff.Cmp(minDiff) < 0 {
			curDiff.Set(minDiff)
		}
	}
	return pow.BigToCompact(curDiff), nil
}
