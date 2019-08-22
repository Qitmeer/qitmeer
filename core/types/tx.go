// Copyright 2017-2018 The qitmeer developers

package types

import (
	"github.com/Qitmeer/qitmeer-lib/common/math"
	"io"
	"github.com/Qitmeer/qitmeer-lib/common/hash"
	s "github.com/Qitmeer/qitmeer-lib/core/serialization"
	"fmt"
	"bytes"
	"encoding/binary"
	"time"
)

type TxType byte

const (
	CoinBase         TxType = 0x01
	Leger            TxType = 0x02
	TxTypeRegular    TxType = 0x03
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
	NullBlockOrder uint32 = 0x00000000

	// NullTxIndex is the null transaction index in a block for an input
	// witness.
	NullTxIndex uint32 = 0xffffffff

	// MaxTxInSequenceNum is the maximum sequence number the sequence field
	// of a transaction input can be.
	MaxTxInSequenceNum uint32 = 0xffffffff

	// MaxPrevOutIndex is the maximum index the index field of a previous
	// outpoint can be.
	MaxPrevOutIndex uint32 = 0xffffffff

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

	// TODO, revisit lock time define
	// SequenceLockTimeGranularity is the defined time based granularity
	// for seconds-based relative time locks.  When converting from seconds
	// to a sequence number, the value is right shifted by this amount,
	// therefore the granularity of relative time locks in 512 or 2^9
	// seconds.  Enforced relative lock times are multiples of 512 seconds.
	SequenceLockTimeGranularity = 9

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

	// NoExpiryValue is the value of expiry that indicates the transaction
	// has no expiry.
	NoExpiryValue uint32 = 0

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
	QitmeerScriptTx ScriptTxType = iota
	BtcScriptTx
)

type Transaction struct {
	Version   uint32
	TxIn 	  []*TxInput
	TxOut 	  []*TxOutput
	LockTime  uint32
	Expire    uint32

	Message   []byte     //a unencrypted/encrypted message if user pay additional fee & limit the max length
	CachedHash *hash.Hash
}

// NewMsgTx returns a new tx message that conforms to the Message interface.
// The return instance has a default version of TxVersion and there
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

// MaxTxPerTxTree returns the maximum number of transactions that could possibly
// fit into a block per ekach merkle root for the given protocol version.
func MaxTxPerTxTree(pver uint32) uint64 {
	return ((MaxBlockPayload / minTxPayload) / 2) + 1
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
	return QitmeerScriptTx
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
	return TxTypeRegular
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
	n = 8 + s.VarIntSerializeSize(uint64(len(tx.TxIn))) +
		s.VarIntSerializeSize(uint64(len(tx.TxOut)))

	for _, txIn := range tx.TxIn {
		n += txIn.SerializeSize()
	}
	for _, txOut := range tx.TxOut {
		n += txOut.SerializeSize()
	}
	return n
}

// mustSerialize returns the serialization of the transaction for the provided
// serialization type without modifying the original transaction.  It will panic
// if any errors occur.
func (tx *Transaction) mustSerialize() []byte {
	serialized, err := tx.Serialize()
	if err != nil {
		panic("tx failed serializing")
	}
	return serialized
}

// serialize returns the serialization of the transaction for the provided
// serialization type without modifying the original transaction.
func (tx *Transaction) Serialize() ([]byte, error) {
	buf := bytes.NewBuffer(make([]byte, 0, tx.SerializeSize()))
	err:=tx.Encode(buf,0)
	return buf.Bytes(),err
}


