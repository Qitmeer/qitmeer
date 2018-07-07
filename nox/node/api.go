package node

import (
	"github.com/noxproject/nox/core/json"
	"github.com/noxproject/nox/common/hash"
	"fmt"
	"github.com/noxproject/nox/rpc"
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

func (api *PublicBlockChainAPI) GetRawTransaction(txId string, verbose bool)(json.TxRawResult, error){

	// Convert the provided transaction hash hex to a Hash.
	txHash, err := hash.NewHashFromStr(txId)
	if err != nil {
		return json.TxRawResult{}, err
	}

	// Try to fetch the transaction from the memory pool and if that fails,
	// try the block database.
	tx, err := api.node.txMemPool.FetchTransaction(txHash, true)
	if tx == nil {
		//not found from mem-pool, try db
		if true {
			return json.TxRawResult{}, fmt.Errorf("The transaction index "+
				"must be enabled to query the blockchain "+
				"(specify --txindex)", "Configuration")
		}

	}
	return json.TxRawResult{},nil
}
