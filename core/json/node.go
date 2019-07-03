package json

// InfoNodeResult models the data returned by the node server getnodeinfo command.
type InfoNodeResult struct {
	Version         int32   `json:"version"`
	ProtocolVersion int32   `json:"protocolversion"`
	Blocks          uint32   `json:"blocks"`
	TimeOffset      int64   `json:"timeoffset"`
	Connections     int32   `json:"connections"`
	Difficulty      float64 `json:"difficulty"`
	TestNet         bool    `json:"testnet"`
	Errors          string  `json:"errors"`
}