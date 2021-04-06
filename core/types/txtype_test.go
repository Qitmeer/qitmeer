// Copyright (c) 2021 The qitmeer developers
// Use of this source code is governed by an ISC
// license that can be found in the LICENSE file.

package types

import (
	"testing"
)

func TestAllTxTypeValues(t *testing.T) {
	for _, test := range []struct {
		txType TxType
		want int
	}{
		{TxTypeRegular, 0},
		{TxTypeCoinbase, 1},
		{TxTypeGenesisLock, 2},
		{TxTypeStakebase, 16},
		{TyTypeStakeReserve,17},
		{TxTypeStakePurchase,18},
		{TxTypeStakeDispose,19},
		{TxTypeTokenRegulation,128},
		{TxTypeTokenNew,129},
		{TxTypeTokenRenew,130},
		{TxTypeTokenValidate,131},
		{TxTypeTokenInvalidate ,132},
		{TxTypeTokenRevoke,143},
		{TxTypeTokenbase, 144},
		{TxTypeTokenMint,145},
		{TxTypeTokenUnmint,146},
	}{
		if test.txType != TxType(test.want) {
			t.Errorf("want %v but got %v", test.want, test.txType)
		}
	}
}
