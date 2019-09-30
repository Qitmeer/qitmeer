// Copyright (c) 2017-2018 The qitmeer developers
package message

import (
	"github.com/Qitmeer/qitmeer/core/types"
	"io"
)

// MsgTx implements the Message interface and represents a transaction message.
// It is used to deliver transaction information in response to a getdata
// message (MsgGetData) for a given transaction.
//
// Use the AddTxIn and AddTxOut functions to build up the list of transaction
// inputs and outputs.

type MsgTx struct {
 	Tx *types.Transaction
}

// Command returns the protocol command string for the message.  This is part
// of the Message interface implementation.
func (msg *MsgTx) Command() string {
	return CmdTx
}

// MaxPayloadLength returns the maximum length the payload can be for the
// receiver.  This is part of the Message interface implementation.
func (msg *MsgTx) MaxPayloadLength(pver uint32) uint32 {
	return types.MaxBlockPayload
}

// Decode decodes r into the receiver.
// This is part of the Message interface implementation.
//
// See Deserialize for decoding transactions stored to disk, such as in a
// database, as opposed to decoding transactions from the wire.
func (msg *MsgTx) Decode(r io.Reader, pver uint32) error {
	msg.Tx =  &types.Transaction{}
	return msg.Tx.Deserialize(r)
}

// Encode encodes the receiver to w.
// This is part of the Message interface implementation.
//
// See Serialize for encoding transactions to be stored to disk, such as in a
// database, as opposed to encoding transactions for the wire.
func (msg *MsgTx) Encode(w io.Writer, pver uint32) error {
	// The serialized encoding of the version includes the real transaction
	// version in the lower 16 bits and the transaction serialization type
	// in the upper 16 bits.
	return msg.Tx.Encode(w,pver,types.TxSerializeFull)
}
