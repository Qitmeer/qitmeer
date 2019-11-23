package message

import (
	"io"
)

type MsgMemPool struct{}

func (msg *MsgMemPool) Decode(r io.Reader, pver uint32) error {
	return nil
}

func (msg *MsgMemPool) Encode(w io.Writer, pver uint32) error {
	return nil
}

func (msg *MsgMemPool) Command() string {
	return CmdMemPool
}

func (msg *MsgMemPool) MaxPayloadLength(pver uint32) uint32 {
	return 0
}

func NewMsgMemPool() *MsgMemPool {
	return &MsgMemPool{}
}
