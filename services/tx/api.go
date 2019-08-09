package tx

import (
	"bytes"
	"encoding/hex"
	"errors"
	"fmt"
	"github.com/HalalChain/qitmeer-lib/common/hash"
	"github.com/HalalChain/qitmeer-lib/common/marshal"
	"github.com/HalalChain/qitmeer-lib/core/address"
	"github.com/HalalChain/qitmeer-lib/core/json"
	"github.com/HalalChain/qitmeer-lib/core/message"
	"github.com/HalalChain/qitmeer-lib/core/protocol"
	"github.com/HalalChain/qitmeer-lib/core/types"
	"github.com/HalalChain/qitmeer-lib/crypto/ecc"
	"github.com/HalalChain/qitmeer-lib/engine/txscript"
	"github.com/HalalChain/qitmeer-lib/params"
	"github.com/HalalChain/qitmeer-lib/rpc"
	"github.com/HalalChain/qitmeer/database"
	"github.com/HalalChain/qitmeer/services/mempool"
	"strconv"
)

func (tm *TxManager) APIs() []rpc.API {
	return []rpc.API{
		{
			NameSpace: rpc.DefaultServiceNameSpace,
			Service:   NewPublicTxAPI(tm),
			Public:    true,
		},
		tm.txMemPool.API(),
	}
}

type PublicTxAPI struct {
	txManager *TxManager
}

func NewPublicTxAPI(tm *TxManager) *PublicTxAPI {
	ptapi:=PublicTxAPI{tm}
	return &ptapi
}

// TransactionInput represents the inputs to a transaction.  Specifically a
// transaction hash and output number pair.
type TransactionInput struct {
	Txid string `json:"txid"`
	Vout uint32 `json:"vout"`
}

type Amounts map[string]float64 //{\"address\":amount,...}


