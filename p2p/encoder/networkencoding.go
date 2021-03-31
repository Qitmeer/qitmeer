/*
 * Copyright (c) 2017-2020 The qitmeer developers
 */

package encoder

import (
	"io"
)

// Defines the different encoding formats
const (
	SSZ       = "ssz"        // SSZ is SSZ only.
	SSZSnappy = "ssz-snappy" // SSZSnappy is SSZ with snappy compression.
)

// NetworkEncoding represents an encoder compatible with Qitmeer 2.0 p2p.
type NetworkEncoding interface {
	// DecodeGossip to the provided gossip message. The interface must be a pointer to the decoding destination.
	DecodeGossip([]byte, interface{}) error
	// DecodeWithMaxLength a bytes from a reader with a varint length prefix. The interface must be a pointer to the
	// decoding destination. The length of the message should not be more than the provided limit.
	DecodeWithMaxLength(io.Reader, interface{}) error
	// EncodeGossip an arbitrary gossip message to the provided writer. The interface must be a pointer object to encode.
	EncodeGossip(io.Writer, interface{}) (int, error)
	// EncodeWithMaxLength an arbitrary message to the provided writer with a varint length prefix. The interface must be
	// a pointer object to encode. The encoded message should not be bigger than the provided limit.
	EncodeWithMaxLength(io.Writer, interface{}) (int, error)
	// ProtocolSuffix returns the last part of the protocol ID to indicate the encoding scheme.
	ProtocolSuffix() string
	// return max chunk size
	GetMaxChunkSize() uint64
	// return msg size
	GetSize(msg interface{}) int
}
