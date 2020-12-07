/*
 * Copyright (c) 2017-2020 The qitmeer developers
 */

package common

import (
	"fmt"
)

// P2PErrorCode identifies a kind of error.
type P2PErrorCode int

const (
	// p2p stream write error
	ErrStreamWrite P2PErrorCode = iota

	// p2p stream read error
	ErrStreamRead

	// p2p stream base error
	ErrStreamBase

	// p2p peer unknown error
	ErrPeerUnknown

	// p2p DAG consensus error
	ErrDAGConsensus

	// p2p message error
	ErrMessage
)

var p2pErrorCodeStrings = map[P2PErrorCode]string{
	ErrStreamWrite:  "ErrStreamWrite",
	ErrStreamRead:   "ErrStreamRead",
	ErrStreamBase:   "ErrStreamBase",
	ErrPeerUnknown:  "ErrPeerUnknown",
	ErrDAGConsensus: "ErrDAGConsensus",
	ErrMessage:      "ErrMessage",
}

func (e P2PErrorCode) String() string {
	if s := p2pErrorCodeStrings[e]; s != "" {
		return s
	}
	return fmt.Sprintf("Unknown P2PErrorCode (%d)", int(e))
}

type P2PError struct {
	Code  P2PErrorCode
	Error error
}

func NewP2PError(code P2PErrorCode, e error) *P2PError {
	return &P2PError{code, e}
}
