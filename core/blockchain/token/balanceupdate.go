package token

import (
	"fmt"
	"github.com/Qitmeer/qitmeer/common/hash"
	"github.com/Qitmeer/qitmeer/core/dbnamespace"
	"github.com/Qitmeer/qitmeer/core/serialization"
	"github.com/Qitmeer/qitmeer/core/types"
)

// balanceUpdate specifies the type and update record of the values that change a token
// balance.
// for TOKON_MINT, the values should add on the meerlock and token balance
// for TOKEN_UNMINT, the values should subtract from the meerlock and token balance
type BalanceUpdate struct {
	*TokenUpdate
	MeerAmount  int64
	TokenAmount types.Amount

	cacheHash *hash.Hash
}

func (bu *BalanceUpdate) Serialize() ([]byte, error) {
	//
	tuSerialized, err := bu.TokenUpdate.Serialize()
	if err != nil {
		return nil, err
	}

	serializeSize := serialization.SerializeSizeVLQ(uint64(bu.MeerAmount))
	serializeSize += serialization.SerializeSizeVLQ(uint64(bu.TokenAmount.Id))
	serializeSize += serialization.SerializeSizeVLQ(uint64(bu.TokenAmount.Value))

	serialized := make([]byte, serializeSize)
	offset := 0

	offset += serialization.PutVLQ(serialized[offset:], uint64(bu.MeerAmount))
	offset += serialization.PutVLQ(serialized[offset:], uint64(bu.TokenAmount.Id))
	offset += serialization.PutVLQ(serialized[offset:], uint64(bu.TokenAmount.Value))

	serialized = append(tuSerialized, serialized...)
	return serialized, nil
}

func (bu *BalanceUpdate) Deserialize(data []byte) (int, error) {
	bytesRead, err := bu.TokenUpdate.Deserialize(data)
	if err != nil {
		return bytesRead, err
	}
	offset := bytesRead
	//meerAmount
	meerAmount, bytesRead := serialization.DeserializeVLQ(data[offset:])
	if bytesRead == 0 {
		return offset, fmt.Errorf("unexpected end of data while reading meer amount at update")
	}
	offset += bytesRead
	//tokenId
	tokenId, bytesRead := serialization.DeserializeVLQ(data[offset:])
	if bytesRead == 0 {
		return offset, fmt.Errorf("unexpected end of data while reading token id at update")
	}
	offset += bytesRead
	//tokenAmount
	tokenAmount, bytesRead := serialization.DeserializeVLQ(data[offset:])
	if bytesRead == 0 {
		return offset, fmt.Errorf("unexpected end of data while reading token amount at update")
	}
	offset += bytesRead

	bu.MeerAmount = int64(meerAmount)
	bu.TokenAmount = types.Amount{Value: int64(tokenAmount), Id: types.CoinID(tokenId)}

	return offset, nil
}

func (bu *BalanceUpdate) GetHash() *hash.Hash {
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

func (bu *BalanceUpdate) CheckSanity() error {
	if bu.Typ != types.TxTypeTokenMint && bu.Typ != types.TxTypeTokenUnmint {
		return fmt.Errorf("invalid token balance update type %v", bu.Typ)
	}
	if bu.TokenAmount.Value <= 0 {
		return fmt.Errorf("invalid token balance update : wrong token amount : %v", bu.TokenAmount.Value)
	}
	if bu.MeerAmount <= 0 {
		return fmt.Errorf("invalid token balance update : wrong meer amount : %v", bu.MeerAmount)
	}
	if bu.TokenAmount.Value > types.MaxAmount {
		return fmt.Errorf("Total token amount value of %v is higher than max allowed value of %v\n", bu.TokenAmount.Value, types.MaxAmount)
	}
	if bu.MeerAmount > types.MaxAmount {
		return fmt.Errorf("Meer amount value of %v is higher than max allowed value of %v\n", bu.MeerAmount, types.MaxAmount)
	}
	return nil
}

func NewBalanceUpdate(tx *types.Transaction) (*BalanceUpdate, error) {
	meerAmount := int64(0)
	tokenAmount := types.Amount{}
	if types.IsTokenMintTx(tx) {
		existingTxOut := make(map[types.TxOutPoint]struct{})
		for _, in := range tx.TxIn {
			if _, exists := existingTxOut[in.PreviousOut]; exists {
				return nil, fmt.Errorf("transaction contains duplicate inputs")
			}
			existingTxOut[in.PreviousOut] = struct{}{}
		}

		meerAmount = int64(dbnamespace.ByteOrder.Uint64(tx.TxIn[0].PreviousOut.Hash[0:8]))

		tokenAmount.Id = tx.TxOut[0].Amount.Id
		for idx, out := range tx.TxOut {
			if tokenAmount.Id != out.Amount.Id {
				return nil, fmt.Errorf("Transaction(%s) output(%d) coin id is invalid\n", tx.TxHash(), idx)
			}
			if out.Amount.Value > types.MaxAmount {
				return nil, fmt.Errorf("transaction output value of %v is "+
					"higher than max allowed value of %v", out.Amount.Value,
					types.MaxAmount)
			}
			if out.Amount.Value < 0 {
				return nil, fmt.Errorf("Transaction output value is less than or equal to zero\n")
			}
			tokenAmount.Value += out.Amount.Value
		}
	}

	if tokenAmount.Value > types.MaxAmount {
		return nil, fmt.Errorf("Total token amount value of %v is higher than max allowed value of %v\n", tokenAmount.Value, types.MaxAmount)
	}

	return &BalanceUpdate{
		TokenUpdate: &TokenUpdate{Typ: types.DetermineTxType(tx)},
		MeerAmount:  meerAmount,
		TokenAmount: tokenAmount,
	}, nil
}
