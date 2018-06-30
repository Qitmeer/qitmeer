// Copyright (c) 2017-2018 The nox developers
package message

import "github.com/noxproject/nox/core/types"

// MsgTx implements the Message interface and represents a transaction message.
// It is used to deliver transaction information in response to a getdata
// message (MsgGetData) for a given transaction.
//
// Use the AddTxIn and AddTxOut functions to build up the list of transaction
// inputs and outputs.

type MsgTx struct {
 	*types.Transaction
}
