package json

type PowDiffReference struct {
	NBits  string `json:"nbits"`
	Target string `json:"target"`
}

//LL(getblocktemplate RPC) 2018-10-28
// TemplateRequest is a request object as defined in BIP22
// (https://en.bitcoin.it/wiki/BIP_0022), it is optionally provided as an
//// argument to GetBlockTemplate RPC.
type TemplateRequest struct {
	Mode         string   `json:"mode,omitempty"`
	PowType      byte     `json:"pow_type"`
	Capabilities []string `json:"capabilities,omitempty"`

	// Optional long polling.
	LongPollID string `json:"longpollid,omitempty"`

	// Optional template tweaking.  SigOpLimit and SizeLimit can be int64
	// or bool.
	SigOpLimit interface{} `json:"sigoplimit,omitempty"`
	SizeLimit  interface{} `json:"sizelimit,omitempty"`
	MaxVersion uint32      `json:"maxversion,omitempty"`

	// Basic pool extension from BIP 0023.
	Target string `json:"target,omitempty"`

	// Block proposal from BIP 0023.  Data is only provided when Mode is
	// "proposal".
	Data   string `json:"data,omitempty"`
	WorkID string `json:"workid,omitempty"`
}

// GetBlockTemplateResultTx models the transactions field of the
// getblocktemplate command.
type GetBlockTemplateResultTx struct {
	Data    string  `json:"data"`
	Hash    string  `json:"hash"`
	Depends []int64 `json:"depends"`
	Fee     int64   `json:"fee"`
	SigOps  int64   `json:"sigops"`
	Weight  int64   `json:"weight"`
}

// GetBlockTemplateResultPt models the parents field of the
// getblocktemplate command.
type GetBlockTemplateResultPt struct {
	Data string `json:"data"`
	Hash string `json:"hash"`
}

// GetBlockTemplateResultAux models the coinbaseaux field of the
// getblocktemplate command.
type GetBlockTemplateResultAux struct {
	Flags string `json:"flags"`
}

// GetBlockTemplateResult models the data returned from the getblocktemplate
type GetBlockTemplateResult struct {
	// Base fields from BIP 0022.  CoinbaseAux is optional.  One of
	// CoinbaseTxn or CoinbaseValue must be specified, but not both.
	StateRoot     string                     `json:"stateroot"`
	CurTime       int64                      `json:"curtime"`
	Height        int64                      `json:"height"`
	Blues         int64                      `json:"blues"`
	PreviousHash  string                     `json:"previousblockhash"`
	SigOpLimit    int64                      `json:"sigoplimit,omitempty"`
	SizeLimit     int64                      `json:"sizelimit,omitempty"`
	WeightLimit   int64                      `json:"weightlimit,omitempty"`
	Parents       []GetBlockTemplateResultPt `json:"parents"`
	Transactions  []GetBlockTemplateResultTx `json:"transactions"`
	Version       uint32                     `json:"version"`
	CoinbaseAux   *GetBlockTemplateResultAux `json:"coinbaseaux,omitempty"`
	CoinbaseTxn   *GetBlockTemplateResultTx  `json:"coinbasetxn,omitempty"`
	CoinbaseValue *uint64                    `json:"coinbasevalue,omitempty"`
	WorkID        string                     `json:"workid,omitempty"`
	NodeInfo      string                     `json:"nodeinfo,omitempty"`

	// Witness commitment defined in BIP 0141.
	DefaultWitnessCommitment string `json:"default_witness_commitment,omitempty"`

	// Optional long polling from BIP 0022.
	LongPollID  string `json:"longpollid,omitempty"`
	LongPollURI string `json:"longpolluri,omitempty"`
	SubmitOld   *bool  `json:"submitold,omitempty"`

	// Basic pool extension from BIP 0023.
	Expires          int64            `json:"expires,omitempty"`
	PowDiffReference PowDiffReference `json:"pow_diff_reference"`
	// Mutations from BIP 0023.
	MaxTime    int64    `json:"maxtime,omitempty"`
	MinTime    int64    `json:"mintime,omitempty"`
	Mutable    []string `json:"mutable,omitempty"`
	NonceRange string   `json:"noncerange,omitempty"`

	// Block proposal from BIP 0023.
	// temp use
	WorkData        string        `json:"workdata"`
	Capabilities    []string      `json:"capabilities,omitempty"`
	RejectReasion   string        `json:"reject-reason,omitempty"`
	BlockFeesMap    map[int]int64 `json:"block_fees_map"`
	CoinbaseVersion string        `json:"coinbase_version"`
}

// GetBlockTemplateResult models the data returned from the getblocktemplate
type SubmitBlockResult struct {
	BlockHash      string `json:"block_hash"`
	CoinbaseTxID   string `json:"coinbase_txid"`
	Order          string `json:"order"`
	Height         int64  `json:"height"`
	CoinbaseAmount uint64 `json:"coinbase_amount"`
	MinerType      string `json:"miner_type"`
}

type MinerInfoResult struct {
	Type          string `json:"type"`
	Pow           string `json:"pow"`
	Running       bool   `json:"running"`
	Coinbase      string `json:"coinbase"`
	Height        uint64 `json:"height"`
	Difficulty    string `json:"difficulty"`
	Target        string `json:"target"`
	Timestamp     string `json:"timestamp"`
	TotalSubmit   int    `json:"totalsubmit"`
	SuccessSubmit int    `json:"successsubmit"`
}
