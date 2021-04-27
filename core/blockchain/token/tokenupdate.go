package token

import (
	"fmt"
	"github.com/Qitmeer/qitmeer/common/hash"
	"github.com/Qitmeer/qitmeer/core/serialization"
	"github.com/Qitmeer/qitmeer/core/types"
)

type ITokenUpdate interface {
	GetType() types.TxType
	GetHash() *hash.Hash
	Serialize() ([]byte, error)
	Deserialize(data []byte) (int, error)
	CheckSanity() error
}

type TokenUpdate struct {
	Typ types.TxType
}

func (tu *TokenUpdate) GetType() types.TxType {
	return tu.Typ
}
func (tu *TokenUpdate) Serialize() ([]byte, error) {
	serializeSize := serialization.SerializeSizeVLQ(uint64(tu.Typ))
	serialized := make([]byte, serializeSize)
	serialization.PutVLQ(serialized[0:], uint64(tu.Typ))
	return serialized, nil
}

func (tu *TokenUpdate) Deserialize(data []byte) (int, error) {
	typ, bytesRead := serialization.DeserializeVLQ(data[0:])
	if bytesRead == 0 {
		return bytesRead, fmt.Errorf("unexpected end of data while reading token update type at update")
	}
	tu.Typ = types.TxType(typ)
	return bytesRead, nil
}

func NewTokenUpdate(typ types.TxType) ITokenUpdate {
	switch typ {
	case types.TxTypeTokenMint, types.TxTypeTokenUnmint:
		return &BalanceUpdate{}
	case types.TxTypeTokenNew, types.TxTypeTokenRenew, types.TxTypeTokenValidate, types.TxTypeTokenInvalidate:
		return &TypeUpdate{}
	}
	return nil
}
