// Copyright (c) 2017-2018 The nox developers
// Copyright (c) 2013-2015 The btcsuite developers
// Copyright (c) 2015-2016 The Decred developers
// Use of this source code is governed by an ISC
// license that can be found in the LICENSE file.

package message

import (
	"io"
)

// MsgGetAddr implements the Message interface and represents a GetAddr message.
// It is used to request a list of known active peers on the network from a peer
// to help identify potential nodes.  The list is returned  via one or more addr
// messages (MsgAddr).
//
// This message has no payload.
type MsgGetAddr struct{}

// Decode decodes r into the receiver.
// This is part of the Message interface implementation.
func (msg *MsgGetAddr) Decode(r io.Reader, pver uint32) error {
	return nil
}

// Encode encodes the receiver to w.
// This is part of the Message interface implementation.
func (msg *MsgGetAddr) Encode(w io.Writer, pver uint32) error {
	return nil
}

// Command returns the protocol command string for the message.  This is part
// of the Message interface implementation.
func (msg *MsgGetAddr) Command() string {
	return CmdGetAddr
}

// MaxPayloadLength returns the maximum length the payload can be for the
// receiver.  This is part of the Message interface implementation.
func (msg *MsgGetAddr) MaxPayloadLength(pver uint32) uint32 {
	return 0
}

// NewMsgGetAddr returns a new GetAddr message that conforms to the
// Message interface.  See MsgGetAddr for details.
func NewMsgGetAddr() *MsgGetAddr {
	return &MsgGetAddr{}
}
