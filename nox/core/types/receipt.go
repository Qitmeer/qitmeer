// Copyright 2017-2018 The nox developers

package types

// block execution event, ex. validator change event
type Event []byte

type BlockReceipt struct{

	BlockNumber uint64

	// The merkle root of storage trie (for kv)
	StorageRoot Hash256

	// The merkle root for change logs trie
	// 1.) ordered storage keys that changed the storage, k also map to tx(one or many) (how them changed)
	// 2.) a parent MT also linked, which representing changes happened over the last change range of blocks.
	//     which refer to block(s) that caused the change rather than the tx included from the current block
	//     which provide proof that any group of change range blocks either don't change a storage item
	//     or can find out which tx exactly did.
	ChangeRoot  Hash256

	// The event digest of a block, which is chain-specific, useful for light-clients.
	// Events are an array of data fragments that runtime encode events for light-client
	// for example authority set changes (ex. validator set changes)
	// fast warp sync by tracking validator set changes trustlessly
	Events      []Event
}







