// Copyright (c) 2013-2016 The btcsuite developers
// Copyright (c) 2015-2016 The Decred developers
// Use of this source code is governed by an ISC
// license that can be found in the LICENSE file.

package chaincfg

import (
	"math"
	"fmt"
	"bytes"
	"io"
	"encoding/binary"
	"github.com/dindinw/dagproject/chaincfg_dcr/chainhash"
)

const (
	// MaxVarIntPayload is the maximum payload size for a variable length integer.
	MaxVarIntPayload = 9
	// binaryFreeListMaxItems is the number of buffers to keep in the free
	// list to use for binary serialization and deserialization.
	binaryFreeListMaxItems = 1024
)

// binaryFreeList defines a concurrent safe free list of byte slices (up to the
// maximum number defined by the binaryFreeListMaxItems constant) that have a
// cap of 8 (thus it supports up to a uint64).  It is used to provide temporary
// buffers for serializing and deserializing primitive numbers to and from their
// binary encoding in order to greatly reduce the number of allocations
// required.
//
// For convenience, functions are provided for each of the primitive unsigned
// integers that automatically obtain a buffer from the free list, perform the
// necessary binary conversion, read from or write to the given io.Reader or
// io.Writer, and return the buffer to the free list.
type binaryFreeList chan []byte

// Borrow returns a byte slice from the free list with a length of 8.  A new
// buffer is allocated if there are not any available on the free list.
func (l binaryFreeList) Borrow() []byte {
	var buf []byte
	select {
	case buf = <-l:
	default:
		buf = make([]byte, 8)
	}
	return buf[:8]
}

// Return puts the provided byte slice back on the free list.  The buffer MUST
// have been obtained via the Borrow function and therefore have a cap of 8.
func (l binaryFreeList) Return(buf []byte) {
	select {
	case l <- buf:
	default:
		// Let it go to the garbage collector.
	}
}

// binarySerializer provides a free list of buffers to use for serializing and
// deserializing primitive integer values to and from io.Readers and io.Writers.
var binarySerializer binaryFreeList = make(chan []byte, binaryFreeListMaxItems)

// PutUint8 copies the provided uint8 into a buffer from the free list and
// writes the resulting byte to the given writer.
func (l binaryFreeList) PutUint8(w io.Writer, val uint8) error {
	buf := l.Borrow()[:1]
	buf[0] = val
	_, err := w.Write(buf)
	l.Return(buf)
	return err
}

// PutUint16 serializes the provided uint16 using the given byte order into a
// buffer from the free list and writes the resulting two bytes to the given
// writer.
func (l binaryFreeList) PutUint16(w io.Writer, byteOrder binary.ByteOrder, val uint16) error {
	buf := l.Borrow()[:2]
	byteOrder.PutUint16(buf, val)
	_, err := w.Write(buf)
	l.Return(buf)
	return err
}

// PutUint32 serializes the provided uint32 using the given byte order into a
// buffer from the free list and writes the resulting four bytes to the given
// writer.
func (l binaryFreeList) PutUint32(w io.Writer, byteOrder binary.ByteOrder, val uint32) error {
	buf := l.Borrow()[:4]
	byteOrder.PutUint32(buf, val)
	_, err := w.Write(buf)
	l.Return(buf)
	return err
}

// PutUint64 serializes the provided uint64 using the given byte order into a
// buffer from the free list and writes the resulting eight bytes to the given
// writer.
func (l binaryFreeList) PutUint64(w io.Writer, byteOrder binary.ByteOrder, val uint64) error {
	buf := l.Borrow()[:8]
	byteOrder.PutUint64(buf, val)
	_, err := w.Write(buf)
	l.Return(buf)
	return err
}

// Serialize encodes the transaction to w using a format that suitable for
// long-term storage such as a database while respecting the Version field in
// the transaction.  This function differs from BtcEncode in that BtcEncode
// encodes the transaction to the decred wire protocol in order to be sent
// across the network.  The wire encoding can technically differ depending on
// the protocol version and doesn't even really need to match the format of a
// stored transaction at all.  As of the time this comment was written, the
// encoded transaction is the same in both instances, but there is a distinct
// difference and separating the two allows the API to be flexible enough to
// deal with changes.
func (msg *MsgTx) Serialize(w io.Writer) error {
	// At the current time, there is no difference between the wire encoding
	// at protocol version 0 and the stable long-term storage format.  As
	// a result, make use of BtcEncode.
	return msg.BtcEncode(w, 0)
}

