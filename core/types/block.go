// Copyright 2017-2018 The qitmeer developers

package types

import (
	"bytes"
	"fmt"
	"github.com/Qitmeer/qitmeer/common/hash"
	s "github.com/Qitmeer/qitmeer/core/serialization"
	"github.com/Qitmeer/qitmeer/core/types/pow"
	"io"
	"math/big"
	"time"
)

// MaxBlockHeaderPayload is the maximum number of bytes a block header can be.
// Version 4 bytes + ParentRoot 32 bytes + TxRoot 32 bytes + StateRoot 32 bytes + Difficulty 4 bytes + Timestamp 4 bytes
// + nonce 8 bytes + powType 1 byte +  edges_bits 1 byte + 42circles 42*4 bytes
// total 117 + 169 = 286
// hash pow only need 117 bytes
// cuckoo need 286 bytes
const MaxBlockHeaderPayload = 4 + (hash.HashSize * 3) + 4 + 4 + 8 + 1 + 1 + 42*4

// MaxBlockPayload is the maximum bytes a block message can be in bytes.
const MaxBlockPayload = 1048576 // 1024*1024 (1MB)

// MaxTxPerBlock is the maximum number of transactions that could
// possibly fit into a block.
const MaxTxPerBlock = (MaxBlockPayload / minTxPayload) + 1

//Limited parents quantity
const MaxParentsPerBlock = 50

// blockHeaderLen is a constant that represents the number of bytes for a block
// header. common header need 117 bytes , proof data need extra 169 bytes
const blockHeaderLen = 117 + 169

// MaxBlocksPerMsg is the maximum number of blocks allowed per message.
const MaxBlocksPerMsg = 500

type BlockHeader struct {

	// block version
	Version uint32

	// The merkle root of the previous parent blocks (the dag layer)
	ParentRoot hash.Hash

	// The merkle root of the tx tree  (tx of the block)
	// included Witness here instead of the separated witness commitment
	TxRoot hash.Hash

	// bip157/158 cbf
	// CompactFilter Hash

	// The merkle root of the stake commits tire
	// for votes/voters/commits/pre-commits/validator/evidence etc, (the pos layer)
	// StakeRoot     Hash

	// The app result/receipt after the tx executed
	// the UTXO commitment also a kind of state result after tx redeemed
	// ResultRoot/ReceiptRoot hash.Hash

	// The Multiset hash of UTXO set or(?) merkle range/path or(?) tire tree root
	// UtxoCommitment      hash.Hash

	// The merkle root of state tire (the app data layer)
	// can all of the state data (stake, receipt, utxo) in state root?
	StateRoot hash.Hash

	// Difficulty target for tx
	Difficulty uint32

	// TimeStamp
	Timestamp time.Time

	// pow blake2bd | cuckaroo | cuckatoo
	Pow pow.IPow

	//might extra data here

	// Size is the size of the serialized block/block-header in its entirety.

	// The variable-sized block might require a size serialized & verify-check
	// BlockSize uint32

}

// BlockHash computes the block identifier hash for the given block header.
func (h *BlockHeader) BlockHash() hash.Hash {
	// Encode the header and hash256 everything prior to the number of
	// transactions.  Ignore the error returns since there is no way the
	// encode could fail except being out of memory which would cause a
	// run-time panic.
	return hash.DoubleHashH(h.BlockData())

}

// BlockData computes the block data for block hash.
// Block data has the dynamic length.
//   - blake2bd data is 117 bytes .
//   - cuckoo data is 282 bytes .
func (bh *BlockHeader) BlockData() []byte {
	// Encode the header and hash256 everything prior to the number of
	// transactions.  Ignore the error returns since there is no way the
	// encode could fail except being out of memory which would cause a
	// run-time panic.
	buf := bytes.NewBuffer(make([]byte, 0, MaxBlockHeaderPayload))
	// TODO, redefine the protocol version and storage
	sec := uint32(bh.Timestamp.Unix())
	_ = s.WriteElements(buf, bh.Version, &bh.ParentRoot, &bh.TxRoot,
		&bh.StateRoot, bh.Difficulty, sec, bh.Pow.BlockData())
	return buf.Bytes()
}

