package node

import (
	"github.com/noxproject/nox/rpc"
	"github.com/noxproject/nox/common/hash"
	"fmt"
	"github.com/noxproject/nox/core/json"
)

func (nf *NoxFull) API() rpc.API {
	return rpc.API{
		NameSpace: rpc.DefaultServiceNameSpace,
		Service:   NewPublicBlockChainAPI(nf),
	}
}

type PublicBlockChainAPI struct{
	node *NoxFull
}

func NewPublicBlockChainAPI(node *NoxFull) *PublicBlockChainAPI {
	return &PublicBlockChainAPI{node}
}

func (api *PublicBlockChainAPI) GetRawTransaction(txHash hash.Hash, verbose bool)(interface{}, error){

	// Try to fetch the transaction from the memory pool and if that fails,
	// try the block database.
	tx, _ := api.node.txMemPool.FetchTransaction(&txHash, true)
	if tx == nil {
		//not found from mem-pool, try db
		txIndex := api.node.txIndex
		if txIndex == nil {
			return nil, fmt.Errorf("the transaction index "+
				"must be enabled to query the blockchain (specify --txindex in configuration)")
		}
		return nil,nil

	}
	return json.TxRawResult{},nil
}