func (api *PublicTxAPI) CreateRawTransaction(inputs []TransactionInput,
	amounts Amounts, lockTime *int64) (interface{}, error) {

	// Validate the locktime, if given.
	if lockTime != nil &&
		(*lockTime < 0 || *lockTime > int64(types.MaxTxInSequenceNum)) {
		return nil, rpc.RpcInvalidError("Locktime out of range")
	}

	// Add all transaction inputs to a new transaction after performing
	// some validity checks.
	mtx := types.NewTransaction()
	for _, input := range inputs {
		txHash, err := hash.NewHashFromStr(input.Txid)
		if err != nil {
			return nil, rpc.RpcDecodeHexError(input.Txid)
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
			return nil, rpc.RpcInvalidError("Invalid amount: 0 >= %v "+
				"> %v", amount, types.MaxAmount)
		}

		// Decode the provided address.
		addr, err := address.DecodeAddress(encodedAddr)
		if err != nil {
			return nil, rpc.RpcAddressKeyError("Could not decode "+
				"address: %v", err)
		}

		// Ensure the address is one of the supported types and that
		// the network encoded with the address matches the network the
		// server is currently on.
		switch addr.(type) {
		case *address.PubKeyHashAddress:
		case *address.ScriptHashAddress:
		default:
			return nil, rpc.RpcAddressKeyError("Invalid type: %T", addr)
		}
		if !address.IsForNetwork(addr, api.txManager.bm.ChainParams()) {
			return nil, rpc.RpcAddressKeyError("Wrong network: %v",
				addr)
		}

		// Create a new script which pays to the provided address.
		pkScript, err := txscript.PayToAddrScript(addr)
		if err != nil {
			return nil, rpc.RpcInternalError(err.Error(),
				"Pay to address script")
		}

		atomic, err := types.NewAmount(amount)
		if err != nil {
			return nil, rpc.RpcInternalError(err.Error(),
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
	mtxHex, err := marshal.MessageToHex(&message.MsgTx{Tx:mtx})
	if err != nil {
		return nil, err
	}
	return mtxHex, nil
}


func (api *PublicTxAPI) DecodeRawTransaction(hexTx string) (interface{}, error) {
	// Deserialize the transaction.
	hexStr := hexTx
	if len(hexStr)%2 != 0 {
		hexStr = "0" + hexStr
	}
	serializedTx, err := hex.DecodeString(hexStr)
	if err != nil {
		return nil, rpc.RpcDecodeHexError(hexStr)
	}
	var mtx types.Transaction
	err = mtx.Deserialize(bytes.NewReader(serializedTx))
	if err != nil {
		return nil, rpc.RpcDeserializationError("Could not decode Tx: %v",
			err)
	}

	log.Trace("decodeRawTx", "hex", hexStr)
	log.Trace("decodeRawTx", "hex", serializedTx)

	// Create and return the result.
	txReply := &json.OrderedResult{
		{Key:"txid", Val:mtx.TxHash().String()},
		{Key:"txhash", Val:mtx.TxHashFull().String()},
		{Key:"version", Val:int32(mtx.Version)},
		{Key:"locktime", Val:mtx.LockTime},
		{Key:"vin", Val:createVinList(&mtx)},
		{Key:"vout", Val:createVoutList(&mtx, api.txManager.bm.ChainParams(), nil)},
	}
	return txReply, nil
}

// createVinList returns a slice of JSON objects for the inputs of the passed
// transaction.
func createVinList(mtx *types.Transaction) []json.Vin {
	// Coinbase transactions only have a single txin by definition.
	vinList := make([]json.Vin, len(mtx.TxIn))
	if mtx.IsCoinBaseTx() {
		txIn := mtx.TxIn[0]
		vinEntry := &vinList[0]
		vinEntry.Coinbase = hex.EncodeToString(txIn.SignScript)
		vinEntry.Sequence = txIn.Sequence
		vinEntry.AmountIn = types.Amount(txIn.AmountIn).ToCoin()
		vinEntry.BlockHeight = txIn.BlockOrder
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
		vinEntry.BlockHeight = txIn.BlockOrder
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
		sc, addrs, reqSigs, _ = txscript.ExtractPkScriptAddrs(v.PkScript, params)
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

func (api *PublicTxAPI) SendRawTransaction(hexTx string, allowHighFees *bool) (interface{}, error) {
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
		return nil, rpc.RpcDecodeHexError(hexStr)
	}
	msgtx := types.NewTransaction()
	err = msgtx.Deserialize(bytes.NewReader(serializedTx))
	if err != nil {
		return nil, rpc.RpcDeserializationError("Could not decode Tx: %v",
			err)
	}

	tx := types.NewTx(msgtx)
	acceptedTxs, err := api.txManager.bm.ProcessTransaction(tx, false,
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
			log.Error("Failed to process transaction", "mempool.RuleError", err)
			txRuleErr, ok := err.(mempool.TxRuleError)
			if ok {
				if txRuleErr.RejectCode == message.RejectDuplicate {
					// return a dublicate tx error
					return nil, rpc.RpcDuplicateTxError("%v", err)
				}
			}

			// return a generic rule error
			return nil, rpc.RpcRuleError("%v", err)
		}

		log.Error("Failed to process transaction", "err", err)
		err = fmt.Errorf("failed to process transaction %v: %v",
			tx.Hash(), err)
		return nil, rpc.RpcDeserializationError("rejected: %v", err)
	}
	//TODO P2P layer announce
	api.txManager.ntmgr.AnnounceNewTransactions(acceptedTxs)

	return tx.Hash().String(), nil
}

func (api *PublicTxAPI) GetRawTransaction(txHash hash.Hash, verbose bool) (interface{}, error) {

	var mtx *types.Transaction
	var blkHash *hash.Hash
	var blkOrder uint64
	var blkHashStr string
	var confirmations int64

	// Try to fetch the transaction from the memory pool and if that fails,
	// try the block database.
	tx, _ := api.txManager.txMemPool.FetchTransaction(&txHash)

	if tx == nil {
		//not found from mem-pool, try db
		txIndex := api.txManager.txIndex
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
			return nil, rpc.RpcNoTxInfoError(&txHash)
		}

		// Load the raw transaction bytes from the database.
		var txBytes []byte
		err = api.txManager.db.View(func(dbTx database.Tx) error {
			var err error
			txBytes, err = dbTx.FetchBlockRegion(blockRegion)
			return err
		})
		if err != nil {
			return nil, rpc.RpcNoTxInfoError(&txHash)
		}

		// When the verbose flag isn't set, simply return the serialized
		// transaction as a hex-encoded string.  This is done here to
		// avoid deserializing it only to reserialize it again later.
		if !verbose {
			return hex.EncodeToString(txBytes), nil
		}

		// Grab the block height.
		blkHash = blockRegion.Hash
		blkOrder, err = api.txManager.bm.GetChain().BlockOrderByHash(blkHash)
		if err != nil {
			context := "Failed to retrieve block height"
			return nil, rpc.RpcInternalError(err.Error(), context)
		}

		// Deserialize the transaction
		var msgTx types.Transaction
		err = msgTx.Deserialize(bytes.NewReader(txBytes))
		log.Trace("GetRawTx", "hex", hex.EncodeToString(txBytes))
		if err != nil {
			context := "Failed to deserialize transaction"
			return nil, rpc.RpcInternalError(err.Error(), context)
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
			hexStr, err := marshal.MessageToHex(&message.MsgTx{Tx:tx.Transaction()})
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
	if tx != nil {
		confirmations = 0
	} else {
		confirmations = 1 + int64(api.txManager.bm.GetChain().BestSnapshot().Order-blkOrder)
	}

	return marshal.MarshalJsonTransaction(mtx, api.txManager.bm.ChainParams(), blkOrder, blkHashStr, confirmations)
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
//  "addresses": ["value",...], (array of string) The qitmeer addresses associated with this script
// },
// "coinbase": true|false,      (boolean)         Whether or not the transaction is a coinbase
//}
func (api *PublicTxAPI) GetUtxo(txHash hash.Hash, vout uint32, includeMempool *bool) (interface{}, error) {

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

	// try mempool by default
	if includeMempoolTx {
		txFromMempool, _ := api.txManager.txMemPool.FetchTransaction(&txHash)
		if txFromMempool != nil {
			tx := txFromMempool.Transaction()
			txOut := tx.TxOut[vout]
			if txOut == nil {
				return nil, nil
			}
			best := api.txManager.bm.GetChain().BestSnapshot()
			bestBlockHash = best.Hash.String()
			confirmations = 0
			txVersion = tx.Version
			amount = txOut.Amount
			pkScript = txOut.PkScript
			isCoinbase = tx.IsCoinBaseTx()
		}
	}

	// otherwise try to lookup utxo set
	if bestBlockHash == "" {
		entry, _ := api.txManager.bm.GetChain().FetchUtxoEntry(&txHash)
		if entry == nil || entry.IsOutputSpent(vout) {
			return nil, nil
		}
		best := api.txManager.bm.GetChain().BestSnapshot()
		bestBlockHash = best.Hash.String()
		confirmations = 1 + int64(best.Order-entry.BlockOrder())
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
	scriptClass, addrs, reqSigs, _ := txscript.ExtractPkScriptAddrs(script, api.txManager.bm.ChainParams())
	addresses := make([]string, len(addrs))
	for i, addr := range addrs {
		addresses[i] = addr.Encode()
	}

	txOutReply := &json.GetUtxoResult{
		BestBlock:     bestBlockHash,
		Confirmations: confirmations,
		Amount:        types.Amount(amount).ToUnit(types.AmountCoin),
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

//TODO: I will remove it soon
func (api *PublicTxAPI) TxSign(privkeyStr string, rawTxStr string) (interface{}, error) {
	privkeyByte, err := hex.DecodeString(privkeyStr)
	if err != nil {
	}
	if len(privkeyByte) != 32 {
	}
	privateKey, pubKey := ecc.Secp256k1.PrivKeyFromBytes(privkeyByte)
	h160 := hash.Hash160(pubKey.SerializeCompressed())

	var param *params.Params= &params.PrivNetParams
	addr, err := address.NewPubKeyHashAddress(h160, param, ecc.ECDSA_Secp256k1)
	if err != nil {
	}
	// Create a new script which pays to the provided address.
	pkScript, err := txscript.PayToAddrScript(addr)
	if err != nil {
	}

	if len(rawTxStr)%2 != 0 {
	}
	serializedTx, err := hex.DecodeString(rawTxStr)
	if err != nil {
	}

	var redeemTx types.Transaction
	err = redeemTx.Deserialize(bytes.NewReader(serializedTx))
	if err != nil {
	}
	var kdb txscript.KeyClosure= func(types.Address) (ecc.PrivateKey, bool, error) {
		return privateKey, true, nil // compressed is true
	}
	var sigScripts [][]byte
	for i:= range redeemTx.TxIn {
		sigScript, err := txscript.SignTxOutput(param, &redeemTx, i, pkScript, txscript.SigHashAll, kdb, nil, nil, ecc.ECDSA_Secp256k1)
		if err != nil {
		}
		sigScripts = append(sigScripts, sigScript)
	}

	for i2:= range sigScripts {
		redeemTx.TxIn[i2].SignScript = sigScripts[i2]
	}

	mtxHex, err := marshal.MessageToHex(&message.MsgTx{Tx:&redeemTx})
	if err != nil {
	}
	return mtxHex, nil
}

// handleSearchRawTransactions implements the searchrawtransactions command.
func (api *PublicTxAPI) GetRawTransactions(addre string,vinext *bool,count *uint,skip *uint,revers *bool,verbose *bool,filterAddrs *[]string) (interface{}, error) {
	addrIndex := api.txManager.addrIndex
	if addrIndex == nil {
		return nil, fmt.Errorf("Address index must be enabled (--addrindex)")
	}
	vinExtra:=false
	if vinext != nil {
		vinExtra = *vinext
	}

	if vinExtra && api.txManager.txIndex == nil {
		return nil,fmt.Errorf("Transaction index must be enabled (--txindex)")
	}
	params := api.txManager.bm.ChainParams()
	addr, err := address.DecodeAddress(addre)
	if err != nil {
		return nil, fmt.Errorf("Invalid address or key: " + err.Error())
	}
	numRequested := uint(100)
	if count != nil {
		numRequested = *count
	}
	if numRequested == 0 {
		return nil, nil
	}

	var numToSkip uint
	if skip != nil {
		numToSkip = *skip
	}

	var reverse bool
	if revers != nil {
		reverse = *revers
	}

	//
	numSkipped := uint32(0)
	addressTxns := make([]retrievedTx, 0, numRequested)
	if reverse {
		mpTxns, mpSkipped := api.fetchMempoolTxnsForAddress(addr,
			uint32(numToSkip), uint32(numRequested))
		numSkipped += mpSkipped
		for _, tx := range mpTxns {
			addressTxns = append(addressTxns, retrievedTx{tx: tx})
		}
	}

	// Fetch transactions from the database in the desired order if more are
	// needed.
	if uint(len(addressTxns)) < numRequested {
		err = api.txManager.db.View(func(dbTx database.Tx) error {
			regions, dbSkipped, err := addrIndex.TxRegionsForAddress(
				dbTx, addr, uint32(numToSkip)-numSkipped,
				uint32(numRequested-uint(len(addressTxns))), reverse)
			if err != nil {
				return err
			}

			// Load the raw transaction bytes from the database.
			serializedTxns, err := dbTx.FetchBlockRegions(regions)
			if err != nil {
				return err
			}

			// Add the transaction and the hash of the block it is
			// contained in to the list.  Note that the transaction
			// is left serialized here since the caller might have
			// requested non-verbose output and hence there would be
			// no point in deserializing it just to reserialize it
			// later.
			for i, serializedTx := range serializedTxns {
				addressTxns = append(addressTxns, retrievedTx{
					txBytes: serializedTx,
					blkHash: regions[i].Hash,
				})
			}
			numSkipped += dbSkipped

			return nil
		})
		if err != nil {
			context := "Failed to load address index entries"
			return nil, fmt.Errorf("%s %s",err.Error(), context)
		}

	}

	// Add transactions from mempool last if client did not request reverse
	// order and the number of results is still under the number requested.
	if !reverse && uint(len(addressTxns)) < numRequested {
		// Transactions in the mempool are not in a block header yet,
		// so the block header field in the retieved transaction struct
		// is left nil.
		mpTxns, mpSkipped := api.fetchMempoolTxnsForAddress(addr,
			uint32(numToSkip)-numSkipped, uint32(numRequested-
				uint(len(addressTxns))))
		numSkipped += mpSkipped
		for _, tx := range mpTxns {
			addressTxns = append(addressTxns, retrievedTx{tx: tx})
		}
	}

	// Address has never been used if neither source yielded any results.
	if len(addressTxns) == 0 {
		return nil, fmt.Errorf("No information available about address")
	}

	// Serialize all of the transactions to hex.
	hexTxns := make([]string, len(addressTxns))
	for i := range addressTxns {
		// Simply encode the raw bytes to hex when the retrieved
		// transaction is already in serialized form.
		rtx := &addressTxns[i]
		if rtx.txBytes != nil {
			hexTxns[i] = hex.EncodeToString(rtx.txBytes)
			continue
		}

		// Serialize the transaction first and convert to hex when the
		// retrieved transaction is the deserialized structure.
		hexTxns[i], err = marshal.MessageToHex(&message.MsgTx{Tx:rtx.tx.Tx})
		if err != nil {
			return nil, err
		}
	}

	// When not in verbose mode, simply return a list of serialized txns.
	if verbose != nil && !(*verbose) {
		return hexTxns, nil
	}

	// Normalize the provided filter addresses (if any) to ensure there are
	// no duplicates.
	filterAddrMap := make(map[string]struct{})
	if filterAddrs != nil && len(*filterAddrs) > 0 {
		for _, addr := range *filterAddrs {
			filterAddrMap[addr] = struct{}{}
		}
	}

	// The verbose flag is set, so generate the JSON object and return it.
	best := api.txManager.bm.GetChain().BestSnapshot()
	srtList := make([]json.GetRawTransactionsResult, len(addressTxns))
	for i := range addressTxns {
		// The deserialized transaction is needed, so deserialize the
		// retrieved transaction if it's in serialized form (which will
		// be the case when it was lookup up from the database).
		// Otherwise, use the existing deserialized transaction.
		rtx := &addressTxns[i]
		var mtx *message.MsgTx
		if rtx.tx == nil {
			// Deserialize the transaction.
			mtx = new(message.MsgTx)
			err := mtx.Decode(bytes.NewReader(rtx.txBytes),protocol.ProtocolVersion)
			if err != nil {
				context := "Failed to deserialize transaction"
				return nil, fmt.Errorf("%s %s",err.Error(), context)
			}
		} else {
			mtx = &message.MsgTx{Tx:rtx.tx.Tx}
		}

		result := &srtList[i]
		result.Hex = hexTxns[i]
		result.Txid = mtx.Tx.TxHash().String()
		result.Vin, err = api.createVinListPrevOut(mtx, params, vinExtra,
			filterAddrMap)
		if err != nil {
			return nil, err
		}
		result.Vout = createVoutList(mtx.Tx, params, filterAddrMap)
		result.Version = mtx.Tx.Version
		result.LockTime = mtx.Tx.LockTime

		// Transactions grabbed from the mempool aren't yet in a block,
		// so conditionally fetch block details here.  This will be
		// reflected in the final JSON output (mempool won't have
		// confirmations or block information).
		var blkHeader *types.BlockHeader
		var blkHashStr string
		var blkOrder uint64
		if blkHash := rtx.blkHash; blkHash != nil {
			// Fetch the header from chain.
			header, err := api.txManager.bm.GetChain().HeaderByHash(blkHash)
			if err != nil {
				return nil, rpc.RpcInternalError("Block not found","")
			}

			// Get the block height from chain.

			order, err := api.txManager.bm.GetChain().BlockOrderByHash(blkHash)
			if err != nil {
				context := "Failed to obtain block height"
				return nil, rpc.RpcInternalError(err.Error(), context)
			}

			blkHeader = &header
			blkHashStr = blkHash.String()
			blkOrder = order
		}

		// Add the block information to the result if there is any.
		if blkHeader != nil {
			// This is not a typo, they are identical in Bitcoin
			// Core as well.
			result.Time = blkHeader.Timestamp.Unix()
			result.Blocktime = blkHeader.Timestamp.Unix()
			result.BlockHash = blkHashStr
			result.Confirmations = uint64(1 + best.Order - blkOrder)
		}
	}

	return srtList, nil
}

func (api *PublicTxAPI) fetchMempoolTxnsForAddress(addr types.Address, numToSkip, numRequested uint32) ([]*types.Tx, uint32) {
	// There are no entries to return when there are less available than the
	// number being skipped.
	mpTxns := api.txManager.addrIndex.UnconfirmedTxnsForAddress(addr)
	numAvailable := uint32(len(mpTxns))
	if numToSkip > numAvailable {
		return nil, numAvailable
	}

	// Filter the available entries based on the number to skip and number
	// requested.
	rangeEnd := numToSkip + numRequested
	if rangeEnd > numAvailable {
		rangeEnd = numAvailable
	}
	return mpTxns[numToSkip:rangeEnd], numToSkip
}

type retrievedTx struct {
	txBytes []byte
	blkHash *hash.Hash // Only set when transaction is in a block.
	tx      *types.Tx
}

func (api *PublicTxAPI) createVinListPrevOut(mtx *message.MsgTx, chainParams *params.Params, vinExtra bool, filterAddrMap map[string]struct{}) ([]json.VinPrevOut, error) {
	// Coinbase transactions only have a single txin by definition.
	if mtx.Tx.IsCoinBaseTx() {
		// Only include the transaction if the filter map is empty
		// because a coinbase input has no addresses and so would never
		// match a non-empty filter.
		if len(filterAddrMap) != 0 {
			return nil, nil
		}

		txIn := mtx.Tx.TxIn[0]
		vinList := make([]json.VinPrevOut, 1)
		vinList[0].Coinbase = hex.EncodeToString(txIn.SignScript)
		vinList[0].Sequence = txIn.Sequence
		return vinList, nil
	}

	// Use a dynamically sized list to accommodate the address filter.
	vinList := make([]json.VinPrevOut, 0, len(mtx.Tx.TxIn))

	// Lookup all of the referenced transaction outputs needed to populate
	// the previous output information if requested.
	var originOutputs map[types.TxOutPoint]types.TxOutput
	if vinExtra || len(filterAddrMap) > 0 {
		var err error
		originOutputs, err = api.fetchInputTxos(mtx)
		if err != nil {
			return nil, err
		}
	}

	for _, txIn := range mtx.Tx.TxIn {
		// The disassembled string will contain [error] inline
		// if the script doesn't fully parse, so ignore the
		// error here.
		disbuf, _ := txscript.DisasmString(txIn.SignScript)

		// Create the basic input entry without the additional optional
		// previous output details which will be added later if
		// requested and available.
		prevOut := &txIn.PreviousOut
		vinEntry := json.VinPrevOut{
			Txid:     prevOut.Hash.String(),
			Vout:     prevOut.OutIndex,
			Sequence: txIn.Sequence,
			ScriptSig: &json.ScriptSig{
				Asm: disbuf,
				Hex: hex.EncodeToString(txIn.SignScript),
			},
		}

		// Add the entry to the list now if it already passed the filter
		// since the previous output might not be available.
		passesFilter := len(filterAddrMap) == 0
		if passesFilter {
			vinList = append(vinList, vinEntry)
		}

		// Only populate previous output information if requested and
		// available.
		if len(originOutputs) == 0 {
			continue
		}
		originTxOut, ok := originOutputs[*prevOut]
		if !ok {
			continue
		}

		// Ignore the error here since an error means the script
		// couldn't parse and there is no additional information about
		// it anyways.
		_, addrs, _, _ := txscript.ExtractPkScriptAddrs(originTxOut.PkScript, chainParams)

		// Encode the addresses while checking if the address passes the
		// filter when needed.
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

		// Ignore the entry if it doesn't pass the filter.
		if !passesFilter {
			continue
		}

		// Add entry to the list if it wasn't already done above.
		if len(filterAddrMap) != 0 {
			vinList = append(vinList, vinEntry)
		}

		// Update the entry with previous output information if
		// requested.
		if vinExtra {
			vinListEntry := &vinList[len(vinList)-1]
			vinListEntry.PrevOut = &json.PrevOut{
				Addresses: encodedAddrs,
				Value:     types.Amount(originTxOut.Amount).ToCoin(),
			}
		}
	}

	return vinList, nil
}

func (api *PublicTxAPI) fetchInputTxos(tx *message.MsgTx) (map[types.TxOutPoint]types.TxOutput, error) {
	mp := api.txManager.txMemPool
	originOutputs := make(map[types.TxOutPoint]types.TxOutput)
	for txInIndex, txIn := range tx.Tx.TxIn {
		// Attempt to fetch and use the referenced transaction from the
		// memory pool.
		origin := &txIn.PreviousOut
		originTx, err := mp.FetchTransaction(&origin.Hash)
		if err == nil {
			txOuts := originTx.Tx.TxOut
			if origin.OutIndex >= uint32(len(txOuts)) {
				return nil, fmt.Errorf("unable to find output "+
					"%v referenced from transaction %s:%d",
					origin, tx.Tx.TxHash(), txInIndex)
			}

			originOutputs[*origin] = *txOuts[origin.OutIndex]
			continue
		}

		// Look up the location of the transaction.
		blockRegion, err := api.txManager.txIndex.TxBlockRegion(origin.Hash)
		if err != nil {
			context := "Failed to retrieve transaction location"
			return nil, rpc.RpcInternalError(err.Error(), context)
		}
		if blockRegion == nil {
			return nil, rpc.RpcNoTxInfoError(&origin.Hash)
		}

		// Load the raw transaction bytes from the database.
		var txBytes []byte
		err = api.txManager.db.View(func(dbTx database.Tx) error {
			var err error
			txBytes, err = dbTx.FetchBlockRegion(blockRegion)
			return err
		})
		if err != nil {
			return nil, rpc.RpcNoTxInfoError(&origin.Hash)
		}

		// Deserialize the transaction
		var msgTx message.MsgTx
		err = msgTx.Decode(bytes.NewReader(txBytes),protocol.ProtocolVersion)
		if err != nil {
			context := "Failed to deserialize transaction"
			return nil, rpc.RpcInternalError(err.Error(), context)
		}

		// Add the referenced output to the map.
		if origin.OutIndex >= uint32(len(msgTx.Tx.TxOut)) {
			errStr := fmt.Sprintf("unable to find output %v "+
				"referenced from transaction %s:%d", origin,
				tx.Tx.TxHash(), txInIndex)
			return nil, rpc.RpcInternalError(errStr, "")
		}
		originOutputs[*origin] = *msgTx.Tx.TxOut[origin.OutIndex]
	}

	return originOutputs, nil
}

// IsInvalidTx
func (api *PublicTxAPI) IsInvalidTx(h hash.Hash) (interface{}, error) {
	return strconv.FormatBool(api.txManager.IsInvalidTx(&h)),nil
}