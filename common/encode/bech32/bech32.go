// Copyright (c) 2017-2018 The nox developers
// Copyright (c) 2017 Takatoshi Nakagawa
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in
// all copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
// THE SOFTWARE.

package bech32

import (
	"bytes"
	"fmt"
	"strings"
)

// The reference implementation for Bech32 and segwit addresses.

var charset = "qpzry9x8gf2tvdw0s3jn54khce6mua7l"

var generator = []int{0x3b6a57b2, 0x26508e6d, 0x1ea119fa, 0x3d4233dd, 0x2a1462b3}

func polymod(values []int) int {
	chk := 1
	for _, v := range values {
		top := chk >> 25
		chk = (chk&0x1ffffff)<<5 ^ v
		for i := 0; i < 5; i++ {
			if (top>>uint(i))&1 == 1 {
				chk ^= generator[i]
			}
		}
	}
	return chk
}

func hrpExpand(hrp string) []int {
	ret := []int{}
	for _, c := range hrp {
		ret = append(ret, int(c>>5))
	}
	ret = append(ret, 0)
	for _, c := range hrp {
		ret = append(ret, int(c&31))
	}
	return ret
}

func verifyChecksum(hrp string, data []int) bool {
	return polymod(append(hrpExpand(hrp), data...)) == 1
}

func createChecksum(hrp string, data []int) []int {
	values := append(append(hrpExpand(hrp), data...), []int{0, 0, 0, 0, 0, 0}...)
	mod := polymod(values) ^ 1
	ret := make([]int, 6)
	for p := 0; p < len(ret); p++ {
		ret[p] = (mod >> uint(5*(5-p))) & 31
	}
	return ret
}

// Encode encodes hrp(human-readable part) and data(32bit data array), returns Bech32 / or error
// if hrp is uppercase, return uppercase Bech32
func Encode(hrp string, data []int) (string, error) {
	if (len(hrp) + len(data) + 7) > 90 {
		return "", fmt.Errorf("too long : hrp length=%d, data length=%d", len(hrp), len(data))
	}
	if len(hrp) < 1 {
		return "", fmt.Errorf("invalid hrp : hrp=%v", hrp)
	}
	for p, c := range hrp {
		if c < 33 || c > 126 {
			return "", fmt.Errorf("invalid character human-readable part : hrp[%d]=%d", p, c)
		}
	}
	if strings.ToUpper(hrp) != hrp && strings.ToLower(hrp) != hrp {
		return "", fmt.Errorf("mix case : hrp=%v", hrp)
	}
	lower := strings.ToLower(hrp) == hrp
	hrp = strings.ToLower(hrp)
	combined := append(data, createChecksum(hrp, data)...)
	var ret bytes.Buffer
	ret.WriteString(hrp)
	ret.WriteString("1")
	for idx, p := range combined {
		if p < 0 || p >= len(charset) {
			return "", fmt.Errorf("invalid data : data[%d]=%d", idx, p)
		}
		ret.WriteByte(charset[p])
	}
	if lower {
		return ret.String(), nil
	}
	return strings.ToUpper(ret.String()), nil
}

// Decode decodes bechString(Bech32) returns hrp(human-readable part) and data(32bit data array) / or error
func Decode(bechString string) (string, []int, error) {
	if len(bechString) > 90 {
		return "", nil, fmt.Errorf("too long : len=%d", len(bechString))
	}
	if strings.ToLower(bechString) != bechString && strings.ToUpper(bechString) != bechString {
		return "", nil, fmt.Errorf("mixed case")
	}
	bechString = strings.ToLower(bechString)
	pos := strings.LastIndex(bechString, "1")
	if pos < 1 || pos+7 > len(bechString) {
		return "", nil, fmt.Errorf("separator '1' at invalid position : pos=%d , len=%d", pos, len(bechString))
	}
	hrp := bechString[0:pos]
	for p, c := range hrp {
		if c < 33 || c > 126 {
			return "", nil, fmt.Errorf("invalid character human-readable part : bechString[%d]=%d", p, c)
		}
	}
	data := []int{}
	for p := pos + 1; p < len(bechString); p++ {
		d := strings.Index(charset, fmt.Sprintf("%c", bechString[p]))
		if d == -1 {
			return "", nil, fmt.Errorf("invalid character data part : bechString[%d]=%d", p, bechString[p])
		}
		data = append(data, d)
	}
	if !verifyChecksum(hrp, data) {
		return "", nil, fmt.Errorf("invalid checksum")
	}
	return hrp, data[:len(data)-6], nil
}

