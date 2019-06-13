// Copyright (c) 2017-2018 The nox developers
// Copyright (c) 2013-2016 The btcsuite developers
// Copyright (c) 2015-2017 The Decred developers
// Use of this source code is governed by an ISC
// license that can be found in the LICENSE file.

package message

import (
	"bytes"
	"github.com/HalalChain/qitmeer/core/types"
	"io"
)

// defaultTransactionAlloc is the default size used for the backing array
// for transactions.  The transaction array will dynamically grow as needed, but
// this figure is intended to provide enough space for the number of
// transactions in the vast majority of blocks without needing to grow the
// backing array multiple times.
const defaultTransactionAlloc = 2048

// MaxBlocksPerMsg is the maximum number of blocks allowed per message.
const MaxBlocksPerMsg = 500

// MsgBlock implements the Message interface and represents a block message.
// It is used to deliver block and transaction information in response
// to a getdata message (MsgGetData) for a given block hash.
type MsgBlock struct {
	*types.Block
}

// MaxPayloadLength returns the maximum length the payload can be for the
// receiver.  This is part of the Message interface implementation.
func (msg *MsgBlock) MaxPayloadLength(pver uint32) uint32 {
	// Block header at 80 bytes + transaction count + max transactions
	// which can vary up to the MaxBlockPayload (including the block header
	// and transaction count).
	return types.MaxBlockPayload
}

// AddTransaction adds a transaction to the message.
func (msg *MsgBlock) AddTransaction(tx *MsgTx) error {
	msg.Block.Transactions = append(msg.Block.Transactions, tx.Tx)
	return nil

}

// ClearTransactions removes all transactions from the message.
func (msg *MsgBlock) ClearTransactions() {
	msg.Block.Transactions = make([]*types.Transaction, 0, defaultTransactionAlloc)
}


// Decode decodes r using the protocol encoding into the receiver.
// This is part of the Message interface implementation.
// See Deserialize for decoding blocks stored to disk, such as in a database, as
// opposed to decoding blocks from the wire.
func (msg *MsgBlock) Decode(r io.Reader, pver uint32) error {
	msg.Block = &types.Block{}
	return msg.Block.Decode(r,pver)
}

// FromBytes deserializes a transaction byte slice.
func (msg *MsgBlock) FromBytes(b []byte) error {
	r := bytes.NewReader(b)
	return msg.Deserialize(r)
}


// Encode encodes the receiver to w using the protocol encoding.
// This is part of the Message interface implementation.
// See Serialize for encoding blocks to be stored to disk, such as in a
// database, as opposed to encoding blocks for the wire.
func (msg *MsgBlock) Encode(w io.Writer, pver uint32) error {
	return msg.Block.Encode(w, pver)
}

// Bytes returns the serialized form of the block in bytes.
func (msg *MsgBlock) Bytes() ([]byte, error) {
	buf := bytes.NewBuffer(make([]byte, 0, msg.SerializeSize()))
	err := msg.Serialize(buf)
	if err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

// SerializeSize returns the number of bytes it would take to serialize the
// the block.
func (msg *MsgBlock) SerializeSize() int {
	// Check to make sure that all transactions have the correct
	// type and version to be included in a block.

	// Block header bytes + Serialized varint size for the number of
	// transactions + Serialized varint size for the number of
	// stake transactions
	return msg.Block.SerializeSize()

}

// Command returns the protocol command string for the message.  This is part
// of the Message interface implementation.
func (msg *MsgBlock) Command() string {
	return CmdBlock
}

// NewMsgBlock returns a new block message that conforms to the
// Message interface.  See MsgBlock for details.
func NewMsgBlock(blockHeader *types.BlockHeader) *MsgBlock {
	block := types.Block{
		Header:        *blockHeader,
		Transactions:  make([]*types.Transaction, 0, defaultTransactionAlloc),
	}
	return &MsgBlock{&block}
}
