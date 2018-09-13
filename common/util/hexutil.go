// Copyright 2017-2018 The nox developers

package util

import (
	"math/big"
	"log"
)

func HasHexPrefix(str string) bool {
	return len(str) >= 2 && str[0] == '0' && (str[1] == 'x' || str[1] == 'X')
}

func FromHex(hex string) *big.Int {
	i, ok := new(big.Int).SetString(hex, 16)
	if !ok {
		log.Fatalln("bad number: " + hex)
	}
	return i
}