func convertbits(data []int, frombits, tobits uint, pad bool) ([]int, error) {
	acc := 0
	bits := uint(0)
	ret := []int{}
	maxv := (1 << tobits) - 1
	for idx, value := range data {
		if value < 0 || (value>>frombits) != 0 {
			return nil, fmt.Errorf("invalid data range : data[%d]=%d (frombits=%d)", idx, value, frombits)
		}
		acc = (acc << frombits) | value
		bits += frombits
		for bits >= tobits {
			bits -= tobits
			ret = append(ret, (acc>>bits)&maxv)
		}
	}
	if pad {
		if bits > 0 {
			ret = append(ret, (acc<<(tobits-bits))&maxv)
		}
	} else if bits >= frombits {
		return nil, fmt.Errorf("illegal zero padding")
	} else if ((acc << (tobits - bits)) & maxv) != 0 {
		return nil, fmt.Errorf("non-zero padding")
	}
	return ret, nil
}

// SegwitAddrDecode decodes hrp(human-readable part) Segwit Address(string), returns version(int) and data(bytes array) / or error
func SegwitAddrDecode(hrp, addr string) (int, []int, error) {
	dechrp, data, err := Decode(addr)
	if err != nil {
		return -1, nil, err
	}
	if dechrp != hrp {
		return -1, nil, fmt.Errorf("invalid human-readable part : %s != %s", hrp, dechrp)
	}
	if len(data) < 1 {
		return -1, nil, fmt.Errorf("invalid decode data length : %d", len(data))
	}
	if data[0] > 16 {
		return -1, nil, fmt.Errorf("invalid witness version : %d", data[0])
	}
	res, err := convertbits(data[1:], 5, 8, false)
	if err != nil {
		return -1, nil, err
	}
	if len(res) < 2 || len(res) > 40 {
		return -1, nil, fmt.Errorf("invalid convertbits length : %d", len(res))
	}
	if data[0] == 0 && len(res) != 20 && len(res) != 32 {
		return -1, nil, fmt.Errorf("invalid program length for witness version 0 (per BIP141) : %d", len(res))
	}
	return data[0], res, nil
}

// SegwitAddrEncode encodes hrp(human-readable part) , version(int) and data(bytes array), returns Segwit Address / or error
func SegwitAddrEncode(hrp string, version int, program []int) (string, error) {
	if version < 0 || version > 16 {
		return "", fmt.Errorf("invalid witness version : %d", version)
	}
	if len(program) < 2 || len(program) > 40 {
		return "", fmt.Errorf("invalid program length : %d", len(program))
	}
	if version == 0 && len(program) != 20 && len(program) != 32 {
		return "", fmt.Errorf("invalid program length for witness version 0 (per BIP141) : %d", len(program))
	}
	data, err := convertbits(program, 8, 5, true)
	if err != nil {
		return "", err
	}
	ret, err := Encode(hrp, append([]int{version}, data...))
	if err != nil {
		return "", err
	}
	return ret, nil
}

// Copyright (c) 2017 The btcsuite developers
// Use of this source code is governed by an ISC
// license that can be found in the LICENSE file.

const bech32_charset = "qpzry9x8gf2tvdw0s3jn54khce6mua7l"

var gen = []int{0x3b6a57b2, 0x26508e6d, 0x1ea119fa, 0x3d4233dd, 0x2a1462b3}

