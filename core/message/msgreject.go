// Copyright (c) 2017-2018 The nox developers
// Copyright (c) 2014-2016 The btcsuite developers
// Copyright (c) 2015-2016 The Decred developers
// Use of this source code is governed by an ISC
// license that can be found in the LICENSE file.

package message

import (
	"fmt"
	"io"

	s "github.com/HalalChain/qitmeer-lib/core/serialization"
	"github.com/HalalChain/qitmeer-lib/common/hash"
)

// RejectCode represents a numeric value by which a remote peer indicates
// why a message was rejected.
type RejectCode uint8

// These constants define the various supported reject codes.
const (
	RejectMalformed       RejectCode = 0x01
	RejectInvalid         RejectCode = 0x10
	RejectObsolete        RejectCode = 0x11
	RejectDuplicate       RejectCode = 0x12
	RejectNonstandard     RejectCode = 0x40
	RejectDust            RejectCode = 0x41
	RejectInsufficientFee RejectCode = 0x42
	RejectCheckpoint      RejectCode = 0x43
)

// Map of reject codes back strings for pretty printing.
var rejectCodeStrings = map[RejectCode]string{
	RejectMalformed:       "REJECT_MALFORMED",
	RejectInvalid:         "REJECT_INVALID",
	RejectObsolete:        "REJECT_OBSOLETE",
	RejectDuplicate:       "REJECT_DUPLICATE",
	RejectNonstandard:     "REJECT_NONSTANDARD",
	RejectDust:            "REJECT_DUST",
	RejectInsufficientFee: "REJECT_INSUFFICIENTFEE",
	RejectCheckpoint:      "REJECT_CHECKPOINT",
}

// String returns the RejectCode in human-readable form.
func (code RejectCode) String() string {
	if s, ok := rejectCodeStrings[code]; ok {
		return s
	}

	return fmt.Sprintf("Unknown RejectCode (%d)", uint8(code))
}

// MsgReject implements the Message interface and represents an reject message.
//
// This message was not added until protocol version RejectVersion.
type MsgReject struct {
	// Cmd is the command for the message which was rejected such as
	// as CmdBlock or CmdTx.  This can be obtained from the Command function
	// of a Message.
	Cmd string

	// RejectCode is a code indicating why the command was rejected.  It
	// is encoded as a uint8 on the wire.
	Code RejectCode

	// Reason is a human-readable string with specific details (over and
	// above the reject code) about why the command was rejected.
	Reason string

	// Hash identifies a specific block or transaction that was rejected
	// and therefore only applies the MsgBlock and MsgTx messages.
	Hash hash.Hash
}

// Decode decodes r encoding into the receiver.
// This is part of the Message interface implementation.
func (msg *MsgReject) Decode(r io.Reader, pver uint32) error {
	// Command that was rejected.
	cmd, err := s.ReadVarString(r, pver)
	if err != nil {
		return err
	}
	msg.Cmd = cmd

	// Code indicating why the command was rejected.
	err = s.ReadElements(r, &msg.Code)
	if err != nil {
		return err
	}

	// Human readable string with specific details (over and above the
	// reject code above) about why the command was rejected.
	reason, err := s.ReadVarString(r, pver)
	if err != nil {
		return err
	}
	msg.Reason = reason

	// CmdBlock and CmdTx messages have an additional hash field that
	// identifies the specific block or transaction.
	if msg.Cmd == CmdBlock || msg.Cmd == CmdTx {
		err := s.ReadElements(r, &msg.Hash)
		if err != nil {
			return err
		}
	}

	return nil
}

// Encode encodes the receiver to w.
// This is part of the Message interface implementation.
func (msg *MsgReject) Encode(w io.Writer, pver uint32) error {
	// Command that was rejected.
	err := s.WriteVarString(w, pver, msg.Cmd)
	if err != nil {
		return err
	}

	// Code indicating why the command was rejected.
	err = s.WriteElements(w, msg.Code)
	if err != nil {
		return err
	}

	// Human readable string with specific details (over and above the
	// reject code above) about why the command was rejected.
	err = s.WriteVarString(w, pver, msg.Reason)
	if err != nil {
		return err
	}

	// CmdBlock and CmdTx messages have an additional hash field that
	// identifies the specific block or transaction.
	if msg.Cmd == CmdBlock || msg.Cmd == CmdTx {
		err := s.WriteElements(w, &msg.Hash)
		if err != nil {
			return err
		}
	}

	return nil
}

// Command returns the protocol command string for the message.  This is part
// of the Message interface implementation.
func (msg *MsgReject) Command() string {
	return CmdReject
}

// MaxPayloadLength returns the maximum length the payload can be for the
// receiver.  This is part of the Message interface implementation.
func (msg *MsgReject) MaxPayloadLength(pver uint32) uint32 {
	// Unfortunately the protocol does not enforce a sane limit on the
	// length of the reason, so the max payload is the overall maximum
	// message payload.
	// TODO, revisit the limit design
	return uint32(MaxMessagePayload)
}

// NewMsgReject returns a new reject message that conforms to the Message
// interface.
// See MsgReject for details.
func NewMsgReject(command string, code RejectCode, reason string) *MsgReject {
	return &MsgReject{
		Cmd:    command,
		Code:   code,
		Reason: reason,
	}
}
