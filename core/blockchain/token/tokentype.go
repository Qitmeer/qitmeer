package token

import (
	"fmt"
	"github.com/Qitmeer/qitmeer/core/types"
)

const MaxTokenNameLength = 6

type TokenType struct {
	Id      types.CoinID
	Owners  []byte
	UpLimit uint64
	Enable  bool
	Name    string
}

type TokenTypesMap map[types.CoinID]TokenType

func (ttm *TokenTypesMap) Update(update *TypeUpdate) error {
	switch update.Typ {
	case types.TxTypeTokenNew:
		_, ok := (*ttm)[update.Tt.Id]
		if ok {
			return fmt.Errorf("It already exists:%s\n", update.Tt.Name)
		}
		(*ttm)[update.Tt.Id] = update.Tt
	default:
		return fmt.Errorf("unknown update type %v", update.Typ)
	}

	return nil
}
