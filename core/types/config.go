// Copyright 2017-2018 The nox developers

package types

import (
	"math/big"
)

type Config struct {
	Id  *big.Int          `json:"Id"  required:"true" min:"0"`
}

type configJSON struct {
	Id  *UInt256
}
