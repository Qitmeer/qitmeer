package types

import (
	"encoding/binary"
	"fmt"
	"strconv"
)

// from 0 ~ 65535
// 0 ~ 255 : Qitmeer reserved
type CoinID uint16

const (
	MEERID CoinID = 0

	QitmeerReservedID CoinID = 255
)

func (c CoinID) Name() string {
	if c == MEERID {
		return "MEER"
	} else if t, ok := CoinNameMap[c]; ok {
		return t
	} else {
		return "Unknown-CoinID:" + strconv.FormatInt(int64(c), 10)
	}
}

func (c CoinID) Bytes() []byte {
	b := [2]byte{}
	binary.LittleEndian.PutUint16(b[:], uint16(c))
	return b[:]
}

// Is it the main coin ?
func (c CoinID) IsBase() bool {
	return c == MEERID
}

func NewCoinID(name string) CoinID {
	for _, coinid := range CoinIDList {
		if name == coinid.Name() {
			return coinid
		}
	}
	// panic(fmt.Sprintf("Unknown-CoinID:%s", name)) // Which way is better ?
	return MEERID
}

var CoinNameMap = map[CoinID]string{}
var CoinIDList = []CoinID{MEERID}

// Check if a valid coinId, current only check if the coinId is known.
func CheckCoinID(id CoinID) error {
	unknownCoin := true
	for _, coinId := range CoinIDList {
		if id == coinId {
			unknownCoin = false
			break
		}
	}
	if unknownCoin {
		return fmt.Errorf("unknown coin id %s", id.Name())
	}
	return nil
}

func IsKnownCoinID(id CoinID) bool {
	return CheckCoinID(id) == nil
}

const (
	// Greater than or equal to
	FloorFeeType = 0

	// Strict equality
	EqualFeeType = 1
)

type FeeType byte
