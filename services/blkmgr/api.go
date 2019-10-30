// Copyright (c) 2017-2018 The qitmeer developers

package blkmgr

import (
	"bytes"
	"encoding/hex"
	"fmt"
	"github.com/Qitmeer/qitmeer/common/hash"
	"github.com/Qitmeer/qitmeer/common/marshal"
	"github.com/Qitmeer/qitmeer/core/json"
	"github.com/Qitmeer/qitmeer/core/types"
	"github.com/Qitmeer/qitmeer/rpc"
	"github.com/Qitmeer/qitmeer/core/blockchain"
	"strconv"
)

func (b *BlockManager) GetChain() *blockchain.BlockChain{
	return b.chain
}
func (b *BlockManager) API() rpc.API {
	return rpc.API{
		NameSpace: rpc.DefaultServiceNameSpace,
		Service:   NewPublicBlockAPI(b),
		Public:    true,
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
 	blockHash,err := api.bm.chain.BlockHashByOrder(uint64(order))
 	if err!=nil {
 		return "",err
	}
	return blockHash.String(),nil
}

// Return the hash range of block from 'start' to 'end'(exclude self)
// if 'end' is equal to zero, 'start' is the number that from the last block to the Gen
// if 'start' is greater than or equal to 'end', it will just return the hash of 'start'
func (api *PublicBlockAPI) GetBlockhashByRange(start uint,end uint) ([]string, error){
	totalOrder:=api.bm.chain.BlockDAG().GetBlockTotal()
	if start>=totalOrder {
		return nil,fmt.Errorf("startOrder(%d) is greater than or equal to the totalOrder(%d)",start,totalOrder)
	}
	result:=[]string{}
	if start>=end && end != 0 {
		block,err := api.bm.chain.BlockByOrder(uint64(start))
		if err!=nil {
			return nil,err
		}
		result=append(result,block.Hash().String())
	}else if end==0 {
		for i:=totalOrder-1;i>=0 ;i--  {
			if uint(len(result))>=start {
				break
			}
			block,err := api.bm.chain.BlockByOrder(uint64(i))
			if err!=nil {
				return nil,err
			}
			result=append(result,block.Hash().String())
		}
	}else {
		for i:=start;i<totalOrder ;i++  {
			if i>=end {
				break
			}
			block,err := api.bm.chain.BlockByOrder(uint64(i))
			if err!=nil {
				return nil,err
			}
			result=append(result,block.Hash().String())
		}
	}
	return result,nil
}

func (api *PublicBlockAPI) GetBlockByOrder(order uint64,verbose *bool,inclTx *bool, fullTx *bool) (interface{}, error){
	if uint(order) > api.bm.chain.BestSnapshot().GraphState.GetMainOrder() {
		return nil,fmt.Errorf("Order is too big")
	}
	blockHash,err := api.bm.chain.BlockHashByOrder(order)
 	if err!=nil {
 		return nil,err
	}
	vb:=false
	if verbose != nil {
		vb=*verbose
	}
	iTx:=true
	if inclTx != nil {
		iTx=*inclTx
	}
	fTx:=true
	if fullTx != nil {
		fTx=*fullTx
	}
	return api.GetBlock(*blockHash,&vb,&iTx,&fTx)
}


func (api *PublicBlockAPI) GetBlock(h hash.Hash, verbose *bool,inclTx *bool, fullTx *bool) (interface{}, error){

	vb:=false
	if verbose != nil {
		vb=*verbose
	}
	iTx:=true
	if inclTx != nil {
		iTx=*inclTx
	}
	fTx:=true
	if fullTx != nil {
		fTx=*fullTx
	}

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
	blk.SetHeight(node.GetHeight())
	// When the verbose flag isn't set, simply return the
	// network-serialized block as a hex-encoded string.
	if !vb {
		blkBytes, err := blk.Bytes()
		if err != nil {
			return nil, rpc.RpcInternalError(err.Error(),
				"Could not serialize block")
		}
		return hex.EncodeToString(blkBytes), nil
	}
	confirmations := int64(api.bm.chain.BlockDAG().GetConfirmations(&h))
	cs:=node.GetChildren()
	children:=[]*hash.Hash{}
	if cs!=nil {
		children=cs.List()
	}
	//TODO, refactor marshal api
	fields, err := marshal.MarshalJsonBlock(blk,iTx, fTx, api.bm.params, confirmations,children,
		api.bm.chain.BlockIndex().NodeStatus(node).KnownValid(),node.IsOrdered())
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
	return best.GraphState.GetMainOrder()+1, nil
}

func (api *PublicBlockAPI) GetBlockTotal() (interface{}, error){
	best := api.bm.chain.BestSnapshot()
	return best.GraphState.GetTotal(), nil
}

// GetBlockHeader implements the getblockheader command.
func (api *PublicBlockAPI) GetBlockHeader(hash hash.Hash, verbose bool) (interface{}, error) {

	// Fetch the block node
	node:=api.bm.chain.BlockIndex().LookupNode(&hash)
	if node==nil {
		return nil, rpc.RpcInternalError(fmt.Errorf("no block").Error(), fmt.Sprintf("Block not found: %v", hash))
	}
	// Fetch the header from chain.
	blockHeader, err := api.bm.chain.HeaderByHash(&hash)
	if err != nil {
		return nil, rpc.RpcInternalError(err.Error(), fmt.Sprintf("Block not found: %v", hash))
	}

	// When the verbose flag isn't set, simply return the serialized block
	// header as a hex-encoded string.
	if !verbose {
		var headerBuf bytes.Buffer
		err := blockHeader.Serialize(&headerBuf)
		if err != nil {
			context := "Failed to serialize block header"
			return nil, rpc.RpcInternalError(err.Error(), context)
		}
		return hex.EncodeToString(headerBuf.Bytes()), nil
	}
	// Get next block hash unless there are none.
	confirmations := int64(api.bm.chain.BlockDAG().GetConfirmations(node.GetHash()))
	layer := api.bm.chain.BlockDAG().GetLayer(node.GetHash())
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
		PowResult:         blockHeader.Pow.GetPowResult(),
	}

	return blockHeaderReply, nil

}

