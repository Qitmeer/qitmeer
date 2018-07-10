package node

import (
	"github.com/noxproject/nox/rpc"
	"github.com/noxproject/nox/common/hash"
	"fmt"
	"encoding/hex"
	"bytes"
	"github.com/noxproject/nox/core/types"
	"errors"
	"github.com/noxproject/nox/database"
	"github.com/noxproject/nox/services/common/marshal"
	"github.com/noxproject/nox/services/common/error"
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



func (api *PublicBlockChainAPI) GetRawTransaction(txHash hash.Hash, verbose bool)(interface{}, error) {

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
			return nil, fmt.Errorf("the transaction index " +
				"must be enabled to query the blockchain (specify --txindex in configuration)")
		}
		// Look up the location of the transaction.
		blockRegion, err := txIndex.TxBlockRegion(txHash)
		if err != nil {
			return nil, errors.New("Failed to retrieve transaction location")
		}
		if blockRegion == nil {
			return nil, er.RpcNoTxInfoError(&txHash)
		}

		// Load the raw transaction bytes from the database.
		var txBytes []byte
		err = api.node.db.View(func(dbTx database.Tx) error {
			var err error
			txBytes, err = dbTx.FetchBlockRegion(blockRegion)
			return err
		})
		if err != nil {
			return nil, er.RpcNoTxInfoError(&txHash)
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
			return nil, er.RpcInternalError(err.Error(), context)
		}

		// Deserialize the transaction
		var msgTx types.Transaction
		err = msgTx.Deserialize(bytes.NewReader(txBytes))
		if err != nil {
			context := "Failed to deserialize transaction"
			return nil, er.RpcInternalError(err.Error(), context)
		}
		mtx = &msgTx
	} else {
		// When the verbose flag isn't set, simply return the
		// network-serialized transaction as a hex-encoded string.
		if !verbose {
			// Note that this is intentionally not directly
			// returning because the first return value is a
			// string and it would result in returning an empty
			// string to the client instead of nothing (nil) in the
			// case of an error.

			buf, err := tx.Transaction().Serialize(types.TxSerializeFull)
			if err != nil {
				return nil, err
			}
			txHex := hex.EncodeToString(buf)
			return txHex, nil
		}

		mtx = tx.Transaction()
	}

	//TODO, refactor the paramas place
	blkHashStr := ""
	if blkHash != nil {
		blkHashStr = blkHash.String()
	}
	return marshal.MarshalJsonTransaction(mtx,api.node.node.Params,blkHeight,blkHashStr)
}



