// Copyright 2017-2018 The nox developers

package types

import (
	"math/big"
	"qitmeer/common/util"
	//"fmt"
	"fmt"
	"strings"
)

type Bytes []byte
type Uint uint
type UInt64 uint64
type UInt128 big.Int
type UInt256 big.Int

func parseUInt64(s string) (UInt64, bool ) {
	var u64 UInt64
	if i,ok := parseBigInt(s,64); ok {
		u64 = UInt64(i.Uint64())
		return u64,true
	}
	return u64,false
}

func parseUInt128(s string) (UInt128, bool ) {
	var u128 UInt128
	if i,ok := parseBigInt(s,128); ok {
		u128 = UInt128(*i)
		return u128,true
	}
	return u128,false
}

func parseUInt256(s string) (UInt256, bool ) {
	var u256 UInt256
	if i,ok := parseBigInt(s,256); ok {
		u256 = UInt256(*i)
		return u256,true
	}
	return u256,false
}

func parseBigInt(s string, size int) (i *big.Int, ok bool ) {
	if s == "" {
		return new(big.Int), true
	}
    i, ok = nil,false
	if util.HasHexPrefix(s) {
		i,ok = new(big.Int).SetString(s[2:], 16)
	} else {
		i,ok = new(big.Int).SetString(s, 10)
	}
	if ok && i.BitLen() > 256 {
		i, ok = nil, false
	}
	return
}

func (i *UInt256) MarshalJSON() ([]byte, error) {
	if i == nil {
		return []byte("0x0"), nil
	}
	return []byte(fmt.Sprintf(`"%#x"`, (*big.Int)(i))), nil
}

func (i *UInt64) UnmarshalJSON(input []byte) error {
	s := strings.Trim(string(input),"\"")
	int64, ok := parseUInt64(s)
	if  !ok {
		return fmt.Errorf("invalid hex or decimal integer %q ", input)
	}
	*i = int64
	return nil
}

func (i *UInt128) UnmarshalJSON(input []byte) error {
	s := strings.Trim(string(input),"\"")
	u128, ok := parseUInt128(s)
	if !ok {
		return fmt.Errorf("invalid hex or decimal integer %q", input)
	}
	*i = u128
	return nil
}

func (i *UInt256) UnmarshalJSON(input []byte) error {
	s := strings.Trim(string(input),"\"")
	u256, ok := parseUInt256(s)
	if !ok {
		return fmt.Errorf("invalid hex or decimal integer %q", input)
	}
	*i = u256
	return nil
}

func NewUInt128(x int64) *UInt128 {
	a := big.NewInt(x)
	i := new(UInt128)
	*i = UInt128(*a)
	return i
}
func NewUInt256(x int64) *UInt256 {
	a := big.NewInt(x)
	i := new(UInt256)
	*i = UInt256(*a)
	return i
}
