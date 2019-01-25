// Copyright (c) 2017-2018 The nox developers
// Copyright (c) 2013-2016 The btcsuite developers
// Copyright (c) 2017-2018 The Decred developers
// Use of this source code is governed by an ISC
// license that can be found in the LICENSE file.
package node

import (
	"bytes"
	"encoding/hex"
	"errors"
	"fmt"
	"github.com/noxproject/nox/common/hash"
	"github.com/noxproject/nox/core/address"
	"github.com/noxproject/nox/core/blockchain"
	"github.com/noxproject/nox/core/json"
	"github.com/noxproject/nox/core/message"
	"github.com/noxproject/nox/core/types"
	"github.com/noxproject/nox/database"
	"github.com/noxproject/nox/engine/txscript"
	"github.com/noxproject/nox/params"
	"github.com/noxproject/nox/rpc"
	"github.com/noxproject/nox/services/common/error"
	"github.com/noxproject/nox/services/common/marshal"
	"github.com/noxproject/nox/services/mempool"
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

// TransactionInput represents the inputs to a transaction.  Specifically a
// transaction hash and output number pair.
type TransactionInput struct {
	Txid string `json:"txid"`
	Vout uint32 `json:"vout"`
}

type Amounts map[string]float64 //{\"address\":amount,...}

func (api *PublicBlockChainAPI) CreateRawTransaction(inputs []TransactionInput,
	amounts Amounts, lockTime *int64)(interface{},error){

	// Validate the locktime, if given.
	if lockTime != nil &&
		(*lockTime < 0 || *lockTime > int64(types.MaxTxInSequenceNum)) {
		return nil, er.RpcInvalidError("Locktime out of range")
	}

	// Add all transaction inputs to a new transaction after performing
	// some validity checks.
	mtx := types.NewTransaction()
	for _, input := range inputs {
		txHash, err := hash.NewHashFromStr(input.Txid)
		if err != nil {
			return nil, er.RpcDecodeHexError(input.Txid)
		}
		prevOut := types.NewOutPoint(txHash, input.Vout)
		txIn := types.NewTxInput(prevOut, types.NullValueIn, []byte{})
		if lockTime != nil && *lockTime != 0 {
			txIn.Sequence = types.MaxTxInSequenceNum - 1
		}
		mtx.AddTxIn(txIn)
	}

		// Add all transaction outputs to the transaction after performing
	// some validity checks.
	for encodedAddr, amount := range amounts {
		// Ensure amount is in the valid range for monetary amounts.
		if amount <= 0 || amount > types.MaxAmount {
			return nil, er.RpcInvalidError("Invalid amount: 0 >= %v "+
				"> %v", amount, types.MaxAmount)
		}

		// Decode the provided address.
		addr, err := address.DecodeAddress(encodedAddr)
		if err != nil {
			return nil, er.RpcAddressKeyError("Could not decode "+
				"address: %v", err)
		}

		// Ensure the address is one of the supported types and that
		// the network encoded with the address matches the network the
		// server is currently on.
		switch addr.(type) {
		case *address.PubKeyHashAddress:
		case *address.ScriptHashAddress:
		default:
			return nil, er.RpcAddressKeyError("Invalid type: %T", addr)
		}
		if !address.IsForNetwork(addr,api.node.node.Params) {
			return nil, er.RpcAddressKeyError("Wrong network: %v",
				addr)
		}

		// Create a new script which pays to the provided address.
		pkScript, err := txscript.PayToAddrScript(addr)
		if err != nil {
			return nil, er.RpcInternalError(err.Error(),
				"Pay to address script")
		}

		atomic, err := types.NewAmount(amount)
		if err != nil {
			return nil, er.RpcInternalError(err.Error(),
				"New amount")
		}

		//TODO fix type conversion
		txOut := types.NewTxOutput(uint64(atomic), pkScript)
		mtx.AddTxOut(txOut)
	}

	// Set the Locktime, if given.
	if lockTime != nil {
		mtx.LockTime = uint32(*lockTime)
	}

	// Return the serialized and hex-encoded transaction.  Note that this
	// is intentionally not directly returning because the first return
	// value is a string and it would result in returning an empty string to
	// the client instead of nothing (nil) in the case of an error.
	mtxHex, err := marshal.MessageToHex(&message.MsgTx{mtx})
	if err != nil {
		return nil, err
	}
	return mtxHex, nil
}



func (api *PublicBlockChainAPI) DecodeRawTransaction(hexTx string)(interface{},error) {
	// Deserialize the transaction.
	hexStr := hexTx
	if len(hexStr)%2 != 0 {
		hexStr = "0" + hexStr
	}
	serializedTx, err := hex.DecodeString(hexStr)
	if err != nil {
		return nil, er.RpcDecodeHexError(hexStr)
	}
	var mtx types.Transaction
	err = mtx.Deserialize(bytes.NewReader(serializedTx))
	if err != nil {
		return nil, er.RpcDeserializationError("Could not decode Tx: %v",
			err)
	}

	log.Trace("decodeRawTx","hex",hexStr)
	log.Trace("decodeRawTx","hex",serializedTx)


	// Create and return the result.
	txReply := &json.OrderedResult{
		{"txid", mtx.TxHash().String()},
		{"txhash", mtx.TxHashFull().String()},
		{"version",  int32(mtx.Version)},
		{"locktime", mtx.LockTime},
		{"vin",      createVinList(&mtx)},
		{"vout",     createVoutList(&mtx, api.node.node.Params, nil)},
	}
	return txReply, nil
}

// createVinList returns a slice of JSON objects for the inputs of the passed
// transaction.
func createVinList(mtx *types.Transaction) []json.Vin {
	// Coinbase transactions only have a single txin by definition.
	vinList := make([]json.Vin, len(mtx.TxIn))
	if blockchain.IsCoinBaseTx(mtx) {
		txIn := mtx.TxIn[0]
		vinEntry := &vinList[0]
		vinEntry.Coinbase = hex.EncodeToString(txIn.SignScript)
		vinEntry.Sequence = txIn.Sequence
		vinEntry.AmountIn = types.Amount(txIn.AmountIn).ToCoin()
		vinEntry.BlockHeight = txIn.BlockHeight
		vinEntry.TxIndex = txIn.TxIndex
		return vinList
	}

	for i, txIn := range mtx.TxIn {
		// The disassembled string will contain [error] inline
		// if the script doesn't fully parse, so ignore the
		// error here.
		disbuf, _ := txscript.DisasmString(txIn.SignScript)

		vinEntry := &vinList[i]
		vinEntry.Txid = txIn.PreviousOut.Hash.String()
		vinEntry.Vout = txIn.PreviousOut.OutIndex
		vinEntry.Sequence = txIn.Sequence
		vinEntry.AmountIn = types.Amount(txIn.AmountIn).ToCoin()
		vinEntry.BlockHeight = txIn.BlockHeight
		vinEntry.TxIndex = txIn.TxIndex
		vinEntry.ScriptSig = &json.ScriptSig{
			Asm: disbuf,
			Hex: hex.EncodeToString(txIn.SignScript),
		}
	}

	return vinList
}

// createVoutList returns a slice of JSON objects for the outputs of the passed
// transaction.
func createVoutList(mtx *types.Transaction, params *params.Params, filterAddrMap map[string]struct{}) []json.Vout {

	voutList := make([]json.Vout, 0, len(mtx.TxOut))
	for _, v := range mtx.TxOut {
		// The disassembled string will contain [error] inline if the
		// script doesn't fully parse, so ignore the error here.
		disbuf, _ := txscript.DisasmString(v.PkScript)
		// Attempt to extract addresses from the public key script.  In
		// the case of stake submission transactions, the odd outputs
		// contain a commitment address, so detect that case
		// accordingly.
		var addrs []types.Address
		var scriptClass string
		var reqSigs int

		// Ignore the error here since an error means the script
		// couldn't parse and there is no additional information
		// about it anyways.
		var sc txscript.ScriptClass
		sc, addrs, reqSigs, _ = txscript.ExtractPkScriptAddrs(
			txscript.DefaultScriptVersion, v.PkScript, params)
			scriptClass = sc.String()

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
		voutSPK := &vout.ScriptPubKey
		vout.Amount = types.Amount(v.Amount).ToCoin()
		voutSPK.Addresses = encodedAddrs
		voutSPK.Asm = disbuf
		voutSPK.Hex = hex.EncodeToString(v.PkScript)
		voutSPK.Type = scriptClass
		voutSPK.ReqSigs = int32(reqSigs)

		voutList = append(voutList, vout)
	}

	return voutList
}

func (api *PublicBlockChainAPI) SendRawTransaction(hexTx string, allowHighFees *bool)(interface{}, error) {
	hexStr := hexTx
	highFees := false
	if allowHighFees != nil {
		highFees = *allowHighFees
	}
	if len(hexStr)%2 != 0 {
		hexStr = "0" + hexStr
	}
	serializedTx, err := hex.DecodeString(hexStr)
	if err != nil {
		return nil, er.RpcDecodeHexError(hexStr)
	}
	msgtx := types.NewTransaction()
	err = msgtx.Deserialize(bytes.NewReader(serializedTx))
	if err != nil {
		return nil, er.RpcDeserializationError("Could not decode Tx: %v",
			err)
	}

	tx := types.NewTx(msgtx)
	acceptedTxs, err := api.node.blockManager.ProcessTransaction(tx, false,
		false, highFees)
	if err != nil {
		// When the error is a rule error, it means the transaction was
		// simply rejected as opposed to something actually going
		// wrong, so log it as such.  Otherwise, something really did
		// go wrong, so log it as an actual error.  In both cases, a
		// JSON-RPC error is returned to the client with the
		// deserialization error code (to match bitcoind behavior).
		if _, ok := err.(mempool.RuleError); ok {
			err = fmt.Errorf("Rejected transaction %v: %v", tx.Hash(),
				err)
			log.Error("Failed to process transaction", "mempool.RuleError",err)
			txRuleErr, ok := err.(mempool.TxRuleError)
			if ok {
				if txRuleErr.RejectCode == message.RejectDuplicate {
					// return a dublicate tx error
					return nil, er.RpcDuplicateTxError("%v", err)
				}
			}

			// return a generic rule error
			return nil, er.RpcRuleError("%v", err)
		}

		log.Error("Failed to process transaction","err", err)
		err = fmt.Errorf("failed to process transaction %v: %v",
			tx.Hash(), err)
		return nil, er.RpcDeserializationError("rejected: %v", err)
	}
	//TODO P2P layer announce
	api.node.nfManager.AnnounceNewTransactions(acceptedTxs)

	return tx.Hash().String(), nil
}


func (api *PublicBlockChainAPI) GetRawTransaction(txHash hash.Hash, verbose bool)(interface{}, error) {

	var mtx *types.Transaction
	var blkHash *hash.Hash
	var blkHeight uint64
	var blkHashStr string
	var confirmations int64

	// Try to fetch the transaction from the memory pool and if that fails,
	// try the block database.
	tx,inRecentBlock,_ := api.node.txMemPool.FetchTransaction(&txHash, true)

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
		log.Trace("GetRawTx","hex",hex.EncodeToString(txBytes))
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
			hexStr, err := marshal.MessageToHex(&message.MsgTx{tx.Transaction()})
			if err != nil {
				return nil, err
			}

			return hexStr, nil
		}

		mtx = tx.Transaction()
	}


	if blkHash != nil {
		blkHashStr = blkHash.String()
	}
	if inRecentBlock{
		blkHeight = api.node.blockManager.GetChain().BestSnapshot().Height
		confirmations = 1
	}else if tx != nil {
		confirmations = 0
	}else {
		confirmations = 1 + int64(api.node.blockManager.GetChain().BestSnapshot().Height - blkHeight)
	}

	return marshal.MarshalJsonTransaction(mtx,api.node.node.Params,blkHeight,blkHashStr,confirmations)
}


