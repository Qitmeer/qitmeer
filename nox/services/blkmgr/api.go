// Copyright (c) 2017-2018 The nox developers

package blkmgr

import (
	"github.com/noxproject/nox/rpc"
	"github.com/noxproject/nox/core/json"
	"github.com/noxproject/nox/core/blockchain"
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

func (api *PublicBlockAPI) GetBlockByHeight(height uint, fullTx bool) (json.OrderedResult, error){
	block,err := api.bm.chain.BlockByHeight(uint64(height))
 	if err!=nil {
 		return nil,err
	}
	fields, err := api.marshalJsonBlock(block, true, fullTx)
	if err != nil {
		return nil, err
	}
	return fields,nil
}

