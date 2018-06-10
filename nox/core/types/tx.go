// Copyright 2017-2018 The nox developers

package types

type TxType byte

const (
	Leger             TxType = 0x01
	CoinBase          TxType = 0xa0
	ContractTransfer  TxType = 0xc0
	ContractCreate    TxType = 0xc1
	ContractExecute   TxType = 0xc2
)

type Transaction struct {
	Id        Hash
	Version   uint32
	LockTime  uint32
	Expire    uint32
	Type      TxType
	Payload   []byte
	Signature []byte
}

type TxOutPoint struct {
	Hash       Hash       //txid
	OutIndex   uint32     //vout
}
type TxInput struct {
	PreviousOut TxOutPoint
	// This number has more historical significance than relevant usage;
	// however, its most relevant purpose is to enable “locking” of payments
	// so that they cannot be redeemed until a certain time.
	Sequence         uint32     //work with LockTime (disable use 0xffffffff, bitcoin historical)
}

// a witness of a txInput
type TxInWitness struct {

	//Fraud proof, input exist to block/tx, useful for SPV
	//see also https://gist.github.com/justusranvier/451616fa4697b5f25f60#invalid-transaction-input-already-spent
	AmountIn         uint64
	BlockHeight      uint32   //might a block hash (?)
	BlockTxIndex     uint32

	//the signature script (or witness script? or redeem script?)
	SignScript       []byte
}

type TxOutput struct {
	Amount     uint64
	PkScript   []byte    //Here, asm/type -> OP_XXX OP_RETURN
}

type LegerTxPayload struct{
	TxIn 	  []TxInput
	TxOut 	  []TxOutput
	Message   []byte     //a unencrypted/encrypted message if user pay additional fee & limit the max length
}

type LegerTxWitnessPayload struct {
	LegerTxPayload
	Witness []TxInWitness
}

type ContractTxHeader struct {
	From Account
	To Account
	GasPrice     uint64
	GasLimit     uint64
	Nonce        uint64
}

type ContractTransferTxPayload struct {
	ContractTxHeader
	Value uint64
}

type ContractCreateTxPayload struct {
	ContractTxHeader
	Code []byte
}

type ContractExecuteTxPayload struct {
	ContractTxHeader
	Input []byte
}
