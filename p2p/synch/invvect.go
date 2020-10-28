/*
 * Copyright (c) 2017-2020 The qitmeer developers
 */

package synch

import (
	"fmt"
	"github.com/Qitmeer/qitmeer/common/hash"
	pb "github.com/Qitmeer/qitmeer/p2p/proto/v1"
)

const (
	// MaxInvPerMsg is the maximum number of inventory vectors that can be in a
	// single inv message.
	MaxInvPerMsg = 50000
)

// InvType represents the allowed types of inventory vectors.  See InvVect.
type InvType uint32

// These constants define the various supported inventory vector types.
const (
	InvTypeError InvType = 0
	InvTypeTx    InvType = 1
	InvTypeBlock InvType = 2
)

// Map of service flags back to their constant names for pretty printing.
var ivStrings = map[InvType]string{
	InvTypeError: "ERROR",
	InvTypeTx:    "MSG_TX",
	InvTypeBlock: "MSG_BLOCK",
}

// String returns the InvType in human-readable form.
func (invtype InvType) String() string {
	if s, ok := ivStrings[invtype]; ok {
		return s
	}
	return fmt.Sprintf("Unknown InvType (%d)", uint32(invtype))
}

func (invtype InvType) Value() uint32 {
	return uint32(invtype)
}

// NewInvVect returns a new InvVect using the provided type and hash.
func NewInvVect(typ InvType, hash *hash.Hash) *pb.InvVect {
	return &pb.InvVect{
		Type: typ.Value(),
		Hash: &pb.Hash{Hash: hash.Bytes()},
	}
}
