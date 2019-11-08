// Copyright (c) 2013-2016 The btcsuite developers
// Copyright (c) 2015-2016 The Decred developers
// Use of this source code is governed by an ISC
// license that can be found in the LICENSE file.

package message

import (
	"fmt"
	"github.com/Qitmeer/qitmeer/common/hash"
	"github.com/Qitmeer/qitmeer/core/blockdag"
	"github.com/Qitmeer/qitmeer/core/protocol"
	s "github.com/Qitmeer/qitmeer/core/serialization"
	"io"
)

// MsgGetHeaders implements the Message interface and represents a getheaders
// message.  It is used to request a list of block headers for blocks starting
// after the last known hash in the slice of block locator hashes.
// The list is returned via a headers message (MsgHeaders) and is limited by a
// specific hash to stop at or the maximum number of block headers per message,
//
// Set the HashStop field to the hash at which to stop and use
// AddBlockLocatorHash to build up the list of block locator hashes.
//
// The algorithm for building the block locator hashes should be to add the
// hashes in reverse order until you reach the genesis block.  In order to keep
// the list of locator hashes to a resonable number of entries, first add the
// most recent 10 block hashes, then double the step each loop iteration to
// exponentially decrease the number of hashes the further away from head and
// closer to the genesis block you get.
type MsgGetHeaders struct {
	ProtocolVersion    uint32
	BlockLocatorHashes []*hash.Hash
	GS                 *blockdag.GraphState
}

// AddBlockLocatorHash adds a new block locator hash to the message.
func (msg *MsgGetHeaders) AddBlockLocatorHash(hash *hash.Hash) error {
	if len(msg.BlockLocatorHashes)+1 > MaxBlockLocatorsPerMsg {
		str := fmt.Sprintf("too many block locator hashes for message [max %v]",
			MaxBlockLocatorsPerMsg)
		return messageError("MsgGetHeaders.AddBlockLocatorHash", str)
	}

	hashValue := *hash
	msg.BlockLocatorHashes = append(msg.BlockLocatorHashes, &hashValue)
	return nil
}

// Decode decodes r using the protocol encoding into the receiver.
// This is part of the Message interface implementation.
func (msg *MsgGetHeaders) Decode(r io.Reader, pver uint32) error {
	err := s.ReadElements(r, &msg.ProtocolVersion)
	if err != nil {
		return err
	}

	// Read num block locator hashes and limit to max.
	count, err := s.ReadVarInt(r, pver)
	if err != nil {
		return err
	}
	if count > MaxBlockLocatorsPerMsg {
		str := fmt.Sprintf("too many block locator hashes for message "+
			"[count %v, max %v]", count, MaxBlockLocatorsPerMsg)
		return messageError("MsgGetHeaders.BtcDecode", str)
	}

	// Create a contiguous slice of hashes to deserialize into in order to
	// reduce the number of allocations.
	locatorHashes := make([]hash.Hash, count)
	msg.BlockLocatorHashes = make([]*hash.Hash, 0, count)
	for i := uint64(0); i < count; i++ {
		hash := &locatorHashes[i]
		err := s.ReadElements(r, hash)
		if err != nil {
			return err
		}
		msg.AddBlockLocatorHash(hash)
	}
	msg.GS = blockdag.NewGraphState()
	err = msg.GS.Decode(r, pver)
	if err != nil {
		return err
	}
	return nil
}

// Encode encodes the receiver to w using the protocol encoding.
// This is part of the Message interface implementation.
func (msg *MsgGetHeaders) Encode(w io.Writer, pver uint32) error {
	// Limit to max block locator hashes per message.
	count := len(msg.BlockLocatorHashes)
	if count > MaxBlockLocatorsPerMsg {
		str := fmt.Sprintf("too many block locator hashes for message "+
			"[count %v, max %v]", count, MaxBlockLocatorsPerMsg)
		return messageError("MsgGetHeaders.BtcEncode", str)
	}

	err := s.WriteElements(w, msg.ProtocolVersion)
	if err != nil {
		return err
	}

	err = s.WriteVarInt(w, pver, uint64(count))
	if err != nil {
		return err
	}

	for _, hash := range msg.BlockLocatorHashes {
		err := s.WriteElements(w, hash)
		if err != nil {
			return err
		}
	}

	err = msg.GS.Encode(w, pver)
	if err != nil {
		return err
	}

	return err
}

// Command returns the protocol command string for the message.  This is part
// of the Message interface implementation.
func (msg *MsgGetHeaders) Command() string {
	return CmdGetHeaders
}

// MaxPayloadLength returns the maximum length the payload can be for the
// receiver.  This is part of the Message interface implementation.
func (msg *MsgGetHeaders) MaxPayloadLength(pver uint32) uint32 {
	// Version 4 bytes + num block locator hashes (varInt) + max allowed block
	// locators + hash stop.
	return 4 + MaxVarIntPayload + (MaxBlockLocatorsPerMsg * hash.HashSize) + msg.GS.MaxPayloadLength()
}

func (msg *MsgGetHeaders) String() string {
	return fmt.Sprintf("ProtocolVersion:%d Blocks:%d GS:%s", msg.ProtocolVersion, len(msg.BlockLocatorHashes), msg.GS.String())
}

// NewMsgGetHeaders returns a new  getheaders message that conforms to
// the Message interface.  See MsgGetHeaders for details.
func NewMsgGetHeaders(gs *blockdag.GraphState) *MsgGetHeaders {
	return &MsgGetHeaders{
		ProtocolVersion:    protocol.ProtocolVersion,
		BlockLocatorHashes: make([]*hash.Hash, 0, MaxBlockLocatorsPerMsg),
		GS:                 gs,
	}
}
