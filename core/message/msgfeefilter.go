/*
 * Copyright (c) 2020.
 * Project:qitmeer
 * File:msgfeefilter.go
 * Date:8/17/20 12:00 PM
 * Author:Jin
 * Email:lochjin@gmail.com
 */

package message

import (
	s "github.com/Qitmeer/qitmeer/core/serialization"
	"io"
)

type MsgFeeFilter struct {
	MinFee int64
}

func (msg *MsgFeeFilter) Decode(r io.Reader, pver uint32) error {
	err := s.ReadElements(r, &msg.MinFee)
	return err
}

func (msg *MsgFeeFilter) Encode(w io.Writer, pver uint32) error {
	err := s.WriteElements(w, msg.MinFee)
	return err
}

func (msg *MsgFeeFilter) Command() string {
	return CmdFeeFilter
}

func (msg *MsgFeeFilter) MaxPayloadLength(pver uint32) uint32 {
	return 8
}

func NewMsgFeeFilter(minfee int64) *MsgFeeFilter {
	return &MsgFeeFilter{
		MinFee: minfee,
	}
}
