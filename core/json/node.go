package json

// for pow diff
type PowDiff struct {
	Blake2bdDiff float64 `json:"blake2bd_diff"`
	CuckarooDiff float64 `json:"cuckaroo_diff"`
	CuckatooDiff float64 `json:"cuckatoo_diff"`
}

// InfoNodeResult models the data returned by the node server getnodeinfo command.
type InfoNodeResult struct {
	ID               string              `json:"ID"`
	Addresss         []string            `json:"address"`
	QNR              string              `json:"QNR,omitempty"`
	Version          int32               `json:"version"`
	BuildVersion     string              `json:"buildversion"`
	ProtocolVersion  int32               `json:"protocolversion"`
	TotalSubsidy     uint64              `json:"totalsubsidy"`
	GraphState       GetGraphStateResult `json:"graphstate"`
	TimeOffset       int64               `json:"timeoffset"`
	Connections      int32               `json:"connections"`
	PowDiff          PowDiff             `json:"pow_diff"`
	TestNet          bool                `json:"testnet"`
	MixNet           bool                `json:"mixnet"`
	Confirmations    int32               `json:"confirmations"`
	CoinbaseMaturity int32               `json:"coinbasematurity"`
	Errors           string              `json:"errors"`
	Modules          []string            `json:"modules"`
	DNS              string              `json:"dns,omitempty"`
}

// GetPeerInfoResult models the data returned from the getpeerinfo command.
type GetPeerInfoResult struct {
	ID         string               `json:"id"`
	QNR        string               `json:"qnr,omitempty"`
	Address    string               `json:"address"`
	State      string               `json:"state"`
	Protocol   uint32               `json:"protocol,omitempty"`
	Genesis    string               `json:"genesis,omitempty"`
	Services   string               `json:"services,omitempty"`
	UserAgent  string               `json:"useragent,omitempty"`
	Direction  string               `json:"direction,omitempty"`
	GraphState *GetGraphStateResult `json:"graphstate,omitempty"`
	SyncNode   bool                 `json:"syncnode,omitempty"`
	TimeOffset int64                `json:"timeoffset"`
}

// GetGraphStateResult data
type GetGraphStateResult struct {
	Tips       []string `json:"tips"`
	MainOrder  uint32   `json:"mainorder"`
	MainHeight uint32   `json:"mainheight"`
	Layer      uint32   `json:"layer"`
}

type GetBanlistResult struct {
	ID   string `json:"id"`
	Bads int    `json:"bads"`
}
