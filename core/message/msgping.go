// Copyright (c) 2013-2015 The btcsuite developers
// Copyright (c) 2015-2016 The Decred developers
// Use of this source code is governed by an ISC
// license that can be found in the LICENSE file.

package message

import (
	"io"
	s "github.com/HalalChain/qitmeer/core/serialization"
)

// MsgPing implements the Message interface and represents a ping message.
//
// The payload for this message just consists of a nonce used for identifying
// it later.
type MsgPing struct {
	// Unique value associated with message that is used to identify
	// specific ping message.
	Nonce uint64
}

// Decode decodes r using the protocol encoding into the receiver.
// This is part of the Message interface implementation.
func (msg *MsgPing) Decode(r io.Reader, pver uint32) error {
	err := s.ReadElements(r, &msg.Nonce)
	return err
}

// Encode encodes the receiver to w using the protocol encoding.
// This is part of the Message interface implementation.
func (msg *MsgPing) Encode(w io.Writer, pver uint32) error {
	err := s.WriteElements(w, msg.Nonce)
	return err
}

// Command returns the protocol command string for the message.  This is part
// of the Message interface implementation.
func (msg *MsgPing) Command() string {
	return CmdPing
}

// MaxPayloadLength returns the maximum length the payload can be for the
// receiver.  This is part of the Message interface implementation.
func (msg *MsgPing) MaxPayloadLength(pver uint32) uint32 {
	plen := uint32(0)

	// Nonce 8 bytes.
	plen += 8

	return plen
}

// NewMsgPing returns a new ping message that conforms to the Message
// interface.  See MsgPing for details.
func NewMsgPing(nonce uint64) *MsgPing {
	return &MsgPing{
		Nonce: nonce,
	}
}
