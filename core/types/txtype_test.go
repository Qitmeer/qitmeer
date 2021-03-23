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
		{TxTypeStakebase, 16},
		{TxTypeTokenbase, 128},
		{TyTypeStakeReserve,17},
		{TxTypeStakePurchase,18},
		{TxTypeStakeDispose,19},
		{TxTypeTokenMint,129},
		{TxTypeTokenUnmint,130},
	}{
		if test.txType != TxType(test.want) {
			t.Errorf("want %v but got %v", test.want, test.txType)
		}
	}
}
