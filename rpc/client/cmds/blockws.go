package cmds

type NotifyBlocksCmd struct{}

func NewNotifyBlocksCmd() *NotifyBlocksCmd {
	return &NotifyBlocksCmd{}
}

type StopNotifyBlocksCmd struct{}

func NewStopNotifyBlocksCmd() *StopNotifyBlocksCmd {
	return &StopNotifyBlocksCmd{}
}

type SessionCmd struct{}

func NewSessionCmd() *SessionCmd {
	return &SessionCmd{}
}

// TODO op
type NotifyReceivedCmd struct {
	Addresses []string
}

func NewNotifyReceivedCmd(addresses []string) *NotifyReceivedCmd {
	return &NotifyReceivedCmd{
		Addresses: addresses,
	}
}

func init() {
	// The commands in this file are only usable by websockets.
	flags := UFWebsocketOnly

	MustRegisterCmd("notifyBlocks", (*NotifyBlocksCmd)(nil), flags, NotifyNameSpace)
	MustRegisterCmd("notifyReceived", (*NotifyReceivedCmd)(nil), flags, NotifyNameSpace)
	MustRegisterCmd("stopNotifyBlocks", (*StopNotifyBlocksCmd)(nil), flags, NotifyNameSpace)
	MustRegisterCmd("session", (*SessionCmd)(nil), flags, NotifyNameSpace)
}
