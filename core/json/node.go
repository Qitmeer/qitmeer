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

// GetPeerInfoResult models the data returned from the getpeerinfo command.
type GetPeerInfoResult struct {
	ID             int32   `json:"id"`
	Addr           string  `json:"addr"`
	AddrLocal      string  `json:"addrlocal,omitempty"`
	Services       string  `json:"services"`
	RelayTxes      bool    `json:"relaytxes"`
	LastSend       int64   `json:"lastsend"`
	LastRecv       int64   `json:"lastrecv"`
	BytesSent      uint64  `json:"bytessent"`
	BytesRecv      uint64  `json:"bytesrecv"`
	ConnTime       int64   `json:"conntime"`
	TimeOffset     int64   `json:"timeoffset"`
	PingTime       float64 `json:"pingtime"`
	PingWait       float64 `json:"pingwait,omitempty"`
	Version        uint32  `json:"version"`
	SubVer         string  `json:"subver"`
	Inbound        bool    `json:"inbound"`
	BanScore       int32   `json:"banscore"`
	SyncNode       bool    `json:"syncnode"`
	GraphState     GetGraphStateResult `json:"graphstate"`
}

// GetGraphStateResult data
type GetGraphStateResult struct {
	Tips []string `json:"tips"`
	Total uint32 `json:"total"`
	Layer uint32 `json:"layer"`
}