package cmds

const (
	BlockConnectedNtfnMethod    = "blockConnected"
	BlockDisconnectedNtfnMethod = "blockDisconnected"
)

type BlockConnectedNtfn struct {
	Hash  string
	Order int64
	Time  int64
}

func NewBlockConnectedNtfn(hash string, order int64, time int64) *BlockConnectedNtfn {
	return &BlockConnectedNtfn{
		Hash:  hash,
		Order: order,
		Time:  time,
	}
}

type BlockDisconnectedNtfn struct {
	Hash  string
	Order int64
	Time  int64
}

func NewBlockDisconnectedNtfn(hash string, order int64, time int64) *BlockDisconnectedNtfn {
	return &BlockDisconnectedNtfn{
		Hash:  hash,
		Order: order,
		Time:  time,
	}
}

func init() {
	flags := UFWebsocketOnly | UFNotification

	MustRegisterCmd(BlockConnectedNtfnMethod, (*BlockConnectedNtfn)(nil), flags, NotifyNameSpace)
	MustRegisterCmd(BlockDisconnectedNtfnMethod, (*BlockDisconnectedNtfn)(nil), flags, NotifyNameSpace)
}
