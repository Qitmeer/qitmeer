// Copyright (c) 2020 The qitmeer developers
// Use of this source code is governed by an ISC
// license that can be found in the LICENSE file.

package blockchain

import (
	"fmt"
	"github.com/Qitmeer/qitmeer/common/hash"
	"github.com/Qitmeer/qitmeer/core/dbnamespace"
	"github.com/Qitmeer/qitmeer/core/json"
	"github.com/Qitmeer/qitmeer/core/types"
	"github.com/Qitmeer/qitmeer/database"
	"strings"
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

	cacheHash *hash.Hash
}

func (bu *balanceUpdate) Serialize() ([]byte, error) {
	if bu.typ != tokenMint && bu.typ != tokenUnMint {
		return nil, fmt.Errorf("invalid token balance update type %v", bu.typ)
	}
	if bu.meerAmount < 0 || bu.tokenAmount.Value < 0 || !types.IsKnownCoinID(bu.tokenAmount.Id) {
		return nil, fmt.Errorf("invalid token balance update %v", bu)
	}
	serializeSize := serializeSizeVLQ(uint64(bu.typ))
	serializeSize += serializeSizeVLQ(uint64(bu.meerAmount))
	serializeSize += serializeSizeVLQ(uint64(bu.tokenAmount.Id))
	serializeSize += serializeSizeVLQ(uint64(bu.tokenAmount.Value))

	serialized := make([]byte, serializeSize)
	offset := 0

	offset += putVLQ(serialized[offset:], uint64(bu.typ))
	offset += putVLQ(serialized[offset:], uint64(bu.meerAmount))
	offset += putVLQ(serialized[offset:], uint64(bu.tokenAmount.Id))
	offset += putVLQ(serialized[offset:], uint64(bu.tokenAmount.Value))
	return serialized, nil
}

func (bu *balanceUpdate) Hash() *hash.Hash {
	if bu.cacheHash != nil {
		return bu.cacheHash
	}
	return bu.CacheHash()
}

func (bu *balanceUpdate) CacheHash() *hash.Hash {
	bu.cacheHash = nil
	bs, err := bu.Serialize()
	if err != nil {
		log.Error(err.Error())
		return bu.cacheHash
	}
	h := hash.DoubleHashH(bs)
	bu.cacheHash = &h
	return bu.cacheHash
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
	prevStateID uint32
	balances    tokenBalances
	updates     []balanceUpdate
}

func (ts *tokenState) GetTokenBalances() []json.TokenBalance {
	tbs := []json.TokenBalance{}
	for k, v := range ts.balances {
		tb := json.TokenBalance{CoinId: uint16(k), CoinName: k.Name(), Balance: v.balance, LockedMeer: v.lockedMeer}
		tbs = append(tbs, tb)
	}
	return tbs
}

type tokenBalances map[types.CoinID]tokenBalance

func (tbs *tokenBalances) UpdateBalance(update *balanceUpdate) error {
	tokenId := update.tokenAmount.Id
	tb := (*tbs)[tokenId]
	switch update.typ {
	case tokenMint:
		tb.balance += update.tokenAmount.Value
		tb.lockedMeer += update.meerAmount
	case tokenUnMint:
		if tb.balance-update.tokenAmount.Value < 0 {
			return fmt.Errorf("can't unmint token %v more than token balance %v", update.tokenAmount, tb)
		}
		tb.balance -= update.tokenAmount.Value
		if tb.lockedMeer-update.meerAmount < 0 {
			return fmt.Errorf("can't unlock %v meer more than locked meer %v", update.meerAmount, tb)
		}
		tb.lockedMeer -= update.meerAmount
	default:
		return fmt.Errorf("unknown balance update type %v", update.typ)
	}
	(*tbs)[tokenId] = tb
	return nil
}

func (tbs *tokenBalances) UpdatesBalance(updates []balanceUpdate) error {
	for _, update := range updates {
		err := tbs.UpdateBalance(&update)
		if err != nil {
			return err
		}
	}
	return nil
}

func (tb *tokenBalances) String() string {
	var b strings.Builder
	fmt.Fprintf(&b, "[")
	for k, v := range *tb {
		b.WriteString(fmt.Sprintf("%v:{balance:%v,locked-meer:%v},", k.Name(), v.balance, v.lockedMeer))
	}
	fmt.Fprintf(&b, "]")
	return b.String()
}
func (tb *tokenBalances) Copy() *tokenBalances {
	newTb := tokenBalances{}
	for k, v := range *tb {
		newTb[k] = v
	}
	return &newTb
}