// readBlockHeader reads a block header from io reader.  See Deserialize for
// decoding block headers stored to disk, such as in a database, as opposed to
// decoding from the type.
// TODO, redefine the protocol version and storage
func readBlockHeader(r io.Reader, pver uint32, bh *BlockHeader) error {
	// TODO fix time ambiguous
	return s.ReadElements(r, &bh.Version, &bh.ParentRoot, &bh.TxRoot,
		&bh.StateRoot, &bh.Difficulty, (*s.Uint32Time)(&bh.Timestamp),
		&bh.Pow)
}

// writeBlockHeader writes a block header to w.  See Serialize for
// encoding block headers to be stored to disk, such as in a database, as
// opposed to encoding for the type.
// TODO, redefine the protocol version and storage
func writeBlockHeader(w io.Writer, pver uint32, bh *BlockHeader) error {
	// TODO fix time ambiguous
	sec := uint32(bh.Timestamp.Unix())
	return s.WriteElements(w, bh.Version, &bh.ParentRoot, &bh.TxRoot,
		&bh.StateRoot, bh.Difficulty, sec, bh.Pow.Bytes())
}

// This function get the simple hash use each parents string, so it can't use to
// check for block body .At present we use the merkles tree.
func GetParentsRoot(parents []*hash.Hash) hash.Hash {
	if len(parents) == 0 {
		return hash.Hash{}
	}
	hashStr := ""
	for _, v := range parents {
		hashStr += v.String()
	}
	return hash.DoubleHashH([]byte(hashStr))
}

// Serialize encodes a block header from r into the receiver using a format
// that is suitable for long-term storage such as a database while respecting
// the Version field.
func (h *BlockHeader) Serialize(w io.Writer) error {
	// At the current time, there is no difference between the wire encoding
	// at protocol version 0 and the stable long-term storage format.  As
	// a result, make use of writeBlockHeader.
	return writeBlockHeader(w, 0, h)
}

// Deserialize decodes a block header from r into the receiver using a format
// that is suitable for long-term storage such as a database while respecting
// the Version field.
func (h *BlockHeader) Deserialize(r io.Reader) error {
	// At the current time, there is no difference between the wire encoding
	// at protocol version 0 and the stable long-term storage format.  As
	// a result, make use of readBlockHeader.
	return readBlockHeader(r, 0, h)
}

type Block struct {
	Header       BlockHeader
	Parents      []*hash.Hash
	Transactions []*Transaction //tx  6
	//Commits     []*StakeCommit    //vote for
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

	n := blockHeaderLen + s.VarIntSerializeSize(uint64(len(block.Parents))) + s.VarIntSerializeSize(uint64(len(block.Transactions)))

	for i := 0; i < len(block.Parents); i++ {
		n += hash.HashSize
	}

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
	err = s.WriteVarInt(w, pver, uint64(len(block.Parents)))
	if err != nil {
		return err
	}
	for _, pb := range block.Parents {
		err = s.WriteElements(w, pb)
		if err != nil {
			return err
		}
	}
	//
	err = s.WriteVarInt(w, pver, uint64(len(block.Transactions)))
	if err != nil {
		return err
	}

	for _, tx := range block.Transactions {
		err = tx.Encode(w, pver, TxSerializeFull)
		if err != nil {
			return err
		}
	}
	return nil
}

// Deserialize decodes a block from r into the receiver using a format that is
// suitable for long-term storage such as a database while respecting the
// Version field in the block.
func (b *Block) Deserialize(r io.Reader) error {
	// At the current time, there is no difference between the wire encoding
	// at protocol version 0 and the stable long-term storage format.  As
	// a result, make use of Decode.
	return b.Decode(r, 0)
}

