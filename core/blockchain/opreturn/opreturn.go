/*
 * Copyright (c) 2017-2020 The qitmeer developers
 */

package opreturn

import (
	"fmt"
	"github.com/Qitmeer/qitmeer/core/types"
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
	if name, ok := OPRNameMap[t]; ok {
		return name
	} else {
		return "Unknown-OPReturn type:" + fmt.Sprintf("%d", t)
	}
}

// Exclusive to Coinbase OP Return function
type IOPReturn interface {
	GetType() OPReturnType
	Verify(tx *types.Transaction) error
	Deserialize(data []byte) error
	PKScript() []byte
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

func GetOPReturnTxOutput(opr IOPReturn) *types.TxOutput {
	return &types.TxOutput{
		Amount:   types.Amount{Value: 0, Id: types.MEERID},
		PkScript: opr.PKScript(),
	}
}
