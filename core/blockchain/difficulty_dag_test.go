// Copyright (c) 2017-2018 The qitmeer developers
// Copyright (c) 2013-2016 The btcsuite developers
// Copyright (c) 2015-2018 The Decred developers
// Use of this source code is governed by an ISC
// license that can be found in the LICENSE file.

package blockchain

import (
	"github.com/Qitmeer/qitmeer/core/types/pow"
	"github.com/Qitmeer/qitmeer/params"
	"github.com/stretchr/testify/assert"
	"math/big"
	"testing"
)

type DiffDagTestData struct {
	CurrentCalcDiff           int64          //main chain calc diff by time
	FirstPastBlueBlockCount   uint           // first dot has blue block count
	CurrentPastBlueBlockCount uint           // current dot has blue block count
	AllBlockSizeInThisPeriod  int            // all block size in this period
	Param                     *params.Params // params
	TargetDiff                int64          // params
}

var TestData = []DiffDagTestData{
	{
		CurrentCalcDiff:           1000,
		FirstPastBlueBlockCount:   100,
		CurrentPastBlueBlockCount: 260,
		AllBlockSizeInThisPeriod:  1000000,
		Param:                     &params.TestNetParams,
		TargetDiff:                1111,
	},
	{
		CurrentCalcDiff:           1000,
		FirstPastBlueBlockCount:   100,
		CurrentPastBlueBlockCount: 200,
		AllBlockSizeInThisPeriod:  1000000,
		Param:                     &params.TestNetParams,
		TargetDiff:                1000,
	},
	{
		CurrentCalcDiff:           1000,
		FirstPastBlueBlockCount:   100,
		CurrentPastBlueBlockCount: 200,
		AllBlockSizeInThisPeriod:  1000000000,
		Param:                     &params.TestNetParams,
		TargetDiff:                694,
	},
}

func TestDagDiffAdjustment(t *testing.T) {
	for _, v := range TestData {
		powInstance := pow.GetInstance(pow.CUCKAROO, 0, []byte{})
		powInstance.SetParams(v.Param.PowConfig)
		currentDiff := big.NewInt(v.CurrentCalcDiff)
		resultCompact, err := CalcDagDiff(powInstance, currentDiff, v.CurrentPastBlueBlockCount,
			v.FirstPastBlueBlockCount,
			v.AllBlockSizeInThisPeriod, v.Param)
		if err != nil {
			t.Fatal(err)
			return
		}
		result := pow.CompactToBig(resultCompact)
		assert.Equal(t, v.TargetDiff, result.Int64())
	}
}
