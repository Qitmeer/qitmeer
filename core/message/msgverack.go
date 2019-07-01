// Copyright (c) 2017-2018 The nox developers
// Copyright (c) 2013-2015 The btcsuite developers
// Copyright (c) 2015-2016 The Decred developers
// Use of this source code is governed by an ISC
// license that can be found in the LICENSE file.

package message

import (
	"io"
)

// MsgVerAck defines a Verack message which is used for a peer to
// acknowledge a version message (MsgVersion) after it has used the
// information to negotiate parameters.
//
// It implements the Message interface.
// This message has no payload.
type MsgVerAck struct{}

// Decode decodes r into the receiver.
// This is part of the Message interface implementation.
func (msg *MsgVerAck) Decode(r io.Reader, pver uint32) error {
	return nil
}

// Encode encodes the receiver to w.
// This is part of the Message interface implementation.
func (msg *MsgVerAck) Encode(w io.Writer, pver uint32) error {
	return nil
}

// Command returns the protocol command string for the message.  This is part
// of the Message interface implementation.
func (msg *MsgVerAck) Command() string {
	return CmdVerAck
}

// MaxPayloadLength returns the maximum length the payload can be for the
// receiver.  This is part of the Message interface implementation.
func (msg *MsgVerAck) MaxPayloadLength(pver uint32) uint32 {
	return 0
}

// NewMsgVerAck returns a new VerAck message that conforms to the
// Message interface.
func NewMsgVerAck() *MsgVerAck {
	return &MsgVerAck{}
}
