// Copyright 2017-2018 The qitmeer developers

package util

import (
	"bytes"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"github.com/stretchr/testify/assert"
	"math/big"
	"testing"
)

func Test_ReadRand(t *testing.T) {
	assert.Equal(t, 32, len(ReadSizedRand(nil, 32)))
	assert.Equal(t, 64, len(ReadSizedRand(nil, 64)))
	println(hex.EncodeToString(ReadSizedRand(nil, 32)))
	println(base64.StdEncoding.EncodeToString(ReadSizedRand(nil, 32)))
	println(hex.EncodeToString(ReadSizedRand(nil, 40)))
	println(base64.StdEncoding.EncodeToString(ReadSizedRand(nil, 40)))
	println(hex.EncodeToString(ReadSizedRand(nil, 64)))
	println(base64.StdEncoding.EncodeToString(ReadSizedRand(nil, 64)))
}

func Test_Padding(t *testing.T) {
	out := PaddedBytes(32, new(big.Int).SetBytes([]byte{0, 1, 2, 3}))
	assert.Equal(t, 32, len(out))
	fmt.Printf("%v\n", out)
	out2 := PaddedAppend(16, []byte{1}, []byte{0, 1, 2, 3})
	fmt.Printf("%v\n", out2)
	assert.Equal(t, 17, len(out2))
	out3 := PaddedAppend(16, nil, []byte{0, 1, 2, 3})
	fmt.Printf("%v\n", out3)
}

var testInt = new(big.Int).SetBytes([]byte{0, 1, 2, 3})

func BenchmarkPaddedBigInt(b *testing.B) {
	for i := 0; i < b.N; i++ {
		PaddedBytes(32, testInt)
	}
}

func BenchmarkPaddedBigInt8(b *testing.B) {
	for i := 0; i < b.N; i++ {
		PaddedBytes(65, testInt)
	}
}

func TestMustDecodeHexPanic(t *testing.T) {
	defer func() {
		r := recover()
		assert.NotNil(t, r)
		errStr, ok := r.(string)
		assert.True(t, ok)
		assert.Equal(t, "invalid hex string in encoding/hex: invalid byte: U+0074 't', hex: test", errStr)
	}()
	MustDecodeHexString("test")
	assert.Fail(t, "should not go here")
}
func TestMustDecodeHex(t *testing.T) {
	b := MustDecodeHexString("123abc")
	assert.NotNil(t, b)
}

var sink []byte

func BenchmarkEncode(b *testing.B) {
	for _, size := range []int{256, 1024, 4096, 16384} {
		src := bytes.Repeat([]byte{2, 3, 5, 7, 9, 11, 13, 17}, size/8)
		sink = make([]byte, 2*size)

		b.Run(fmt.Sprintf("%v", size), func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				hex.Encode(sink, src)
			}
		})
	}
}
