// Copyright 2017-2018 The nox developers

package types

import (
	"math/big"
	"time"
	"bytes"
	"github.com/noxproject/nox/common/hash"
	"io"
	s "github.com/noxproject/nox/core/serialization"
	"fmt"
)

// MaxBlockHeaderPayload is the maximum number of bytes a block header can be.
// Version 4 bytes + ParentRoot 32 bytes + TxRoot 32 bytes + StateRoot 32 bytes
// Difficulty 4 bytes + Height 4 bytes  + Timestamp 4 bytes + Nonce 8 bytes
// --> Total 120 bytes.
const MaxBlockHeaderPayload = 4 + 4 + 4 + 4 + 8 + (hash.HashSize * 3)

// MaxBlockPayload is the maximum bytes a block message can be in bytes.
// After Segregated Witness, the max block payload has been raised to 4MB.
const MaxBlockPayload = 4000000

// maxTxPerBlock is the maximum number of transactions that could
// possibly fit into a block.
const maxTxPerBlock = (MaxBlockPayload / minTxPayload) + 1

// blockHeaderLen is a constant that represents the number of bytes for a block
// header.
const blockHeaderLen = 180

type BlockHeader struct {

	// block version
	Version   uint32

	// The merkle root of the previous parent blocks
	ParentRoot    hash.Hash

	// The merkle root of the tx tree  (tx of the block)
	// included Witness here instead of the separated witness commitment
	TxRoot      hash.Hash

	// The merkle root of the stake tx tree
	// STxRoot     Hash

	// The Multiset hash of UTXO set or(?) merkle range/path or(?) tire tree root
	// UtxoCommitment Hash

	// bip157/158 cbf
	// CompactFilter Hash

	// The merkle root of state tire
	StateRoot	hash.Hash

	// Difficulty target for tx
	Difficulty  uint32

	// block number
	Height uint64

	// TimeStamp
	Timestamp   time.Time

	// Nonce
	Nonce       uint64

	//might extra data here

}

// BlockHash computes the block identifier hash for the given block header.
func (h *BlockHeader) BlockHash() hash.Hash {
	// Encode the header and hash256 everything prior to the number of
	// transactions.  Ignore the error returns since there is no way the
	// encode could fail except being out of memory which would cause a
	// run-time panic.
	buf := bytes.NewBuffer(make([]byte, 0, MaxBlockHeaderPayload))
	// TODO, redefine the protocol version and storage
	_ = writeBlockHeader(buf,0, h)
	// TODO, add an abstract layer of hash func
	// TODO, double sha256 or other crypto hash
	return hash.DoubleHashH(buf.Bytes())
}

// readBlockHeader reads a block header from io reader.  See Deserialize for
// decoding block headers stored to disk, such as in a database, as opposed to
// decoding from the type.
// TODO, redefine the protocol version and storage
func readBlockHeader(r io.Reader,pver uint32, bh *BlockHeader) error {
	// TODO fix time ambiguous
	return s.ReadElements(r, &bh.Version, &bh.ParentRoot, &bh.TxRoot,
		&bh.StateRoot, &bh.Difficulty, &bh.Height, (*s.Uint32Time)(&bh.Timestamp),
		&bh.Nonce)
}

// writeBlockHeader writes a block header to w.  See Serialize for
// encoding block headers to be stored to disk, such as in a database, as
// opposed to encoding for the type.
// TODO, redefine the protocol version and storage
func writeBlockHeader(w io.Writer, pver uint32, bh *BlockHeader) error {
	// TODO fix time ambiguous
	sec := uint32(bh.Timestamp.Unix())
	return s.WriteElements(w, bh.Version, &bh.ParentRoot, &bh.TxRoot,
		&bh.StateRoot, bh.Difficulty, bh.Height, sec, bh.Nonce)
}

type Block struct {
	Header        BlockHeader
	Parents       []*hash.Hash
	Transactions  []*Transaction    //tx
	// STransactions []Transaction    //stx
}

