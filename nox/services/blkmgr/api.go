// Copyright (c) 2017-2018 The nox developers

package blkmgr

import (
	"github.com/noxproject/nox/rpc"
	"github.com/noxproject/nox/core/json"
	"github.com/noxproject/nox/core/blockchain"
	"github.com/noxproject/nox/services/common/marshal"
)

func (b *BlockManager) GetChain() *blockchain.BlockChain{
	return b.chain
}
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

func (api *PublicBlockAPI) GetBlockByHeight(height uint64, fullTx bool) (json.OrderedResult, error){
	block,err := api.bm.chain.BlockByHeight(height)
 	if err!=nil {
 		return nil,err
	}

	best := api.bm.chain.BestSnapshot()

	// See if this block is an orphan and adjust Confirmations accordingly.
	onMainChain, _ := api.bm.chain.MainChainHasBlock(block.Hash())

	// Get next block hash unless there are none.
	var nextHashString string
	confirmations := int64(-1)

	if onMainChain {
		if height < best.Height {
			nextHash, err := api.bm.chain.BlockHashByHeight(height + 1)
			if err != nil {
				return nil, err
			}
			nextHashString = nextHash.String()
		}
		confirmations = 1 + int64(best.Height) - int64(height)
	}
	fields, err := marshal.MarshalJsonBlock(block, true, fullTx, api.bm.params, confirmations, nextHashString)
	if err != nil {
		return nil, err
	}
	return fields,nil
}

