/*
 * Copyright (c) 2017-2020 The qitmeer developers
 */

package opreturn

import (
	"fmt"
	"github.com/Qitmeer/qitmeer/engine/txscript"
)

type OPReturnType byte

const (
	// Show locked currency amount
	ShowAmountType = 0x01

	// ...
)

var OPRNameMap = map[OPReturnType]string{
	ShowAmountType: "Show locked amount",
}

func (t OPReturnType) Name() string {
	if t, ok := OPRNameMap[t]; ok {
		return t
	} else {
		return "Unknown-OPReturn type:" + fmt.Sprintf("%x", t)
	}
}

// Exclusive to Coinbase OP Return function
type IOPReturn interface {
	GetType() OPReturnType
	Verify() error
	Deserialize(data []byte) error
	Serialize() ([]byte, error)
}

func IsOPReturn(pks []byte) bool {
	if len(pks) <= 0 {
		return false
	}
	return txscript.GetScriptClass(txscript.DefaultScriptVersion, pks) == txscript.NullDataTy
}

func NewOPReturnFrom(pks []byte) (IOPReturn, error) {
	opData, err := txscript.ExtractCoinbaseNullData(pks)
	if err != nil {
		return nil, err
	}
	if len(opData) <= 0 {
		return nil, fmt.Errorf("Is is not coinbase opreturn")
	}
	opType := OPReturnType(opData[0])
	switch opType {
	case ShowAmountType:
		sa := ShowAmount{}
		sa.Deserialize(opData[1:])
		return &sa, nil
	}
	return nil, fmt.Errorf("No support %s", opType.Name())
}
