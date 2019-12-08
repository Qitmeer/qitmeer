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

type MsgSyncDAG struct {
	MainLocator []*hash.Hash
	GS          *blockdag.GraphState
}

func (msg *MsgSyncDAG) Decode(r io.Reader, pver uint32) error {
	var count uint64
	err := s.ReadElements(r, &count)
	if err != nil {
		return err
	}
	msg.MainLocator = []*hash.Hash{}
	for i := uint64(0); i < count; i++ {
		var blockHash hash.Hash
		err := s.ReadElements(r, &blockHash)
		if err != nil {
			return err
		}
		msg.MainLocator = append(msg.MainLocator, &blockHash)
	}
	msg.GS = blockdag.NewGraphState()
	err = msg.GS.Decode(r, pver)
	if err != nil {
		return err
	}
	return nil
}

func (msg *MsgSyncDAG) Encode(w io.Writer, pver uint32) error {
	err := s.WriteElements(w, uint64(len(msg.MainLocator)))
	if err != nil {
		return err
	}

	for _, v := range msg.MainLocator {
		err = s.WriteElements(w, v)
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

func (msg *MsgSyncDAG) Command() string {
	return CmdSyncDAG
}

func (msg *MsgSyncDAG) MaxPayloadLength(pver uint32) uint32 {
	return (blockdag.MaxMainLocatorNum * hash.HashSize) + msg.GS.MaxPayloadLength()
}

func NewMsgSyncDAG(gs *blockdag.GraphState, locator []*hash.Hash) *MsgSyncDAG {
	return &MsgSyncDAG{
		MainLocator: locator,
		GS:          gs,
	}
}
