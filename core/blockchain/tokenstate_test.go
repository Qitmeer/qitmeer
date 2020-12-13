// Copyright (c) 2020 The qitmeer developers
// Use of this source code is governed by an ISC
// license that can be found in the LICENSE file.

package blockchain

import (
	"bytes"
	"encoding/hex"
	"github.com/Qitmeer/qitmeer/core/types"
	"reflect"
	"testing"
)

func TestTokeStateSerialization(t *testing.T) {
	tests := []struct {
		name  string
		state *tokenState
		bytes []byte
		ok    bool
	}{
		{name: "test1",
			state: &tokenState{
				balances: map[types.CoinID]tokenBalance{
					types.QITID: tokenBalance{200 * 1e8, 100 * 1e8}},
				updates: []balanceUpdate{
					{typ: tokenMint,
						meerAmount:  100 * 1e8,
						tokenAmount: types.Amount{200 * 1e8, types.QITID}},
				},
			},
			bytes: bytesFromStr("0101c9bfde8f00a49faec7000101a49faec70001c9bfde8f00"),
			ok:    true,
		},
	}
	for i, test := range tests {
		serialized, err := serializeTokeState(*test.state)
		if !bytes.Equal(serialized, test.bytes) {
			t.Errorf("test[%d][%s] failed: want %x but got %x : %w", i, test.name, test.bytes, serialized, err)
		}
		deserialized, err := deserializeTokenState(test.bytes)
		if !reflect.DeepEqual(deserialized, test.state) {
			t.Errorf("test[%d][%s] failed: want %v but got %v", i, test.name, test.state, deserialized)
		}
	}

}

func bytesFromStr(s string) []byte {
	b, _ := hex.DecodeString(s)
	return b
}
