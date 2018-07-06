// Copyright (c) 2017-2018 The nox developers

package blkmgr

import (
	"github.com/noxproject/nox/rpc"
	"github.com/noxproject/nox/core/types"
	"github.com/noxproject/nox/common/util"
	"strconv"
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

func (api *PublicBlockAPI) GetBlockByHeight(height uint, fullTx bool) (util.OrderedMap, error){
	block,err := api.bm.chain.BlockByHeight(uint64(height))
 	if err!=nil {
 		return nil,err
	}
	fields, err := api.rpcMarshalBlock(block, true, fullTx)
	if err != nil {
		return nil, err
	}
	return fields,nil
}



// RPCMarshalBlock converts the given block to the RPC output which depends on fullTx. If inclTx is true transactions are
// returned. When fullTx is true the returned block contains full transaction details, otherwise it will only contain
// transaction hashes.
func (api *PublicBlockAPI) rpcMarshalBlock(b *types.SerializedBlock, inclTx bool, fullTx bool) (util.OrderedMap, error) {

	head := b.Block().Header // copies the header once


	best := api.bm.chain.BestSnapshot()

	// See if this block is an orphan and adjust Confirmations accordingly.
	onMainChain, _ := api.bm.chain.MainChainHasBlock(b.Hash())

	// Get next block hash unless there are none.
	var nextHashString string
	confirmations := int64(-1)
	height := uint64(head.Height)
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
	fields := util.OrderedMap{
		{"hash",         b.Hash().String()},
		{"confirmations",confirmations},
		{"version",      head.Version},
		{"height",       height},
		{"txRoot",       head.TxRoot.String()},
		{"stateRoot",    head.StateRoot.String()},
		{"bits",         strconv.FormatUint(uint64(head.Difficulty),16)},
		{"difficulty",   head.Difficulty},
		{"nonce",        head.Nonce},
		{"timestamp",    head.Timestamp.Format("2006-01-02 15:04:05.0000")},
		{"parentHash",   head.ParentRoot.String()},
		{"childrenHash", nextHashString}	,
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
