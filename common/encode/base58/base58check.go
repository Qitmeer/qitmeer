// Copyright (c) 2017-2018 The qitmeer developers
// Copyright (c) 2013-2014 The btcsuite developers
// Copyright (c) 2015 The Decred developers
// Use of this source code is governed by an ISC
// license that can be found in the LICENSE file.

package base58

import (
	"errors"
	"github.com/Qitmeer/qitmeer/common/hash"
	"github.com/Qitmeer/qitmeer/common/hash/btc"
	"github.com/Qitmeer/qitmeer/common/hash/dcr"
	"reflect"
)

// ErrChecksum indicates that the checksum of a check-encoded string does not verify against
// the checksum.
var ErrChecksum = errors.New("checksum error")

// ErrInvalidFormat indicates that the check-encoded string has an invalid format.
var ErrInvalidFormat = errors.New("invalid format: version and/or checksum bytes missing")

// btc checksum: first four bytes of double-sha256.
func checksum_btc(input []byte) []byte {
	h := btc.DoubleHashB(input)
	var cksum [4]byte
	copy(cksum[:], h[:])
	return cksum[:]
}

// dcr checksum: first four bytes of double-BLAKE256.
func checksum_dcr(input []byte) []byte {
	h := dcr.DoubleHashB(input)
	var cksum [4]byte
	copy(cksum[:], h[:])
	return cksum[:]
}

// checksum: first four bytes of double-BLAKEb.
func checksum_qitmeer(input []byte) []byte {
	h := hash.DoubleHashB(input)
	var cksum [4]byte
	copy(cksum[:], h[:])
	return cksum[:]
}

func checksum_ss(input []byte) []byte {
	return SingleHashChecksumFunc(hash.GetHasher(hash.Blake2b_512), 2)(input)
}

func SingleHashChecksumFunc(hasher hash.Hasher, cksum_size int) func([]byte) []byte {
	return func(input []byte) []byte {
		h := hash.CalcHash(input, hasher)
		var cksum []byte
		cksum = append(cksum, h[:cksum_size]...)
		return cksum[:]
	}
}

func DoubleHashChecksumFunc(hasher hash.Hasher, cksum_size int) func([]byte) []byte {
	return func(input []byte) []byte {
		first := hash.CalcHash(input, hasher)
		second := hash.CalcHash(first[:], hasher)
		var cksum []byte
		cksum = append(cksum, second[:cksum_size]...)
		return cksum[:]
	}
}

// CheckEncode prepends two version bytes and appends a four byte checksum.
func QitmeerCheckEncode(input []byte, version []byte) ([]byte, error) {
	return CheckEncode(input, version[:], 4, checksum_qitmeer)
}

func DcrCheckEncode(input []byte, version [2]byte) ([]byte, error) {
	return CheckEncode(input, version[:], 4, checksum_dcr)
}
func BtcCheckEncode(input []byte, version byte) ([]byte, error) {
	var ver []byte
	ver = append(ver, version)
	return CheckEncode(input, ver[:], 4, checksum_btc)
}

func checkInputOverflow(input []byte) ([]byte, error){
	if len(input) > 64*1024*1024 {
		return nil, errors.New("value too large")
	}
	return input,nil
}

func CheckEncode(input []byte, version []byte, cksum_size int, cksumfunc func([]byte) []byte) ([]byte, error) {
	input, err := checkInputOverflow(input)
	if err != nil {
		return nil, err
	}
	b := make([]byte, 0, len(version)+len(input)+cksum_size)
	b = append(b, version[:]...)
	b = append(b, input[:]...)
	var cksum []byte = cksumfunc(b)
	b = append(b, cksum[:]...)
	enc,_ := Encode(b)  //need not check input overflow again, ignore err here
	return enc,nil
}

func CheckDecode(input []byte, version_size, cksum_size int, cksumfunc func([]byte) []byte) (result []byte, version []byte, err error) {
	input,err = checkInputOverflow(input)
	if err != nil {
		return nil, nil, err
	}
	decoded := Decode(input)
	if len(decoded) < cksum_size+version_size {
		return nil, []byte{}, ErrInvalidFormat
	}
	version = append(version, decoded[:version_size]...)
	var cksum []byte
	cksum = append(cksum, decoded[len(decoded)-cksum_size:]...)
	if !reflect.DeepEqual(cksumfunc(decoded[:len(decoded)-cksum_size]), cksum[:]) {
		return nil, []byte{}, ErrChecksum
	}
	payload := decoded[version_size : len(decoded)-cksum_size]
	result = append(result, payload...)
	return
}

// QitmeerCheckDecode decodes a string that was encoded with 2 bytes version and verifies
// the checksum using blake2b-256 hash.
func QitmeerCheckDecode(input string) (result []byte, version [2]byte, err error) {
	r, v, err := CheckDecode([]byte(input), 2, 4, checksum_qitmeer)
	if err != nil {
		return nil, [2]byte{}, err
	}
	return r, [2]byte{v[0], v[1]}, nil
}

func BtcCheckDecode(input string) (result []byte, version byte, err error) {
	r, v, err := CheckDecode([]byte(input), 1, 4, checksum_btc)
	if err != nil {
		return nil, 0, err
	}
	return r, v[0], err
}

func DcrCheckDecode(input string) (result []byte, version [2]byte, err error) {
	r, v, err := CheckDecode([]byte(input), 2, 4, checksum_dcr)
	if err != nil {
		return nil, [2]byte{}, err
	}
	return r, [2]byte{v[0], v[1]}, nil
}
