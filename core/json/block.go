// Copyright (c) 2017-2018 The qitmeer developers

package json

type ProofData struct {
	EdgeBits int `json:"edge_bits"`
	CircleNonces []uint32 `json:"circle_nonces"`
}

// pow json result
type PowResult struct {
	Nonce     uint32     `json:"nonce"`
	PowName   string     `json:"pow_name"`
	PowType   uint8     `json:"pow_type"`
	ProofData ProofData     `json:"proof_data"`
}

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
	PowResult     PowResult  `json:"pow_result"`
	Difficulty    uint32       `json:"difficulty"`
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
	PowResult     PowResult  `json:"pow_result"`
}
