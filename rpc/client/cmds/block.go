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

func init() {
	flags := UsageFlag(0)

	MustRegisterCmd("getBlockCount", (*GetBlockCountCmd)(nil), flags, DefaultServiceNameSpace)
	MustRegisterCmd("getBlockhash", (*GetBlockhashCmd)(nil), flags, DefaultServiceNameSpace)
	MustRegisterCmd("getBlockhashByRange", (*GetBlockhashByRangeCmd)(nil), flags, DefaultServiceNameSpace)
}
