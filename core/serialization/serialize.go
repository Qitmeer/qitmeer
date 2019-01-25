// Copyright 2017-2018 The nox developers

package serialization

import "io"

// TODO, redefine the protocol version and storage

type Serializable interface {

	Serialize(w io.Writer) error

	Deserialize(r io.Reader) error
}