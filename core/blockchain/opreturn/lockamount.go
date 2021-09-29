package opreturn

import (
	"fmt"
	"github.com/Qitmeer/qitmeer/core/types"
	"github.com/Qitmeer/qitmeer/engine/txscript"
)

type LockAmount struct {
	amount int64
}

func (a *LockAmount) GetType() OPReturnType {
	return OPReturnType(txscript.OP_MEER_LOCK)
}

func (a *LockAmount) Verify(tx *types.Transaction) error {
	if len(tx.TxOut) <= 0 {
		return fmt.Errorf("Coinbase tx is error")
	}
	amount := tx.TxOut[0].Amount.Value
	if amount == a.amount {
		return nil
	}
	return fmt.Errorf("It is not equal in %s:%d != %d ", a.GetType().Name(), a.amount, amount)
}

func (a *LockAmount) Init(ops []txscript.ParsedOpcode) error {
	if len(ops) < 3 {
		return fmt.Errorf("Illegal %s", a.GetType().Name())
	}
	amount, err := txscript.MakeScriptNum(ops[2].GetData(), true, 8)
	if err != nil {
		log.Error(err.Error())
		return err
	}
	a.amount = int64(amount)
	return nil
}

func (a *LockAmount) PKScript() []byte {
	pks, err := txscript.NewScriptBuilder().AddOp(txscript.OP_RETURN).AddOp(txscript.OP_MEER_LOCK).AddInt64(a.amount).Script()
	if err != nil {
		log.Error(err.Error())
		return nil
	}
	return pks
}

func (a *LockAmount) GetAmount() int64 {
	return a.amount
}

func NewShowAmount(amount int64) *LockAmount {
	return &LockAmount{amount: amount}
}