// Serialize encodes the block to w using a format that suitable for long-term
// storage such as a database while respecting the Version field in the block.
// This function differs from BtcEncode in that BtcEncode encodes the block to
// the decred wire protocol in order to be sent across the network.  The wire
// encoding can technically differ depending on the protocol version and doesn't
// even really need to match the format of a stored block at all.  As of the
// time this comment was written, the encoded block is the same in both
// instances, but there is a distinct difference and separating the two allows
// the API to be flexible enough to deal with changes.
func (msg *MsgBlock) Serialize(w io.Writer) error {
	// At the current time, there is no difference between the wire encoding
	// at protocol version 0 and the stable long-term storage format.  As
	// a result, make use of BtcEncode.
	return msg.BtcEncode(w, 0)
}

// BtcEncode encodes the receiver to w using the decred protocol encoding.
// This is part of the Message interface implementation.
// See Serialize for encoding blocks to be stored to disk, such as in a
// database, as opposed to encoding blocks for the wire.
func (msg *MsgBlock) BtcEncode(w io.Writer, pver uint32) error {
	err := writeBlockHeader(w, pver, &msg.Header)
	if err != nil {
		return err
	}

	err = WriteVarInt(w, pver, uint64(len(msg.Transactions)))
	if err != nil {
		return err
	}

	for _, tx := range msg.Transactions {
		err = tx.BtcEncode(w, pver)
		if err != nil {
			return err
		}
	}

	err = WriteVarInt(w, pver, uint64(len(msg.STransactions)))
	if err != nil {
		return err
	}

	for _, tx := range msg.STransactions {
		err = tx.BtcEncode(w, pver)
		if err != nil {
			return err
		}
	}

	return nil
}

// BtcEncode encodes the receiver to w using the Decred protocol encoding.
// This is part of the Message interface implementation.
// See Serialize for encoding transactions to be stored to disk, such as in a
// database, as opposed to encoding transactions for the wire.
func (msg *MsgTx) BtcEncode(w io.Writer, pver uint32) error {
	// The serialized encoding of the version includes the real transaction
	// version in the lower 16 bits and the transaction serialization type
	// in the upper 16 bits.
	serializedVersion := uint32(msg.Version) | uint32(msg.SerType)<<16
	err := binarySerializer.PutUint32(w, binary.LittleEndian, serializedVersion)
	if err != nil {
		return err
	}

	switch msg.SerType {
	case TxSerializeNoWitness:
		err := msg.encodePrefix(w, pver)
		if err != nil {
			return err
		}

	case TxSerializeOnlyWitness:
		err := msg.encodeWitness(w, pver)
		if err != nil {
			return err
		}

	case TxSerializeWitnessSigning:
		err := msg.encodeWitnessSigning(w, pver)
		if err != nil {
			return err
		}

	case TxSerializeWitnessValueSigning:
		err := msg.encodeWitnessValueSigning(w, pver)
		if err != nil {
			return err
		}

	case TxSerializeFull:
		err := msg.encodePrefix(w, pver)
		if err != nil {
			return err
		}
		err = msg.encodeWitness(w, pver)
		if err != nil {
			return err
		}

	default:
		return messageError("MsgTx.BtcEncode", "unsupported transaction type")
	}

	return nil
}

// encodePrefix encodes a transaction prefix into a writer.
func (msg *MsgTx) encodePrefix(w io.Writer, pver uint32) error {
	count := uint64(len(msg.TxIn))
	err := WriteVarInt(w, pver, count)
	if err != nil {
		return err
	}

	for _, ti := range msg.TxIn {
		err = writeTxInPrefix(w, pver, msg.Version, ti)
		if err != nil {
			return err
		}
	}

	count = uint64(len(msg.TxOut))
	err = WriteVarInt(w, pver, count)
	if err != nil {
		return err
	}

	for _, to := range msg.TxOut {
		err = writeTxOut(w, pver, msg.Version, to)
		if err != nil {
			return err
		}
	}

	err = binarySerializer.PutUint32(w, binary.LittleEndian, msg.LockTime)
	if err != nil {
		return err
	}

	return binarySerializer.PutUint32(w, binary.LittleEndian, msg.Expiry)
}

