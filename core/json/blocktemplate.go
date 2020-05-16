package json

// for miner template
type PowDiffReference struct {
	//blake2bd diff
	Blake2bDBits string `json:"blake2bd_bits"`
	//blake2bd hash diff compare target
	Blake2bTarget          string `json:"blake2bd_target"`
	X16rv3Bits             string `json:"x_16_rv_3_bits"`
	X16rv3Target           string `json:"x_16_rv_3_target"`
	X8r16Bits              string `json:"x8r16_bits"`
	X8r16Target            string `json:"x8r16_target"`
	QitmeerKeccak256Bits   string `json:"qitmeer_keccak256_bits"`
	QitmeerKeccak256Target string `json:"qitmeer_keccak256_target"`

	//cuckoo mining min diff
	CuckarooMinDiff  uint64 `json:"cuckaroo_min_diff,omitempty"`
	CuckaroomMinDiff uint64 `json:"cuckaroom_min_diff,omitempty"`
	CuckatooMinDiff  uint64 `json:"cuckatoo_min_diff,omitempty"`
}

//LL(getblocktemplate RPC) 2018-10-28
// TemplateRequest is a request object as defined in BIP22
// (https://en.bitcoin.it/wiki/BIP_0022), it is optionally provided as an
//// argument to GetBlockTemplate RPC.
type TemplateRequest struct {
	Mode         string   `json:"mode,omitempty"`
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
	Capabilities  []string `json:"capabilities,omitempty"`
	RejectReasion string   `json:"reject-reason,omitempty"`
}
