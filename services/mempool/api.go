package mempool

import (
	"fmt"
	"github.com/Qitmeer/qitmeer/rpc"
	"github.com/Qitmeer/qitmeer/rpc/client/cmds"
	"sort"
)

func (t *TxPool) API() rpc.API {
	return rpc.API{
		NameSpace: cmds.DefaultServiceNameSpace,
		Service:   NewPublicMempoolAPI(t),
		Public:    true,
	}
}

type PublicMempoolAPI struct {
	txPool *TxPool
}

func NewPublicMempoolAPI(txPool *TxPool) *PublicMempoolAPI {
	return &PublicMempoolAPI{txPool}
}

func (api *PublicMempoolAPI) GetMempool(txType *string, verbose bool) (interface{}, error) {
	log.Trace("GetMempool called")
	// TODO verbose
	// The response is simply an array of the transaction hashes if the
	// verbose flag is not set.
	descs := api.txPool.TxDescs()
	hashStrings := make([]string, 0, len(descs))
	for i := range descs {
		hashStrings = append(hashStrings, descs[i].Tx.Hash().String())
	}
	sort.Strings(hashStrings)
	return hashStrings, nil
}

func (api *PublicMempoolAPI) GetMempoolCount() (interface{}, error) {
	return fmt.Sprintf("%d", api.txPool.Count()), nil
}

func (api *PublicMempoolAPI) SaveMempool() (interface{}, error) {
	num, err := api.txPool.Perisit()
	if err != nil {
		return nil, err
	}
	return fmt.Sprintf("Mempool persist:%d transactions", num), nil
}