// decodes r into the receiver.
// See Deserialize for decoding blocks stored to disk, such as in a database, as
// opposed to decoding blocks .
func (b *Block) Decode(r io.Reader, pver uint32) error {
	err := readBlockHeader(r, pver, &b.Header)
	if err != nil {
		return err
	}
	//
	pbCount, err := s.ReadVarInt(r, pver)
	if err != nil {
		return err
	}
	if pbCount > MaxParentsPerBlock {
		str := fmt.Sprintf("too many parents to fit into a block "+
			"[count %d, max %d]", pbCount, MaxParentsPerBlock)
		return fmt.Errorf("MsgBlock.BtcDecode %s", str)
	}
	b.Parents = make([]*hash.Hash, 0, pbCount)
	phash := hash.Hash{}
	for i := uint64(0); i < pbCount; i++ {
		err = s.ReadElements(r, &phash)
		if err != nil {
			return err
		}
		ph := phash
		b.Parents = append(b.Parents, &ph)
	}
	//
	txCount, err := s.ReadVarInt(r, pver)
	if err != nil {
		return err
	}

	b.Transactions = make([]*Transaction, 0, txCount)
	for i := uint64(0); i < txCount; i++ {
		var tx Transaction
		err := tx.Deserialize(r)
		if err != nil {
			return err
		}
		b.Transactions = append(b.Transactions, &tx)
	}

	return nil
}

// DeserializeTxLoc decodes r in the same manner Deserialize does, but it takes
// a byte buffer instead of a generic reader and returns a slice containing the
// start and length of each transaction within the raw data that is being
// deserialized.
func (b *Block) DeserializeTxLoc(r *bytes.Buffer) ([]TxLoc, error) {
	fullLen := r.Len()

	// At the current time, there is no difference between the wire encoding
	// at protocol version 0 and the stable long-term storage format.  As
	// a result, make use of existing wire protocol functions.
	err := readBlockHeader(r, 0, &b.Header)
	if err != nil {
		return nil, err
	}
	//
	pbCount, err := s.ReadVarInt(r, 0)
	if err != nil {
		return nil, err
	}
	if pbCount > MaxParentsPerBlock {
		str := fmt.Sprintf("too many parents to fit into a block "+
			"[count %d, max %d]", pbCount, MaxParentsPerBlock)
		return nil, fmt.Errorf("MsgBlock.BtcDecode %s", str)
	}
	b.Parents = make([]*hash.Hash, 0, pbCount)
	phash := hash.Hash{}
	for i := uint64(0); i < pbCount; i++ {
		err = s.ReadElements(r, &phash)
		if err != nil {
			return nil, err
		}
		ph := phash
		b.Parents = append(b.Parents, &ph)
	}

	//

	txCount, err := s.ReadVarInt(r, 0)
	if err != nil {
		return nil, err
	}
	// Prevent more transactions than could possibly fit into a block.
	// It would be possible to cause memory exhaustion and panics without
	// a sane upper bound on this count.
	if txCount > MaxTxPerBlock {
		return nil, fmt.Errorf("Block.DeserializeTxLoc: too many transactions to fit into a block "+
			"[count %d, max %d]", txCount, MaxTxPerBlock)
	}

	// Deserialize each transaction while keeping track of its location
	// within the byte stream.
	b.Transactions = make([]*Transaction, 0, txCount)
	txLocs := make([]TxLoc, txCount)
	for i := uint64(0); i < txCount; i++ {
		txLocs[i].TxStart = fullLen - r.Len()
		var tx Transaction
		err := tx.Deserialize(r)
		if err != nil {
			return nil, err
		}
		b.Transactions = append(b.Transactions, &tx)
		txLocs[i].TxLen = (fullLen - r.Len()) - txLocs[i].TxStart
	}
	return txLocs, nil
}

// AddTransaction adds a transaction to the message.
func (b *Block) AddTransaction(tx *Transaction) error {
	b.Transactions = append(b.Transactions, tx)
	return nil

}

// AddTransaction adds a transaction to the message.
func (b *Block) AddParent(h *hash.Hash) error {
	b.Parents = append(b.Parents, h)
	return nil

}

// SerializedBlock provides easier and more efficient manipulation of raw blocks.
// It also memorizes hashes for the block and its transactions on their first
// access so subsequent accesses don't have to  repeat the relatively expensive
// hashing operations.
type SerializedBlock struct {
	block           *Block    // Underlying Block
	hash            hash.Hash // Cached block hash
	serializedBytes []byte    // Serialized bytes for the block
	transactions    []*Tx     // Transactions
	txnsGenerated   bool      // ALL wrapped transactions generated
	order           uint64    //order is in the position of whole block chain.
	height          uint      //height is in the sub dag chain.
}

