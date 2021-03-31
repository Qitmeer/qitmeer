package cmds

type GetBlockCountCmd struct{}

func NewGetBlockCountCmd() *GetBlockCountCmd {
	return &GetBlockCountCmd{}
}

type GetBlockhashCmd struct {
	Order uint
}

// BlockDetails describes details of a tx in a block.
type BlockDetails struct {
	Order uint64 `json:"order"`
	Hash  string `json:"hash"`
	Index int    `json:"index"`
	Time  int64  `json:"time"`
}

// RescanFinishedNtfn defines the rescanfinished JSON-RPC notification.
//
type RescanFinishedNtfn struct {
	Hash       string
	Order      uint64
	Time       int64
	LastTxHash string
}

// RescanProgressNtfn defines the rescanprogress JSON-RPC notification.
//
type RescanProgressNtfn struct {
	Hash  string
	Order uint64
	Time  int64
}

// NewRescanProgressNtfn returns a new instance which can be used to issue a
// rescanprogress JSON-RPC notification.
//
func NewRescanProgressNtfn(hash string, order uint64, time int64) *RescanProgressNtfn {
	return &RescanProgressNtfn{
		Hash:  hash,
		Order: order,
		Time:  time,
	}
}

// NewRescanFinishedNtfn returns a new instance which can be used to issue a
// rescanfinished JSON-RPC notification.
//
func NewRescanFinishedNtfn(hash, txHash string, order uint64, time int64) *RescanFinishedNtfn {
	return &RescanFinishedNtfn{
		Hash:       hash,
		Order:      order,
		Time:       time,
		LastTxHash: txHash,
	}
}

// RedeemingTxNtfn defines the redeemingtx JSON-RPC notification.
//
type RedeemingTxNtfn struct {
	HexTx string
	Block *BlockDetails
}

func NewGetBlockhashCmd(order uint) *GetBlockhashCmd {
	return &GetBlockhashCmd{
		Order: order,
	}
}

type GetBlockhashByRangeCmd struct {
	Start uint
	End   uint
}

func NewGetBlockhashByRangeCmd(start uint, end uint) *GetBlockhashByRangeCmd {
	return &GetBlockhashByRangeCmd{
		Start: start,
		End:   end,
	}
}

type GetBlockCmd struct {
	H       string
	Verbose bool
	InclTx  bool
	FullTx  bool
}

func NewGetBlockCmd(h string, verbose bool, inclTx bool, fullTx bool) *GetBlockCmd {
	return &GetBlockCmd{
		H:       h,
		Verbose: verbose,
		InclTx:  inclTx,
		FullTx:  fullTx,
	}
}

type GetBlockByOrderCmd struct {
	Order   uint
	Verbose bool
	InclTx  bool
	FullTx  bool
}

func NewGetBlockByOrderCmd(order uint, verbose bool, inclTx bool, fullTx bool) *GetBlockByOrderCmd {
	return &GetBlockByOrderCmd{
		Order:   order,
		Verbose: verbose,
		InclTx:  inclTx,
		FullTx:  fullTx,
	}
}

type GetBlockV2Cmd struct {
	H       string
	Verbose bool
	InclTx  bool
	FullTx  bool
}

func NewGetBlockV2Cmd(h string, verbose bool, inclTx bool, fullTx bool) *GetBlockCmd {
	return &GetBlockCmd{
		H:       h,
		Verbose: verbose,
		InclTx:  inclTx,
		FullTx:  fullTx,
	}
}

type GetBestBlockHashCmd struct{}

func NewGetBestBlockHashCmd() *GetBestBlockHashCmd {
	return &GetBestBlockHashCmd{}
}

type GetBlockTotalCmd struct{}

func NewGetBlockTotalCmd() *GetBlockTotalCmd {
	return &GetBlockTotalCmd{}
}

type GetBlockHeaderCmd struct {
	Hash    string
	Verbose bool
}

func NewGetBlockHeaderCmd(hash string, verbose bool) *GetBlockHeaderCmd {
	return &GetBlockHeaderCmd{
		Hash:    hash,
		Verbose: verbose,
	}
}

type IsOnMainChainCmd struct {
	H string
}

func NewIsOnMainChainCmd(h string) *IsOnMainChainCmd {
	return &IsOnMainChainCmd{
		H: h,
	}
}