func (tx *Transaction) Encode(w io.Writer, pver uint32) error {
	// serialize version using Full
	serializedVersion := uint32(tx.Version) | uint32(TxSerializeFull)<<16
	err := s.BinarySerializer.PutUint32(w, binary.LittleEndian, serializedVersion)
	if err != nil {
		return err
	}

	count := uint64(len(tx.TxIn))
	err = s.WriteVarInt(w, pver, count)
	if err != nil {
		return err
	}

	for _, ti := range tx.TxIn {
		err = writeTxIn(w, pver, tx.Version, ti)
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

// writeTxInPrefixs encodes for a transaction input (TxIn) prefix to w.
func writeTxIn(w io.Writer, pver uint32, version uint32, ti *TxInput) error {
	err := WriteOutPoint(w, pver, version, &ti.PreviousOut)
	if err != nil {
		return err
	}
	err =s.WriteVarBytes(w, pver, ti.SignScript)
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

// Deserialize decodes a transaction from r into the receiver using a format
// that is suitable for long-term storage such as a database while respecting
// the Version field in the transaction.
func (tx *Transaction) Deserialize(r io.Reader) error {
	return tx.Decode(r,0)
}

func (tx *Transaction) Decode(r io.Reader, pver uint32) error {
	// The serialized encoding of the version includes the real transaction
	// version in the lower 16 bits and the transaction serialization type
	// in the upper 16 bits.
	version, err := s.BinarySerializer.Uint32(r, binary.LittleEndian)
	if err != nil {
		return err
	}
	tx.Version = uint32(version & 0xffff)
	//serType := TxSerializeType(version >> 16)

	count, err := s.ReadVarInt(r,0)
	if err != nil {
		return err
	}

	// Prevent more input transactions than could possibly fit into a
	// message.  It would be possible to cause memory exhaustion and panics
	// without a sane upper bound on this count.
	if count > uint64(maxTxInPerMessage) {
		return fmt.Errorf("Tx.decodePrefix: too many input " +
			"transactions to fit into max message size [count %d, max %d]",
			count,
			maxTxInPerMessage)
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

	// Deserialize the inputs.
	var totalScriptSize uint64
	txIns := make([]TxInput, count)
	tx.TxIn = make([]*TxInput, count)
	for i := uint64(0); i < count; i++ {
		// The pointer is set now in case a script buffer is borrowed
		// and needs to be returned to the pool on error.
		ti := &txIns[i]
		tx.TxIn[i] = ti
		err = readTxIn(r, tx.Version, ti)
		if err != nil {
			returnScriptBuffers()
			return err
		}
		totalScriptSize += uint64(len(ti.SignScript))
	}

	count, err = s.ReadVarInt(r, 0)
	if err != nil {
		returnScriptBuffers()
		return err
	}

	// Prevent more output transactions than could possibly fit into a
	// message.  It would be possible to cause memory exhaustion and panics
	// without a sane upper bound on this count.
	if count > uint64(maxTxOutPerMessage) {

		return fmt.Errorf("Tx.decodePrefix too many output transactions" +
			" to fit into max message size [count %d, max %d]", count,
			maxTxOutPerMessage)
	}


	// Deserialize the outputs.
	txOuts := make([]TxOutput, count)
	tx.TxOut = make([]*TxOutput, count)
	for i := uint64(0); i < count; i++ {
		// The pointer is set now in case a script buffer is borrowed
		// and needs to be returned to the pool on error.
		to := &txOuts[i]
		tx.TxOut[i] = to
		err = readTxOut(r, to)
		if err != nil {
			returnScriptBuffers()
			return err
		}
		totalScriptSize += uint64(len(to.PkScript))
	}

	// Locktime and expiry.
	tx.LockTime, err = s.BinarySerializer.Uint32(r, binary.LittleEndian)
	if err != nil {
		returnScriptBuffers()
		return err
	}

	tx.Expire, err = s.BinarySerializer.Uint32(r, binary.LittleEndian)
	if err != nil {
		returnScriptBuffers()
		return err
	}

	// Create a single allocation to house all of the scripts and set each
	// input signature script and output public key script to the
	// appropriate subslice of the overall contiguous buffer.  Then, return
	// each individual script buffer back to the pool so they can be reused
	// for future deserializations.  This is done because it significantly
	// reduces the number of allocations the garbage collector needs to
	// track, which in turn improves performance and drastically reduces the
	// amount of runtime overhead that would otherwise be needed to keep
	// track of millions of small allocations.
	//
	// NOTE: It is no longer valid to call the returnScriptBuffers closure
	// after these blocks of code run because it is already done and the
	// scripts in the transaction inputs and outputs no longer point to the
	// buffers.
	var offset uint64
	scripts := make([]byte, totalScriptSize)
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
	return nil
}

func readTxIn(r io.Reader, version uint32, ti *TxInput) error {
	// Outpoint.
	err := ReadOutPoint(r,version, &ti.PreviousOut)
	if err != nil {
		return err
	}
	ti.SignScript, err = readScript(r)
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
	return hash.DoubleHashH(tx.mustSerialize())
}

// IsCoinBaseTx determines whether or not a transaction is a coinbase.  A
// coinbase is a special transaction created by miners that has no inputs.
// This is represented in the block chain by a transaction with a single input
// that has a previous output transaction index set to the maximum value along
// with a zero hash.
//
// This function only differs from IsCoinBase in that it works with a raw wire
// transaction as opposed to a higher level util transaction.
func (tx *Transaction) IsCoinBase() bool {
	// A coin base must only have one transaction input.
	if len(tx.TxIn) != 1 {
		return false
	}
	// The previous output of a coin base must have a max value index and a
	// zero hash.
	prevOut := &tx.TxIn[0].PreviousOut
	if prevOut.OutIndex != math.MaxUint32 || !prevOut.Hash.IsEqual(&hash.ZeroHash) {
		return false
	}
	return true
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

// NewTxDeep returns a new instance of a transaction given an underlying
// wire.MsgTx.  Until NewTx, it completely copies the data in the msgTx
// so that there are new memory allocations, in case you were to somewhere
// else modify the data assigned to these pointers.
func NewTxDeep(msgTx *Transaction) *Tx {
	txIns := make([]*TxInput, len(msgTx.TxIn))
	txOuts := make([]*TxOutput, len(msgTx.TxOut))

	for i, txin := range msgTx.TxIn {
		sigScript := make([]byte, len(txin.SignScript))
		copy(sigScript[:], txin.SignScript[:])

		txIns[i] = &TxInput{
			PreviousOut: TxOutPoint{
				Hash:     txin.PreviousOut.Hash,
				OutIndex: txin.PreviousOut.OutIndex,
			},
			Sequence:        txin.Sequence,
			SignScript:      sigScript,
		}
	}

	for i, txout := range msgTx.TxOut {
		pkScript := make([]byte, len(txout.PkScript))
		copy(pkScript[:], txout.PkScript[:])

		txOuts[i] = &TxOutput{
			Amount:   txout.Amount,
			PkScript: pkScript,
		}
	}

	mtx := &Transaction{
		CachedHash: nil,
		Version:    msgTx.Version,
		TxIn:       txIns,
		TxOut:      txOuts,
		LockTime:   msgTx.LockTime,
		Expire:     msgTx.Expire,
	}

	return &Tx{
		hash:    mtx.TxHash(),
		Tx:      mtx,
		txIndex: TxIndexUnknown,
	}
}


// NewTxDeepTxIns is used to deep copy a transaction, maintaining the old
// pointers to the TxOuts while replacing the old pointers to the TxIns with
// deep copies. This is to prevent races when the fraud proofs for the
// transactions are set by the miner.
func NewTxDeepTxIns(msgTx *Transaction) *Tx {
	if msgTx == nil {
		return nil
	}

	newMsgTx := new(Transaction)

	// Copy the fixed fields.
	newMsgTx.Version = msgTx.Version
	newMsgTx.LockTime = msgTx.LockTime
	newMsgTx.Expire = msgTx.Expire

	// Copy the TxIns deeply.
	for _, txIn := range msgTx.TxIn {
		sigScrLen := len(txIn.SignScript)
		sigScrCopy := make([]byte, sigScrLen)

		txInCopy := new(TxInput)
		txInCopy.PreviousOut.Hash = txIn.PreviousOut.Hash
		txInCopy.PreviousOut.OutIndex = txIn.PreviousOut.OutIndex

		txInCopy.Sequence = txIn.Sequence

		txInCopy.SignScript = sigScrCopy

		newMsgTx.AddTxIn(txIn)
	}

	// Shallow copy the TxOuts.
	for _, txOut := range msgTx.TxOut {
		newMsgTx.AddTxOut(txOut)
	}

	return &Tx{
		hash:    msgTx.TxHash(),
		Tx:   msgTx,
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
	//the signature script (or witness script? or redeem script?)
	SignScript       []byte
	// This number has more historical significance than relevant usage;
	// however, its most relevant purpose is to enable “locking” of payments
	// so that they cannot be redeemed until a certain time.
	Sequence         uint32     //work with LockTime (disable use 0xffffffff, bitcoin historical)
}

// NewTxIn returns a new transaction input with the provided  previous outpoint
// point and signature script with a default sequence of MaxTxInSequenceNum.
func NewTxInput(prevOut *TxOutPoint,  signScript []byte) *TxInput {
	return &TxInput{
		PreviousOut:   *prevOut,
		Sequence:      MaxTxInSequenceNum,
		SignScript:    signScript,
	}
}

func (ti *TxInput) GetSignScript() []byte{
	return ti.SignScript
}

// SerializeSizeWitness returns the number of bytes it would take to serialize the
// transaction input for a witness.
func (ti *TxInput) SerializeSize() int {
	// Outpoint Hash 32 bytes + Outpoint Index 4 bytes + Sequence 4 bytes +
	// serialized varint size for the length of SignatureScript +
	// SignatureScript bytes.
	return 40 + s.VarIntSerializeSize(uint64(len(ti.SignScript))) +
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
	// Value 8 bytes + serialized varint size for
	// the length of PkScript + PkScript bytes.
	return 8 + s.VarIntSerializeSize(uint64(len(to.PkScript))) + len(to.PkScript)
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

	// Added is the time when the entry was added to the source pool.
	Added time.Time

	// Height is the block height when the entry was added to the the source
	// pool.
	Height int64

	// Fee is the total fee the transaction associated with the entry pays.
	Fee int64

	// FeePerKB is the fee the transaction pays in meer per 1000 bytes.
	FeePerKB int64
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
