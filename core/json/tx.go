// Copyright (c) 2017-2018 The qitmeer developers

package json

import "encoding/json"

// TxRawResult models the data from the getrawtransaction command.
type TxRawResult struct {
	Hex           string `json:"hex"`
	Txid          string `json:"txid"`
	TxHash        string `json:"txhash,omitempty"`
	Size          int32  `json:"size,omitempty"`
	Version       uint32 `json:"version"`
	LockTime      uint32 `json:"locktime"`
	Timestamp     string `json:"timestamp,omitempty"`
	Expire        uint32 `json:"expire"`
	Vin           []Vin  `json:"vin"`
	Vout          []Vout `json:"vout"`
	BlockHash     string `json:"blockhash,omitempty"`
	BlockOrder    uint64 `json:"blockorder,omitempty"`
	TxIndex       uint32 `json:"txindex,omitempty"`
	Confirmations int64  `json:"confirmations"`
	Time          int64  `json:"time,omitempty"`
	Blocktime     int64  `json:"blocktime,omitempty"`
	Duplicate     bool   `json:"duplicate,omitempty"`
	Txsvalid      bool   `json:"txsvalid"`
}

// Vin models parts of the tx data.  It is defined separately since
// getrawtransaction, decoderawtransaction, and searchrawtransaction use the
// same structure.
type Vin struct {
	Coinbase  string     `json:"coinbase"`
	Txid      string     `json:"txid"`
	Vout      uint32     `json:"vout"`
	Sequence  uint32     `json:"sequence"`
	ScriptSig *ScriptSig `json:"scriptSig"`
	TxType    string     `json:"type,omitempty"`
}

// IsCoinBase returns a bool to show if a Vin is a Coinbase one or not.
func (v *Vin) IsCoinBase() bool {
	return len(v.Coinbase) > 0
}

func (v *Vin) IsNonStd() bool {
	return len(v.TxType) > 0
}

// MarshalJSON provides a custom Marshal method for Vin.
func (v *Vin) MarshalJSON() ([]byte, error) {
	if v.IsCoinBase() {
		coinbaseStruct := struct {
			Coinbase string `json:"coinbase"`
			Sequence uint32 `json:"sequence"`
		}{
			Coinbase: v.Coinbase,
			Sequence: v.Sequence,
		}
		return json.Marshal(coinbaseStruct)
	}
	if v.IsNonStd() {
		coinbaseStruct := struct {
			Type      string     `json:"type"`
			ScriptSig *ScriptSig `json:"scriptSig"`
		}{
			Type:      v.TxType,
			ScriptSig: v.ScriptSig,
		}
		return json.Marshal(coinbaseStruct)
	}

	txStruct := struct {
		Txid      string     `json:"txid"`
		Vout      uint32     `json:"vout"`
		Sequence  uint32     `json:"sequence"`
		ScriptSig *ScriptSig `json:"scriptSig"`
	}{
		Txid:      v.Txid,
		Vout:      v.Vout,
		Sequence:  v.Sequence,
		ScriptSig: v.ScriptSig,
	}
	return json.Marshal(txStruct)
}

// Vout models parts of the tx data.  It is defined separately since both
// getrawtransaction and decoderawtransaction use the same structure.
type Vout struct {
	Coin         string             `json:"coin,omitempty"`
	CoinId       uint16             `json:"coinid,omitempty"`
	Amount       uint64             `json:"amount,omitempty"`
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
	CoinId        uint16             `json:"coinId"`
	Amount        float64            `json:"amount"`
	ScriptPubKey  ScriptPubKeyResult `json:"scriptPubKey"`
	Version       int32              `json:"version"`
	Coinbase      bool               `json:"coinbase"`
}

// GetRawTransactionsResult models the data from the getrawtransactions
// command.
type GetRawTransactionsResult struct {
	Hex           string       `json:"hex,omitempty"`
	Txid          string       `json:"txid"`
	Hash          string       `json:"hash"`
	Size          string       `json:"size"`
	Vsize         string       `json:"vsize"`
	Version       uint32       `json:"version"`
	LockTime      uint32       `json:"locktime"`
	Vin           []VinPrevOut `json:"vin"`
	Vout          []Vout       `json:"vout"`
	BlockHash     string       `json:"blockhash,omitempty"`
	Confirmations uint64       `json:"confirmations,omitempty"`
	Time          int64        `json:"time,omitempty"`
	Blocktime     int64        `json:"blocktime,omitempty"`
}

type VinPrevOut struct {
	Coinbase  string     `json:"coinbase"`
	Txid      string     `json:"txid"`
	Vout      uint32     `json:"vout"`
	ScriptSig *ScriptSig `json:"scriptSig"`
	PrevOut   *PrevOut   `json:"prevOut"`
	Sequence  uint32     `json:"sequence"`
}

type PrevOut struct {
	Addresses []string `json:"addresses,omitempty"`
	CoinId    uint16   `json:"coinId"`
	Value     float64  `json:"value"`
}

type DecodeRawTransactionResult struct {
	Order      uint64 `json:"order"`
	BlockHash  string `json:"blockhash"`
	Txvalid    bool   `json:"txvalid"`
	Duplicate  bool   `json:"duplicate,omitempty"`
	IsCoinbase bool   `json:"is_coinbase"`
	Confirms   uint64 `json:"confirms"`
	IsBlue     bool   `json:"is_blue"`
	Txid       string `json:"txid"`
	Hash       string `json:"txhash"`
	Version    uint32 `json:"version"`
	LockTime   uint32 `json:"locktime"`
	Time       string `json:"timestamp"`
	Vin        []Vin  `json:"vin"`
	Vout       []Vout `json:"vout"`
}

// TransactionInput represents the inputs to a transaction.  Specifically a
// transaction hash and output number pair.
type TransactionInput struct {
	Txid string `json:"txid"`
	Vout uint32 `json:"vout"`
}

type Amounts map[string]uint64 //{\"address\":amount,...}

type Amout struct {
	CoinId uint16 `json:"coinid"`
	Amount int64  `json:"amount"`
}

type AdreesAmount map[string]Amout
