package marshal

import (
	"bytes"
	"encoding/hex"
	"errors"
	"fmt"
	"strconv"
	"qitmeer/core/blockchain"
	"qitmeer/core/json"
	"qitmeer/core/message"
	"qitmeer/core/protocol"
	"qitmeer/core/types"
	"qitmeer/engine/txscript"
	"qitmeer/params"
	"qitmeer/services/common/error"
	"qitmeer/common/hash"
)

// messageToHex serializes a message to the wire protocol encoding using the
// latest protocol version and returns a hex-encoded string of the result.
func MessageToHex(msg message.Message) (string, error) {
	var buf bytes.Buffer
	if err := msg.Encode(&buf, protocol.ProtocolVersion); err != nil {
		context := fmt.Sprintf("Failed to encode msg of type %T", msg)
		return "", er.RpcInternalError(err.Error(), context)
	}

	return hex.EncodeToString(buf.Bytes()), nil
}

func MarshalJsonTx(tx *types.Tx, params *params.Params, blkHeight uint64,blkHashStr string,
	confirmations int64) (json.TxRawResult, error){
	if tx == nil {
		return json.TxRawResult{}, errors.New("can't marshal nil transaction")
	}
	return MarshalJsonTransaction(tx.Transaction(), params, blkHeight,blkHashStr, confirmations)
}


func MarshalJsonTransaction(tx *types.Transaction, params *params.Params, blkHeight uint64,blkHashStr string,
	confirmations int64) (json.TxRawResult, error){

	hexStr, err := MessageToHex(&message.MsgTx{Tx:tx})
	if err!=nil {
		return json.TxRawResult{}, err
	}

	bufWit,_ :=tx.Serialize(types.TxSerializeOnlyWitness)
	hexStrWit:=hex.EncodeToString(bufWit)

	bufNoWit,_:=tx.Serialize(types.TxSerializeNoWitness)
	hexStrNoWit:=hex.EncodeToString(bufNoWit)

	//TODO, handle the blkHeight/blkHash

	return json.TxRawResult{
		Hex : hexStr,
		HexNoWit: hexStrNoWit,
		HexWit : hexStrWit,
		Txid : tx.TxHash().String(),
		TxHash : tx.TxHashFull().String(),
		Version:tx.Version,
		LockTime:tx.LockTime,
		Expire:tx.Expire,
		Vin: MarshJsonVin(tx),
		Vout:MarshJsonVout(tx,nil, params),
		BlockHash:blkHashStr,
		BlockHeight:blkHeight,
		Confirmations: confirmations,
	},nil

}

func  MarshJsonVin(tx *types.Transaction)([]json.Vin) {
	// Coinbase transactions only have a single txin by definition.
	vinList := make([]json.Vin, len(tx.TxIn))
	if blockchain.IsCoinBaseTx(tx) {
		txIn := tx.TxIn[0]
		vinEntry := &vinList[0]
		vinEntry.Coinbase = hex.EncodeToString(txIn.SignScript)
		vinEntry.Sequence = txIn.Sequence
		vinEntry.AmountIn = float64(txIn.AmountIn) //TODO coin conversion
		vinEntry.BlockHeight = txIn.BlockHeight
		vinEntry.TxIndex = txIn.TxIndex
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
		vinEntry.TxIndex = txIn.TxIndex
		vinEntry.ScriptSig = &json.ScriptSig{
			Asm: disbuf,
			Hex: hex.EncodeToString(txIn.SignScript),
		}
	}
	return vinList
}

func  MarshJsonVout(tx *types.Transaction,filterAddrMap map[string]struct{}, params *params.Params)([]json.Vout) {
	voutList := make([]json.Vout, 0, len(tx.TxOut))
	for _, v := range tx.TxOut {
		// The disassembled string will contain [error] inline if the
		// script doesn't fully parse, so ignore the error here.
		disbuf, _ := txscript.DisasmString(v.PkScript)


		// Ignore the error here since an error means the script
		// couldn't parse and there is no additional information
		// about it anyways.
		sc , addrs, reqSigs, _ := txscript.ExtractPkScriptAddrs(0,
				v.PkScript, params)
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
func MarshalJsonBlock(b *types.SerializedBlock, inclTx bool, fullTx bool,
	params *params.Params, confirmations int64,children []*hash.Hash) (json.OrderedResult, error) {

	head := b.Block().Header // copies the header once
	// Get next block hash unless there are none.
	height := uint64(b.Height())

	fields := json.OrderedResult{
		{Key:"hash",         Val:b.Hash().String()},
		{Key:"confirmations",Val:confirmations},
		{Key:"version",      Val:head.Version},
		{Key:"height",       Val:height},
		{Key:"txRoot",       Val:head.TxRoot.String()},
	}

	if inclTx {
		formatTx := func(tx *types.Tx) (interface{}, error) {
			return tx.Hash().String(), nil
		}
		if fullTx {
			formatTx = func(tx *types.Tx) (interface{}, error) {
				return MarshalJsonTx(tx,params,height,"",confirmations)
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
		fields = append(fields, json.KV{Key:"transactions", Val:transactions})
	}
	fields = append(fields, json.OrderedResult{
		{Key:"stateRoot", Val:head.StateRoot.String()},
		{Key:"bits", Val:strconv.FormatUint(uint64(head.Difficulty), 16)},
		{Key:"difficulty", Val:head.Difficulty},
		{Key:"nonce", Val:head.Nonce},
		{Key:"timestamp", Val:head.Timestamp.Format("2006-01-02 15:04:05.0000")},
	}...)
	tempArr:=[]string{}
	if b.Block().Parents!=nil&&len(b.Block().Parents)>0 {

		for i:=0;i<len(b.Block().Parents);i++  {
			tempArr=append(tempArr,b.Block().Parents[i].String())
		}
	}else {
		tempArr=append(tempArr,"null")
	}
	fields = append(fields, json.KV{Key:"parents", Val:tempArr})

	tempArr=[]string{}
	if len(children)>0 {

		for i:=0;i<len(children);i++  {
			tempArr=append(tempArr,children[i].String())
		}
	}else {
		tempArr=append(tempArr,"null")
	}
	fields = append(fields, json.KV{Key:"children", Val:tempArr})

	return fields, nil
}
