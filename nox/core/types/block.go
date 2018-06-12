// Copyright 2017-2018 The nox developers

package types

import "math/big"

type BlockHeader struct {

	// block id/hash
	Hash   Hash

	// block number
	Height uint64

	// DAG references to previous blocks
	Parents     []Hash

	// The merkle root of the tx tree  (tx of the block)
	// included Witness here instead of the separated witness commitment
	TxRoot      Hash

	// The merkle root of the stake tx tree
	// STxRoot     Hash

	// The Multiset hash of UTXO set or(?) merkle range/path or(?) tire tree root
	UtxoCommitment Hash

	// bip157/158 cbf
	CompactFilter Hash

	// The merkle root of state tire
	StateRoot	Hash

	// Nonce
	Nonce       uint64
	// Difficulty target for tx
	Difficulty  uint32

	// TimeStamp (might big.Int, if we don't want limitation in the future)
	Timestamp   uint64

	//might extra data here

}

type Block struct {
	Header        BlockHeader
	Transactions  []Transaction    //tx
	// STransactions []Transaction    //stx
}

// Contract block header
type CBlockHeader struct {

	//Contract block number
	CBlockNum      *big.Int

	//Parent block hash
    CBlockParent   Hash

	// The merkle root of contract storage
	ContractRoot Hash

	// The merkle root the ctx receipt trie  (proof of changes)
	// receipt generated after ctx processed (aka. post-tx info)
	ReceiptRoot Hash

	// bloom filter for log entry of ctx receipt
	// can we remove/combine with cbf ?
	LogBloom    Hash

	// Difficulty target for ctx
	CDifficulty  uint32
	// Nonce for ctx
	CNonce       uint64

	//Do we need to add Coinbase address here?
}

type CBlock struct {
	Header        CBlockHeader
	CTransactions []ContractTransaction    //ctx
}
