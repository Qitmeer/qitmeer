package cmds

type GetBlockCountCmd struct{}

func NewGetBlockCountCmd() *GetBlockCountCmd {
	return &GetBlockCountCmd{}
}

func init() {
	// The commands in this file are only usable by websockets.
	flags := UsageFlag(0)

	MustRegisterCmd("getBlockCount", (*GetBlockCountCmd)(nil), flags, DefaultServiceNameSpace)
}
