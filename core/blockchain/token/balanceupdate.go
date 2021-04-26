package token

import (
	"fmt"
	"github.com/Qitmeer/qitmeer/common/hash"
	"github.com/Qitmeer/qitmeer/core/serialization"
	"github.com/Qitmeer/qitmeer/core/types"
)

// balanceUpdateType specifies the possible types of updates that might
// change the token balance
type BalanceUpdateType byte

// The following constants define the known type of balanceUpdateType
const (
	TokenMint   BalanceUpdateType = 0x01
	TokenUnMint BalanceUpdateType = 0x02
)

// balanceUpdate specifies the type and update record of the values that change a token
// balance.
// for TOKON_MINT, the values should add on the meerlock and token balance
// for TOKEN_UNMINT, the values should subtract from the meerlock and token balance
type BalanceUpdate struct {
	Typ         BalanceUpdateType
	MeerAmount  int64
	TokenAmount types.Amount

	cacheHash *hash.Hash
}

func (bu *BalanceUpdate) Serialize() ([]byte, error) {
	if bu.Typ != TokenMint && bu.Typ != TokenUnMint {
		return nil, fmt.Errorf("invalid token balance update type %v", bu.Typ)
	}
	if bu.MeerAmount < 0 || bu.TokenAmount.Value < 0 || !types.IsKnownCoinID(bu.TokenAmount.Id) {
		return nil, fmt.Errorf("invalid token balance update %v", bu)
	}
	serializeSize := serialization.SerializeSizeVLQ(uint64(bu.Typ))
	serializeSize += serialization.SerializeSizeVLQ(uint64(bu.MeerAmount))
	serializeSize += serialization.SerializeSizeVLQ(uint64(bu.TokenAmount.Id))
	serializeSize += serialization.SerializeSizeVLQ(uint64(bu.TokenAmount.Value))

	serialized := make([]byte, serializeSize)
	offset := 0

	offset += serialization.PutVLQ(serialized[offset:], uint64(bu.Typ))
	offset += serialization.PutVLQ(serialized[offset:], uint64(bu.MeerAmount))
	offset += serialization.PutVLQ(serialized[offset:], uint64(bu.TokenAmount.Id))
	offset += serialization.PutVLQ(serialized[offset:], uint64(bu.TokenAmount.Value))
	return serialized, nil
}

func (bu *BalanceUpdate) Hash() *hash.Hash {
	if bu.cacheHash != nil {
		return bu.cacheHash
	}
	return bu.CacheHash()
}

func (bu *BalanceUpdate) CacheHash() *hash.Hash {
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
