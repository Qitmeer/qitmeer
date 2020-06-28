package marshal

import (
	"bytes"
	"encoding/hex"
	"errors"
	"fmt"
	"github.com/Qitmeer/qitmeer/common/hash"
	"github.com/Qitmeer/qitmeer/core/json"
	"github.com/Qitmeer/qitmeer/core/message"
	"github.com/Qitmeer/qitmeer/core/protocol"
	"github.com/Qitmeer/qitmeer/core/types"
	"github.com/Qitmeer/qitmeer/engine/txscript"
	"github.com/Qitmeer/qitmeer/params"
	"github.com/Qitmeer/qitmeer/rpc"
	"strconv"
	"time"
)

// messageToHex serializes a message to the wire protocol encoding using the
// latest protocol version and returns a hex-encoded string of the result.
func MessageToHex(msg message.Message) (string, error) {
	var buf bytes.Buffer
	if err := msg.Encode(&buf, protocol.ProtocolVersion); err != nil {
		context := fmt.Sprintf("Failed to encode msg of type %T", msg)
		return "", rpc.RpcInternalError(err.Error(), context)
	}

	return hex.EncodeToString(buf.Bytes()), nil
}

func MarshalJsonTx(tx *types.Tx, params *params.Params, blkHashStr string,
	confirmations int64, coinbaseAmout uint64) (json.TxRawResult, error) {
	if tx == nil {
		return json.TxRawResult{}, errors.New("can't marshal nil transaction")
	}
	txr, err := MarshalJsonTransaction(tx.Transaction(), params, blkHashStr, confirmations, coinbaseAmout)
	if err == nil {
		txr.Duplicate = tx.IsDuplicate
	}
	return txr, err
}

func MarshalJsonTransaction(tx *types.Transaction, params *params.Params, blkHashStr string,
	confirmations int64, coinbaseAmout uint64) (json.TxRawResult, error) {

	hexStr, err := MessageToHex(&message.MsgTx{Tx: tx})
	if err != nil {
		return json.TxRawResult{}, err
	}
	txr := json.TxRawResult{
		Hex:       hexStr,
		Txid:      tx.TxHash().String(),
		TxHash:    tx.TxHashFull().String(),
		Size:      int32(tx.SerializeSize()),
		Version:   tx.Version,
		LockTime:  tx.LockTime,
		Timestamp: tx.Timestamp.Format(time.RFC3339),
		Expire:    tx.Expire,
		Vin:       MarshJsonVin(tx),
		Vout:      MarshJsonVout(tx, nil, params),
	}
	if tx.IsCoinBase() {
		txr.Vout[0].Amount = coinbaseAmout
	}
	if blkHashStr != "" {
		txr.BlockHash = blkHashStr
		txr.Confirmations = confirmations
	}
	return txr, nil
}

func MarshJsonVin(tx *types.Transaction) []json.Vin {
	// Coinbase transactions only have a single txin by definition.
	vinList := make([]json.Vin, len(tx.TxIn))
	if tx.IsCoinBase() {
		txIn := tx.TxIn[0]
		vinEntry := &vinList[0]
		vinEntry.Coinbase = hex.EncodeToString(txIn.SignScript)
		vinEntry.Sequence = txIn.Sequence
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
		vinEntry.ScriptSig = &json.ScriptSig{
			Asm: disbuf,
			Hex: hex.EncodeToString(txIn.SignScript),
		}
	}
	return vinList
}

func MarshJsonVout(tx *types.Transaction, filterAddrMap map[string]struct{}, params *params.Params) []json.Vout {
	voutList := make([]json.Vout, 0, len(tx.TxOut))
	for _, v := range tx.TxOut {
		// The disassembled string will contain [error] inline if the
		// script doesn't fully parse, so ignore the error here.
		disbuf, _ := txscript.DisasmString(v.PkScript)

		// Ignore the error here since an error means the script
		// couldn't parse and there is no additional information
		// about it anyways.
		sc, addrs, reqSigs, _ := txscript.ExtractPkScriptAddrs(
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
		voutSPK := &vout.ScriptPubKey
		vout.Amount = v.Amount
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
	params *params.Params, confirmations int64, children []*hash.Hash, state bool, isOrdered bool, coinbaseAmout uint64) (json.OrderedResult, error) {

	head := b.Block().Header // copies the header once
	// Get next block hash unless there are none.
	fields := json.OrderedResult{
		{Key: "hash", Val: b.Hash().String()},
		{Key: "txsvalid", Val: state},
		{Key: "confirmations", Val: confirmations},
		{Key: "version", Val: head.Version},
		{Key: "weight", Val: types.GetBlockWeight(b.Block())},
		{Key: "height", Val: b.Height()},
		{Key: "txRoot", Val: head.TxRoot.String()},
	}
	if isOrdered {
		fields = append(fields, json.KV{Key: "order", Val: b.Order()})
	}
	if inclTx {
		formatTx := func(tx *types.Tx) (interface{}, error) {
			return tx.Hash().String(), nil
		}
		if fullTx {
			formatTx = func(tx *types.Tx) (interface{}, error) {
				return MarshalJsonTx(tx, params, b.Hash().String(), confirmations, coinbaseAmout)
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
		fields = append(fields, json.KV{Key: "transactions", Val: transactions})
	}
	fields = append(fields, json.OrderedResult{
		{Key: "stateRoot", Val: head.StateRoot.String()},
		{Key: "bits", Val: strconv.FormatUint(uint64(head.Difficulty), 16)},
		{Key: "difficulty", Val: head.Difficulty},
		{Key: "pow", Val: head.Pow.GetPowResult()},
		{Key: "timestamp", Val: head.Timestamp.Format(time.RFC3339)},
		{Key: "parentroot", Val: head.ParentRoot.String()},
	}...)
	tempArr := []string{}
	if b.Block().Parents != nil && len(b.Block().Parents) > 0 {

		for i := 0; i < len(b.Block().Parents); i++ {
			tempArr = append(tempArr, b.Block().Parents[i].String())
		}
	} else {
		tempArr = append(tempArr, "null")
	}
	fields = append(fields, json.KV{Key: "parents", Val: tempArr})

	tempArr = []string{}
	if len(children) > 0 {

		for i := 0; i < len(children); i++ {
			tempArr = append(tempArr, children[i].String())
		}
	} else {
		tempArr = append(tempArr, "null")
	}
	fields = append(fields, json.KV{Key: "children", Val: tempArr})

	return fields, nil
}
