package json

// for pow diff
type PowDiff struct {
	CurrentDiff float64 `json:"current_diff"`
}

// InfoNodeResult models the data returned by the node server getnodeinfo command.
type InfoNodeResult struct {
	ID                  string                              `json:"ID"`
	Addresss            []string                            `json:"address"`
	QNR                 string                              `json:"QNR,omitempty"`
	Version             int32                               `json:"version"`
	BuildVersion        string                              `json:"buildversion"`
	ProtocolVersion     int32                               `json:"protocolversion"`
	TotalSubsidy        uint64                              `json:"totalsubsidy,omitempty"`
	GraphState          *GetGraphStateResult                `json:"graphstate,omitempty"`
	TimeOffset          int64                               `json:"timeoffset,omitempty"`
	PowDiff             *PowDiff                            `json:"pow_diff,omitempty"`
	Confirmations       int32                               `json:"confirmations,omitempty"`
	CoinbaseMaturity    int32                               `json:"coinbasematurity,omitempty"`
	Errors              string                              `json:"errors,omitempty"`
	Modules             []string                            `json:"modules,omitempty"`
	DNS                 string                              `json:"dns,omitempty"`
	ConsensusDeployment map[string]*ConsensusDeploymentDesc `json:"consensusdeployment,omitempty"`
	Network             string                              `json:"network"`
	Connections         int32                               `json:"connections"`
}

// GetPeerInfoResult models the data returned from the getpeerinfo command.
type GetPeerInfoResult struct {
	ID         string               `json:"id"`
	QNR        string               `json:"qnr,omitempty"`
	Address    string               `json:"address"`
	State      string               `json:"state,omitempty"`
	Protocol   uint32               `json:"protocol,omitempty"`
	Genesis    string               `json:"genesis,omitempty"`
	Services   string               `json:"services,omitempty"`
	Name       string               `json:"name,omitempty"`
	Direction  string               `json:"direction,omitempty"`
	GraphState *GetGraphStateResult `json:"graphstate,omitempty"`
	GSUpdate   string               `json:"gsupdate,omitempty"`
	SyncNode   bool                 `json:"syncnode,omitempty"`
	TimeOffset int64                `json:"timeoffset"`
	LastSend   string               `json:"lastsend,omitempty"`
	LastRecv   string               `json:"lastrecv,omitempty"`
	BytesSent  uint64               `json:"bytessent,omitempty"`
	BytesRecv  uint64               `json:"bytesrecv,omitempty"`
	ConnTime   string               `json:"conntime,omitempty"`
	Version    string               `json:"version,omitempty"`
	Network    string               `json:"network,omitempty"`
	Circuit    bool                 `json:"circuit,omitempty"`
	Bads       int                  `json:"bads,omitempty"`
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

type ConsensusDeploymentDesc struct {
	Status    string `json:"status"`
	Bit       uint8  `json:"bit"`
	StartTime int64  `json:"startTime"`
	Timeout   int64  `json:"timeout"`
	Perform   int64  `json:"perform,omitempty"`
	Since     string `json:"since,omitempty"`
}

type NetworkStat struct {
	TotalPeers     int            `json:"totalpeers"`
	MaxConnected   uint           `json:"maxconnected"`
	MaxInbound     int            `json:"maxinbound"`
	TotalConnected int            `json:"totalconnected"`
	TotalRelays    int            `json:"totalrelays"`
	Infos          []*NetworkInfo `json:"infos"`
}

type NetworkInfo struct {
	Name       string `json:"name"`
	Peers      int    `json:"peers"`
	Relays     int    `json:"relays"`
	Connecteds int    `json:"connecteds"`
	AverageGS  string `json:"averagegs,omitempty"`
	MaxGS      string `json:"maxgs,omitempty"`
	MinGS      string `json:"mings,omitempty"`
}

type SubsidyInfo struct {
	Mode               string `json:"mode"`
	TotalSubsidy       uint64 `json:"totalsubsidy"`
	TargetTotalSubsidy int64  `json:"targettotalsubsidy,omitempty"`
	LeftTotalSubsidy   int64  `json:"lefttotalsubsidy,omitempty"`
	TotalTime          string `json:"totalTime,omitempty"`
	LeftTotalTime      string `json:"lefttotalTime,omitempty"`
	BaseSubsidy        int64  `json:"basesubsidy"`
	NextSubsidy        int64  `json:"nextsubsidy"`
}
