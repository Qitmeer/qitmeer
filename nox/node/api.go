package node

import (
	"github.com/noxproject/nox/rpc"
	"github.com/noxproject/nox/common/hash"
	"fmt"
	"github.com/noxproject/nox/core/json"
	"encoding/hex"
	"bytes"
	"github.com/noxproject/nox/core/types"
	"errors"
	"github.com/noxproject/nox/database"
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

	var mtx *types.Transaction
	var blkHash *hash.Hash
	var blkHeight uint64
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
		// Look up the location of the transaction.
		blockRegion, err := txIndex.TxBlockRegion(txHash)
		if err != nil {
			return nil,errors.New("Failed to retrieve transaction location")
		}
		if blockRegion == nil {
			return nil, rpcNoTxInfoError(&txHash)
		}

		// Load the raw transaction bytes from the database.
		var txBytes []byte
		err = api.node.db.View(func(dbTx database.Tx) error {
			var err error
			txBytes, err = dbTx.FetchBlockRegion(blockRegion)
			return err
		})
		if err != nil {
			return nil, rpcNoTxInfoError(&txHash)
		}

		// When the verbose flag isn't set, simply return the serialized
		// transaction as a hex-encoded string.  This is done here to
		// avoid deserializing it only to reserialize it again later.
		if !verbose {
			return hex.EncodeToString(txBytes), nil
		}

		// Grab the block height.
		blkHash = blockRegion.Hash
		blkHeight, err = api.node.blockManager.GetChain().BlockHeightByHash(blkHash)
		if err != nil {
			context := "Failed to retrieve block height"
			return nil, rpcInternalError(err.Error(), context)
		}

		// Deserialize the transaction
		var msgTx types.Transaction
		err = msgTx.Deserialize(bytes.NewReader(txBytes))
		if err != nil {
			context := "Failed to deserialize transaction"
			return nil, rpcInternalError(err.Error(), context)
		}
		mtx = &msgTx
	}
	return json.TxRawResult{
		BlockHash:blkHash.String(),
		BlockIndex:uint32(blkHeight),  //TODO, remove type conversion
		Txid:mtx.TxHash().String(),
	},nil
}


// rpcNoTxInfoError is a convenience function for returning a nicely formatted
// RPC error which indicates there is no information available for the provided
// transaction hash.
func rpcNoTxInfoError(txHash *hash.Hash) error {
	return errors.New(
		fmt.Sprintf("No information available about transaction %v",
			txHash))
}

func rpcInternalError(err, context string) error{
	return errors.New(
		fmt.Sprintf("%s:%s",err,context))
}