type GetMainChainHeightCmd struct{}

func NewGetMainChainHeightCmd() *GetMainChainHeightCmd {
	return &GetMainChainHeightCmd{}
}

type GetBlockWeightCmd struct {
	H string
}

func NewGetBlockWeightCmd(h string) *GetBlockWeightCmd {
	return &GetBlockWeightCmd{
		H: h,
	}
}

type GetOrphansTotalCmd struct{}

func NewGetOrphansTotalCmd() *GetOrphansTotalCmd {
	return &GetOrphansTotalCmd{}
}

type GetBlockByNumCmd struct {
	ID      uint
	Verbose bool
	InclTx  bool
	FullTx  bool
}

func NewGetBlockByNumCmd(id uint, verbose bool, inclTx bool, fullTx bool) *GetBlockByNumCmd {
	return &GetBlockByNumCmd{
		ID:      id,
		Verbose: verbose,
		InclTx:  inclTx,
		FullTx:  fullTx,
	}
}

type IsBlueCmd struct {
	H string
}

func NewIsBlueCmd(h string) *IsBlueCmd {
	return &IsBlueCmd{
		H: h,
	}
}

type IsCurrentCmd struct {
}

func NewIsCurrentCmd() *IsCurrentCmd {
	return &IsCurrentCmd{}
}

type TipsCmd struct {
}

func NewTipsCmd() *TipsCmd {
	return &TipsCmd{}
}

type GetCoinbaseCmd struct {
}

func NewGetCoinbaseCmd() *GetCoinbaseCmd {
	return &GetCoinbaseCmd{}
}

type GetFeesCmd struct {
	H string
}

func NewGetFeesCmd(h string) *GetFeesCmd {
	return &GetFeesCmd{
		H: h,
	}
}

func init() {
	flags := UsageFlag(0)

	MustRegisterCmd("getBlockCount", (*GetBlockCountCmd)(nil), flags, DefaultServiceNameSpace)
	MustRegisterCmd("getBlockhash", (*GetBlockhashCmd)(nil), flags, DefaultServiceNameSpace)
	MustRegisterCmd("getBlockhashByRange", (*GetBlockhashByRangeCmd)(nil), flags, DefaultServiceNameSpace)
	MustRegisterCmd("getBlock", (*GetBlockCmd)(nil), flags, DefaultServiceNameSpace)
	MustRegisterCmd("getBlockV2", (*GetBlockV2Cmd)(nil), flags, DefaultServiceNameSpace)
	MustRegisterCmd("getBlockByOrder", (*GetBlockByOrderCmd)(nil), flags, DefaultServiceNameSpace)
	MustRegisterCmd("getBestBlockHash", (*GetBestBlockHashCmd)(nil), flags, DefaultServiceNameSpace)
	MustRegisterCmd("getBlockTotal", (*GetBlockTotalCmd)(nil), flags, DefaultServiceNameSpace)
	MustRegisterCmd("getBlockHeader", (*GetBlockHeaderCmd)(nil), flags, DefaultServiceNameSpace)
	MustRegisterCmd("isOnMainChain", (*IsOnMainChainCmd)(nil), flags, DefaultServiceNameSpace)
	MustRegisterCmd("getMainChainHeight", (*GetMainChainHeightCmd)(nil), flags, DefaultServiceNameSpace)
	MustRegisterCmd("getBlockWeight", (*GetBlockWeightCmd)(nil), flags, DefaultServiceNameSpace)
	MustRegisterCmd("getOrphansTotal", (*GetOrphansTotalCmd)(nil), flags, DefaultServiceNameSpace)
	MustRegisterCmd("getBlockByNum", (*GetBlockByNumCmd)(nil), flags, DefaultServiceNameSpace)
	MustRegisterCmd("isBlue", (*IsBlueCmd)(nil), flags, DefaultServiceNameSpace)
	MustRegisterCmd("isCurrent", (*IsCurrentCmd)(nil), flags, DefaultServiceNameSpace)
	MustRegisterCmd("tips", (*TipsCmd)(nil), flags, DefaultServiceNameSpace)
	MustRegisterCmd("getCoinbase", (*GetCoinbaseCmd)(nil), flags, DefaultServiceNameSpace)
	MustRegisterCmd("getFees", (*GetFeesCmd)(nil), flags, DefaultServiceNameSpace)
}
