package marshal

import (
	"bytes"
	"encoding/hex"
	"errors"
	"fmt"
	"strconv"
	"github.com/noxproject/nox/core/blockchain"
	"github.com/noxproject/nox/core/json"
	"github.com/noxproject/nox/core/message"
	"github.com/noxproject/nox/core/protocol"
	"github.com/noxproject/nox/core/types"
	"github.com/noxproject/nox/engine/txscript"
	"github.com/noxproject/nox/params"
	"github.com/noxproject/nox/services/common/error"
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

func MarshalJsonTx(tx *types.Tx, params *params.Params, blkHeight uint64,blkHashStr string) (json.TxRawResult, error){
	if tx == nil {
		return json.TxRawResult{}, errors.New("can't marshal nil transaction")
	}
	return MarshalJsonTransaction(tx.Transaction(), params, blkHeight,blkHashStr)
}


func MarshalJsonTransaction(tx *types.Transaction, params *params.Params, blkHeight uint64,blkHashStr string) (json.TxRawResult, error){

	hexStr, err := MessageToHex(&message.MsgTx{tx})
	if err!=nil {
		return json.TxRawResult{}, err
	}

	bufWit,_ :=tx.Serialize(types.TxSerializeOnlyWitness)
	hexStrWit:=hex.EncodeToString(bufWit)

	//TODO, handle the blkHeight/blkHash

	return json.TxRawResult{
		Hex : hexStr,
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
	params *params.Params, confirmations int64, nextHashString string) (json.OrderedResult, error) {

	head := b.Block().Header // copies the header once

	// Get next block hash unless there are none.
	height := uint64(head.Height)

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
				return MarshalJsonTx(tx,params,height,"")
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