// Query whether a given block is on the main chain.
// Note that some DAG protocols may not support this feature.
func (api *PublicBlockAPI) IsOnMainChain(h hash.Hash) (interface{}, error){
	node:=api.bm.chain.BlockIndex().LookupNode(&h)
	if node==nil {
		return nil, rpc.RpcInternalError(fmt.Errorf("no block").Error(), fmt.Sprintf("Block not found: %v", h))
	}
	isOn:=api.bm.chain.BlockDAG().IsOnMainChain(&h)

	return strconv.FormatBool(isOn),nil
}

// Return the current height of DAG main chain
func (api *PublicBlockAPI) GetMainChainHeight() (interface{}, error){
	return strconv.FormatUint(uint64(api.bm.GetChain().BlockDAG().GetMainChainTip().GetHeight()),10),nil
}

// Return the weight of block
func (api *PublicBlockAPI) GetBlockWeight(h hash.Hash) (interface{}, error){
	block,err:=api.bm.chain.FetchBlockByHash(&h)
	if err != nil {
		return nil, rpc.RpcInternalError(fmt.Errorf("no block").Error(), fmt.Sprintf("Block not found: %v", h))
	}
	return strconv.FormatInt(int64(types.GetBlockWeight(block.Block())),10),nil
}

// Return the total of orphans
func (api *PublicBlockAPI) GetOrphansTotal() (interface{}, error) {
	return api.bm.GetChain().GetOrphansTotal(),nil
}

func (api *PublicBlockAPI) GetBlockByID(id uint64,verbose *bool,inclTx *bool, fullTx *bool) (interface{}, error) {
	blockHash:= api.bm.GetChain().BlockDAG().GetBlockHash(uint(id))
	if blockHash == nil {
		return nil, rpc.RpcInternalError(fmt.Errorf("no block").Error(), fmt.Sprintf("Block not found: %v", id))
	}
	vb:=false
	if verbose != nil {
		vb=*verbose
	}
	iTx:=true
	if inclTx != nil {
		iTx=*inclTx
	}
	fTx:=true
	if fullTx != nil {
		fTx=*fullTx
	}
	return api.GetBlock(*blockHash,&vb,&iTx,&fTx)
}

// IsBlue:0:not blue;  1：blue  2：Cannot confirm
func (api *PublicBlockAPI) IsBlue(h hash.Hash) (interface{}, error){
	ib:=api.bm.chain.BlockDAG().GetBlock(&h)
	if ib == nil {
		return 2, rpc.RpcInternalError(fmt.Errorf("no block").Error(), fmt.Sprintf("Block not found: %s", h.String()))
	}
	confirmations := api.bm.chain.BlockDAG().GetConfirmations(&h)
	if confirmations ==0 {
		return 2,nil
	}
	if api.bm.chain.BlockDAG().IsBlue(&h) {
		return 1,nil
	}
	return 0,nil
}