// Decode decodes a bech32 encoded string, returning the human-readable
// part and the data part excluding the checksum.
func DecodeBech32(bech string) (string, []byte, error) {
	// The maximum allowed length for a bech32 string is 90. It must also
	// be at least 8 characters, since it needs a non-empty HRP, a
	// separator, and a 6 character checksum.
	if len(bech) < 8 || len(bech) > 90 {
		return "", nil, fmt.Errorf("invalid bech32 string length %d",
			len(bech))
	}
	// Only	ASCII characters between 33 and 126 are allowed.
	for i := 0; i < len(bech); i++ {
		if bech[i] < 33 || bech[i] > 126 {
			return "", nil, fmt.Errorf("invalid character in "+
				"string: '%c'", bech[i])
		}
	}

	// The characters must be either all lowercase or all uppercase.
	lower := strings.ToLower(bech)
	upper := strings.ToUpper(bech)
	if bech != lower && bech != upper {
		return "", nil, fmt.Errorf("string not all lowercase or all " +
			"uppercase")
	}

	// We'll work with the lowercase string from now on.
	bech = lower

	// The string is invalid if the last '1' is non-existent, it is the
	// first character of the string (no human-readable part) or one of the
	// last 6 characters of the string (since checksum cannot contain '1'),
	// or if the string is more than 90 characters in total.
	one := strings.LastIndexByte(bech, '1')
	if one < 1 || one+7 > len(bech) {
		return "", nil, fmt.Errorf("invalid index of 1")
	}

	// The human-readable part is everything before the last '1'.
	hrp := bech[:one]
	data := bech[one+1:]

	// Each character corresponds to the byte with value of the index in
	// 'charset'.
	decoded, err := toBytes(data)
	if err != nil {
		return "", nil, fmt.Errorf("failed converting data to bytes: "+
			"%v", err)
	}

	if !bech32VerifyChecksum(hrp, decoded) {
		moreInfo := ""
		checksum := bech[len(bech)-6:]
		expected, err := toChars(bech32Checksum(hrp,
			decoded[:len(decoded)-6]))
		if err == nil {
			moreInfo = fmt.Sprintf("Expected %v, got %v.",
				expected, checksum)
		}
		return "", nil, fmt.Errorf("checksum failed. " + moreInfo)
	}

	// We exclude the last 6 bytes, which is the checksum.
	return hrp, decoded[:len(decoded)-6], nil
}

// Encode encodes a byte slice into a bech32 string with the
// human-readable part hrb. Note that the bytes must each encode 5 bits
// (base32).
func EncodeBech32(hrp string, data []byte) (string, error) {
	// Calculate the checksum of the data and append it at the end.
	checksum := bech32Checksum(hrp, data)
	combined := append(data, checksum...)

	// The resulting bech32 string is the concatenation of the hrp, the
	// separator 1, data and checksum. Everything after the separator is
	// represented using the specified charset.
	dataChars, err := toChars(combined)
	if err != nil {
		return "", fmt.Errorf("unable to convert data bytes to chars: "+
			"%v", err)
	}
	return hrp + "1" + dataChars, nil
}

// toBytes converts each character in the string 'chars' to the value of the
// index of the correspoding character in 'charset'.
func toBytes(chars string) ([]byte, error) {
	decoded := make([]byte, 0, len(chars))
	for i := 0; i < len(chars); i++ {
		index := strings.IndexByte(charset, chars[i])
		if index < 0 {
			return nil, fmt.Errorf("invalid character not part of "+
				"charset: %v", chars[i])
		}
		decoded = append(decoded, byte(index))
	}
	return decoded, nil
}

// toChars converts the byte slice 'data' to a string where each byte in 'data'
// encodes the index of a character in 'charset'.
func toChars(data []byte) (string, error) {
	result := make([]byte, 0, len(data))
	for _, b := range data {
		if int(b) >= len(charset) {
			return "", fmt.Errorf("invalid data byte: %v", b)
		}
		result = append(result, charset[b])
	}
	return string(result), nil
}

