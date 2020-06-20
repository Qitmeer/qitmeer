package main

import (
	"github.com/Qitmeer/qitmeer/core/dbnamespace"
	"io"
)

type IBDBlock struct {
	length uint32
	bytes  []byte
}

func (b *IBDBlock) Encode(w io.Writer) error {
	var serializedLen [4]byte
	dbnamespace.ByteOrder.PutUint32(serializedLen[:], b.length)
	_, err := w.Write(serializedLen[:])
	if err != nil {
		return err
	}
	_, err = w.Write(b.bytes)
	return err
}
