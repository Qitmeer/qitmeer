package synch

import (
	"io"
)

// MsgFilterClear implements the Message interface and represents a qitmeer
// filterclear message which is used to reset a Bloom filter.
//
// This message was not added until protocol version BIP0037Version and has
// no payload.
type MsgFilterClear struct{}

// QitmeerDecode decodes r using the qitmeer protocol encoding into the receiver.
// This is part of the Message interface implementation.
func (msg *MsgFilterClear) QitmeerDecode(r io.Reader, pver uint32, enc MessageEncoding) error {
	return nil
}

// QitmeerEncode encodes the receiver to w using the qitmeer protocol encoding.
// This is part of the Message interface implementation.
func (msg *MsgFilterClear) QitmeerEncode(w io.Writer, pver uint32, enc MessageEncoding) error {
	return nil
}

// Command returns the protocol command string for the message.  This is part
// of the Message interface implementation.
func (msg *MsgFilterClear) Command() string {
	return CmdFilterClear
}

// MaxPayloadLength returns the maximum length the payload can be for the
// receiver.  This is part of the Message interface implementation.
func (msg *MsgFilterClear) MaxPayloadLength(pver uint32) uint32 {
	return 0
}

// NewMsgFilterClear returns a new qitmeer filterclear message that conforms to the Message
// interface.  See MsgFilterClear for details.
func NewMsgFilterClear() *MsgFilterClear {
	return &MsgFilterClear{}
}
