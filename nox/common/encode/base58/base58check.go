// Copyright (c) 2017-2018 The nox developers
// Copyright (c) 2013-2014 The btcsuite developers
// Copyright (c) 2015 The Decred developers
// Use of this source code is governed by an ISC
// license that can be found in the LICENSE file.

package base58

import (
	"errors"
	"github.com/noxproject/nox/common/hash"
	"github.com/noxproject/nox/common/hash/dcr"
	"github.com/noxproject/nox/common/hash/btc"
)

// ErrChecksum indicates that the checksum of a check-encoded string does not verify against
// the checksum.
var ErrChecksum = errors.New("checksum error")

// ErrInvalidFormat indicates that the check-encoded string has an invalid format.
var ErrInvalidFormat = errors.New("invalid format: version and/or checksum bytes missing")

// btc checksum: first four bytes of double-sha256.
func checksum_btc(input []byte) (cksum [4]byte) {
	h := btc.DoubleHashB(input)
	copy(cksum[:],h[:])
	return
}
// dcr checksum: first four bytes of double-BLAKE256.
func checksum_dcr(input []byte) (cksum [4]byte) {
	h := dcr.DoubleHashB(input)
	copy(cksum[:],h[:])
	return
}
// checksum: first four bytes of double-BLAKEb.
func checksum(input []byte) (cksum [4]byte) {
	h := hash.DoubleHashB(input)
	copy(cksum[:],h[:])
	return
}

// CheckEncode prepends two version bytes and appends a four byte checksum.
func CheckEncode(input []byte, version [2]byte) string {
	switch version[0] {
		case 0x42 :  //BTC
			return checkEncode(input,version[1:],checksum_btc)
		case 0x44 :  //DCR
			return checkEncode(input,version[:],checksum_dcr)
		default:
			return checkEncode(input,version[:],checksum)
	}
}

func checkEncode(input []byte, version []byte, cksumfunc func([]byte) [4]byte) string{
	b := make([]byte, 0, len(version)+len(input)+4)
	b = append(b, version[:]...)
	b = append(b, input[:]...)
	var cksum [4]byte
	cksum = cksumfunc(b)
	b = append(b, cksum[:]...)
	return Encode(b)
}

func checkDecode(input string, version_size int, cksumfunc func([]byte) [4]byte) (result []byte, version [2]byte, err error) {
	decoded := Decode(input)
	if len(decoded) < 4 + version_size {
		return nil, [2]byte{0, 0}, ErrInvalidFormat
	}
	if version_size == 1 {
		version = [2]byte{decoded[0], 0}
	}else{
		version = [2]byte{decoded[0],decoded[1]}
	}
	var cksum [4]byte
	copy(cksum[:], decoded[len(decoded)-4:])
	if cksumfunc(decoded[:len(decoded)-4]) != cksum {
		return nil, [2]byte{0, 0}, ErrChecksum
	}
	payload := decoded[version_size : len(decoded)-4]
	result = append(result, payload...)
	return
}

// CheckDecode decodes a string that was encoded with CheckEncode and verifies
// the checksum.
func CheckDecode(input string) (result []byte, version [2]byte, err error) {
	return checkDecode(input,2, checksum)
}

func BtcCheckDecode(input string) (result []byte, version byte, err error) {
	r,v,err := checkDecode(input, 1, checksum_btc)
	if err!=nil{
		return nil,0,err
	}
	return r, v[0],err
}


func DcrCheckDecode(input string) (result []byte, version [2]byte, err error) {
	return checkDecode(input,2, checksum_dcr)
}

