package cmds

const (
	BlockConnectedNtfnMethod    = "blockConnected"
	BlockDisconnectedNtfnMethod = "blockDisconnected"
)

type BlockConnectedNtfn struct {
	Hash  string
	Order int64
	Time  int64
	Txs   []string
}

func NewBlockConnectedNtfn(hash string, order int64, time int64, txs []string) *BlockConnectedNtfn {
	return &BlockConnectedNtfn{
		Hash:  hash,
		Order: order,
		Time:  time,
		Txs:   txs,
	}
}

type BlockDisconnectedNtfn struct {
	Hash  string
	Order int64
	Time  int64
	Txs   []string
}

func NewBlockDisconnectedNtfn(hash string, order int64, time int64, txs []string) *BlockDisconnectedNtfn {
	return &BlockDisconnectedNtfn{
		Hash:  hash,
		Order: order,
		Time:  time,
		Txs:   txs,
	}
}

func init() {
	flags := UFWebsocketOnly | UFNotification

	MustRegisterCmd(BlockConnectedNtfnMethod, (*BlockConnectedNtfn)(nil), flags, NotifyNameSpace)
	MustRegisterCmd(BlockDisconnectedNtfnMethod, (*BlockDisconnectedNtfn)(nil), flags, NotifyNameSpace)
}
