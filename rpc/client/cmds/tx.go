package cmds

import (
	"github.com/Qitmeer/qitmeer/core/json"
)

type CreateRawTransactionCmd struct {
	Inputs   []json.TransactionInput
	Amounts  json.Amounts
	LockTime int64
}

func NewCreateRawTransactionCmd(inputs []json.TransactionInput, amounts json.Amounts, lockTime int64) *CreateRawTransactionCmd {
	return &CreateRawTransactionCmd{
		Inputs:   inputs,
		Amounts:  amounts,
		LockTime: lockTime,
	}
}

type DecodeRawTransactionCmd struct {
	HexTx string
}

func NewDecodeRawTransactionCmd(hexTx string) *DecodeRawTransactionCmd {
	return &DecodeRawTransactionCmd{
		HexTx: hexTx,
	}
}

type SendRawTransactionCmd struct {
	HexTx         string
	AllowHighFees bool
}

func NewSendRawTransactionCmd(hexTx string, allowHighFees bool) *SendRawTransactionCmd {
	return &SendRawTransactionCmd{
		HexTx:         hexTx,
		AllowHighFees: allowHighFees,
	}
}

type GetRawTransactionCmd struct {
	TxHash  string
	Verbose bool
}

func NewGetRawTransactionCmd(txHash string, verbose bool) *GetRawTransactionCmd {
	return &GetRawTransactionCmd{
		TxHash:  txHash,
		Verbose: verbose,
	}
}

type GetUtxoCmd struct {
	TxHash         string
	Vout           uint32
	IncludeMempool bool
}

func NewGetUtxoCmd(txHash string, vout uint32, includeMempool bool) *GetUtxoCmd {
	return &GetUtxoCmd{
		TxHash:         txHash,
		Vout:           vout,
		IncludeMempool: includeMempool,
	}
}

type GetRawTransactionsCmd struct {
	Addre       string
	Vinext      bool
	Count       uint
	Skip        uint
	Revers      bool
	Verbose     bool
	FilterAddrs []string
}

func NewGetRawTransactionsCmd(addre string, vinext bool, count uint, skip uint, revers bool, verbose bool, filterAddrs []string) *GetRawTransactionsCmd {
	return &GetRawTransactionsCmd{
		Addre:       addre,
		Vinext:      vinext,
		Count:       count,
		Skip:        skip,
		Revers:      revers,
		Verbose:     verbose,
		FilterAddrs: filterAddrs,
	}
}

type TxSignCmd struct {
	PrivkeyStr string
	RawTxStr   string
}

func NewTxSignCmd(privkeyStr string, rawTxStr string) *TxSignCmd {
	return &TxSignCmd{
		PrivkeyStr: privkeyStr,
		RawTxStr:   rawTxStr,
	}
}

type GetMempoolCmd struct {
	TxType  string
	Verbose bool
}

func NewGetMempoolCmd(txType string, verbose bool) *GetMempoolCmd {
	return &GetMempoolCmd{
		TxType:  txType,
		Verbose: verbose,
	}
}

// ws
type NotifyNewTransactionsCmd struct {
	Verbose bool
}

// OutPoint describes a transaction outpoint that will be marshalled to and
// from JSON.
type OutPoint struct {
	Hash  string `json:"hash"`
	Index uint32 `json:"index"`
}

// RescanCmd defines the rescan JSON-RPC command.
//
type RescanCmd struct {
	BeginBlock uint64
	Addresses  []string
	OutPoints  []OutPoint
	EndBlock   uint64
}

type TxConfirm struct {
	Txid          string
	Order         uint64
	Confirmations int32
	EndHeight     uint64
}

type NotifyTxsConfirmedCmd struct {
	Txs []TxConfirm
}

type RemoveTxsConfirmedCmd struct {
	Txs []TxConfirm
}

// ws
type NotifyTxsByAddrCmd struct {
	Reload    bool
	Addresses []string
	OutPoints []OutPoint
}

// ws
type UnNotifyTxsByAddrCmd struct {
	Addresses []string
}

func NewNotifyNewTransactionsCmd(verbose bool) *NotifyNewTransactionsCmd {
	return &NotifyNewTransactionsCmd{
		Verbose: verbose,
	}
}

func NewNotifyTxsConfirmed(txs []TxConfirm) *NotifyTxsConfirmedCmd {
	return &NotifyTxsConfirmedCmd{
		Txs: txs,
	}
}

func NewRemoveTxsConfirmed(txs []TxConfirm) *RemoveTxsConfirmedCmd {
	return &RemoveTxsConfirmedCmd{
		Txs: txs,
	}
}

type StopNotifyNewTransactionsCmd struct{}

func NewStopNotifyNewTransactionsCmd() *StopNotifyNewTransactionsCmd {
	return &StopNotifyNewTransactionsCmd{}
}

func NewNotifyTxsByAddrCmd(reload bool, addr []string, outpoint []OutPoint) *NotifyTxsByAddrCmd {
	return &NotifyTxsByAddrCmd{
		Reload:    reload,
		Addresses: addr,
		OutPoints: outpoint,
	}
}

func StopNotifyTxsByAddrCmd(addr []string) *UnNotifyTxsByAddrCmd {
	return &UnNotifyTxsByAddrCmd{
		Addresses: addr,
	}
}

func init() {
	flags := UsageFlag(0)

	MustRegisterCmd("createRawTransaction", (*CreateRawTransactionCmd)(nil), flags, DefaultServiceNameSpace)
	MustRegisterCmd("decodeRawTransaction", (*DecodeRawTransactionCmd)(nil), flags, DefaultServiceNameSpace)
	MustRegisterCmd("sendRawTransaction", (*SendRawTransactionCmd)(nil), flags, DefaultServiceNameSpace)
	MustRegisterCmd("getRawTransaction", (*GetRawTransactionCmd)(nil), flags, DefaultServiceNameSpace)
	MustRegisterCmd("getUtxo", (*GetUtxoCmd)(nil), flags, DefaultServiceNameSpace)
	MustRegisterCmd("getRawTransactions", (*GetRawTransactionsCmd)(nil), flags, DefaultServiceNameSpace)
	MustRegisterCmd("txSign", (*TxSignCmd)(nil), flags, TestNameSpace)

	MustRegisterCmd("getMempool", (*GetMempoolCmd)(nil), flags, DefaultServiceNameSpace)

	// ws
	MustRegisterCmd("notifynewtransactions", (*NotifyNewTransactionsCmd)(nil), UFWebsocketOnly, NotifyNameSpace)
	MustRegisterCmd("stopnotifynewtransactions", (*StopNotifyNewTransactionsCmd)(nil), UFWebsocketOnly, NotifyNameSpace)

	// ws
	MustRegisterCmd("notifyTxsByAddr", (*NotifyTxsByAddrCmd)(nil), UFWebsocketOnly, NotifyNameSpace)
	MustRegisterCmd("stopnotifyTxsByAddr", (*UnNotifyTxsByAddrCmd)(nil), UFWebsocketOnly, NotifyNameSpace)

	MustRegisterCmd("notifyTxsConfirmed", (*NotifyTxsConfirmedCmd)(nil), flags, NotifyNameSpace)

	MustRegisterCmd("removeTxsConfirmed", (*RemoveTxsConfirmedCmd)(nil), flags, NotifyNameSpace)
}