// BlockHash computes the block identifier hash for this block.
func (block *Block) BlockHash() hash.Hash {
	return block.Header.BlockHash()
}

// SerializeSize returns the number of bytes it would take to serialize the
// the block.
func (block *Block) SerializeSize() int {
	// Check to make sure that all transactions have the correct
	// type and version to be included in a block.

	// Block header bytes + Serialized varint size for the number of
	// transactions + Serialized varint size for the number of
	// stake transactions

	n := blockHeaderLen + s.VarIntSerializeSize(uint64(len(block.Transactions))) +
		s.VarIntSerializeSize(uint64(len(block.Parents)))

	for _, tx := range block.Transactions {
		n += tx.SerializeSize()
	}

	//TODO, handle parents

	return n
}
// Serialize encodes the block to w using a format that suitable for long-term
// storage such as a database while respecting the Version field in the block.
func (block *Block) Serialize(w io.Writer) error {
	// At the current time, there is no difference between the wire encoding
	// at protocol version 0 and the stable long-term storage format.
	// TODO, redefine the protocol version and storage
	return block.Encode(w, 0)
}

// Encode encodes the receiver to w using the protocol encoding.
// This is part of the Message interface implementation.
// See Serialize for encoding blocks to be stored to disk, such as in a
// database, as opposed to encoding blocks for the wire.
func (block *Block) Encode(w io.Writer, pver uint32) error {
	err := writeBlockHeader(w, pver, &block.Header)
	if err != nil {
		return err
	}

	//TODO, write block.Parents

	err = s.WriteVarInt(w, pver, uint64(len(block.Transactions)))
	if err != nil {
		return err
	}

	for _, tx := range block.Transactions {
		err = tx.Encode(w, pver)
		if err != nil {
			return err
		}
	}
	return nil
}

// DeserializeTxLoc decodes r in the same manner Deserialize does, but it takes
// a byte buffer instead of a generic reader and returns a slice containing the
// start and length of each transaction within the raw data that is being
// deserialized.
func (msg *Block) DeserializeTxLoc(r *bytes.Buffer) ([]TxLoc, error) {
	fullLen := r.Len()

	// At the current time, there is no difference between the wire encoding
	// at protocol version 0 and the stable long-term storage format.  As
	// a result, make use of existing wire protocol functions.
	err := readBlockHeader(r, 0, &msg.Header)
	if err != nil {
		return nil, err
	}

	txCount, err := s.ReadVarInt(r, 0)
	if err != nil {
		return nil, err
	}
	// Prevent more transactions than could possibly fit into a block.
	// It would be possible to cause memory exhaustion and panics without
	// a sane upper bound on this count.
	if txCount > maxTxPerBlock {
		return nil, fmt.Errorf("Block.DeserializeTxLoc: too many transactions to fit into a block "+
			"[count %d, max %d]", txCount, maxTxPerBlock)
	}

	// Deserialize each transaction while keeping track of its location
	// within the byte stream.
	msg.Transactions = make([]*Transaction, 0, txCount)
	txLocs := make([]TxLoc, txCount)
	for i := uint64(0); i < txCount; i++ {
		txLocs[i].TxStart = fullLen - r.Len()
		var tx Transaction
		err := tx.Deserialize(r)
		if err != nil {
			return nil, err
		}
		msg.Transactions = append(msg.Transactions, &tx)
		txLocs[i].TxLen = (fullLen - r.Len()) - txLocs[i].TxStart
	}
	return txLocs, nil
}

// SerializedBlock provides easier and more efficient manipulation of raw blocks.
// It also memorizes hashes for the block and its transactions on their first
// access so subsequent accesses don't have to  repeat the relatively expensive
// hashing operations.
type SerializedBlock struct {
	block                 *Block          // Underlying Block
	hash                  hash.Hash       // Cached block hash
	serializedBytes       []byte          // Serialized bytes for the block
	transactions          []*Tx           // Transactions
	txnsGenerated         bool            // ALL wrapped transactions generated
}

