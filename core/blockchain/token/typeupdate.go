package token

import (
	"fmt"
	"github.com/Qitmeer/qitmeer/common/hash"
	"github.com/Qitmeer/qitmeer/core/types"
	"github.com/Qitmeer/qitmeer/engine/txscript"
	"github.com/Qitmeer/qitmeer/params"
)

type TypeUpdate struct {
	*TokenUpdate
	Tt TokenType

	cacheHash *hash.Hash
}

func (tu *TypeUpdate) Serialize() ([]byte, error) {
	tuSerialized, err := tu.TokenUpdate.Serialize()
	if err != nil {
		return nil, err
	}

	serialized, err := tu.Tt.Serialize()
	if err != nil {
		return nil, err
	}

	serialized = append(tuSerialized, serialized...)
	return serialized, nil
}

func (tu *TypeUpdate) Deserialize(data []byte) (int, error) {
	bytesRead, err := tu.TokenUpdate.Deserialize(data)
	if err != nil {
		return bytesRead, err
	}
	offset := bytesRead

	bytesRead, err = tu.Tt.Deserialize(data[offset:])
	if err != nil {
		return bytesRead, err
	}
	offset += bytesRead

	return offset, nil
}

func (tu *TypeUpdate) GetHash() *hash.Hash {
	if tu.cacheHash != nil {
		return tu.cacheHash
	}
	return tu.CacheHash()
}

func (tu *TypeUpdate) CacheHash() *hash.Hash {
	tu.cacheHash = nil
	bs, err := tu.Serialize()
	if err != nil {
		log.Error(err.Error())
		return tu.cacheHash
	}
	h := hash.DoubleHashH(bs)
	tu.cacheHash = &h
	return tu.cacheHash
}

func (tu *TypeUpdate) CheckSanity() error {
	if tu.Tt.Id <= types.QitmeerReservedID {
		return fmt.Errorf("Coin ID (%d) is qitmeer reserved. It has to be greater than %d for token type update.\n", tu.Tt.Id, types.QitmeerReservedID)
	}
	class, _, _, err := txscript.ExtractPkScriptAddrs(tu.Tt.Owners, params.ActiveNetParams.Params)
	if err != nil || class != txscript.TokenPubKeyHashTy {
		return err
	}
	if len(tu.Tt.Name) > MaxTokenNameLength {
		return fmt.Errorf("Token name (%s) exceeds the maximum length (%d).\n", tu.Tt.Name, MaxTokenNameLength)
	}
	if tu.GetType() == types.TxTypeTokenNew || tu.GetType() == types.TxTypeTokenRenew {
		if tu.Tt.UpLimit == 0 {
			return fmt.Errorf("UpLimit cannot be zero")
		}
		if len(tu.Tt.Name) <= 0 {
			return fmt.Errorf("Must have token name.\n")
		}
		if tu.GetType() == types.TxTypeTokenRenew {
			err := types.CheckCoinID(tu.Tt.Id)
			if err != nil {
				return err
			}
		}
	} else {
		return fmt.Errorf("This type (%v) is not supported\n", tu.GetType())
	}
	return nil
}

func NewTypeUpdateFromScript(tx *types.Transaction) (*TypeUpdate, error) {
	script, err := txscript.ParsePkScript(tx.TxOut[0].PkScript)
	if err != nil {
		log.Error(err.Error())
		return nil, err
	}

	if tnScript, ok := script.(*txscript.TokenNewScript); ok {
		return &TypeUpdate{
			TokenUpdate: &TokenUpdate{Typ: types.DetermineTxType(tx)},
			Tt: TokenType{
				Id:      tnScript.GetCoinId(),
				Owners:  tx.TxOut[0].PkScript,
				UpLimit: tnScript.GetUpLimit(),
				Enable:  false,
				Name:    tnScript.GetName(),
			},
		}, nil
	}
	return nil, fmt.Errorf("Not supported:%v\n", tx.TxOut[0].PkScript)
}
