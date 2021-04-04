package synch

import (
	"fmt"
	s "github.com/Qitmeer/qitmeer/core/serialization"
	"io"
)

const (
	// MaxFilterAddDataSize is the maximum byte size of a data
	// element to add to the Bloom filter.  It is equal to the
	// maximum element size of a script.
	MaxFilterAddDataSize = 520
)

// MsgFilterAdd implements the Message interface and represents a qitmeer
// filteradd message.  It is used to add a data element to an existing Bloom
// filter.
//
// This message was not added until protocol version BIP0037Version.
type MsgFilterAdd struct {
	Data []byte
}

// QitmeerDecode decodes r using the qitmeer protocol encoding into the receiver.
// This is part of the Message interface implementation.
func (msg *MsgFilterAdd) BtcDecode(r io.Reader, pver uint32, enc MessageEncoding) error {
	var err error
	msg.Data, err = s.ReadVarBytes(r, pver, MaxFilterAddDataSize,
		"filteradd data")
	return err
}

// QitmeerEncode encodes the receiver to w using the qitmeer protocol encoding.
// This is part of the Message interface implementation.
func (msg *MsgFilterAdd) QitmeerEncode(w io.Writer, pver uint32, enc MessageEncoding) error {
	size := len(msg.Data)
	if size > MaxFilterAddDataSize {
		str := fmt.Sprintf("filteradd size too large for message "+
			"[size %v, max %v]", size, MaxFilterAddDataSize)
		return messageError("MsgFilterAdd.BtcEncode", str)
	}

	return s.WriteVarBytes(w, pver, msg.Data)
}

// Command returns the protocol command string for the message.  This is part
// of the Message interface implementation.
func (msg *MsgFilterAdd) Command() string {
	return CmdFilterAdd
}

// MaxPayloadLength returns the maximum length the payload can be for the
// receiver.  This is part of the Message interface implementation.
func (msg *MsgFilterAdd) MaxPayloadLength(pver uint32) uint32 {
	return uint32(s.VarIntSerializeSize(MaxFilterAddDataSize)) +
		MaxFilterAddDataSize
}

// NewMsgFilterAdd returns a new qitmeer filteradd message that conforms to the
// Message interface.  See MsgFilterAdd for details.
func NewMsgFilterAdd(data []byte) *MsgFilterAdd {
	return &MsgFilterAdd{
		Data: data,
	}
}
