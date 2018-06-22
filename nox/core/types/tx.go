// Copyright 2017-2018 The nox developers

package types

import (
	"io"
	"github.com/noxproject/nox/common/hash"
	s "github.com/noxproject/nox/core/serialization"
	"fmt"
	"bytes"
	"encoding/binary"
)

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

// TxIndexUnknown is the value returned for a transaction index that is unknown.
// This is typically because the transaction has not been inserted into a block
// yet.
const TxIndexUnknown = -1

// TxSerializeType represents the serialized type of a transaction.
type TxSerializeType uint16

const (
	// TxSerializeFull indicates a transaction be serialized with the prefix
	// and all witness data.
	TxSerializeFull TxSerializeType = iota
	// TxSerializeNoWitness indicates a transaction be serialized with only
	// the prefix.
	TxSerializeNoWitness
	// TxSerializeOnlyWitness indicates a transaction be serialized with
	// only the witness data.
	TxSerializeOnlyWitness
)

type Transaction struct {
	Version   uint32
	LockTime  uint32
	Expire    uint32
	Type      TxType
	TxIn 	  []*TxInput
	TxOut 	  []*TxOutput
	Message   []byte     //a unencrypted/encrypted message if user pay additional fee & limit the max length
}

// SerializeSize returns the number of bytes it would take to serialize the
// the transaction. (full size)
func (tx *Transaction) SerializeSize() int {
	// Unknown type return 0.
	n := 0

	// Version 4 bytes + LockTime 4 bytes + Expire 4 bytes + Serialized
	// varint size for the number of transaction inputs (x2) and
	// outputs. The number of inputs is added twice because it's
	// encoded once in both the witness and the prefix.
	n = 12 + s.VarIntSerializeSize(uint64(len(tx.TxIn))) +
		s.VarIntSerializeSize(uint64(len(tx.TxIn))) +
		s.VarIntSerializeSize(uint64(len(tx.TxOut)))

	for _, txIn := range tx.TxIn {
		n += txIn.SerializeSizePrefix()
	}
	for _, txIn := range tx.TxIn {
		n += txIn.SerializeSizeWitness()
	}
	for _, txOut := range tx.TxOut {
		n += txOut.SerializeSize()
	}
	return n
}
func (tx *Transaction) SerializeSizeNoWitness() int {
	// Unknown type return 0.
	n := 0
	// Version 4 bytes + LockTime 4 bytes + Expiry 4 bytes +
	// Serialized varint size for the number of transaction
	// inputs and outputs.
	n = 12 + s.VarIntSerializeSize(uint64(len(tx.TxIn))) +
			s.VarIntSerializeSize(uint64(len(tx.TxOut)))

	for _, txIn := range tx.TxIn {
		n += txIn.SerializeSizePrefix()
	}
	for _, txOut := range tx.TxOut {
		n += txOut.SerializeSize()
	}
	return n
}
func (tx *Transaction) SerializeSizeOnlyWitness() int {
	// Unknown type return 0.
	n := 0

	// Version 4 bytes + Serialized varint size for the
	// number of transaction signatures.
	n = 4 + s.VarIntSerializeSize(uint64(len(tx.TxIn)))

	for _, txIn := range tx.TxIn {
		n += txIn.SerializeSizeWitness()
	}
	return n
}

// mustSerialize returns the serialization of the transaction for the provided
// serialization type without modifying the original transaction.  It will panic
// if any errors occur.
func (tx *Transaction) mustSerialize(serType TxSerializeType) []byte {
	serialized, err := tx.serialize(serType)
	if err != nil {
		panic(fmt.Sprintf("MsgTx failed serializing for type %v",
			serType))
	}
	return serialized
}

// serialize returns the serialization of the transaction for the provided
// serialization type without modifying the original transaction.
func (tx *Transaction) serialize(serType TxSerializeType) ([]byte, error) {
	// Shallow copy so the serialization type can be changed without
	// modifying the original transaction.
	txCopy := *tx
	switch serType {
	case TxSerializeNoWitness:
		buf := bytes.NewBuffer(make([]byte, 0, txCopy.SerializeSizeNoWitness()))
		err := tx.encodePrefix(buf,0)
		if err != nil {
			return nil,err
		}
		return buf.Bytes(), nil
	case TxSerializeOnlyWitness:
		buf := bytes.NewBuffer(make([]byte, 0, txCopy.SerializeSizeOnlyWitness()))
		err := tx.encodeWitness(buf, 0)
		if err != nil {
			return nil,err
		}
		return buf.Bytes(), nil
	case TxSerializeFull:
		buf := bytes.NewBuffer(make([]byte, 0, txCopy.SerializeSize()))
		err := tx.encodePrefix(buf,0)
		if err != nil {
			return nil,err
		}
		err = tx.encodeWitness(buf,0)
		if err != nil {
			return nil,err
		}
		return buf.Bytes(), nil
	default:
		return nil, fmt.Errorf("unknown TxSerializeType")
	}
}

