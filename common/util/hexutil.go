// Copyright 2017-2018 The qitmeer developers

package util

import (
	"encoding/hex"
	"log"
	"math/big"
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

// MustHex2Bytes returns the bytes represented by the hexadecimal string str. Must means panic when err
func MustHex2Bytes(str string) []byte {
	h, err := hex.DecodeString(str)
	if err != nil {
		panic(err)
	}
	return h
}
