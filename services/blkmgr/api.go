// Copyright (c) 2017-2018 The nox developers

package blkmgr

import (
	"github.com/noxproject/nox/rpc"
	"github.com/noxproject/nox/core/json"
	"github.com/noxproject/nox/core/blockchain"
	"github.com/noxproject/nox/services/common/marshal"
	"github.com/noxproject/nox/common/hash"
	"encoding/hex"
	"github.com/noxproject/nox/services/common/error"
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

//TODO, refactor BlkMgr API
func (api *PublicBlockAPI) GetBlockhash(height uint) (string, error){
 	block,err := api.bm.chain.BlockByHeight(uint64(height))
 	if err!=nil {
 		return "",err
	}
	return block.Hash().String(),nil
}

//TODO, refactor BlkMgr API
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
	//TODO, refactor marshal api
	fields, err := marshal.MarshalJsonBlock(block, true, fullTx, api.bm.params, confirmations, nextHashString)
	if err != nil {
		return nil, err
	}
	return fields,nil
}


func (api *PublicBlockAPI) GetBlock(hash hash.Hash, verbose bool) (interface{}, error){

	// Load the raw block bytes from the database.
	// Note :
	// FetchBlockByHash differs from BlockByHash in that this one also returns blocks
	// that are not part of the main chain (if they are known).
	blk, err := api.bm.chain.FetchBlockByHash(&hash)
	if err != nil {
		return nil,err
	}

	// When the verbose flag isn't set, simply return the
	// network-serialized block as a hex-encoded string.
	if !verbose {
		blkBytes, err := blk.Bytes()
		if err != nil {
			return nil, er.RpcInternalError(err.Error(),
				"Could not serialize block")
		}
		return hex.EncodeToString(blkBytes), nil
	}
	best := api.bm.chain.BestSnapshot()

	// See if this block is an orphan and adjust Confirmations accordingly.
	onMainChain, _ := api.bm.chain.MainChainHasBlock(&hash)

	// Get next block hash unless there are none.
	var nextHashString string
	//blockHeader := &blk.Block().Header
	height := blk.Height()
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
	//TODO, refactor marshal api
	fields, err := marshal.MarshalJsonBlock(blk, true, verbose, api.bm.params, confirmations, nextHashString)
	if err != nil {
		return nil, err
	}
	return fields,nil

}

func (api *PublicBlockAPI) GetBestBlockHash() (interface{}, error){
	best := api.bm.chain.BestSnapshot()
	return best.Hash.String(), nil
}

func (api *PublicBlockAPI) GetBlockCount() (interface{}, error){
	best := api.bm.chain.BestSnapshot()
	return best.Height, nil
}