// encodeWitness encodes a transaction witness into a writer.
func (msg *MsgTx) encodeWitness(w io.Writer, pver uint32) error {
	count := uint64(len(msg.TxIn))
	err := WriteVarInt(w, pver, count)
	if err != nil {
		return err
	}

	for _, ti := range msg.TxIn {
		err = writeTxInWitness(w, pver, msg.Version, ti)
		if err != nil {
			return err
		}
	}

	return nil
}

// writeTxWitness encodes ti to the decred protocol encoding for a transaction
// input (TxIn) witness to w.
func writeTxInWitness(w io.Writer, pver uint32, version uint16, ti *TxIn) error {
	// ValueIn.
	err := binarySerializer.PutUint64(w, binary.LittleEndian, uint64(ti.ValueIn))
	if err != nil {
		return err
	}

	// BlockHeight.
	err = binarySerializer.PutUint32(w, binary.LittleEndian, ti.BlockHeight)
	if err != nil {
		return err
	}

	// BlockIndex.
	binarySerializer.PutUint32(w, binary.LittleEndian, ti.BlockIndex)
	if err != nil {
		return err
	}

	// Write the signature script.
	return WriteVarBytes(w, pver, ti.SignatureScript)
}

// WriteVarBytes serializes a variable length byte array to w as a varInt
// containing the number of bytes, followed by the bytes themselves.
func WriteVarBytes(w io.Writer, pver uint32, bytes []byte) error {
	slen := uint64(len(bytes))
	err := WriteVarInt(w, pver, slen)
	if err != nil {
		return err
	}

	_, err = w.Write(bytes)
	return err
}

// mustSerialize returns the serialization of the transaction for the provided
// serialization type without modifying the original transaction.  It will panic
// if any errors occur.
func (msg *MsgTx) mustSerialize(serType TxSerializeType) []byte {
	serialized, err := msg.serialize(serType)
	if err != nil {
		panic(fmt.Sprintf("MsgTx failed serializing for type %v",
			serType))
	}
	return serialized
}

