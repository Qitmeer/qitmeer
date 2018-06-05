// Copyright 2017-2018 The nox developers

package types


type TxType byte

const (
	CoinBase         TxType = 0x01
	Leger            TxType = 0x02
	ContractCreate   TxType = 0x03
	ContractExecute  TxType = 0x04
)

type Transaction struct {
	Id  Hash
	Version byte
	Type TxType
	Nonce uint64
	Message []byte
	Payload []byte
	Signature []byte
}

type TxInput []byte
type TxOutput []byte

type LegerTxPayload struct{
	Inputs []TxInput
	Outputs []TxOutput
}

type ContractTxPayLoad struct {
	ContractAddr Account
	GasPrice     uint64
	GasLimit     uint64
	Payload      []byte
}

type ContractCreatePayload struct {
	Code []byte
}

type ContractExecutePayload struct {
	Code   []byte
	Method []byte
	Params []byte
}



