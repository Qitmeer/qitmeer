package message

import (
	"github.com/Qitmeer/qitmeer/core/blockdag"
	"io"
)

type MsgGraphState struct {
	GS *blockdag.GraphState
}

func (msg *MsgGraphState) Decode(r io.Reader, pver uint32) error {
	msg.GS = blockdag.NewGraphState()
	err := msg.GS.Decode(r, pver)
	if err != nil {
		return err
	}
	return nil
}

func (msg *MsgGraphState) Encode(w io.Writer, pver uint32) error {
	err := msg.GS.Encode(w, pver)
	if err != nil {
		return err
	}
	return nil
}

func (msg *MsgGraphState) Command() string {
	return CmdGraphState
}

func (msg *MsgGraphState) MaxPayloadLength(pver uint32) uint32 {
	return msg.GS.MaxPayloadLength()
}

func NewMsgGraphState(gs *blockdag.GraphState) *MsgGraphState {
	return &MsgGraphState{
		GS: gs,
	}
}