// TxHash generates the hash for the transaction prefix.  Since it does not
// contain any witness data, it is not malleable and therefore is stable for
// use in unconfirmed transaction chains.
func (tx *Transaction) TxHash() hash.Hash {
	// TxHash should always calculate a non-witnessed hash.
	return hash.DoubleHashH(tx.mustSerialize(TxSerializeNoWitness))
}

// TxHashWitness generates the hash for the transaction witness.
func (tx *Transaction) TxHashWitness() hash.Hash {
	// TxHashWitness should always calculate a witnessed hash.
	return hash.DoubleHashH(tx.mustSerialize(TxSerializeOnlyWitness))
}

// TxHashFull generates the hash for the transaction prefix || witness. It first
// obtains the hashes for both the transaction prefix and witness, then
// concatenates them and hashes the result.
func (tx *Transaction) TxHashFull() hash.Hash {
	// Note that the inputs to the hashes, the serialized prefix and
	// witness, have different serialized versions because the serialized
	// encoding of the version includes the real transaction version in the
	// lower 16 bits and the transaction serialization type in the upper 16
	// bits.  The real transaction version (lower 16 bits) will be the same
	// in both serializations.
	concat := make([]byte, hash.HashSize*2)
	prefixHash := tx.TxHash()
	witnessHash := tx.TxHashWitness()
	copy(concat[0:], prefixHash[:])
	copy(concat[hash.HashSize:], witnessHash[:])

	return hash.DoubleHashH(concat)
}

func (tx *Transaction) Encode(w io.Writer, pver uint32) error {
	err := s.BinarySerializer.PutUint32(w, binary.LittleEndian, uint32(tx.Version))
	if err != nil {
		return err
	}

	//TODO, make sure the segwit work
	//full serialization here (with witness)

	err = tx.encodePrefix(w, pver)
	if err != nil {
		return err
	}

	err = tx.encodeWitness(w, pver)
	if err != nil {
		return err
	}

	return nil
}

// encodePrefix encodes a transaction prefix into a writer.
func (tx *Transaction) encodePrefix(w io.Writer, pver uint32) error {
	count := uint64(len(tx.TxIn))
	err := s.WriteVarInt(w, pver, count)
	if err != nil {
		return err
	}

	for _, ti := range tx.TxIn {
		err = writeTxInPrefix(w, pver, tx.Version, ti)
		if err != nil {
			return err
		}
	}

	count = uint64(len(tx.TxOut))
	err = s.WriteVarInt(w, pver, count)
	if err != nil {
		return err
	}

	for _, to := range tx.TxOut {
		err = writeTxOut(w, pver, to)
		if err != nil {
			return err
		}
	}

	err = s.BinarySerializer.PutUint32(w, binary.LittleEndian, tx.LockTime)
	if err != nil {
		return err
	}

	return s.BinarySerializer.PutUint32(w, binary.LittleEndian, tx.Expire)
}

// encodeWitness encodes a transaction witness into a writer.
func (tx *Transaction) encodeWitness(w io.Writer, pver uint32) error {
	count := uint64(len(tx.TxIn))
	err := s.WriteVarInt(w, pver, count)
	if err != nil {
		return err
	}

	for _, ti := range tx.TxIn {
		err = writeTxInWitness(w, tx.Version, ti)
		if err != nil {
			return err
		}
	}

	return nil
}

// writeTxInPrefixs encodes for a transaction input (TxIn) prefix to w.
func writeTxInPrefix(w io.Writer, pver uint32, version uint32, ti *TxInput) error {
	err := WriteOutPoint(w, pver, version, &ti.PreviousOut)
	if err != nil {
		return err
	}

	return s.BinarySerializer.PutUint32(w, binary.LittleEndian, ti.Sequence)
}

