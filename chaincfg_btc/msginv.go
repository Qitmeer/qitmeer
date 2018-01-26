// Copyright (c) 2013-2015 The btcsuite developers
// Use of this source code is governed by an ISC
// license that can be found in the LICENSE file.

package chaincfg

import (
	"fmt"
	"io"

	"github.com/dindinw/dagproject/chaincfg_btc/chainhash"
)

const (
	// MaxInvPerMsg is the maximum number of inventory vectors that can be in a
	// single bitcoin inv message.
	MaxInvPerMsg = 50000

	// Maximum payload size for an inventory vector.
	maxInvVectPayload = 4 + chainhash.HashSize

	// InvWitnessFlag denotes that the inventory vector type is requesting,
	// or sending a version which includes witness data.
	InvWitnessFlag = 1 << 30
)

// InvType represents the allowed types of inventory vectors.  See InvVect.
type InvType uint32

// These constants define the various supported inventory vector types.
const (
	InvTypeError                InvType = 0
	InvTypeTx                   InvType = 1
	InvTypeBlock                InvType = 2
	InvTypeFilteredBlock        InvType = 3
	InvTypeWitnessBlock         InvType = InvTypeBlock | InvWitnessFlag
	InvTypeWitnessTx            InvType = InvTypeTx | InvWitnessFlag
	InvTypeFilteredWitnessBlock InvType = InvTypeFilteredBlock | InvWitnessFlag
)

// Map of service flags back to their constant names for pretty printing.
var ivStrings = map[InvType]string{
	InvTypeError:                "ERROR",
	InvTypeTx:                   "MSG_TX",
	InvTypeBlock:                "MSG_BLOCK",
	InvTypeFilteredBlock:        "MSG_FILTERED_BLOCK",
	InvTypeWitnessBlock:         "MSG_WITNESS_BLOCK",
	InvTypeWitnessTx:            "MSG_WITNESS_TX",
	InvTypeFilteredWitnessBlock: "MSG_FILTERED_WITNESS_BLOCK",
}

// String returns the InvType in human-readable form.
func (invtype InvType) String() string {
	if s, ok := ivStrings[invtype]; ok {
		return s
	}

	return fmt.Sprintf("Unknown InvType (%d)", uint32(invtype))
}

// InvVect defines a bitcoin inventory vector which is used to describe data,
// as specified by the Type field, that a peer wants, has, or does not have to
// another peer.
type InvVect struct {
	Type InvType        // Type of data
	Hash chainhash.Hash // Hash of the data
}

// NewInvVect returns a new InvVect using the provided type and hash.
func NewInvVect(typ InvType, hash *chainhash.Hash) *InvVect {
	return &InvVect{
		Type: typ,
		Hash: *hash,
	}
}

// readInvVect reads an encoded InvVect from r depending on the protocol
// version.
func readInvVect(r io.Reader, pver uint32, iv *InvVect) error {
	return readElements(r, &iv.Type, &iv.Hash)
}

// writeInvVect serializes an InvVect to w depending on the protocol version.
func writeInvVect(w io.Writer, pver uint32, iv *InvVect) error {
	return writeElements(w, iv.Type, &iv.Hash)
}

// defaultInvListAlloc is the default size used for the backing array for an
// inventory list.  The array will dynamically grow as needed, but this
// figure is intended to provide enough space for the max number of inventory
// vectors in a *typical* inventory message without needing to grow the backing
// array multiple times.  Technically, the list can grow to MaxInvPerMsg, but
// rather than using that large figure, this figure more accurately reflects the
// typical case.
const defaultInvListAlloc = 1000

// MsgInv implements the Message interface and represents a bitcoin inv message.
// It is used to advertise a peer's known data such as blocks and transactions
// through inventory vectors.  It may be sent unsolicited to inform other peers
// of the data or in response to a getblocks message (MsgGetBlocks).  Each
// message is limited to a maximum number of inventory vectors, which is
// currently 50,000.
//
// Use the AddInvVect function to build up the list of inventory vectors when
// sending an inv message to another peer.
type MsgInv struct {
	InvList []*InvVect
}

