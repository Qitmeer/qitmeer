package message

import (
	"github.com/noxproject/nox/core/types"
	"io"
	"fmt"
)

// MsgTx implements the Message interface and represents a Decred tx message.
// It is used to deliver transaction information in response to a getdata
// message (MsgGetData) for a given transaction.
//
// Use the AddTxIn and AddTxOut functions to build up the list of transaction
// inputs and outputs.

const (

)

type MsgTx struct {
	*types.Transaction
}

// decodePrefix decodes a transaction prefix and stores the contents
// in the embedded msgTx.
func (msg *MsgTx) decodePrefix(r io.Reader) (uint64, error) {
	count, err := s.ReadVarInt(r,0)
	if err != nil {
		return 0, err
	}

}
