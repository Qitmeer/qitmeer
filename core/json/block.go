// Copyright (c) 2017-2018 The nox developers

package json


// BlockVerboseResult models the data from the getblock command when the
// verbose flag is set.  When the verbose flag is not set, getblock returns a
// hex-encoded string.

type BlockVerboseResult struct {
	Hash          string        `json:"hash"`
	Confirmations int64         `json:"confirmations"`
	Size          int32         `json:"size"`
	Height        int64         `json:"height"`
	Version       int32         `json:"version"`
	TxRoot        string        `json:"txRoot"`
	Tx            []string      `json:"tx,omitempty"`
	RawTx         []TxRawResult `json:"rawtx,omitempty"`
	Time          int64         `json:"time"`
	Nonce         uint32        `json:"nonce"`
	Bits          string        `json:"bits"`
	Difficulty    float64       `json:"difficulty"`
	PreviousHash  string        `json:"previousblockhash"`
	NextHash      string        `json:"nextblockhash,omitempty"`
}

// GetBlockHeaderVerboseResult models the data from the getblockheader command when
// the verbose flag is set.  When the verbose flag is not set, getblockheader
// returns a hex-encoded string.
type GetBlockHeaderVerboseResult struct {
	Hash          string  `json:"hash"`
	Confirmations int64   `json:"confirmations"`
	Version       int32   `json:"version"`
	ParentRoot    string  `json:"parentroot"`
	TxRoot        string  `json:"txRoot"`
	StateRoot     string  `json:"stateRoot"`
	Difficulty    uint32  `json:"difficulty"`
	Layer         uint32  `json:"layer"`
	Time          int64   `json:"time"`
	Nonce         uint64  `json:"nonce"`
}
