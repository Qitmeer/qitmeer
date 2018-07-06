// Copyright (c) 2017-2018 The nox developers

package blkmgr

import (
	"github.com/noxproject/nox/rpc"
	"github.com/noxproject/nox/core/types"
	"strconv"
	"errors"
	"github.com/noxproject/nox/core/json"
	"encoding/hex"
	"github.com/noxproject/nox/engine/txscript"
	"github.com/noxproject/nox/core/blockchain"
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



func (api *PublicBlockAPI) marshalJsonTransaction(tx *types.Tx) (json.TxRawResult, error){
	if tx == nil {
		return json.TxRawResult{}, errors.New("can't marshal nil transaction")
	}

	//TODO, refactor the hex serialize code
	buf,_ :=tx.Transaction().Serialize(types.TxSerializeFull)
	hexStr:=hex.EncodeToString(buf)

	return json.TxRawResult{
		Hex : hexStr,
		Txid : tx.Hash().String(),
		Version: tx.Transaction().Version,
		LockTime:tx.Transaction().LockTime,
		Expire:tx.Transaction().Expire,
		Vin: api.marshJsonVin(tx.Transaction()),
		Vout:api.marshJsonVout(tx.Transaction(),nil),
	},nil
}

func (api *PublicBlockAPI) marshJsonVin(tx *types.Transaction)([]json.Vin) {
	// Coinbase transactions only have a single txin by definition.
	vinList := make([]json.Vin, len(tx.TxIn))
	if blockchain.IsCoinBaseTx(tx) {
		txIn := tx.TxIn[0]
		vinEntry := &vinList[0]
		vinEntry.Coinbase = hex.EncodeToString(txIn.SignScript)
		vinEntry.Sequence = txIn.Sequence
		vinEntry.AmountIn = float64(txIn.AmountIn) //TODO coin conversion
		vinEntry.BlockHeight = txIn.BlockHeight
		vinEntry.BlockIndex = txIn.BlockTxIndex
		return vinList
	}

	for i, txIn := range tx.TxIn {
		// The disassembled string will contain [error] inline
		// if the script doesn't fully parse, so ignore the
		// error here.
		disbuf, _ := txscript.DisasmString(txIn.SignScript)

		vinEntry := &vinList[i]
		vinEntry.Txid = txIn.PreviousOut.Hash.String()
		vinEntry.Vout = txIn.PreviousOut.OutIndex
		vinEntry.Sequence = txIn.Sequence
		vinEntry.AmountIn = float64(txIn.AmountIn) //TODO coin conversion
		vinEntry.BlockHeight = txIn.BlockHeight
		vinEntry.BlockIndex = txIn.BlockTxIndex
		vinEntry.ScriptSig = &json.ScriptSig{
			Asm: disbuf,
			Hex: hex.EncodeToString(txIn.SignScript),
		}
	}
	return vinList
}

func (api *PublicBlockAPI) marshJsonVout(tx *types.Transaction,filterAddrMap map[string]struct{})([]json.Vout) {
	voutList := make([]json.Vout, 0, len(tx.TxOut))
	for _, v := range tx.TxOut {
		// The disassembled string will contain [error] inline if the
		// script doesn't fully parse, so ignore the error here.
		disbuf, _ := txscript.DisasmString(v.PkScript)


		// Ignore the error here since an error means the script
		// couldn't parse and there is no additional information
		// about it anyways.
		sc , addrs, reqSigs, _ := txscript.ExtractPkScriptAddrs(0,
				v.PkScript, api.bm.params)
		scriptClass := sc.String()

		// Encode the addresses while checking if the address passes the
		// filter when needed.
		passesFilter := len(filterAddrMap) == 0
		encodedAddrs := make([]string, len(addrs))
		for j, addr := range addrs {
			encodedAddr := addr.Encode()
			encodedAddrs[j] = encodedAddr

			// No need to check the map again if the filter already
			// passes.
			if passesFilter {
				continue
			}
			if _, exists := filterAddrMap[encodedAddr]; exists {
				passesFilter = true
			}
		}

		if !passesFilter {
			continue
		}

		var vout json.Vout
		vout.Amount = float64(v.Amount) //TODO coin conversion
		voutSPK := &vout.ScriptPubKey
		voutSPK.Addresses = encodedAddrs
		voutSPK.Asm = disbuf
		voutSPK.Hex = hex.EncodeToString(v.PkScript)
		voutSPK.Type = scriptClass
		voutSPK.ReqSigs = int32(reqSigs)
		voutList = append(voutList, vout)
	}

	return voutList
}

// RPCMarshalBlock converts the given block to the RPC output which depends on fullTx. If inclTx is true transactions are
// returned. When fullTx is true the returned block contains full transaction details, otherwise it will only contain
// transaction hashes.
func (api *PublicBlockAPI) marshalJsonBlock(b *types.SerializedBlock, inclTx bool, fullTx bool) (json.OrderedResult, error) {

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
	fields := json.OrderedResult{
		{"hash",         b.Hash().String()},
		{"confirmations",confirmations},
		{"version",      head.Version},
		{"height",       height},
		{"txRoot",       head.TxRoot.String()},
	}

	if inclTx {
		formatTx := func(tx *types.Tx) (interface{}, error) {
			return tx.Hash().String(), nil
		}
		if fullTx {
			formatTx = func(tx *types.Tx) (interface{}, error) {
				return api.marshalJsonTransaction(tx)
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
		fields = append(fields, json.KV{"transactions", transactions})
	}
	fields = append(fields, json.OrderedResult{
		{"stateRoot", head.StateRoot.String()},
		{"bits", strconv.FormatUint(uint64(head.Difficulty), 16)},
		{"difficulty", head.Difficulty},
		{"nonce", head.Nonce},
		{"timestamp", head.Timestamp.Format("2006-01-02 15:04:05.0000")},
		{"parentHash", head.ParentRoot.String()},
		{"childrenHash", nextHashString},
	}...)
	return fields, nil
}
