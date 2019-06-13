// Copyright (c) 2017-2018 The nox developers
// Copyright (c) 2013-2016 The btcsuite developers
// Copyright (c) 2015-2016 The Decred developers
// Use of this source code is governed by an ISC
// license that can be found in the LICENSE file.

package message

import (
	"fmt"
	"github.com/HalalChain/qitmeer/common/hash"
	"github.com/HalalChain/qitmeer/core/blockdag"
	"github.com/HalalChain/qitmeer/core/protocol"
	"io"
	s "github.com/HalalChain/qitmeer/core/serialization"

)

// MaxBlockLocatorsPerMsg is the maximum number of block locator hashes allowed
// per message.
const MaxBlockLocatorsPerMsg = 500

// MsgGetBlocks implements the Message interface and represents a getblocks
// message.  It is used to request a list of blocks starting after the last
// known hash in the slice of block locator hashes.  The list is returned
// via an inv message (MsgInv) and is limited by a specific hash to stop at
// or the maximum number of blocks per message.
//
// Set the HashStop field to the hash at which to stop and use
// AddBlockLocatorHash to build up the list of block locator hashes.
//
// The algorithm for building the block locator hashes should be to add the
// hashes in reverse order until you reach the genesis block.  In order to keep
// the list of locator hashes to a reasonable number of entries, first add the
// most recent 10 block hashes, then double the step each loop iteration to
// exponentially decrease the number of hashes the further away from head and
// closer to the genesis block you get.
type MsgGetBlocks struct {
	ProtocolVersion    uint32
	BlockLocatorHashes []*hash.Hash
	GS                 *blockdag.GraphState
}

// AddBlockLocatorHash adds a new block locator hash to the message.
func (msg *MsgGetBlocks) AddBlockLocatorHash(hash *hash.Hash) error {
	if len(msg.BlockLocatorHashes)+1 > MaxBlockLocatorsPerMsg {
		str := fmt.Sprintf("too many block locator hashes for message [max %v]",
			MaxBlockLocatorsPerMsg)
		return messageError("MsgGetBlocks.AddBlockLocatorHash", str)
	}

	msg.BlockLocatorHashes = append(msg.BlockLocatorHashes, hash)
	return nil
}

// Decode decodes r using the protocol encoding into the receiver.
// This is part of the Message interface implementation.
func (msg *MsgGetBlocks) Decode(r io.Reader, pver uint32) error {
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
		return messageError("MsgGetBlocks.BtcDecode", str)
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
	msg.GS=blockdag.NewGraphState()
	err=msg.GS.Decode(r,pver)
	if err != nil {
		return err
	}
	return nil
}

// Encode encodes the receiver to w using the protocol encoding.
// This is part of the Message interface implementation.
func (msg *MsgGetBlocks) Encode(w io.Writer, pver uint32) error {
	count := len(msg.BlockLocatorHashes)
	if count > MaxBlockLocatorsPerMsg {
		str := fmt.Sprintf("too many block locator hashes for message "+
			"[count %v, max %v]", count, MaxBlockLocatorsPerMsg)
		return messageError("MsgGetBlocks.BtcEncode", str)
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
		err = s.WriteElements(w, hash)
		if err != nil {
			return err
		}
	}

	err = msg.GS.Encode(w,pver)
	if err != nil {
		return err
	}

	return err
}

// Command returns the protocol command string for the message.  This is part
// of the Message interface implementation.
func (msg *MsgGetBlocks) Command() string {
	return CmdGetBlocks
}

// MaxPayloadLength returns the maximum length the payload can be for the
// receiver.  This is part of the Message interface implementation.
func (msg *MsgGetBlocks) MaxPayloadLength(pver uint32) uint32 {
	// Protocol version 4 bytes + num hashes (varInt) + max block locator
	// hashes + hash stop.
	return 4 + MaxVarIntPayload + (MaxBlockLocatorsPerMsg * hash.HashSize) + msg.GS.MaxPayloadLength()
}

// NewMsgGetBlocks returns a new getblocks message that conforms to the
// Message interface using the passed parameters and defaults for the remaining
// fields.
func NewMsgGetBlocks(gs *blockdag.GraphState) *MsgGetBlocks {
	return &MsgGetBlocks{
		ProtocolVersion:    protocol.ProtocolVersion,
		BlockLocatorHashes: make([]*hash.Hash, 0, MaxBlockLocatorsPerMsg),
		GS:gs,
	}
}
