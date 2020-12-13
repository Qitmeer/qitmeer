// Copyright (c) 2020 The qitmeer developers
// Use of this source code is governed by an ISC
// license that can be found in the LICENSE file.

package blockchain

import (
	"fmt"
	"github.com/Qitmeer/qitmeer/core/types"
)

// balanceUpdateType specifies the possible types of updates that might
// change the token balance
type balanceUpdateType byte

// The following constants define the known type of balanceUpdateType
const (
	tokenMint   balanceUpdateType = 0x01
	tokenUnMint balanceUpdateType = 0x02
)

// balanceUpdate specifies the type and update record of the values that change a token
// balance.
// for TOKON_MINT, the values should add on the meerlock and token balance
// for TOKEN_UNMINT, the values should subtract from the meerlock and token balance
type balanceUpdate struct {
	typ         balanceUpdateType
	meerAmount  int64
	tokenAmount types.Amount
}

// tokenBalance specifies the token balance and the locked meer amount
type tokenBalance struct {
	balance    int64
	lockedMeer int64
}

// tokenState specifies the token balance of the current block.
// the updates are written in the same order as the tx in the block, which is
// used to verify the correctness of the token balance
type tokenState struct {
	balances map[types.CoinID]tokenBalance
	updates  []balanceUpdate
}

// serializeTokeState function will serialize the token state into byte slice
func serializeTokeState(ts tokenState) ([]byte, error) {
	// total number of bytes to serialize
	serializeSize := serializeSizeVLQ(uint64(len(ts.balances)))
	for id, b := range ts.balances {
		// sanity check
		if id == types.MEERID || b.balance < 0 || b.lockedMeer < 0 {
			return nil, fmt.Errorf("invalid token balance {%v, %v}", id, b)
		}
		serializeSize += serializeSizeVLQ(uint64(id))
		serializeSize += serializeSizeVLQ(uint64(b.balance))
		serializeSize += serializeSizeVLQ(uint64(b.lockedMeer))
	}
	serializeSize += serializeSizeVLQ(uint64(len(ts.updates)))
	for _, v := range ts.updates {
		if v.typ != tokenMint && v.typ != tokenUnMint {
			return nil, fmt.Errorf("invalid token balance update type %v", v.typ)
		}
		if v.meerAmount < 0 || v.tokenAmount.Value < 0 || !types.IsKnownCoinID(v.tokenAmount.Id) {
			return nil, fmt.Errorf("invalid token balance update %v", v)
		}
		serializeSize += 1 // balanceUpdateType takes 1 byte
		serializeSize += serializeSizeVLQ(uint64(v.meerAmount))
		serializeSize += serializeSizeVLQ(uint64(v.tokenAmount.Id))
		serializeSize += serializeSizeVLQ(uint64(v.tokenAmount.Value))
	}
	serialized := make([]byte, serializeSize)
	offset := 0
	offset = putVLQ(serialized, uint64(len(ts.balances)))
	for id, b := range ts.balances {
		offset += putVLQ(serialized[offset:], uint64(id))
		offset += putVLQ(serialized[offset:], uint64(b.balance))
		offset += putVLQ(serialized[offset:], uint64(b.lockedMeer))
	}

	offset += putVLQ(serialized[offset:], uint64(len(ts.updates)))
	for _, v := range ts.updates {
		offset += putVLQ(serialized[offset:], uint64(v.typ))
		offset += putVLQ(serialized[offset:], uint64(v.meerAmount))
		offset += putVLQ(serialized[offset:], uint64(v.tokenAmount.Id))
		offset += putVLQ(serialized[offset:], uint64(v.tokenAmount.Value))
	}
	return serialized, nil
}

// deserializeTokenState function will deserializes token state from the byte slice
func deserializeTokenState(data []byte) (*tokenState, error) {
	// Deserialize the balance.
	var balances map[types.CoinID]tokenBalance
	numOfBalances, offset := deserializeVLQ(data)
	if offset == 0 {
		return nil, errDeserialize("unexpected end of data while reading number of balances")
	}
	if numOfBalances > 0 {
		balances = make(map[types.CoinID]tokenBalance, numOfBalances)
		for i := uint64(0); i < numOfBalances; i++ {
			// token id
			derId, bytesRead := deserializeVLQ(data[offset:])
			if bytesRead == 0 {
				return nil, fmt.Errorf("unexpected end of data while reading token id at balances{%d}", i)
			}
			offset += bytesRead

			// token balance
			balance, bytesRead := deserializeVLQ(data[offset:])
			if bytesRead == 0 {
				return nil, fmt.Errorf("unexpected end of data while reading balance at balances{%d}", i)
			}
			offset += bytesRead

			// locked meer
			lockedMeer, bytesRead := deserializeVLQ(data[offset:])
			if bytesRead == 0 {
				return nil, fmt.Errorf("unexpected end of data while reading balance at balances{%d}", i)
			}
			offset += bytesRead

			id := types.CoinID(uint16(derId))
			balances[id] = tokenBalance{int64(balance), int64(lockedMeer)}
		}
	}
	updates := []balanceUpdate{}
	numOfUpdates, bytesRead := deserializeVLQ(data[offset:])
	if bytesRead == 0 {
		return nil, errDeserialize("unexpected end of data while reading number of balances")
	}
	offset += bytesRead
	if numOfUpdates > 0 {
		updates = make([]balanceUpdate, numOfUpdates)
		for i := uint64(0); i < numOfUpdates; i++ {
			//type
			updateType, bytesRead := deserializeVLQ(data[offset:])
			if bytesRead == 0 {
				return nil, fmt.Errorf("unexpected end of data while reading balance update type at update[%d]", i)
			}
			offset += bytesRead
			//meerAmount
			meerAmount, bytesRead := deserializeVLQ(data[offset:])
			if bytesRead == 0 {
				return nil, fmt.Errorf("unexpected end of data while reading meer amount at update[%d]", i)
			}
			offset += bytesRead
			//tokenId
			tokenId, bytesRead := deserializeVLQ(data[offset:])
			if bytesRead == 0 {
				return nil, fmt.Errorf("unexpected end of data while reading token id at update[%d]", i)
			}
			offset += bytesRead
			//tokenAmount
			tokenAmount, bytesRead := deserializeVLQ(data[offset:])
			if bytesRead == 0 {
				return nil, fmt.Errorf("unexpected end of data while reading token amount at update[%d]", i)
			}
			offset += bytesRead

			updates[i] = balanceUpdate{
				typ:         balanceUpdateType(updateType),
				meerAmount:  int64(meerAmount),
				tokenAmount: types.Amount{int64(tokenAmount), types.CoinID(uint16(tokenId))},
			}
		}
	}
	return &tokenState{balances: balances, updates: updates}, nil
}