// WriteOutPoint encodes for an OutPoint to w.
func WriteOutPoint(w io.Writer, pver uint32, version uint32, op *TxOutPoint) error {
	_, err := w.Write(op.Hash[:])
	if err != nil {
		return err
	}

	return s.BinarySerializer.PutUint32(w, binary.LittleEndian, op.OutIndex)
}

// writeTxOut encodes for a transaction output (TxOut) to w.
func writeTxOut(w io.Writer, pver uint32, to *TxOutput) error {
	err := s.BinarySerializer.PutUint64(w, binary.LittleEndian, uint64(to.Amount))
	if err != nil {
		return err
	}
	return s.WriteVarBytes(w, pver, to.PkScript)
}

// writeTxWitness encodes for a transaction input (TxIn) witness to w.
func writeTxInWitness(w io.Writer, pver uint32, ti *TxInput) error {
	// ValueIn.
	err := s.BinarySerializer.PutUint64(w, binary.LittleEndian, uint64(ti.AmountIn))
	if err != nil {
		return err
	}

	// BlockHeight.
	err = s.BinarySerializer.PutUint32(w, binary.LittleEndian, ti.BlockHeight)
	if err != nil {
		return err
	}

	// BlockIndex.
	s.BinarySerializer.PutUint32(w, binary.LittleEndian, ti.BlockTxIndex)
	if err != nil {
		return err
	}

	// Write the signature script.
	return s.WriteVarBytes(w, pver, ti.SignScript)
}

// Tx defines a transaction that provides easier and more efficient manipulation
// of raw transactions.  It also memorizes the hash for the transaction on its
// first access so subsequent accesses don't have to repeat the relatively
// expensive hashing operations.
type Tx struct {
	Tx      *Transaction   // Underlying Transaction
	hash    hash.Hash           // Cached transaction hash
	txIndex int            // Position within a block or TxIndexUnknown
}

// Transaction() returns the underlying Tx for the transaction.
func (t *Tx) Transaction() *Transaction {
	// Return the cached transaction.
	return t.Tx
}
// NewTx returns a new instance of a transaction given an underlying
// wire.MsgTx.  See Tx.
func NewTx(t *Transaction) *Tx {
	return &Tx{
		hash:    t.TxHash(),
		Tx:   t,
		txIndex: TxIndexUnknown,
	}
}

type TxOutPoint struct {
	Hash       hash.Hash       //txid
	OutIndex   uint32     //vout
}
type TxInput struct {
	PreviousOut TxOutPoint
	// This number has more historical significance than relevant usage;
	// however, its most relevant purpose is to enable “locking” of payments
	// so that they cannot be redeemed until a certain time.
	Sequence         uint32     //work with LockTime (disable use 0xffffffff, bitcoin historical)

	// Witness part
	// the Fraud proof, input exist to block/tx, useful for SPV
	// see also https://gist.github.com/justusranvier/451616fa4697b5f25f60#invalid-transaction-input-already-spent
	AmountIn         uint64
	BlockHeight      uint32   //might a block hash (?)
	BlockTxIndex     uint32

	//the signature script (or witness script? or redeem script?)
	SignScript       []byte
}

// SerializeSizePrefix returns the number of bytes it would take to serialize
// the transaction input for a prefix.
func (ti *TxInput) SerializeSizePrefix() int {
	// Outpoint Hash 32 bytes + Outpoint Index 4 bytes + Sequence 4 bytes.
	return 40
}

// SerializeSizeWitness returns the number of bytes it would take to serialize the
// transaction input for a witness.
func (ti *TxInput) SerializeSizeWitness() int {
	// ValueIn (8 bytes) + BlockHeight (4 bytes) + BlockTxIndex (4 bytes) +
	// serialized varint size for the length of SignScript +
	// SignScript bytes.
	return 8 + 4 + 4 + s.VarIntSerializeSize(uint64(len(ti.SignScript))) +
		len(ti.SignScript)
}

type TxOutput struct {
	Amount     uint64
	PkScript   []byte    //Here, asm/type -> OP_XXX OP_RETURN
}

// SerializeSize returns the number of bytes it would take to serialize the
// the transaction output.
func (to *TxOutput) SerializeSize() int {
	// Value 8 bytes + Version 2 bytes + serialized varint size for
	// the length of PkScript + PkScript bytes.
	return 8 + 2 + s.VarIntSerializeSize(uint64(len(to.PkScript))) + len(to.PkScript)
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
