package blockchain

import "github.com/HalalChain/qitmeer-lib/common/hash"

type TxManager interface {

	RemoveInvalidTx(bh *hash.Hash)

	GetInvalidTxFromBlock(bh *hash.Hash) []*hash.Hash

	IsInvalidTx(txh *hash.Hash) bool

	AddInvalidTx(txh *hash.Hash, bh *hash.Hash)
}