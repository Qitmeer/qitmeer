package consensus

import "github.com/Qitmeer/qitmeer/core/types"

const (
	TxTypeCrossChainExport types.TxType = 0x0101 // Cross chain by export tx
	TxTypeCrossChainImport types.TxType = 0x0102 // Cross chain by import tx
)

type Tx interface {
	GetType() types.TxType
	GetFrom() string
	GetTo() string
	GetValue() uint64
	GetData() []byte
}
