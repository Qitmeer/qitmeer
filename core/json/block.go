// Copyright (c) 2017-2018 The qitmeer developers

package json

type ProofData struct {
	EdgeBits     int    `json:"edge_bits"`
	CircleNonces string `json:"circle_nonces"`
}

// pow json result
type PowResult struct {
	PowName   string     `json:"pow_name"`
	PowType   uint8      `json:"pow_type"`
	Nonce     uint64     `json:"nonce"`
	ProofData *ProofData `json:"proof_data,omitempty"`
}

// BlockVerboseResult models the data from the getblock command when the
// verbose flag is set.  When the verbose flag is not set, getblock returns a
// hex-encoded string.

type BlockVerboseResult struct {
	Hash          string        `json:"hash"`
	Txsvalid      bool          `json:"txsvalid"`
	Confirmations int64         `json:"confirmations"`
	Version       int32         `json:"version"`
	Weight        int64         `json:"weight"`
	Height        int64         `json:"height"`
	TxRoot        string        `json:"txRoot"`
	Order         int64         `json:"order,omitempty"`
	Tx            []TxRawResult `json:"transactions,omitempty"`
	TxFee         int64         `json:"transactionfee,omitempty"`
	StateRoot     string        `json:"stateRoot"`
	Bits          string        `json:"bits"`
	Difficulty    uint32        `json:"difficulty"`
	PowResult     PowResult     `json:"pow"`
	Time          string        `json:"timestamp"`
	ParentRoot    string        `json:"parentroot"`
	Parents       []string      `json:"parents"`
	Children      []string      `json:"children"`
}

type BlockResult struct {
	Hash          string    `json:"hash"`
	Txsvalid      bool      `json:"txsvalid"`
	Confirmations int64     `json:"confirmations"`
	Version       int32     `json:"version"`
	Weight        int64     `json:"weight"`
	Height        int64     `json:"height"`
	TxRoot        string    `json:"txRoot"`
	Order         int64     `json:"order,omitempty"`
	Tx            []string  `json:"transactions,omitempty"`
	TxFee         int64     `json:"transactionfee,omitempty"`
	StateRoot     string    `json:"stateRoot"`
	Bits          string    `json:"bits"`
	Difficulty    uint32    `json:"difficulty"`
	PowResult     PowResult `json:"pow"`
	Time          string    `json:"timestamp"`
	ParentRoot    string    `json:"parentroot"`
	Parents       []string  `json:"parents"`
	Children      []string  `json:"children"`
}

// GetBlockHeaderVerboseResult models the data from the getblockheader command when
// the verbose flag is set.  When the verbose flag is not set, getblockheader
// returns a hex-encoded string.
type GetBlockHeaderVerboseResult struct {
	Hash          string    `json:"hash"`
	Confirmations int64     `json:"confirmations"`
	Version       int32     `json:"version"`
	ParentRoot    string    `json:"parentroot"`
	TxRoot        string    `json:"txRoot"`
	StateRoot     string    `json:"stateRoot"`
	Difficulty    uint32    `json:"difficulty"`
	Layer         uint32    `json:"layer"`
	Time          int64     `json:"time"`
	PowResult     PowResult `json:"pow"`
}

type TokenState struct {
	CoinId     uint16 `json:"coinid"`
	CoinName   string `json:"coinname"`
	Owners     string `json:"owners"`
	UpLimit    uint64 `json:"uplimit,omitempty"`
	Enable     bool   `json:"enable,omitempty"`
	Balance    int64  `json:"balance,omitempty"`
	LockedMeer int64  `json:"lockedMEER,omitempty"`
}