// AddInvVect adds an inventory vector to the message.
func (msg *MsgInv) AddInvVect(iv *InvVect) error {
	if len(msg.InvList)+1 > MaxInvPerMsg {
		str := fmt.Sprintf("too many invvect in message [max %v]",
			MaxInvPerMsg)
		return messageError("MsgInv.AddInvVect", str)
	}

	msg.InvList = append(msg.InvList, iv)
	return nil
}

// BtcDecode decodes r using the bitcoin protocol encoding into the receiver.
// This is part of the Message interface implementation.
func (msg *MsgInv) BtcDecode(r io.Reader, pver uint32, enc MessageEncoding) error {
	count, err := ReadVarInt(r, pver)
	if err != nil {
		return err
	}

	// Limit to max inventory vectors per message.
	if count > MaxInvPerMsg {
		str := fmt.Sprintf("too many invvect in message [%v]", count)
		return messageError("MsgInv.BtcDecode", str)
	}

	// Create a contiguous slice of inventory vectors to deserialize into in
	// order to reduce the number of allocations.
	invList := make([]InvVect, count)
	msg.InvList = make([]*InvVect, 0, count)
	for i := uint64(0); i < count; i++ {
		iv := &invList[i]
		err := readInvVect(r, pver, iv)
		if err != nil {
			return err
		}
		msg.AddInvVect(iv)
	}

	return nil
}

// BtcEncode encodes the receiver to w using the bitcoin protocol encoding.
// This is part of the Message interface implementation.
func (msg *MsgInv) BtcEncode(w io.Writer, pver uint32, enc MessageEncoding) error {
	// Limit to max inventory vectors per message.
	count := len(msg.InvList)
	if count > MaxInvPerMsg {
		str := fmt.Sprintf("too many invvect in message [%v]", count)
		return messageError("MsgInv.BtcEncode", str)
	}

	err := WriteVarInt(w, pver, uint64(count))
	if err != nil {
		return err
	}

	for _, iv := range msg.InvList {
		err := writeInvVect(w, pver, iv)
		if err != nil {
			return err
		}
	}

	return nil
}

// Command returns the protocol command string for the message.  This is part
// of the Message interface implementation.
func (msg *MsgInv) Command() string {
	return CmdInv
}

// MaxPayloadLength returns the maximum length the payload can be for the
// receiver.  This is part of the Message interface implementation.
func (msg *MsgInv) MaxPayloadLength(pver uint32) uint32 {
	// Num inventory vectors (varInt) + max allowed inventory vectors.
	return MaxVarIntPayload + (MaxInvPerMsg * maxInvVectPayload)
}

// NewMsgInv returns a new bitcoin inv message that conforms to the Message
// interface.  See MsgInv for details.
func NewMsgInv() *MsgInv {
	return &MsgInv{
		InvList: make([]*InvVect, 0, defaultInvListAlloc),
	}
}

// NewMsgInvSizeHint returns a new bitcoin inv message that conforms to the
// Message interface.  See MsgInv for details.  This function differs from
// NewMsgInv in that it allows a default allocation size for the backing array
// which houses the inventory vector list.  This allows callers who know in
// advance how large the inventory list will grow to avoid the overhead of
// growing the internal backing array several times when appending large amounts
// of inventory vectors with AddInvVect.  Note that the specified hint is just
// that - a hint that is used for the default allocation size.  Adding more
// (or less) inventory vectors will still work properly.  The size hint is
// limited to MaxInvPerMsg.
func NewMsgInvSizeHint(sizeHint uint) *MsgInv {
	// Limit the specified hint to the maximum allow per message.
	if sizeHint > MaxInvPerMsg {
		sizeHint = MaxInvPerMsg
	}

	return &MsgInv{
		InvList: make([]*InvVect, 0, sizeHint),
	}
}
