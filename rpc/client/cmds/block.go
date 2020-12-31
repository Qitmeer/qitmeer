package cmds

type GetBlockCountCmd struct{}

func NewGetBlockCountCmd() *GetBlockCountCmd {
	return &GetBlockCountCmd{}
}

type GetBlockhashCmd struct {
	Order uint
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
	MustRegisterCmd("getBlockCount", (*GetBlockCountCmd)(nil), flags, DefaultServiceNameSpace)
	MustRegisterCmd("getOrphansTotal", (*GetOrphansTotalCmd)(nil), flags, DefaultServiceNameSpace)
}
