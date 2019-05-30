// Copyright (c) 2013-2015 The btcsuite developers
// Copyright (c) 2015-2016 The Decred developers
// Use of this source code is governed by an ISC
// license that can be found in the LICENSE file.

package message

import (
	"fmt"
	"io"
	"qitmeer/common/hash"
	s "qitmeer/core/serialization"

)

// MaxMSBlocksAtHeadPerMsg is the maximum number of block hashes allowed
// per message.
const MaxMSBlocksAtHeadPerMsg = 8

// MsgMiningState implements the Message interface and represents a mining state
// message.  It is used to request a list of blocks located at the chain tip
// The list is returned is limited by the maximum number of blocks per message
// message.
type MsgMiningState struct {
	Version     uint32
	Height      uint32
	BlockHashes []*hash.Hash
}

// AddBlockHash adds a new block hash to the message.
func (msg *MsgMiningState) AddBlockHash(hash *hash.Hash) error {
	if len(msg.BlockHashes)+1 > MaxMSBlocksAtHeadPerMsg {
		str := fmt.Sprintf("too many block hashes for message [max %v]",
			MaxMSBlocksAtHeadPerMsg)
		return messageError("MsgMiningState.AddBlockHash", str)
	}

	msg.BlockHashes = append(msg.BlockHashes, hash)
	return nil
}

// Decode decodes r using the protocol encoding into the receiver.
// This is part of the Message interface implementation.
func (msg *MsgMiningState) Decode(r io.Reader, pver uint32) error {
	err := s.ReadElements(r, &msg.Version)
	if err != nil {
		return err
	}

	err = s.ReadElements(r, &msg.Height)
	if err != nil {
		return err
	}

	// Read num block hashes and limit to max.
	count, err := s.ReadVarInt(r, pver)
	if err != nil {
		return err
	}
	if count > MaxMSBlocksAtHeadPerMsg {
		str := fmt.Sprintf("too many block hashes for message "+
			"[count %v, max %v]", count, MaxMSBlocksAtHeadPerMsg)
		return messageError("MsgMiningState.BtcDecode", str)
	}

	msg.BlockHashes = make([]*hash.Hash, 0, count)
	for i := uint64(0); i < count; i++ {
		hash := hash.Hash{}
		err := s.ReadElements(r, &hash)
		if err != nil {
			return err
		}
		msg.AddBlockHash(&hash)
	}

	return nil
}

// Encode encodes the receiver to w using the protocol encoding.
// This is part of the Message interface implementation.
func (msg *MsgMiningState) Encode(w io.Writer, pver uint32) error {
	err := s.WriteElements(w, msg.Version)
	if err != nil {
		return err
	}

	err = s.WriteElements(w, msg.Height)
	if err != nil {
		return err
	}

	// Write block hashes.
	count := len(msg.BlockHashes)
	if count > MaxMSBlocksAtHeadPerMsg {
		str := fmt.Sprintf("too many block hashes for message "+
			"[count %v, max %v]", count, MaxMSBlocksAtHeadPerMsg)
		return messageError("MsgMiningState.BtcEncode", str)
	}

	err = s.WriteVarInt(w, pver, uint64(count))
	if err != nil {
		return err
	}

	for _, hash := range msg.BlockHashes {
		err = s.WriteElements(w, hash)
		if err != nil {
			return err
		}
	}

	return nil
}

// Command returns the protocol command string for the message.  This is part
// of the Message interface implementation.
func (msg *MsgMiningState) Command() string {
	return CmdMiningState
}

// MaxPayloadLength returns the maximum length the payload can be for the
// receiver.  This is part of the Message interface implementation.
func (msg *MsgMiningState) MaxPayloadLength(pver uint32) uint32 {
	// Protocol version 4 bytes + Height 4 bytes + num block hashes (varInt) +
	// block hashes
	return 4 + 4 + MaxVarIntPayload + (MaxMSBlocksAtHeadPerMsg *
		hash.HashSize)
}

// NewMsgMiningState returns a new miningstate message that conforms to
// the Message interface using the defaults for the fields.
func NewMsgMiningState() *MsgMiningState {
	return &MsgMiningState{
		Version:     1,
		Height:      0,
		BlockHashes: make([]*hash.Hash, 0, MaxMSBlocksAtHeadPerMsg),
	}
}