// serializeTokeState function will serialize the token state into byte slice
func serializeTokeState(ts tokenState) ([]byte, error) {
	// total number of bytes to serialize
	serializeSize := serializeSizeVLQ(uint64(ts.prevStateID))

	serializeSize += serializeSizeVLQ(uint64(len(ts.balances)))
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
	offset = putVLQ(serialized, uint64(ts.prevStateID))

	offset += putVLQ(serialized[offset:], uint64(len(ts.balances)))
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
	prevStateID, offset := deserializeVLQ(data)
	if offset == 0 {
		return nil, errDeserialize("unexpected end of data while reading prevStateID")
	}
	// Deserialize the balance.
	var balances map[types.CoinID]tokenBalance
	numOfBalances, bytesRead := deserializeVLQ(data[offset:])
	if bytesRead == 0 {
		return nil, fmt.Errorf("unexpected end of data while reading number of balances")
	}
	offset += bytesRead

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
				tokenAmount: types.Amount{Value: int64(tokenAmount), Id: types.CoinID(uint16(tokenId))},
			}
		}
	}
	return &tokenState{prevStateID: uint32(prevStateID), balances: balances, updates: updates}, nil
}

// dbPutTokenState put a token balance record into the token state database.
// the key is the provided block hash
func dbPutTokenState(dbTx database.Tx, bid uint32, ts tokenState) error {
	// Serialize the current token state.
	serializedData, err := serializeTokeState(ts)
	if err != nil {
		return err
	}
	// Store the current token balance record into the token state database.
	meta := dbTx.Metadata()
	bucket := meta.Bucket(dbnamespace.TokenBucketName)
	var serializedID [4]byte
	dbnamespace.ByteOrder.PutUint32(serializedID[:], bid)
	return bucket.Put(serializedID[:], serializedData)
}

// dbFetchTokenState fetch the token balance record from the token state database.
// the key is the input block hash.
func dbFetchTokenState(dbTx database.Tx, bid uint32) (*tokenState, error) {
	// if it is genesis hash, return empty tokenState directly
	if bid == 0 {
		return &tokenState{}, nil
	}
	// Fetch record from the token state database by block hash
	meta := dbTx.Metadata()
	bucket := meta.Bucket(dbnamespace.TokenBucketName)

	var serializedID [4]byte
	dbnamespace.ByteOrder.PutUint32(serializedID[:], bid)
	v := bucket.Get(serializedID[:])
	if v == nil {
		return nil, fmt.Errorf("tokenstate db can't find record from block id : %v", bid)
	}
	// deserialize the fetched token state record
	return deserializeTokenState(v)
}

func dbRemoveTokenState(dbTx database.Tx, id uint32) error {
	bucket := dbTx.Metadata().Bucket(dbnamespace.TokenBucketName)
	var serializedID [4]byte
	dbnamespace.ByteOrder.PutUint32(serializedID[:], id)

	key := serializedID[:]
	return bucket.Delete(key)
}

func checkUnMintUpdate(update *balanceUpdate) error {
	if update.typ != tokenUnMint {
		return fmt.Errorf("checkUnMintUpdate : wrong update type %v", update.typ)
	}
	if err := checkUpdateCommon(update); err != nil {
		return err
	}
	return nil
}

func checkMintUpdate(update *balanceUpdate) error {
	if update.typ != tokenMint {
		return fmt.Errorf("checkUnMintUpdate : wrong update type %v", update.typ)
	}
	if err := checkUpdateCommon(update); err != nil {
		return err
	}
	return nil
}

func checkUpdateCommon(update *balanceUpdate) error {
	if !types.IsKnownCoinID(update.tokenAmount.Id) {
		return fmt.Errorf("checkUpdateCommon : unknown token id %v", update.tokenAmount.Id.Name())
	}
	if update.tokenAmount.Value <= 0 {
		return fmt.Errorf("checkUpdateCommon : wrong token amount : %v", update.tokenAmount.Value)
	}
	if update.meerAmount <= 0 {
		return fmt.Errorf("checkUpdateCommon : wrong meer amount : %v", update.meerAmount)
	}
	return nil
}
