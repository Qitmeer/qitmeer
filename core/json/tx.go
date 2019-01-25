// Copyright (c) 2017-2018 The nox developers

package json

import "encoding/json"

// TxRawResult models the data from the getrawtransaction command.
type TxRawResult struct {
	Hex           string `json:"hex"`
	HexNoWit      string `json:"hexnowit"`
	HexWit        string `json:"hexwit"`
	Txid          string `json:"txid"`
	TxHash        string `json:"txhash"`
	Version       uint32 `json:"version"`
	LockTime      uint32 `json:"locktime"`
	Expire        uint32 `json:"expire"`
	Vin           []Vin  `json:"vin"`
	Vout          []Vout `json:"vout"`
	BlockHash     string `json:"blockhash,omitempty"`
	BlockHeight   uint64 `json:"blockheight"`
	TxIndex       uint32 `json:"txindex,omitempty"`
	Confirmations int64  `json:"confirmations"`
	Time          int64  `json:"time,omitempty"`
	Blocktime     int64  `json:"blocktime,omitempty"`
}

// Vin models parts of the tx data.  It is defined separately since
// getrawtransaction, decoderawtransaction, and searchrawtransaction use the
// same structure.
type Vin struct {
	Coinbase    string     `json:"coinbase"`
	Txid        string     `json:"txid"`
	Vout        uint32     `json:"vout"`
	Sequence    uint32     `json:"sequence"`
	AmountIn    float64    `json:"amountin"`
	BlockHeight uint32     `json:"blockheight"`
	TxIndex     uint32     `json:"txindex"`
	ScriptSig   *ScriptSig `json:"scriptSig"`
}

// IsCoinBase returns a bool to show if a Vin is a Coinbase one or not.
func (v *Vin) IsCoinBase() bool {
	return len(v.Coinbase) > 0
}

// MarshalJSON provides a custom Marshal method for Vin.
func (v *Vin) MarshalJSON() ([]byte, error) {
	if v.IsCoinBase() {
		coinbaseStruct := struct {
			AmountIn    float64 `json:"amountin"`
			BlockHeight uint32  `json:"blockheight"`
			TxIndex  uint32     `json:"txindex"`
			Coinbase    string  `json:"coinbase"`
			Sequence    uint32  `json:"sequence"`
		}{
			AmountIn:    v.AmountIn,
			BlockHeight: v.BlockHeight,
			TxIndex:     v.TxIndex,
			Coinbase:    v.Coinbase,
			Sequence:    v.Sequence,
		}
		return json.Marshal(coinbaseStruct)
	}

	txStruct := struct {
		Txid        string     `json:"txid"`
		Vout        uint32     `json:"vout"`
		Sequence    uint32     `json:"sequence"`
		AmountIn    float64    `json:"amountin"`
		BlockHeight uint32     `json:"blockheight"`
		TxIndex     uint32     `json:"txindex"`
		ScriptSig   *ScriptSig `json:"scriptSig"`
	}{
		Txid:        v.Txid,
		Vout:        v.Vout,
		Sequence:    v.Sequence,
		AmountIn:    v.AmountIn,
		BlockHeight: v.BlockHeight,
		TxIndex:     v.TxIndex,
		ScriptSig:   v.ScriptSig,
	}
	return json.Marshal(txStruct)
}

// Vout models parts of the tx data.  It is defined separately since both
// getrawtransaction and decoderawtransaction use the same structure.
type Vout struct {
	Amount       float64            `json:"amount"`
	ScriptPubKey ScriptPubKeyResult `json:"scriptPubKey"`
}

// ScriptPubKeyResult models the scriptPubKey data of a tx script.  It is
// defined separately since it is used by multiple commands.
type ScriptPubKeyResult struct {
	Asm       string   `json:"asm"`
	Hex       string   `json:"hex,omitempty"`
	ReqSigs   int32    `json:"reqSigs,omitempty"`
	Type      string   `json:"type"`
	Addresses []string `json:"addresses,omitempty"`
}

// ScriptSig models a signature script.  It is defined separately since it only
// applies to non-coinbase.  Therefore the field in the Vin structure needs
// to be a pointer.
type ScriptSig struct {
	Asm string `json:"asm"`
	Hex string `json:"hex"`
}

// GetUtxoResult models the data from the GetUtxo command.
type GetUtxoResult struct {
	BestBlock     string             `json:"bestblock"`
	Confirmations int64              `json:"confirmations"`
	Amount        float64            `json:"amount"`
	ScriptPubKey  ScriptPubKeyResult `json:"scriptPubKey"`
	Version       int32              `json:"version"`
	Coinbase      bool               `json:"coinbase"`
}

