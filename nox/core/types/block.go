// Copyright 2017-2018 The nox developers

package types

type BlockHeader struct {
	Hash   Hash

	// block number
	Height uint64

	// DAG references to previous blocks
	Parents     []Hash

	// state
	// The merkle root of the leger tx tree    (tx of the block)
	// included Witness here instead of the separated witness commitment
	TxRoot      Hash
	// The merkle root of the stake tx tree
	STxRoot     Hash
	// The merkle root of the contact tx tree
	CTxRoot     Hash

	// The Multiset hash of UTXO set or(?) merkle range/path or(?) tire tree root
	Utxo        Hash

	// The merkle root of state tire
	StateRoot	Hash
	// The merkle root the receipt trie  (proof of changes)
	ReceiptRoot Hash

	// Nonce
	Nonce       uint64

	// TimeStamp (might big.Int, if we don't want limitation in the future)
	Timestamp   uint64

	//might extra data here
}

type Block struct {
	Header       BlockHeader
	Transactions []Transaction    //tx
	STransctions []Transaction    //stx
	CTransctions []Transaction    //ctx
}
