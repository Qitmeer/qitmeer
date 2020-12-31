package cmds

import "github.com/Qitmeer/qitmeer/services/tx"

type CreateRawTransactionCmd struct {
	Inputs   []tx.TransactionInput
	Amounts  tx.Amounts
	LockTime int64
}

func NewCreateRawTransactionCmd(inputs []tx.TransactionInput, amounts tx.Amounts, lockTime int64) *CreateRawTransactionCmd {
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
}
