package cmds

type GetNodeInfoCmd struct{}

func NewGetNodeInfoCmd() *GetNodeInfoCmd {
	return &GetNodeInfoCmd{}
}

type GetPeerInfoCmd struct{}

func NewGetPeerInfoCmd() *GetPeerInfoCmd {
	return &GetPeerInfoCmd{}
}

func init() {
	flags := UsageFlag(0)

	MustRegisterCmd("getNodeInfo", (*GetNodeInfoCmd)(nil), flags, DefaultServiceNameSpace)
	MustRegisterCmd("getPeerInfo", (*GetPeerInfoCmd)(nil), flags, DefaultServiceNameSpace)
}
