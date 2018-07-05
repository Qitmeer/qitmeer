package blkmgr

import (
	"github.com/noxproject/nox/rpc"
	"github.com/noxproject/nox/core/types"
)

func (b *BlockManager) API() rpc.API {
	return rpc.API{
		NameSpace: rpc.DefaultServiceNameSpace,
		Service:   NewPublicBlockAPI(b),
	}
}

type PublicBlockAPI struct{
	bm *BlockManager
}

func NewPublicBlockAPI(bm *BlockManager) *PublicBlockAPI {
	return &PublicBlockAPI{bm}
}

func (api *PublicBlockAPI) GetBlockhash(height uint) (string, error){
 	block,err := api.bm.chain.BlockByHeight(uint64(height))
 	if err!=nil {
 		return "",err
	}
	return block.Hash().String(),nil
}

func (api *PublicBlockAPI) GetBlockByHeight(height uint, fullTx bool) (map[string]interface{}, error){
	block,err := api.bm.chain.BlockByHeight(uint64(height))
 	if err!=nil {
 		return nil,err
	}
	fields, err := RPCMarshalBlock(block, true, fullTx)
	if err != nil {
		return nil, err
	}
	return fields,nil
}

// RPCMarshalBlock converts the given block to the RPC output which depends on fullTx. If inclTx is true transactions are
// returned. When fullTx is true the returned block contains full transaction details, otherwise it will only contain
// transaction hashes.
func RPCMarshalBlock(b *types.SerializedBlock, inclTx bool, fullTx bool) (map[string]interface{}, error) {
	head := b.Block().Header // copies the header once
	fields := map[string]interface{}{
		"number":           head.Height,
		"hash":             b.Hash().String(),
		"parentRoot":       head.ParentRoot.String(),
		"transactionsRoot": head.TxRoot.String(),
		"stateRoot":        head.StateRoot.String(),
		"nonce":            head.Nonce,
		"difficulty":       head.Difficulty,
		"version":			head.Version,
		"timestamp":        head.Timestamp.Format("2006-01-02|15:04:05.0000"),
	}

	/*
	if inclTx {
		formatTx := func(tx *types.Transaction) (interface{}, error) {
			return tx.Hash(), nil
		}
		if fullTx {
			formatTx = func(tx *types.Transaction) (interface{}, error) {
				return newRPCTransactionFromBlockHash(b, tx.Hash()), nil
			}
		}
		txs := b.Transactions()
		transactions := make([]interface{}, len(txs))
		var err error
		for i, tx := range txs {
			if transactions[i], err = formatTx(tx); err != nil {
				return nil, err
			}
		}
		fields["transactions"] = transactions
	}
	*/
	return fields, nil
}
