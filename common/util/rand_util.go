// Copyright 2017-2018 The qitmeer developers

package util

import (
	cryptorand "crypto/rand"
	"encoding/hex"
	"io"
	"math/big"
)

// ReadRand read size bytes from input rand
// if input rand is nil, use crypto/rand
func ReadSizedRand(rand io.Reader, size uint) []byte {
	readBuff := make([]byte, size)
	if rand == nil {
		rand = cryptorand.Reader
	}
	_, err := io.ReadFull(rand, readBuff)
	if err != nil {
		panic("reading from crypto/rand failed: " + err.Error())
	}
	return readBuff
}

// PaddedAppend appends the src byte slice to dst, returning the new slice.
// If the length of the source is smaller than the passed size, leading zero
// bytes are appended to the dst slice before appending src.
// Example :
// Bitcoin uncompressed pubkey
//    uncompressed := make([]byte, 0, 65)
//    uncompressed = append(uncompressed, 0x01)
//    uncompressed = PaddedAppend(32, uncompressed, p.X.Bytes())
//    uncompressed = PaddedAppend(32, uncompressed, p.Y.Bytes())
func PaddedAppend(size uint, dst, src []byte) []byte {
	for i := 0; i < int(size)-len(src); i++ {
		dst = append(dst, 0)
	}
	return append(dst, src...)
}

func RightPadBytes(slice []byte, l int) []byte {
	if l <= len(slice) {
		return slice
	}

	padded := make([]byte, l)
	copy(padded, slice)

	return padded
}

func LeftPadBytes(slice []byte, l int) []byte {
	if l <= len(slice) {
		return slice
	}

	padded := make([]byte, l)
	copy(padded[l-len(slice):], slice)

	return padded
}

// PaddedBytes encodes a big integer as a big-endian byte slice, if the length of
// byte slice is smaller than the passed size, leading zero bytes will be added.
// Example :
// Ethereum privatekey
//   seckey := PaddedBigInt(32, prv.D)
// Ethereum pubkey
func PaddedBytes(size uint, n *big.Int) []byte {
	if n.BitLen()/8 >= int(size) {
		return n.Bytes()
	}
	return PaddedAppend(size, nil, n.Bytes())
}

// MustDecodeHex wrap the calling to hex.DecodeString() method to return the bytes
// represented by the hexadecimal string. It panics if an error occurs.
// This is useful in the tests or some special cases.
func MustDecodeHexString(hexStr string) []byte {
	bytes, err := hex.DecodeString(hexStr)
	if err != nil {
		panic("invalid hex string in " + err.Error() + ", hex: " + hexStr)
	}
	return bytes
}
