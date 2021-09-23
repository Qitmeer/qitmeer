package opreturn

import (
	"fmt"
	"github.com/Qitmeer/qitmeer/core/dbnamespace"
	"github.com/Qitmeer/qitmeer/core/types"
	"github.com/Qitmeer/qitmeer/engine/txscript"
)

type ShowAmount struct {
	subsidy int64
}

func (a *ShowAmount) GetType() OPReturnType {
	return ShowAmountType
}

func (a *ShowAmount) Verify(tx *types.Transaction) error {
	if len(tx.TxOut) <= 0 {
		return fmt.Errorf("Coinbase tx is error")
	}
	subsidy := tx.TxOut[0].Amount.Value
	if subsidy == a.subsidy {
		return nil
	}
	return fmt.Errorf("It is not equal in %s:%d != %d ", a.GetType().Name(), a.subsidy, subsidy)
}

func (a *ShowAmount) Deserialize(data []byte) error {
	a.subsidy = int64(dbnamespace.ByteOrder.Uint64(data))
	return nil
}

func (a *ShowAmount) PKScript() []byte {
	var subsidydata [8]byte
	dbnamespace.ByteOrder.PutUint64(subsidydata[:], uint64(a.subsidy))

	data := []byte{byte(a.GetType())}
	data = append(data, subsidydata[:]...)
	pks, err := txscript.GenerateProvablyPruneableOut(data)
	if err != nil {
		log.Error(err.Error())
		return nil
	}
	return pks
}

func (a *ShowAmount) GetAmount() int64 {
	return a.subsidy
}

func NewShowAmount(subsidy int64) *ShowAmount {
	return &ShowAmount{subsidy: subsidy}
}
