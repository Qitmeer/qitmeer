// Copyright (c) 2017-2018 The nox developers

package blkmgr

import (
	"bytes"
	"encoding/hex"
	"qitmeer/common/hash"
	"qitmeer/core/blockchain"
	"qitmeer/core/json"
	"qitmeer/rpc"
	"qitmeer/services/common/error"
	"fmt"
	"qitmeer/services/common/marshal"
	"strconv"
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
func (api *PublicBlockAPI) GetBlockhash(order uint) (string, error){
 	block,err := api.bm.chain.BlockByOrder(uint64(order))
 	if err!=nil {
 		return "",err
	}
	return block.Hash().String(),nil
}

//TODO, refactor BlkMgr API
func (api *PublicBlockAPI) GetBlockByOrder(order uint64, fullTx bool) (json.OrderedResult, error){
	block,err := api.bm.chain.BlockByOrder(order)

 	if err!=nil {
 		return nil,err
	}
	node:=api.bm.chain.BlockIndex().LookupNode(block.Hash())
	if node==nil {
		return nil,fmt.Errorf("no node")
	}
	// Update the source block order
	block.SetOrder(node.GetOrder())

	best := api.bm.chain.BestSnapshot()

	// See if this block is an orphan and adjust Confirmations accordingly.
	onMainChain, _ := api.bm.chain.MainChainHasBlock(block.Hash())

	// Get next block hash unless there are none.
	confirmations := int64(-1)

	if onMainChain {
		confirmations = 1 + int64(best.Order) - int64(order)
	}
	cs:=node.GetChildren()
	children:=[]*hash.Hash{}
	if cs!=nil {
		children=cs.List()
	}
	//TODO, refactor marshal api

	fields, err := marshal.MarshalJsonBlock(block, true, fullTx, api.bm.params, confirmations,children)
	if err != nil {
		return nil, err
	}
	return fields,nil
}


func (api *PublicBlockAPI) GetBlock(h hash.Hash, verbose bool) (interface{}, error){

	// Load the raw block bytes from the database.
	// Note :
	// FetchBlockByHash differs from BlockByHash in that this one also returns blocks
	// that are not part of the main chain (if they are known).
	blk, err := api.bm.chain.FetchBlockByHash(&h)
	if err != nil {
		return nil,err
	}
	node:=api.bm.chain.BlockIndex().LookupNode(&h)
	if node==nil {
		return nil,fmt.Errorf("no node")
	}
	// Update the source block order
	blk.SetOrder(node.GetOrder())
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
	onMainChain, _ := api.bm.chain.MainChainHasBlock(&h)

	//blockHeader := &blk.Block().Header
	order := blk.Order()
	confirmations := int64(-1)

	if onMainChain {
		confirmations = 1 + int64(best.Order) - int64(order)
	}
	cs:=node.GetChildren()
	children:=[]*hash.Hash{}
	if cs!=nil {
		children=cs.List()
	}
	//TODO, refactor marshal api
	fields, err := marshal.MarshalJsonBlock(blk, true, verbose, api.bm.params, confirmations,children)
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
	return best.Order, nil
}

// GetBlockHeader implements the getblockheader command.
func (api *PublicBlockAPI) GetBlockHeader(hash hash.Hash, verbose bool) (interface{}, error) {

	// Fetch the block node
	node:=api.bm.chain.BlockIndex().LookupNode(&hash)
	if node==nil {
		return nil, er.RpcInternalError(fmt.Errorf("no block").Error(), fmt.Sprintf("Block not found: %v", hash))
	}
	// Fetch the header from chain.
	blockHeader, err := api.bm.chain.HeaderByHash(&hash)
	if err != nil {
		return nil, er.RpcInternalError(err.Error(), fmt.Sprintf("Block not found: %v", hash))
	}

	// When the verbose flag isn't set, simply return the serialized block
	// header as a hex-encoded string.
	if !verbose {
		var headerBuf bytes.Buffer
		err := blockHeader.Serialize(&headerBuf)
		if err != nil {
			context := "Failed to serialize block header"
			return nil, er.RpcInternalError(err.Error(), context)
		}
		return hex.EncodeToString(headerBuf.Bytes()), nil
	}

	best := api.bm.chain.BestSnapshot()
	bestNode:=api.bm.chain.BlockIndex().LookupNode(&best.Hash)

	// Get next block hash unless there are none.
	confirmations := int64(-1)
	layer := api.bm.chain.BlockDAG().GetLayer(node.GetHash())
	if bestNode!=nil {
		confirmations=1+int64(api.bm.chain.BlockDAG().GetLayer(bestNode.GetHash()))-int64(layer)
		if confirmations<1 {
			confirmations=1
		}
	}
	blockHeaderReply := json.GetBlockHeaderVerboseResult{
		Hash:          hash.String(),
		Confirmations: confirmations,
		Version:       int32(blockHeader.Version),
		ParentRoot:    blockHeader.ParentRoot.String(),
		TxRoot:        blockHeader.TxRoot.String(),
		StateRoot:     blockHeader.StateRoot.String(),
		Difficulty:    blockHeader.Difficulty,
		Layer:         uint32(layer),
		Time:          blockHeader.Timestamp.Unix(),
		Nonce:         blockHeader.Nonce,
	}

	return blockHeaderReply, nil

}

// Query whether a given block is on the main chain.
// Note that some DAG protocols may not support this feature.
func (api *PublicBlockAPI) IsOnMainChain(h hash.Hash) (interface{}, error){
	node:=api.bm.chain.BlockIndex().LookupNode(&h)
	if node==nil {
		return nil, er.RpcInternalError(fmt.Errorf("no block").Error(), fmt.Sprintf("Block not found: %v", h))
	}
	isOn:=api.bm.chain.BlockDAG().IsOnMainChain(&h)

	return strconv.FormatBool(isOn),nil
}