// serialize returns the serialization of the transaction for the provided
// serialization type without modifying the original transaction.
func (msg *MsgTx) serialize(serType TxSerializeType) ([]byte, error) {
	// Shallow copy so the serialization type can be changed without
	// modifying the original transaction.
	mtxCopy := *msg
	mtxCopy.SerType = serType
	buf := bytes.NewBuffer(make([]byte, 0, mtxCopy.SerializeSize()))
	err := mtxCopy.Serialize(buf)
	if err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

// SerializeSize returns the number of bytes it would take to serialize the
// the transaction.
func (msg *MsgTx) SerializeSize() int {
	// Unknown type return 0.
	n := 0
	switch msg.SerType {
	case TxSerializeNoWitness:
		// Version 4 bytes + LockTime 4 bytes + Expiry 4 bytes +
		// Serialized varint size for the number of transaction
		// inputs and outputs.
		n = 12 + VarIntSerializeSize(uint64(len(msg.TxIn))) +
			VarIntSerializeSize(uint64(len(msg.TxOut)))

		for _, txIn := range msg.TxIn {
			n += txIn.SerializeSizePrefix()
		}
		for _, txOut := range msg.TxOut {
			n += txOut.SerializeSize()
		}

	case TxSerializeOnlyWitness:
		// Version 4 bytes + Serialized varint size for the
		// number of transaction signatures.
		n = 4 + VarIntSerializeSize(uint64(len(msg.TxIn)))

		for _, txIn := range msg.TxIn {
			n += txIn.SerializeSizeWitness()
		}

	case TxSerializeWitnessSigning:
		// Version 4 bytes + Serialized varint size for the
		// number of transaction signatures.
		n = 4 + VarIntSerializeSize(uint64(len(msg.TxIn)))

		for _, txIn := range msg.TxIn {
			n += txIn.SerializeSizeWitnessSigning()
		}

	case TxSerializeWitnessValueSigning:
		// Version 4 bytes + Serialized varint size for the
		// number of transaction signatures.
		n = 4 + VarIntSerializeSize(uint64(len(msg.TxIn)))

		for _, txIn := range msg.TxIn {
			n += txIn.SerializeSizeWitnessValueSigning()
		}

	case TxSerializeFull:
		// Version 4 bytes + LockTime 4 bytes + Expiry 4 bytes + Serialized
		// varint size for the number of transaction inputs (x2) and
		// outputs. The number of inputs is added twice because it's
		// encoded once in both the witness and the prefix.
		n = 12 + VarIntSerializeSize(uint64(len(msg.TxIn))) +
			VarIntSerializeSize(uint64(len(msg.TxIn))) +
			VarIntSerializeSize(uint64(len(msg.TxOut)))

		for _, txIn := range msg.TxIn {
			n += txIn.SerializeSizePrefix()
		}
		for _, txIn := range msg.TxIn {
			n += txIn.SerializeSizeWitness()
		}
		for _, txOut := range msg.TxOut {
			n += txOut.SerializeSize()
		}
	}

	return n
}

// SerializeSizePrefix returns the number of bytes it would take to serialize
// the transaction input for a prefix.
func (t *TxIn) SerializeSizePrefix() int {
	// Outpoint Hash 32 bytes + Outpoint Index 4 bytes + Outpoint Tree 1 byte +
	// Sequence 4 bytes.
	return 41
}

// SerializeSizeWitness returns the number of bytes it would take to serialize the
// transaction input for a witness.
func (t *TxIn) SerializeSizeWitness() int {
	// ValueIn (8 bytes) + BlockHeight (4 bytes) + BlockIndex (4 bytes) +
	// serialized varint size for the length of SignatureScript +
	// SignatureScript bytes.
	return 8 + 4 + 4 + VarIntSerializeSize(uint64(len(t.SignatureScript))) +
		len(t.SignatureScript)
}

// SerializeSizeWitnessSigning returns the number of bytes it would take to
// serialize the transaction input for a witness used in signing.
func (t *TxIn) SerializeSizeWitnessSigning() int {
	// Serialized varint size for the length of SignatureScript +
	// SignatureScript bytes.
	return VarIntSerializeSize(uint64(len(t.SignatureScript))) +
		len(t.SignatureScript)
}

// SerializeSizeWitnessValueSigning returns the number of bytes it would take to
// serialize the transaction input for a witness used in signing with value
// included.
func (t *TxIn) SerializeSizeWitnessValueSigning() int {
	// ValueIn (8 bytes) + serialized varint size for the length of
	// SignatureScript + SignatureScript bytes.
	return 8 + VarIntSerializeSize(uint64(len(t.SignatureScript))) +
		len(t.SignatureScript)
}

// SerializeSize returns the number of bytes it would take to serialize the
// the transaction output.
func (t *TxOut) SerializeSize() int {
	// Value 8 bytes + Version 2 bytes + serialized varint size for
	// the length of PkScript + PkScript bytes.
	return 8 + 2 + VarIntSerializeSize(uint64(len(t.PkScript))) + len(t.PkScript)
}

// WriteVarInt serializes val to w using a variable number of bytes depending
// on its value.
func WriteVarInt(w io.Writer, pver uint32, val uint64) error {
	if val < 0xfd {
		return binarySerializer.PutUint8(w, uint8(val))
	}

	if val <= math.MaxUint16 {
		err := binarySerializer.PutUint8(w, 0xfd)
		if err != nil {
			return err
		}
		return binarySerializer.PutUint16(w, binary.LittleEndian, uint16(val))
	}

	if val <= math.MaxUint32 {
		err := binarySerializer.PutUint8(w, 0xfe)
		if err != nil {
			return err
		}
		return binarySerializer.PutUint32(w, binary.LittleEndian, uint32(val))
	}

	err := binarySerializer.PutUint8(w, 0xff)
	if err != nil {
		return err
	}
	return binarySerializer.PutUint64(w, binary.LittleEndian, val)
}

// VarIntSerializeSize returns the number of bytes it would take to serialize
// val as a variable length integer.
func VarIntSerializeSize(val uint64) int {
	// The value is small enough to be represented by itself, so it's
	// just 1 byte.
	if val < 0xfd {
		return 1
	}

	// Discriminant 1 byte plus 2 bytes for the uint16.
	if val <= math.MaxUint16 {
		return 3
	}

	// Discriminant 1 byte plus 4 bytes for the uint32.
	if val <= math.MaxUint32 {
		return 5
	}

	// Discriminant 1 byte plus 8 bytes for the uint64.
	return 9
}


// writeBlockHeader writes a decred block header to w.  See Serialize for
// encoding block headers to be stored to disk, such as in a database, as
// opposed to encoding for the wire.
func writeBlockHeader(w io.Writer, pver uint32, bh *BlockHeader) error {
	sec := uint32(bh.Timestamp.Unix())
	return writeElements(w, bh.Version, &bh.PrevBlock, &bh.MerkleRoot,
		&bh.StakeRoot, bh.VoteBits, bh.FinalState, bh.Voters,
		bh.FreshStake, bh.Revocations, bh.PoolSize, bh.Bits, bh.SBits,
		bh.Height, bh.Size, sec, bh.Nonce, bh.ExtraData,
		bh.StakeVersion)
}

// writeElements writes multiple items to w.  It is equivalent to multiple
// calls to writeElement.
func writeElements(w io.Writer, elements ...interface{}) error {
	for _, element := range elements {
		err := writeElement(w, element)
		if err != nil {
			return err
		}
	}
	return nil
}

const CommandSize = 12

// writeElement writes the little endian representation of element to w.
func writeElement(w io.Writer, element interface{}) error {
	// Attempt to write the element based on the concrete type via fast
	// type assertions first.
	switch e := element.(type) {
	case int32:
		err := binarySerializer.PutUint32(w, binary.LittleEndian, uint32(e))
		if err != nil {
			return err
		}
		return nil

	case uint32:
		err := binarySerializer.PutUint32(w, binary.LittleEndian, e)
		if err != nil {
			return err
		}
		return nil

	case int64:
		err := binarySerializer.PutUint64(w, binary.LittleEndian, uint64(e))
		if err != nil {
			return err
		}
		return nil

	case uint64:
		err := binarySerializer.PutUint64(w, binary.LittleEndian, e)
		if err != nil {
			return err
		}
		return nil

	case bool:
		var err error
		if e {
			err = binarySerializer.PutUint8(w, 0x01)
		} else {
			err = binarySerializer.PutUint8(w, 0x00)
		}
		if err != nil {
			return err
		}
		return nil

	// Message header checksum.
	case [4]byte:
		_, err := w.Write(e[:])
		if err != nil {
			return err
		}
		return nil

	// Message header command.
	case [CommandSize]uint8:
		_, err := w.Write(e[:])
		if err != nil {
			return err
		}
		return nil

	// IP address.
	case [16]byte:
		_, err := w.Write(e[:])
		if err != nil {
			return err
		}
		return nil

	case *chainhash.Hash:
		_, err := w.Write(e[:])
		if err != nil {
			return err
		}
		return nil

	case ServiceFlag:
		err := binarySerializer.PutUint64(w, binary.LittleEndian, uint64(e))
		if err != nil {
			return err
		}
		return nil

	case InvType:
		err := binarySerializer.PutUint32(w, binary.LittleEndian, uint32(e))
		if err != nil {
			return err
		}
		return nil

	case CurrencyNet:
		err := binarySerializer.PutUint32(w, binary.LittleEndian, uint32(e))
		if err != nil {
			return err
		}
		return nil

	case BloomUpdateType:
		err := binarySerializer.PutUint8(w, uint8(e))
		if err != nil {
			return err
		}
		return nil

	case RejectCode:
		err := binarySerializer.PutUint8(w, uint8(e))
		if err != nil {
			return err
		}
		return nil
	}

	// Fall back to the slower binary.Write if a fast path was not available
	// above.
	return binary.Write(w, binary.LittleEndian, element)
}

// writeTxInPrefixs encodes ti to the decred protocol encoding for a transaction
// input (TxIn) prefix to w.
func writeTxInPrefix(w io.Writer, pver uint32, version uint16, ti *TxIn) error {
	err := WriteOutPoint(w, pver, version, &ti.PreviousOutPoint)
	if err != nil {
		return err
	}

	return binarySerializer.PutUint32(w, binary.LittleEndian, ti.Sequence)
}

// WriteOutPoint encodes op to the decred protocol encoding for an OutPoint
// to w.
func WriteOutPoint(w io.Writer, pver uint32, version uint16, op *OutPoint) error {
	_, err := w.Write(op.Hash[:])
	if err != nil {
		return err
	}

	err = binarySerializer.PutUint32(w, binary.LittleEndian, op.Index)
	if err != nil {
		return err
	}

	return binarySerializer.PutUint8(w, uint8(op.Tree))
}

// encodeWitnessSigning encodes a transaction witness into a writer for signing.
func (msg *MsgTx) encodeWitnessSigning(w io.Writer, pver uint32) error {
	count := uint64(len(msg.TxIn))
	err := WriteVarInt(w, pver, count)
	if err != nil {
		return err
	}

	for _, ti := range msg.TxIn {
		err = writeTxInWitnessSigning(w, pver, msg.Version, ti)
		if err != nil {
			return err
		}
	}

	return nil
}

// writeTxInWitnessSigning encodes ti to the decred protocol encoding for a
// transaction input (TxIn) witness to w for signing.
func writeTxInWitnessSigning(w io.Writer, pver uint32, version uint16, ti *TxIn) error {
	// Only write the signature script.
	return WriteVarBytes(w, pver, ti.SignatureScript)
}

// writeTxInWitnessValueSigning encodes ti to the decred protocol encoding for a
// transaction input (TxIn) witness to w for signing with value included.
func writeTxInWitnessValueSigning(w io.Writer, pver uint32, version uint16, ti *TxIn) error {
	// ValueIn.
	err := binarySerializer.PutUint64(w, binary.LittleEndian, uint64(ti.ValueIn))
	if err != nil {
		return err
	}

	// Signature script.
	return WriteVarBytes(w, pver, ti.SignatureScript)
}

// encodeWitnessValueSigning encodes a transaction witness into a writer for
// signing, with the value included.
func (msg *MsgTx) encodeWitnessValueSigning(w io.Writer, pver uint32) error {
	count := uint64(len(msg.TxIn))
	err := WriteVarInt(w, pver, count)
	if err != nil {
		return err
	}

	for _, ti := range msg.TxIn {
		err = writeTxInWitnessValueSigning(w, pver, msg.Version, ti)
		if err != nil {
			return err
		}
	}

	return nil
}

// MessageError describes an issue with a message.
// An example of some potential issues are messages from the wrong decred
// network, invalid commands, mismatched checksums, and exceeding max payloads.
//
// This provides a mechanism for the caller to type assert the error to
// differentiate between general io errors such as io.EOF and issues that
// resulted from malformed messages.
type MessageError struct {
	Func        string // Function name
	Description string // Human readable description of the issue
}

// Error satisfies the error interface and prints human-readable errors.
func (e *MessageError) Error() string {
	if e.Func != "" {
		return fmt.Sprintf("%v: %v", e.Func, e.Description)
	}
	return e.Description
}

// messageError creates an error for the given function and description.
func messageError(f string, desc string) *MessageError {
	return &MessageError{Func: f, Description: desc}
}


// writeTxOut encodes to into the decred protocol encoding for a transaction
// output (TxOut) to w.
func writeTxOut(w io.Writer, pver uint32, version uint16, to *TxOut) error {
	err := binarySerializer.PutUint64(w, binary.LittleEndian, uint64(to.Value))
	if err != nil {
		return err
	}

	err = binarySerializer.PutUint16(w, binary.LittleEndian, to.Version)
	if err != nil {
		return err
	}

	return WriteVarBytes(w, pver, to.PkScript)
}


