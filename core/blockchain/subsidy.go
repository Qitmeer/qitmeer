// Copyright (c) 2017-2018 The nox developers
// Copyright (c) 2013-2015 The btcsuite developers
// Copyright (c) 2015-2018 The Decred developers
// Use of this source code is governed by an ISC
// license that can be found in the LICENSE file.

// TODO decoupling subsidy from bm, might move to protocol/params

package blockchain

import (
	"sync"

	"qitmeer/params"
	"qitmeer/core/types"
)

// The number of values to precalculate on initialization of the subsidy
// cache.
const subsidyCacheInitWidth = 4

// SubsidyCache is a structure that caches calculated values of subsidy so that
// they're not constantly recalculated. The blockchain struct itself possesses a
// pointer to a preinitialized SubsidyCache.
type SubsidyCache struct {
	subsidyCache     map[uint64]int64
	subsidyCacheLock sync.RWMutex

	params           *params.Params
}

// NewSubsidyCache initializes a new subsidy cache for a given height. It
// precalculates the values of the subsidy that are most likely to be seen by
// the client when it connects to the network.
func NewSubsidyCache(height int64, params *params.Params) *SubsidyCache {
	scm := make(map[uint64]int64)
	sc := SubsidyCache{
		subsidyCache: scm,
		params:       params,
	}

	iteration := uint64(height / params.SubsidyReductionInterval)
	if iteration < subsidyCacheInitWidth {
		return &sc
	}

	for i := iteration - 4; i <= iteration; i++ {
		sc.CalcBlockSubsidy(int64(iteration) * params.SubsidyReductionInterval)
	}

	return &sc
}

// CalcBlockSubsidy returns the subsidy amount a block at the provided height
// should have. This is mainly used for determining how much the coinbase for
// newly generated blocks awards as well as validating the coinbase for blocks
// has the expected value.
//
// Subsidy calculation for exponential reductions:
// 0 for i in range (0, height / SubsidyReductionInterval):
// 1     subsidy *= MulSubsidy
// 2     subsidy /= DivSubsidy
//
// Safe for concurrent access.
func (s *SubsidyCache) CalcBlockSubsidy(height int64) int64 {
	// TODO, impl ICO using block-one subsidy
	// Block height 1 subsidy is 'special' and used to
	// distribute initial tokens, if any.
	/*
	if height == 1 {
		return s.params.BlockOneSubsidy()

	}
	*/

	iteration := uint64(height / s.params.SubsidyReductionInterval)

	if iteration == 0 {
		return s.params.BaseSubsidy
	}

	// First, check the cache.
	s.subsidyCacheLock.RLock()
	cachedValue, existsInCache := s.subsidyCache[iteration]
	s.subsidyCacheLock.RUnlock()
	if existsInCache {
		return cachedValue
	}

	// Is the previous one in the cache? If so, calculate
	// the subsidy from the previous known value and store
	// it in the database and the cache.
	s.subsidyCacheLock.RLock()
	cachedValue, existsInCache = s.subsidyCache[iteration-1]
	s.subsidyCacheLock.RUnlock()
	if existsInCache {
		cachedValue *= s.params.MulSubsidy
		cachedValue /= s.params.DivSubsidy

		s.subsidyCacheLock.Lock()
		s.subsidyCache[iteration] = cachedValue
		s.subsidyCacheLock.Unlock()

		return cachedValue
	}

	// Calculate the subsidy from scratch and store in the
	// cache. TODO If there's an older item in the cache,
	// calculate it from that to save time.
	subsidy := s.params.BaseSubsidy
	for i := uint64(0); i < iteration; i++ {
		subsidy *= s.params.MulSubsidy
		subsidy /= s.params.DivSubsidy
	}

	s.subsidyCacheLock.Lock()
	s.subsidyCache[iteration] = subsidy
	s.subsidyCacheLock.Unlock()

	return subsidy
}

// CalcBlockWorkSubsidy calculates the proof of work subsidy for a block as a
// proportion of the total subsidy. (aka, the coinbase subsidy)
// TODO refactor CalcBlockWorkSubsidy
func CalcBlockWorkSubsidy(subsidyCache *SubsidyCache, height int64, voters uint16, params *params.Params) uint64 {

	subsidy := subsidyCache.CalcBlockSubsidy(height)
	proportionWork := int64(params.WorkRewardProportion)
	proportions := int64(params.TotalSubsidyProportions())
	subsidy *= proportionWork
	subsidy /= proportions

	return uint64(subsidy) //TODO remove type conversion
}

// CalculateAddedSubsidy calculates the amount of subsidy added by a block
// and its parent. The blocks passed to this function MUST be valid blocks
// that have already been confirmed to abide by the consensus rules of the
// network, or the function might panic.
// TODO refactor CalculateAddedSubsidy
func CalculateAddedSubsidy(block *types.SerializedBlock) int64 {
	return 0
}

// CalcBlockTaxSubsidy calculates the subsidy for the organization address in the
// coinbase.
// TODO refactor CalcBlockTaxSubsidy
func CalcBlockTaxSubsidy(subsidyCache *SubsidyCache, height int64, voters uint16, params *params.Params) int64 {

	subsidy := subsidyCache.CalcBlockSubsidy(height)
	proportionTax := int64(params.BlockTaxProportion)
	proportions := int64(params.TotalSubsidyProportions())
	subsidy *= proportionTax
	subsidy /= proportions

	return subsidy
}

