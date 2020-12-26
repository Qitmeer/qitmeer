package cmds

type NotifyBlocksCmd struct{}

func NewNotifyBlocksCmd() *NotifyBlocksCmd {
	return &NotifyBlocksCmd{}
}

type StopNotifyBlocksCmd struct{}

func NewStopNotifyBlocksCmd() *StopNotifyBlocksCmd {
	return &StopNotifyBlocksCmd{}
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

	MustRegisterCmd("notifyblocks", (*NotifyBlocksCmd)(nil), flags)
	MustRegisterCmd("notifyreceived", (*NotifyReceivedCmd)(nil), flags)
	MustRegisterCmd("stopnotifyblocks", (*StopNotifyBlocksCmd)(nil), flags)
}