// NewBlock returns a new instance of the serialized block given an underlying Block.
// the block hash has been calculated and cached
func NewBlock(block *Block) *SerializedBlock {
	return &SerializedBlock{
		hash:   block.BlockHash(),
		block: 	block,
	}
}

// Hash returns the block identifier hash for the Block.  This is equivalent to
// calling BlockHash on the underlying Block, however it caches the
// result so subsequent calls are more efficient.
func (sb *SerializedBlock) Hash() *hash.Hash {
	//TODO, might need to assertBlockImmutability
	return &sb.hash
}

func (sb *SerializedBlock) Block() *Block {
	return sb.block
}

// Bytes returns the serialized bytes for the Block.  This is equivalent to
// calling Serialize on the underlying Block, however it caches the
// result so subsequent calls are more efficient.
func (sb *SerializedBlock) Bytes() ([]byte, error) {
	// Return the cached serialized bytes if it has already been generated.
	if len(sb.serializedBytes) != 0 {
		return sb.serializedBytes, nil
	}

	// Serialize the MsgBlock.
	var w bytes.Buffer
	w.Grow(sb.block.SerializeSize())
	err := sb.block.Serialize(&w)
	if err != nil {
		return nil, err
	}
	serialized := w.Bytes()

	// Cache the serialized bytes and return them.
	sb.serializedBytes = serialized
	return serialized, nil
}

// TxLoc returns the offsets and lengths of each transaction in a raw block.
// It is used to allow fast indexing into transactions within the raw byte
// stream.
func (sb *SerializedBlock) TxLoc() ([]TxLoc, error) {
	rawMsg, err := sb.Bytes()
	if err != nil {
		return nil, err
	}
	rbuf := bytes.NewBuffer(rawMsg)

	var mblock Block
	txLocs, err := mblock.DeserializeTxLoc(rbuf)
	if err != nil {
		return nil, err
	}
	return txLocs, err
}

// Height returns a casted int64 height from the block header.
//
// This function should not be used for new code and will be
// removed in the future.
func (sb *SerializedBlock) Height() int64 {
	return int64(sb.block.Header.Height)
}

// Transactions returns a slice of wrapped transactions (btcutil.Tx) for all
// transactions in the Block.  This is nearly equivalent to accessing the raw
// transactions (wire.MsgTx) in the underlying wire.MsgBlock, however it
// instead provides easy access to wrapped versions (btcutil.Tx) of them.
func (sb *SerializedBlock) Transactions() []*Tx {
	// Return transactions if they have ALL already been generated.  This
	// flag is necessary because the wrapped transactions are lazily
	// generated in a sparse fashion.
	if sb.txnsGenerated {
		return sb.transactions
	}

	// Generate slice to hold all of the wrapped transactions if needed.
	if len(sb.transactions) == 0 {
		sb.transactions = make([]*Tx, len(sb.block.Transactions))
	}

	// Generate and cache the wrapped transactions for all that haven't
	// already been done.
	for i, tx := range sb.transactions {
		if tx == nil {
			newTx := NewTx(sb.block.Transactions[i])
			newTx.SetIndex(i)
			sb.transactions[i] = newTx
		}
	}

	sb.txnsGenerated = true
	return sb.transactions
}


// Contract block header
type CBlockHeader struct {

	//Contract block number
	CBlockNum      *big.Int

	//Parent block hash
    CBlockParent   hash.Hash

	// The merkle root of contract storage
	ContractRoot hash.Hash

	// The merkle root the ctx receipt trie  (proof of changes)
	// receipt generated after ctx processed (aka. post-tx info)
	ReceiptRoot hash.Hash

	// bloom filter for log entry of ctx receipt
	// can we remove/combine with cbf ?
	LogBloom    hash.Hash

	// Difficulty target for ctx
	CDifficulty  uint32
	// Nonce for ctx
	CNonce       uint64

	//Do we need to add Coinbase address here?
}

type CBlock struct {
	Header        CBlockHeader
	CTransactions []*ContractTransaction    //ctx
}
