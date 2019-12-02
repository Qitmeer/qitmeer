// Copyright (c) 2017-2018 The qitmeer developers
// Use of this source code is governed by an ISC
// license that can be found in the LICENSE file.

package message

import (
	"github.com/Qitmeer/qitmeer/common/hash"
	"github.com/Qitmeer/qitmeer/core/blockdag"
	s "github.com/Qitmeer/qitmeer/core/serialization"
	"io"
)

type MsgSyncPoint struct {
	SyncPoint *hash.Hash
	GS        *blockdag.GraphState
}

func (msg *MsgSyncPoint) Decode(r io.Reader, pver uint32) error {
	var point hash.Hash
	err := s.ReadElements(r, &point)
	if err != nil {
		return err
	}
	msg.SyncPoint = &point

	msg.GS = blockdag.NewGraphState()
	err = msg.GS.Decode(r, pver)
	if err != nil {
		return err
	}
	return nil
}

func (msg *MsgSyncPoint) Encode(w io.Writer, pver uint32) error {
	err := s.WriteElements(w, msg.SyncPoint)
	if err != nil {
		return err
	}
	err = msg.GS.Encode(w, pver)
	if err != nil {
		return err
	}

	return err
}

func (msg *MsgSyncPoint) Command() string {
	return CmdSyncPoint
}

func (msg *MsgSyncPoint) MaxPayloadLength(pver uint32) uint32 {
	return hash.HashSize + msg.GS.MaxPayloadLength()
}

func NewMsgSyncPoint(gs *blockdag.GraphState, point *hash.Hash) *MsgSyncPoint {
	return &MsgSyncPoint{
		SyncPoint: point,
		GS:        gs,
	}
}
