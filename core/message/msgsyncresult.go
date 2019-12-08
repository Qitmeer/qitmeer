package message

import (
	"github.com/Qitmeer/qitmeer/core/blockdag"
	s "github.com/Qitmeer/qitmeer/core/serialization"
	"io"
)

type MsgSyncResult struct {
	GS   *blockdag.GraphState
	Mode blockdag.SyncMode
}

func (msg *MsgSyncResult) Decode(r io.Reader, pver uint32) error {
	msg.GS = blockdag.NewGraphState()
	err := msg.GS.Decode(r, pver)
	if err != nil {
		return err
	}

	var mode byte
	err = s.ReadElements(r, &mode)
	if err != nil {
		return err
	}
	msg.Mode = blockdag.SyncMode(mode)
	return nil
}

func (msg *MsgSyncResult) Encode(w io.Writer, pver uint32) error {
	err := msg.GS.Encode(w, pver)
	if err != nil {
		return err
	}
	err = s.WriteElements(w, byte(msg.Mode))
	if err != nil {
		return err
	}
	return nil
}

func (msg *MsgSyncResult) Command() string {
	return CmdSyncResult
}

func (msg *MsgSyncResult) MaxPayloadLength(pver uint32) uint32 {
	return msg.GS.MaxPayloadLength() + 1
}

func NewMsgSyncResult(gs *blockdag.GraphState, mode blockdag.SyncMode) *MsgSyncResult {
	return &MsgSyncResult{
		GS:   gs,
		Mode: mode,
	}
}
