// Copyright 2017-2018 The nox developers

package types

import (
	"io"
	"github.com/noxproject/nox/common/hash"
	s "github.com/noxproject/nox/core/serialization"
	"fmt"
	"bytes"
	"encoding/binary"
	"time"
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

const (

	// TxVersion is the current latest supported transaction version.
	TxVersion uint32 = 1

	// defaultTxInOutAlloc is the default size used for the backing array
	// for transaction inputs and outputs.  The array will dynamically grow
	// as needed, but this figure is intended to provide enough space for
	// the number of inputs and outputs in a typical transaction without
	// needing to grow the backing array multiple times.
	defaultTxInOutAlloc = 15

	// NullValueIn is a null value for an input witness.
	NullValueIn uint64 = 0

	// NullBlockHeight is the null value for an input witness. It references
	// the genesis block.
	NullBlockHeight uint32 = 0x00000000

	// NullBlockIndex is the null transaction index in a block for an input
	// witness.
	NullBlockIndex uint32 = 0xffffffff

	// MaxTxInSequenceNum is the maximum sequence number the sequence field
	// of a transaction input can be.
	MaxTxInSequenceNum uint32 = 0xffffffff

	// SequenceLockTimeDisabled is a flag that if set on a transaction
	// input's sequence number, the sequence number will not be interpreted
	// as a relative locktime.
	SequenceLockTimeDisabled = 1 << 31

	// SequenceLockTimeIsSeconds is a flag that if set on a transaction
	// input's sequence number, the relative locktime has units of 512
	// seconds.
	SequenceLockTimeIsSeconds = 1 << 22

	// SequenceLockTimeMask is a mask that extracts the relative locktime
	// when masked against the transaction input sequence number.
	SequenceLockTimeMask = 0x0000ffff

	// minTxPayload is the minimum payload size for a transaction.  Note
	// that any realistically usable transaction must have at least one
	// input or output, but that is a rule enforced at a higher layer, so
	// it is intentionally not included here.
	// Version 4 bytes + Varint number of transaction inputs 1 byte + Varint
	// number of transaction outputs 1 byte + LockTime 4 bytes + min input
	// payload + min output payload.
	minTxPayload = 10

	// minTxInPayload is the minimum payload size for a transaction input.
	// PreviousOutPoint.Hash + PreviousOutPoint.Index 4 bytes +
	// PreviousOutPoint.Tree 1 byte + Varint for SignatureScript length 1
	// byte + Sequence 4 bytes.
	minTxInPayload = 11 + hash.HashSize

	// MaxMessagePayload is the maximum bytes a message can be regardless of other
	// individual limits imposed by messages themselves.
	MaxMessagePayload = (1024 * 1024 * 32) // 32MB

	// maxTxInPerMessage is the maximum number of transactions inputs that
	// a transaction which fits into a message could possibly have.
	maxTxInPerMessage = (MaxMessagePayload / minTxInPayload) + 1

	// minTxOutPayload is the minimum payload size for a transaction output.
	// Value 8 bytes + Varint for PkScript length 1 byte.
	minTxOutPayload = 9

	// maxTxOutPerMessage is the maximum number of transactions outputs that
	// a transaction which fits into a message could possibly have.
	maxTxOutPerMessage = (MaxMessagePayload / minTxOutPayload) + 1


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

type Input interface{
	GetSignScript() []byte
}

type Output interface{
	GetPkScript() []byte
}
type ScriptTx interface {
	GetInput()     []Input
	GetOutput()    []Output
	GetVersion()   uint32
	GetLockTime()  uint32
	GetType()      ScriptTxType
}
type ScriptTxType int

const (
	NoxScriptTx ScriptTxType = iota
	BtcScriptTx
)

type Transaction struct {
	CachedHash *hash.Hash
	Version   uint32
	LockTime  uint32
	Expire    uint32
	Type      TxType
	TxIn 	  []*TxInput
	TxOut 	  []*TxOutput
	Message   []byte     //a unencrypted/encrypted message if user pay additional fee & limit the max length
}

// NewMsgTx returns a new Decred tx message that conforms to the Message
// interface.  The return instance has a default version of TxVersion and there
// are no transaction inputs or outputs.  Also, the lock time is set to zero
// to indicate the transaction is valid immediately as opposed to some time in
// future.
func NewTransaction() *Transaction {
	return &Transaction{
		Version: TxVersion,
		TxIn:    make([]*TxInput, 0, defaultTxInOutAlloc),
		TxOut:   make([]*TxOutput, 0, defaultTxInOutAlloc),
	}
}

 func (t *Transaction) GetInput() []Input {
	txIns := make([]Input,len(t.TxIn))
	for i, txIn := range t.TxIn {
		txIns[i] = txIn
	}
	return txIns
}
func (t *Transaction) GetVersion() uint32{
	return uint32(t.Version)
}
func (t *Transaction) GetLockTime() uint32{
	return t.LockTime
}
func (t *Transaction) GetType() ScriptTxType{
	return NoxScriptTx
}
func (t *Transaction) GetOutput() []Output{
	txOuts := make([]Output,len(t.TxOut))
	for i, txOut := range t.TxOut{
		txOuts[i] = txOut
	}
	return txOuts
}

// AddTxIn adds a transaction input to the message.
func (t *Transaction) AddTxIn(ti *TxInput) {
	t.TxIn = append(t.TxIn, ti)
}

// AddTxOut adds a transaction output to the message.
func (t *Transaction) AddTxOut(to *TxOutput) {
	t.TxOut = append(t.TxOut, to)
}

// DetermineTxType determines the type of stake transaction a transaction is; if
// none, it returns that it is an assumed regular tx.
func DetermineTxType(tx *Transaction) TxType {
	//TODO txType
	return Leger
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
	serialized, err := tx.Serialize(serType)
	if err != nil {
		panic(fmt.Sprintf("MsgTx failed serializing for type %v",
			serType))
	}
	return serialized
}

// serialize returns the serialization of the transaction for the provided
// serialization type without modifying the original transaction.
func (tx *Transaction) Serialize(serType TxSerializeType) ([]byte, error) {
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

// Deserialize decodes a transaction from r into the receiver using a format
// that is suitable for long-term storage such as a database while respecting
// the Version field in the transaction.
func (tx *Transaction) Deserialize(r io.Reader) error {

	// The serialized encoding of the version includes the real transaction
	// version in the lower 16 bits and the transaction serialization type
	// in the upper 16 bits.
	version, err := s.BinarySerializer.Uint32(r, binary.LittleEndian)
	if err != nil {
		return err
	}
	tx.Version = uint32(version & 0xffff)
	serType := TxSerializeType(version >> 16)

 	if serType != TxSerializeFull {
		return fmt.Errorf("Transaction.Deserialize : wrong transaction serializetion type [%d]",serType)
	}

	// returnScriptBuffers is a closure that returns any script buffers that
	// were borrowed from the pool when there are any deserialization
	// errors.  This is only valid to call before the final step which
	// replaces the scripts with the location in a contiguous buffer and
	// returns them.
	returnScriptBuffers := func() {
		for _, txIn := range tx.TxIn {
			if txIn == nil || txIn.SignScript == nil {
				continue
			}
			scriptPool.Return(txIn.SignScript)
		}
		for _, txOut := range tx.TxOut {
			if txOut == nil || txOut.PkScript == nil {
				continue
			}
			scriptPool.Return(txOut.PkScript)
		}
	}

	totalScriptSizeIns, err := tx.decodePrefix(r)
		if err != nil {
			returnScriptBuffers()
			return err
		}
	totalScriptSizeOuts, err := tx.decodeWitness(r)
		if err != nil {
			returnScriptBuffers()
			return err
		}
	writeTxScriptsToMsgTx(tx, totalScriptSizeIns+
			totalScriptSizeOuts, TxSerializeFull)

	return nil
}

// writeTxScriptsToMsgTx allocates the memory for variable length fields in a
// Tx TxIns, TxOuts, or both as a contiguous chunk of memory, then fills
// in these fields for the Tx by copying to a contiguous piece of memory
// and setting the pointer.
//
// NOTE: It is no longer valid to return any previously borrowed script
// buffers after this function has run because it is already done and the
// scripts in the transaction inputs and outputs no longer point to the
// buffers.
func writeTxScriptsToMsgTx(tx *Transaction, totalScriptSize uint64, serType TxSerializeType) {
	// Create a single allocation to house all of the scripts and set each
	// input signature scripts and output public key scripts to the
	// appropriate subslice of the overall contiguous buffer.  Then, return
	// each individual script buffer back to the pool so they can be reused
	// for future deserialization.  This is done because it significantly
	// reduces the number of allocations the garbage collector needs to track,
	// which in turn improves performance and drastically reduces the amount
	// of runtime overhead that would otherwise be needed to keep track of
	// millions of small allocations.
	//
	// using Closures to write the TxIn and TxOut scripts because, depending
	// on the serialization type desired, only input or output scripts may
	// be required.
	var offset uint64
	scripts := make([]byte, totalScriptSize)
	writeTxIns := func() {
		for i := 0; i < len(tx.TxIn); i++ {
			// Copy the signature script into the contiguous buffer at the
			// appropriate offset.
			signatureScript := tx.TxIn[i].SignScript
			copy(scripts[offset:], signatureScript)

			// Reset the signature script of the transaction input to the
			// slice of the contiguous buffer where the script lives.
			scriptSize := uint64(len(signatureScript))
			end := offset + scriptSize
			tx.TxIn[i].SignScript = scripts[offset:end:end]
			offset += scriptSize

			// Return the temporary script buffer to the pool.
			scriptPool.Return(signatureScript)
		}
	}
	writeTxOuts := func() {
		for i := 0; i < len(tx.TxOut); i++ {
			// Copy the public key script into the contiguous buffer at the
			// appropriate offset.
			pkScript := tx.TxOut[i].PkScript
			copy(scripts[offset:], pkScript)

			// Reset the public key script of the transaction output to the
			// slice of the contiguous buffer where the script lives.
			scriptSize := uint64(len(pkScript))
			end := offset + scriptSize
			tx.TxOut[i].PkScript = scripts[offset:end:end]
			offset += scriptSize

			// Return the temporary script buffer to the pool.
			scriptPool.Return(pkScript)
		}
	}

	// Handle the serialization types accordingly.
	writeTxIns()
	writeTxOuts()

}

func (tx *Transaction) decodeWitness(r io.Reader) (uint64, error) {
	// Witness only; generate the TxIn list and fill out only the
	// sigScripts.
	var totalScriptSize uint64

	// We're decoding witnesses from a full transaction, so read in
	// the number of signature scripts, check to make sure it's the
	// same as the number of TxIns we currently have, then fill in
	// the signature scripts.
	count, err := s.ReadVarInt(r, 0)
	if err != nil {
		return 0, err
	}

	// Don't allow the deserializer to panic by accessing memory
	// that doesn't exist.
	if int(count) != len(tx.TxIn) {
		return 0, fmt.Errorf("Tx.decodeWitness: non equal witness and prefix txin quantities "+
			"(witness %v, prefix %v)", count,
			len(tx.TxIn))
	}

	// Prevent more input transactions than could possibly fit into a
	// message.  It would be possible to cause memory exhaustion and panics
	// without a sane upper bound on this count.
	if count > uint64(maxTxInPerMessage) {
		return 0, fmt.Errorf("MsgTx.decodeWitness:too many input transactions to fit into "+
			"max message size [count %d, max %d]", count,
			maxTxInPerMessage)
	}

	// Read in the witnesses, and copy them into the already generated
	// by decodePrefix TxIns.
	txIns := make([]TxInput, count)
	for i := uint64(0); i < count; i++ {
		ti := &txIns[i]
		err = readTxInWitness(r, ti)
		if err != nil {
			return 0, err
		}
		totalScriptSize += uint64(len(ti.SignScript))

		tx.TxIn[i].AmountIn = ti.AmountIn
		tx.TxIn[i].BlockHeight = ti.BlockHeight
		tx.TxIn[i].BlockTxIndex = ti.BlockTxIndex
		tx.TxIn[i].SignScript = ti.SignScript
	}
	return totalScriptSize, nil
}

// readTxInWitness reads the next sequence of bytes from r as a transaction input
// (TxIn) in the transaction witness.
func readTxInWitness(r io.Reader, ti *TxInput) error {
	// ValueIn.
	valueIn, err := s.BinarySerializer.Uint64(r, binary.LittleEndian)
	if err != nil {
		return err
	}
	ti.AmountIn = uint64(valueIn)

	// BlockHeight.
	ti.BlockHeight, err = s.BinarySerializer.Uint32(r, binary.LittleEndian)
	if err != nil {
		return err
	}

	// BlockIndex.
	ti.BlockTxIndex, err = s.BinarySerializer.Uint32(r, binary.LittleEndian)
	if err != nil {
		return err
	}

	// Signature script.
	ti.SignScript, err = readScript(r)
	return err
}

// decodePrefix decodes a transaction prefix and stores the contents
// in the embedded msgTx.
func (tx *Transaction) decodePrefix(r io.Reader) (uint64, error) {

	count, err := s.ReadVarInt(r,0)

	// Prevent more input transactions than could possibly fit into a
	// message.  It would be possible to cause memory exhaustion and panics
	// without a sane upper bound on this count.
	if count > uint64(maxTxInPerMessage) {
		return 0, fmt.Errorf("Tx.decodePrefix: too many input " +
			"transactions to fit into max message size [count %d, max %d]",
			count,
			maxTxInPerMessage)
	}

	// TxIns.
	txIns := make([]TxInput, count)
	tx.TxIn = make([]*TxInput, count)
	for i := uint64(0); i < count; i++ {
		// The pointer is set now in case a script buffer is borrowed
		// and needs to be returned to the pool on error.
		ti := &txIns[i]
		tx.TxIn[i] = ti
		err = readTxInPrefix(r,tx.Version, ti)
		if err != nil {
			return 0, err
		}
	}

	count, err = s.ReadVarInt(r, 0)
	if err != nil {
		return 0, err
	}

	// Prevent more output transactions than could possibly fit into a
	// message.  It would be possible to cause memory exhaustion and panics
	// without a sane upper bound on this count.
	if count > uint64(maxTxOutPerMessage) {

		return 0, fmt.Errorf("Tx.decodePrefix too many output transactions" +
			" to fit into max message size [count %d, max %d]", count,
			maxTxOutPerMessage)
	}

	// TxOuts.
	var totalScriptSize uint64
	txOuts := make([]TxOutput, count)
	tx.TxOut = make([]*TxOutput, count)
	for i := uint64(0); i < count; i++ {
		// The pointer is set now in case a script buffer is borrowed
		// and needs to be returned to the pool on error.
		to := &txOuts[i]
		tx.TxOut[i] = to
		err = readTxOut(r, to)
		if err != nil {
			return 0, err
		}
		totalScriptSize += uint64(len(to.PkScript))
	}

	// Locktime and expiry.
	tx.LockTime, err = s.BinarySerializer.Uint32(r, binary.LittleEndian)
	if err != nil {
		return 0, err
	}

	tx.Expire, err = s.BinarySerializer.Uint32(r, binary.LittleEndian)
	if err != nil {
		return 0, err
	}

	return totalScriptSize, nil
}

// readTxInPrefix reads the next sequence of bytes from r as a transaction input
// (TxIn) in the transaction prefix.
func readTxInPrefix(r io.Reader, version uint32, ti *TxInput) error {

	// Outpoint.
	err := ReadOutPoint(r,version, &ti.PreviousOut)
	if err != nil {
		return err
	}

	// Sequence.
	ti.Sequence, err = s.BinarySerializer.Uint32(r, binary.LittleEndian)
	return err
}

// ReadOutPoint reads the next sequence of bytes from r as an OutPoint.
func ReadOutPoint(r io.Reader,version uint32, op *TxOutPoint) error {
	_, err := io.ReadFull(r, op.Hash[:])
	if err != nil {
		return err
	}
	op.OutIndex, err = s.BinarySerializer.Uint32(r, binary.LittleEndian)
	if err != nil {
		return err
	}
	return nil
}

// readTxOut reads the next sequence of bytes from r as a transaction output
// (TxOut).
func readTxOut(r io.Reader, to *TxOutput) error {
	value, err := s.BinarySerializer.Uint64(r, binary.LittleEndian)
	if err != nil {
		return err
	}
	to.Amount = uint64(value)

	to.PkScript, err = readScript(r)
	return err
}

// readScript reads a variable length byte array that represents a transaction
// script.  It is encoded as a varInt containing the length of the array
// followed by the bytes themselves.  An error is returned if the length is
// greater than the passed maxAllowed parameter which helps protect against
// memory exhaustion attacks and forced panics thorugh malformed messages.  The
// fieldName parameter is only used for the error message so it provides more
// context in the error.
func readScript(r io.Reader) ([]byte, error) {
	count, err := s.ReadVarInt(r, 0)
	if err != nil {
		return nil, err
	}
	// Prevent byte array larger than the max message size.  It would
	// be possible to cause memory exhaustion and panics without a sane
	// upper bound on this count.
	if count > uint64(MaxMessagePayload) {
		return nil,fmt.Errorf("readScript: larger than the max allowed size "+
			"[count %d, max %d]", count, MaxMessagePayload)
	}

	b := scriptPool.Borrow(count)
	_, err = io.ReadFull(r, b)
	if err != nil {
		scriptPool.Return(b)
		return nil, err
	}
	return b, nil
}

// CachedTxHash is equivalent to calling TxHash, however it caches the result so
// subsequent calls do not have to recalculate the hash.  It can be recalculated
// later with RecacheTxHash.
func (t *Transaction) CachedTxHash() *hash.Hash {
	if t.CachedHash == nil {
		h := t.TxHash()
		t.CachedHash = &h
	}

	return t.CachedHash
}

// RecacheTxHash is equivalent to calling TxHash, however it replaces the cached
// result so future calls to CachedTxHash will return this newly calculated
// hash.
func (t *Transaction) RecacheTxHash() *hash.Hash {
	h := t.TxHash()
	t.CachedHash = &h
	return t.CachedHash
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
	// serialize version using Full
	serializedVersion := uint32(tx.Version) | uint32(TxSerializeFull)<<16
	err := s.BinarySerializer.PutUint32(w, binary.LittleEndian, serializedVersion)
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
	hash    hash.Hash      // Cached transaction hash
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

// Hash returns the hash of the transaction.  This is equivalent to
// calling TxHash on the underlying wire.MsgTx, however it caches the
// result so subsequent calls are more efficient.
func (t *Tx) Hash() *hash.Hash {
	return &t.hash
}

// SetIndex sets the index of the transaction in within a block.
func (t *Tx) SetIndex(index int) {
	t.txIndex = index
}

type TxOutPoint struct {
	Hash       hash.Hash       //txid
	OutIndex   uint32          //vout
}

// NewOutPoint returns a new transaction outpoint point with the
// provided hash and index.
func NewOutPoint(hash *hash.Hash, index uint32) *TxOutPoint {
	return &TxOutPoint{
		Hash:  *hash,
		OutIndex: index,
	}
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

// NewTxIn returns a new Decred transaction input with the provided
// previous outpoint point and signature script with a default sequence of
// MaxTxInSequenceNum.
func NewTxInput(prevOut *TxOutPoint, amountIn uint64, signScript []byte) *TxInput {
	return &TxInput{
		PreviousOut:   *prevOut,
		Sequence:      MaxTxInSequenceNum,
		SignScript:    signScript,
		AmountIn:      amountIn,
		BlockHeight:   NullBlockHeight,
		BlockTxIndex:  NullBlockIndex,
	}
}

func (ti *TxInput) GetSignScript() []byte{
	return ti.SignScript
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

// NewTxOutput returns a new bitcoin transaction output with the provided
// transaction value and public key script.
func NewTxOutput(amount uint64, pkScript []byte) *TxOutput {
	return &TxOutput{
		Amount:    amount,
		PkScript: pkScript,
	}
}

func (to *TxOutput) GetPkScript() []byte {
	return to.PkScript
}

// SerializeSize returns the number of bytes it would take to serialize the
// the transaction output.
func (to *TxOutput) SerializeSize() int {
	// Value 8 bytes + Version 2 bytes + serialized varint size for
	// the length of PkScript + PkScript bytes.
	return 8 + 2 + s.VarIntSerializeSize(uint64(len(to.PkScript))) + len(to.PkScript)
}

type ContractTransaction struct {
	From         Account
	To           Account
	Value        uint64
	GasPrice     uint64
	GasLimit     uint64
	Nonce        uint64
	Input        []byte
	Signature    []byte
}


// TxDesc is a descriptor about a transaction in a transaction source along with
// additional metadata.
type TxDesc struct {
	// Tx is the transaction associated with the entry.
	Tx *Tx

	// Type is the type of the transaction associated with the entry.
	Type TxType

	// Added is the time when the entry was added to the source pool.
	Added time.Time

	// Height is the block height when the entry was added to the the source
	// pool.
	Height int64

	// Fee is the total fee the transaction associated with the entry pays.
	Fee int64
}

// TxLoc holds locator data for the offset and length of where a transaction is
// located within a MsgBlock data buffer.
type TxLoc struct {
	TxStart int
	TxLen   int
}


const (
	// freeListMaxScriptSize is the size of each buffer in the free list
	// that	is used for deserializing scripts from the wire before they are
	// concatenated into a single contiguous buffers.  This value was chosen
	// because it is slightly more than twice the size of the vast majority
	// of all "standard" scripts.  Larger scripts are still deserialized
	// properly as the free list will simply be bypassed for them.
	freeListMaxScriptSize = 512

	// freeListMaxItems is the number of buffers to keep in the free list
	// to use for script deserialization.  This value allows up to 100
	// scripts per transaction being simultaneously deserialized by 125
	// peers.  Thus, the peak usage of the free list is 12,500 * 512 =
	// 6,400,000 bytes.
	freeListMaxItems = 12500
)
// scriptFreeList defines a free list of byte slices (up to the maximum number
// defined by the freeListMaxItems constant) that have a cap according to the
// freeListMaxScriptSize constant.  It is used to provide temporary buffers for
// deserializing scripts in order to greatly reduce the number of allocations
// required.
//
// The caller can obtain a buffer from the free list by calling the Borrow
// function and should return it via the Return function when done using it.
type scriptFreeList chan []byte

// Borrow returns a byte slice from the free list with a length according the
// provided size.  A new buffer is allocated if there are any items available.
//
// When the size is larger than the max size allowed for items on the free list
// a new buffer of the appropriate size is allocated and returned.  It is safe
// to attempt to return said buffer via the Return function as it will be
// ignored and allowed to go the garbage collector.
func (c scriptFreeList) Borrow(size uint64) []byte {
	if size > freeListMaxScriptSize {
		return make([]byte, size)
	}

	var buf []byte
	select {
	case buf = <-c:
	default:
		buf = make([]byte, freeListMaxScriptSize)
	}
	return buf[:size]
}

// Return puts the provided byte slice back on the free list when it has a cap
// of the expected length.  The buffer is expected to have been obtained via
// the Borrow function.  Any slices that are not of the appropriate size, such
// as those whose size is greater than the largest allowed free list item size
// are simply ignored so they can go to the garbage collector.
func (c scriptFreeList) Return(buf []byte) {
	// Ignore any buffers returned that aren't the expected size for the
	// free list.
	if cap(buf) != freeListMaxScriptSize {
		return
	}

	// Return the buffer to the free list when it's not full.  Otherwise let
	// it be garbage collected.
	select {
	case c <- buf:
	default:
		// Let it go to the garbage collector.
	}
}

// Create the concurrent safe free list to use for script deserialization.  As
// previously described, this free list is maintained to significantly reduce
// the number of allocations.
var scriptPool scriptFreeList = make(chan []byte, freeListMaxItems)
