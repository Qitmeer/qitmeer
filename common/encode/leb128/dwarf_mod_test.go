// Copyright (c) 2017-2018 The qitmeer developers

package leb128

import (
	"fmt"
	"reflect"
	"strings"
	"testing"
)

func TestInt64(t *testing.T) {
	tests := []struct {
		n    int64
		want []byte
	}{
		{0, []byte{0x00}},
		{1, []byte{0x01}},
		{2, []byte{0x02}},
		{3, []byte{0x03}},
		{4, []byte{0x04}},
		{5, []byte{0x05}},
		{63, []byte{0x3F}},
		{64, []byte{0x40}},
		{65, []byte{0x41}},
		{100, []byte{0x64}},
		{127, []byte{0x7F}},
		{128, []byte{0x80, 0x01}},
		{129, []byte{0x81, 0x01}},
		{2141192192, []byte{0x80, 0x80, 0x80, 0xFD, 0x07}},
		{^int64(0), []byte{0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0x01}},
	}
	for _, tt := range tests {
		t.Run("TestFromInt64", func(t *testing.T) {
			if got := Uleb128FromInt64(tt.n); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Uleb128FromInt64(%v) = %0x, want %0x", tt.n, got, tt.want)
			}
		})
		t.Run("TestToInt64", func(t *testing.T) {
			if got, _ := Uleb128ToUint64(tt.want); !reflect.DeepEqual(int64(got), tt.n) {
				t.Errorf("Uleb128ToUint64(%v) = %0x, want %0x", tt.want, got, tt.n)
			}
		})
	}
}

// Note
// The following test case copied & modified from
// https://github.com/Equim-chan/leb128/blob/master/leb128_test.go

var (
	uleb128Spec = map[uint64]string{
		0:                    "00000000",
		0x7f:                 "01111111",
		0x80:                 "00000001 10000000",
		0xff:                 "00000001 11111111",
		2333:                 "00010010 10011101",
		0xef17:               "00000011 11011110 10010111",
		624485:               "00100110 10001110 11100101",
		0xffff:               "00000011 11111111 11111111",
		18446744073709551615: "00000001 11111111 11111111 11111111 11111111 11111111 11111111 11111111 11111111 11111111",
	}
	sleb128Spec = map[int64]string{
		-9223372036854775808: "01111111 10000000 10000000 10000000 10000000 10000000 10000000 10000000 10000000 10000000",
		-624485:              "01011001 11110001 10011011",
		^0x40:                "01111111 10111111",
		^0x3f:                "01000000",
		-1:                   "01111111",
		0:                    "00000000",
		1:                    "00000001",
		0x3f:                 "00111111",
		0x40:                 "00000000 11000000",
		0xef17:               "00000011 11011110 10010111",
		9223372036854775807:  "00000000 11111111 11111111 11111111 11111111 11111111 11111111 11111111 11111111 11111111",
	}
	errorSpec = map[string]uint8{
		"11111111":          0,
		"10000000 10000000": 0,
		"01111111 10000000 10000000 10000000 10000000 10000000 10000000 10000000 10000000 10000000 11111111": 0,
	}
)

// bytesToBinary encodes b into a base 2 string in MSB to LSB order.
func bytesToBinary(b []byte) string {
	l := len(b)
	parts := make([]string, l)

	for i, v := range b {
		parts[l-i-1] = fmt.Sprintf("%08b", v)
	}

	return strings.Join(parts, " ")
}

// binaryToBytes decodes base 2 string in MSB to LSB order into bytes.
func binaryToBytes(s string) []byte {
	parts := strings.Split(s, " ")
	l := len(parts)
	b := make([]byte, l)

	for i, v := range parts {
		fmt.Sscanf(v, "%b", &b[l-i-1])
	}

	return b
}

func TestEncodeUleb128(t *testing.T) {
	count := 0
	for input, expected := range uleb128Spec {
		var b []byte
		if count%2 == 0 {
			b = AppendUleb128([]byte("magic"), input)[5:]
		} else {
			b = AppendUleb128(nil, input)
		}

		if actual := bytesToBinary(b); actual != expected {
			t.Fatalf("\nInput: %v\nExpected: %v\n     Got: %v\n", input, expected, actual)
		}

		count++
	}
}

func TestEncodeSleb128(t *testing.T) {
	count := 0
	for input, expected := range sleb128Spec {
		var b []byte
		if count%2 == 0 {
			b = AppendSleb128([]byte("equim"), input)[5:]
		} else {
			b = AppendSleb128(nil, input)
		}

		if actual := bytesToBinary(b); actual != expected {
			t.Fatalf("\nInput: %v\nExpected: %v\n     Got: %v\n", input, expected, actual)
		}

		count++
	}
}

func TestDecodeUleb128(t *testing.T) {
	count := 0
	for expected, binStr := range uleb128Spec {
		input := binaryToBytes(binStr)

		var actual uint64
		var n uint8
		if count%2 == 0 {
			actual, n = Uleb128ToUint64(append(input, []byte("magic")...))
		} else {
			actual, n = Uleb128ToUint64(input)
		}

		if expectedN := len(input); int(n) != expectedN {
			t.Fatalf("\nInput: %v\nExpected: n == %v\n     Got: n == %v\n", input, expectedN, n)
		}

		if actual != expected {
			t.Fatalf("\nInput: %v\nExpected: %v\n     Got: %v\n", input, expected, actual)
		}

		count++
	}

	for binStr, expected := range errorSpec {
		input := binaryToBytes(binStr)

		_, actual := Uleb128ToUint64(input)

		if actual != expected {
			t.Fatalf("\nInput: %v\nExpected: %v\n     Got: %v\n", input, expected, actual)
		}
	}
}

func TestDecodeSleb128(t *testing.T) {
	count := 0
	for expected, binStr := range sleb128Spec {
		input := binaryToBytes(binStr)

		var actual int64
		var n uint8
		if count%2 == 0 {
			actual, n = Sleb128ToInt64(append(input, []byte("magic")...))
		} else {
			actual, n = Sleb128ToInt64(input)
		}

		if expectedN := len(input); int(n) != expectedN {
			t.Fatalf("\nInput: %v\nExpected: n == %v\n     Got: n == %v\n", input, expectedN, n)
		}

		if actual != expected {
			t.Fatalf("\nInput: %v\nExpected: %v\n     Got: %v\n", input, expected, actual)
		}

		count++
	}

	for binStr, expected := range errorSpec {
		input := binaryToBytes(binStr)

		_, actual := Sleb128ToInt64(input)

		if actual != expected {
			t.Fatalf("\nInput: %v\nExpected: %v\n     Got: %v\n", input, expected, actual)
		}
	}
}
