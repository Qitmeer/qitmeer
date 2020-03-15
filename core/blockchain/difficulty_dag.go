// Copyright (c) 2017-2018 The qitmeer developers
// Copyright (c) 2013-2016 The btcsuite developers
// Copyright (c) 2015-2018 The Decred developers
// Use of this source code is governed by an ISC
// license that can be found in the LICENSE file.

package blockchain

import (
	"github.com/Qitmeer/qitmeer/core/types"
	"github.com/Qitmeer/qitmeer/core/types/pow"
	"github.com/Qitmeer/qitmeer/params"
	"math/big"
)

// find block node by pow type
func (b *BlockChain) GetDagDiff(curNode *blockNode, powInstance pow.IPow, curDiff *big.Int) (uint32, error) {
	if !b.needAjustPowDifficulty(curNode, powInstance.GetPowType(), b.params.DagDiffAdjustmentConfig.WorkDiffWindowSize) {
		return pow.BigToCompact(curDiff), nil
	}
	lastPastBlockCount, err := b.bd.GetBlockConcurrency(curNode.GetHash())
	if err != nil {
		return 0, err
	}
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
	firstPastBlockCount, err := b.bd.GetBlockConcurrency(curNode.GetHash())
	if err != nil {
		return 0, err
	}
	return CalcDagDiff(powInstance, curDiff, uint(lastPastBlockCount), uint(firstPastBlockCount), allBlockSize, b.params)
}

// find block node by pow type
func CalcDagDiff(powInstance pow.IPow, curDiff *big.Int,
	lastPastBlockCount, firstPastBlockCount uint, allBlockSize int, p *params.Params) (uint32, error) {
	// adjust max value
	oldDiffMax := big.NewInt(0).Add(big.NewInt(0), curDiff)
	// adjust min value
	oldDiffMin := big.NewInt(0).Add(big.NewInt(0), curDiff)
	//get all blocks between b.params.DagDiffAdjustmentConfig.MaxConcurrencyCount blocks
	allBlockDagCount := int64(lastPastBlockCount - firstPastBlockCount)
	targetAllowMaxBlockCount := p.DagDiffAdjustmentConfig.WorkDiffWindowSize * p.DagDiffAdjustmentConfig.MaxConcurrencyCount
	if allBlockDagCount == targetAllowMaxBlockCount {
		return pow.BigToCompact(curDiff), nil
	}
	needChangeWeight := big.NewInt(0)
	// fault-tolerant
	allAllowMaxBLockSize := p.DagDiffAdjustmentConfig.MaxConcurrencyCount * p.DagDiffAdjustmentConfig.WorkDiffWindowSize * (types.MaxBlockPayload - p.DagDiffAdjustmentConfig.FaultTolerantBlockSize)
	if allBlockDagCount < targetAllowMaxBlockCount {
		if int64(allBlockSize) < allAllowMaxBLockSize {
			return pow.BigToCompact(curDiff), nil
		}
	}
	targetAllowMaxBlockCountWei := big.NewInt(targetAllowMaxBlockCount).Lsh(big.NewInt(targetAllowMaxBlockCount), 32)
	needChangeWeight = needChangeWeight.Add(needChangeWeight, targetAllowMaxBlockCountWei.Div(targetAllowMaxBlockCountWei, big.NewInt(allBlockDagCount)))
	percent := big.NewInt(1).Lsh(big.NewInt(1), 32)
	curDiff = powInstance.GetNextDiffBig(needChangeWeight, curDiff, percent)
	maxDiff := oldDiffMax.Mul(oldDiffMax, big.NewInt(p.DagDiffAdjustmentConfig.RetargetAdjustmentFactor))
	minDiff := oldDiffMin.Div(oldDiffMin, big.NewInt(p.DagDiffAdjustmentConfig.RetargetAdjustmentFactor))
	if curDiff.Cmp(maxDiff) > 0 {
		curDiff.Set(maxDiff)
	}
	if curDiff.Cmp(minDiff) < 0 {
		curDiff.Set(minDiff)
	}
	// Limit new value to the proof of work limit.
	curDiff = powInstance.GetSafeDiff(curDiff.Uint64())
	return pow.BigToCompact(curDiff), nil
}