// Returns information about an unspent transaction output
// 1. txid           (string, required)                The hash of the transaction
// 2. vout           (numeric, required)               The index of the output
// 3. includemempool (boolean, optional, default=true) Include the mempool when true
//
//Result:
//{
// "bestblock": "value",        (string)          The block hash that contains the transaction output
// "confirmations": n,          (numeric)         The number of confirmations
// "amount": n.nnn,             (numeric)         The transaction amount
// "scriptPubKey": {            (object)          The public key script used to pay coins as a JSON object
//  "asm": "value",             (string)          Disassembly of the script
//  "hex": "value",             (string)          Hex-encoded bytes of the script
//  "reqSigs": n,               (numeric)         The number of required signatures
//  "type": "value",            (string)          The type of the script (e.g. 'pubkeyhash')
//  "addresses": ["value",...], (array of string) The nox addresses associated with this script
// },
// "coinbase": true|false,      (boolean)         Whether or not the transaction is a coinbase
//}
func (api *PublicBlockChainAPI) GetUtxo(txHash hash.Hash, vout uint32, includeMempool *bool )(interface{}, error) {

	// If requested and the tx is available in the mempool try to fetch it
	// from there, otherwise attempt to fetch from the block database.
	var bestBlockHash string
	var confirmations int64
	var txVersion uint32
	var amount uint64
	var pkScript []byte
	var isCoinbase bool

	// by default try to search mempool tx
	includeMempoolTx := true
	if includeMempool != nil {
		includeMempoolTx = *includeMempool
	}

	var txFromMempool *types.Tx

	// try mempool by default
	if includeMempoolTx {
		txFromMempool, inRecentBlock, _ := api.node.txMemPool.FetchTransaction(&txHash,true)
		if txFromMempool != nil {
			tx := txFromMempool.Transaction()
			txOut := tx.TxOut[vout]
			if txOut == nil {
				return nil,nil
			}
			best := api.node.blockManager.GetChain().BestSnapshot()
			bestBlockHash = best.Hash.String()
			if inRecentBlock {
				confirmations = 1
			}else {
				confirmations = 0
			}
			txVersion = tx.Version
			amount = txOut.Amount
			pkScript = txOut.PkScript
			isCoinbase = blockchain.IsCoinBaseTx(tx)
		}
	}

	// otherwise try to lookup utxo set
	if txFromMempool == nil {
		entry, _:= api.node.blockManager.GetChain().FetchUtxoEntry(&txHash)
		if entry == nil || entry.IsOutputSpent(vout) {
			return nil,nil
		}
		best := api.node.blockManager.GetChain().BestSnapshot()
		bestBlockHash = best.Hash.String()
		confirmations = 1 + int64(best.Height - entry.BlockHeight())
		txVersion = entry.TxVersion()
		amount = entry.AmountByIndex(vout)
		pkScript = entry.PkScriptByIndex(vout)
		isCoinbase = entry.IsCoinBase()
	}

	// Disassemble script into single line printable format.  The
	// disassembled string will contain [error] inline if the script
	// doesn't fully parse, so ignore the error here.
	script := pkScript
	disbuf, _ := txscript.DisasmString(script)

	// Get further info about the script.  Ignore the error here since an
	// error means the script couldn't parse and there is no additional
	// information about it anyways.
	scriptClass, addrs, reqSigs, _ := txscript.ExtractPkScriptAddrs(txscript.DefaultScriptVersion,
		script, api.node.node.Params)
	addresses := make([]string, len(addrs))
	for i, addr := range addrs {
		addresses[i] = addr.Encode()
	}

	txOutReply := &json.GetUtxoResult{
		BestBlock:     bestBlockHash,
		Confirmations: confirmations,
		Amount:         types.Amount(amount).ToUnit(types.AmountCoin),
		Version:       int32(txVersion),
		ScriptPubKey: json.ScriptPubKeyResult{
			Asm:       disbuf,
			Hex:       hex.EncodeToString(pkScript),
			ReqSigs:   int32(reqSigs),
			Type:      scriptClass.String(),
			Addresses: addresses,
		},
		Coinbase: isCoinbase,
	}
	return txOutReply, nil
}