// The stringer method makes SerializedBlock satisfy the Stringer interface.
// It simplifies the message printing in the trace logs.
func (sb *SerializedBlock) String() string {
	return fmt.Sprintf("blockhash: %v transactions:%d txnsGenerated:%v order:%d height:%d",
		sb.hash.String(), len(sb.transactions), sb.txnsGenerated, sb.order, sb.height)
}

// NewBlock returns a new instance of the serialized block given an underlying Block.
// the block hash has been calculated and cached
func NewBlock(block *Block) *SerializedBlock {
	return &SerializedBlock{
		hash:  block.BlockHash(),
		block: block,
	}
}

// NewBlockFromBlockAndBytes returns a new instance of a block given
// an underlying Block and the serialized bytes for it.
func NewBlockFromBlockAndBytes(block *Block, serializedBytes []byte) *SerializedBlock {
	return &SerializedBlock{
		hash:            block.BlockHash(),
		block:           block,
		serializedBytes: serializedBytes,
	}
}

// NewBlockDeepCopyCoinbase returns a new instance of a block given an underlying
// wire.MsgBlock, but makes a deep copy of the coinbase transaction since it's
// sometimes mutable.
func NewBlockDeepCopyCoinbase(msgBlock *Block) *SerializedBlock {
	// Copy the msgBlock and the pointers to all the transactions.
	msgBlockCopy := new(Block)

	msgBlockCopy.Parents = msgBlock.Parents

	lenTxs := len(msgBlock.Transactions)
	mtxsCopy := make([]*Transaction, lenTxs)
	copy(mtxsCopy, msgBlock.Transactions)

	msgBlockCopy.Transactions = mtxsCopy

	msgBlockCopy.Header = msgBlock.Header

	// Deep copy the first transaction. Also change the coinbase pointer.
	msgBlockCopy.Transactions[0] =
		NewTxDeep(msgBlockCopy.Transactions[0]).Transaction()

	bl := &SerializedBlock{
		block: msgBlockCopy,
	}
	bl.hash = msgBlock.BlockHash()

	return bl
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

// NewBlockFromBytes returns a new instance of a block given the
// serialized bytes.  See Block.
func NewBlockFromBytes(serializedBytes []byte) (*SerializedBlock, error) {
	br := bytes.NewReader(serializedBytes)
	b, err := NewBlockFromReader(br)
	if err != nil {
		return nil, err
	}
	b.serializedBytes = serializedBytes

	return b, nil
}

// NewBlockFromReader returns a new instance of a block given a
// Reader to deserialize the block.  See Block.
func NewBlockFromReader(r io.Reader) (*SerializedBlock, error) {
	// Deserialize the bytes into a MsgBlock.
	var block Block
	err := block.Deserialize(r)
	if err != nil {
		return nil, err
	}
	sb := NewBlock(&block)
	return sb, nil
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

// Order returns a casted int64 order from the block header.
//
// This function should not be used for new code and will be
// removed in the future.
func (sb *SerializedBlock) Order() uint64 {
	return sb.order
}

func (sb *SerializedBlock) SetOrder(order uint64) {
	sb.order = order
}

func (sb *SerializedBlock) Height() uint {
	return sb.height
}

func (sb *SerializedBlock) SetHeight(height uint) {
	sb.height = height
}

// Transactions returns a slice of wrapped transactions for all
// transactions in the Block.  This is nearly equivalent to accessing the raw
// transactions (types.Transaction) in the underlying types.Block, however it
// instead provides easy access to wrapped versions of them.
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
	CBlockNum *big.Int

	//Parent block hash
	CBlockParent hash.Hash

	// The merkle root of contract storage
	ContractRoot hash.Hash

	// The merkle root the ctx receipt trie  (proof of changes)
	// receipt generated after ctx processed (aka. post-tx info)
	ReceiptRoot hash.Hash

	// bloom filter for log entry of ctx receipt
	// can we remove/combine with cbf ?
	LogBloom hash.Hash

	// Difficulty target for ctx
	CDifficulty uint32
	// Nonce for ctx
	CNonce uint64

	//Do we need to add Coinbase address here?
}

type CBlock struct {
	Header        CBlockHeader
	CTransactions []*ContractTransaction //ctx
}
