package token

import "github.com/Qitmeer/qitmeer/core/types"

type TokenType struct {
	Id      types.CoinID
	Owners  []byte
	UpLimit uint64
	Enable  bool
}

type TokenTypesMap map[types.CoinID]TokenType
