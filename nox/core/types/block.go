// Copyright 2017-2018 The nox developers

package types

type BlockHeader struct {
	Hash   Hash

	// block number
	Height uint64

	// DAG references to previous blocks
	ParentHashes     []Hash

	// state
	// The merkle root of the tx tree    (tx of the block)
	TxRoot      Hash
	// The merkle root of the stake tx tree
	StakeRoot   Hash
	// The merkle root of UTXO set
	UtxoRoot    Hash
	// The merkle root of state tire
	StorageRoot	Hash
	// The merkle root the receipt trie  (proof of changes)
	ReceiptRoot Hash

	// coinbase
	Coinbase    Address

	// sign
	Signature   []byte
}

type Block struct {
	Header       BlockHeader
	Transactions []Transaction    //tx
	STransctions []Transaction    //stx
	CTransctions []Transaction    //ctx
}
