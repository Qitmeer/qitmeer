// Copyright 2017-2018 The nox developers

package types

type TxType byte

const (
	CoinBase         TxType = 0x01
	Leger            TxType = 0x02
	AssetIssue       TxType = 0xa0
	AssetRevoke      TxType = 0xa1
	ContractCreate   TxType = 0xc0
	ContractDestroy  TxType = 0xc1
	ContractUpdate   TxType = 0xc2
)

type Transaction struct {
	Id        Hash
	Version   uint32
	LockTime  uint32
	Expire    uint32
	Type      TxType
	TxIn 	  []*TxInput
	TxOut 	  []*TxOutput
	Witness   []*TxInWitness
	Message   []byte     //a unencrypted/encrypted message if user pay additional fee & limit the max length
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

type ContractTransaction struct {
	From Account
	To Account
	Value        uint64
	GasPrice     uint64
	GasLimit     uint64
	Nonce        uint64
	Input        []byte
	Signature    []byte
}
