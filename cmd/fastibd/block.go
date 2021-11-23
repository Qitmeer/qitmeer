package main

import (
	"github.com/Qitmeer/qitmeer/core/dbnamespace"
	"github.com/Qitmeer/qng-core/core/types"
	"io"
)

type IBDBlock struct {
	length uint32
	bytes  []byte
	blk    *types.SerializedBlock
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

func (b *IBDBlock) Decode(bytes []byte) error {
	b.length = dbnamespace.ByteOrder.Uint32(bytes[:4])

	block, err := types.NewBlockFromBytes(bytes[4 : b.length+4])
	if err != nil {
		return err
	}
	b.blk = block
	return nil
}
