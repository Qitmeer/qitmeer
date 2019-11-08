// Copyright (c) 2017-2020 The qitmeer developers
// license that can be found in the LICENSE file.
// Reference resources of rust bitVector
package util

import "errors"

//invalid bit length
var errInvalidLength = errors.New("invalid length")

//BitVector this is design for cuckoo hash bytes
type BitVector struct {
	len int
	b   []byte
}

// init
func New(l int) (bv *BitVector, err error) {
	return NewFromBytes(make([]byte, l/8+1), l)
}

// convert from bytes
func NewFromBytes(b []byte, l int) (bv *BitVector, err error) {
	if l <= 0 {
		return nil, errInvalidLength
	}
	if len(b)*8 < l {
		return nil, errInvalidLength
	}
	return &BitVector{
		len: l,
		b:   b,
	}, nil
}

//get BitVector
func (bv *BitVector) Get(i int) bool {
	bi := i / 8
	return bv.b[bi]&(0x1<<uint(i%8)) != 0
}

//set BitVector
func (bv *BitVector) Set(i int, v bool) {
	bi := i / 8
	cv := bv.Get(i)
	if cv != v {
		bv.b[bi] ^= 0x1 << uint8(i%8)
	}
}

// set position
func (bv *BitVector) SetBitAt(pos int) {
	bv.b[pos/8] |= 1 << (uint(pos) % 8)
}

//return bytes
func (bv *BitVector) Bytes() []byte {
	return bv.b
}
