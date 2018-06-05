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
	Id        Hash
	Version   uint32
	LockTime  uint32
	Expire    uint32
	Type      TxType
	Nonce     uint64
	Message   []byte
	Payload   []byte
	Signature []byte
}

type TxOutRef struct {
	Hash       Hash       //txid
	OutIndex   uint32     //vout
}
type TxInput struct {


	// This number has more historical significance than relevant usage;
	// however, its most relevant purpose is to enable “locking” of payments
	// so that they cannot be redeemed until a certain time.
	/*
	Sequence         uint32     //work with LockTime (disable use 0xffffffff, bitcoin historical)
	*/

	//ref to the block/tx,
	AmountIn         uint64
	BlockHeight      uint32
	TxIndex          uint32
	TxOutRef

	SignScript       []byte
}

type TxOutput struct {
	Amount     uint64
	PkScript   []byte    //Here, asm/type -> OP_XXX OP_RETURN
}

type LegerTxPayload struct{
	Inputs []TxInput
	Outputs []TxOutput
}

type ContractTxPayload struct {
	LegerTxPayload
	GasPrice     uint64
	GasLimit     uint64
}

type ContractCreateTxPayload struct {
	ContractTxPayload
	Code []byte
}

type ContractExecuteTxPayload struct {
	ContractTxPayload
	ContractAddr Account
	Method []byte
	Params []byte
}