// ConvertBits converts a byte slice where each byte is encoding fromBits bits,
// to a byte slice where each byte is encoding toBits bits.
func ConvertBits(data []byte, fromBits, toBits uint8, pad bool) ([]byte, error) {
	if fromBits < 1 || fromBits > 8 || toBits < 1 || toBits > 8 {
		return nil, fmt.Errorf("only bit groups between 1 and 8 allowed")
	}

	// The final bytes, each byte encoding toBits bits.
	var regrouped []byte

	// Keep track of the next byte we create and how many bits we have
	// added to it out of the toBits goal.
	nextByte := byte(0)
	filledBits := uint8(0)

	for _, b := range data {

		// Discard unused bits.
		b = b << (8 - fromBits)

		// How many bits remaining to extract from the input data.
		remFromBits := fromBits
		for remFromBits > 0 {
			// How many bits remaining to be added to the next byte.
			remToBits := toBits - filledBits

			// The number of bytes to next extract is the minimum of
			// remFromBits and remToBits.
			toExtract := remFromBits
			if remToBits < toExtract {
				toExtract = remToBits
			}

			// Add the next bits to nextByte, shifting the already
			// added bits to the left.
			nextByte = (nextByte << toExtract) | (b >> (8 - toExtract))

			// Discard the bits we just extracted and get ready for
			// next iteration.
			b = b << toExtract
			remFromBits -= toExtract
			filledBits += toExtract

			// If the nextByte is completely filled, we add it to
			// our regrouped bytes and start on the next byte.
			if filledBits == toBits {
				regrouped = append(regrouped, nextByte)
				filledBits = 0
				nextByte = 0
			}
		}
	}

	// We pad any unfinished group if specified.
	if pad && filledBits > 0 {
		nextByte = nextByte << (toBits - filledBits)
		regrouped = append(regrouped, nextByte)
		filledBits = 0
		nextByte = 0
	}

	// Any incomplete group must be <= 4 bits, and all zeroes.
	if filledBits > 0 && (filledBits > 4 || nextByte != 0) {
		return nil, fmt.Errorf("invalid incomplete group")
	}

	return regrouped, nil
}

// For more details on the checksum calculation, please refer to BIP 173.
func bech32Checksum(hrp string, data []byte) []byte {
	// Convert the bytes to list of integers, as this is needed for the
	// checksum calculation.
	integers := make([]int, len(data))
	for i, b := range data {
		integers[i] = int(b)
	}
	values := append(bech32HrpExpand(hrp), integers...)
	values = append(values, []int{0, 0, 0, 0, 0, 0}...)
	polymod := bech32Polymod(values) ^ 1
	var res []byte
	for i := 0; i < 6; i++ {
		res = append(res, byte((polymod>>uint(5*(5-i)))&31))
	}
	return res
}

// For more details on the polymod calculation, please refer to BIP 173.
func bech32Polymod(values []int) int {
	chk := 1
	for _, v := range values {
		b := chk >> 25
		chk = (chk&0x1ffffff)<<5 ^ v
		for i := 0; i < 5; i++ {
			if (b>>uint(i))&1 == 1 {
				chk ^= gen[i]
			}
		}
	}
	return chk
}

// For more details on HRP expansion, please refer to BIP 173.
func bech32HrpExpand(hrp string) []int {
	v := make([]int, 0, len(hrp)*2+1)
	for i := 0; i < len(hrp); i++ {
		v = append(v, int(hrp[i]>>5))
	}
	v = append(v, 0)
	for i := 0; i < len(hrp); i++ {
		v = append(v, int(hrp[i]&31))
	}
	return v
}

// For more details on the checksum verification, please refer to BIP 173.
func bech32VerifyChecksum(hrp string, data []byte) bool {
	integers := make([]int, len(data))
	for i, b := range data {
		integers[i] = int(b)
	}
	concat := append(bech32HrpExpand(hrp), integers...)
	return bech32Polymod(concat) == 1
}
