// Copyright (c) 2017-2018 The qitmeer developers

package leb128

import (
	"math/big"
	"reflect"
	"testing"

	"encoding/hex"
	"fmt"
	"github.com/stretchr/testify/assert"
)

func TestFromUInt64(t *testing.T) {
	tests := []struct {
		n    uint64
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
		{^uint64(0), []byte{0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0x01}},
	}
	for _, tt := range tests {
		t.Run("TestFromUInt64", func(t *testing.T) {
			if got := FromUInt64(tt.n); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("FromUInt64(%v) = %0x, want %0x", tt.n, got, tt.want)
			}
		})
	}
}

func TestLEB128UInt64RoundTrip(t *testing.T) {
	assert := assert.New(t)

	for pow := uint(0); pow < 64; pow++ {
		var x uint64 = 1 << pow
		for v := uint64(x - 10); v <= x+10; v++ {
			got := ToUInt64(FromUInt64(v))
			assert.Equal(v, got)
		}
	}
}

func TestFromBigInt(t *testing.T) {
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
		{64, []byte{0xC0, 0x00}},
		{65, []byte{0xC1, 0x00}},
		{100, []byte{0xE4, 0x00}},
		{127, []byte{0xFF, 0x00}},
		{128, []byte{0x80, 0x01}},
		{129, []byte{0x81, 0x01}},
		{2141192192, []byte{0x80, 0x80, 0x80, 0xFD, 0x07}},

		{-1, []byte{0x7F}},
		{-2, []byte{0x7E}},
		{-3, []byte{0x7D}},
		{-4, []byte{0x7C}},
		{-5, []byte{0x7B}},
		{-63, []byte{0x41}},
		{-64, []byte{0x40}},
		{-65, []byte{0xBF, 0x7F}},
		{-100, []byte{0x9C, 0x7F}},
		{-127, []byte{0x81, 0x7F}},
		{-128, []byte{0x80, 0x7F}},
		{-129, []byte{0xFF, 0x7E}},
		{-624485, []byte{0x9B, 0xF1, 0x59}},
	}

	for _, tt := range tests {
		t.Run("TestFromBigInt", func(t *testing.T) {
			if gotOut := FromBigInt(big.NewInt(tt.n)); !reflect.DeepEqual(gotOut, tt.want) {
				t.Errorf("FromBigInt(%v) = %0x, want %0x", tt.n, gotOut, tt.want)
			}
		})
	}
}

func TestFromBigIntIsNotDestructive(t *testing.T) {
	assert := assert.New(t)

	v := big.NewInt(-10)
	vOrig := big.NewInt(0).Set(v)
	_ = FromBigInt(v)
	assert.True(vOrig.Cmp(v) == 0)

	v = big.NewInt(10)
	vOrig = big.NewInt(0).Set(v)
	_ = FromBigInt(v)
	assert.True(vOrig.Cmp(v) == 0)
}

func TestLEB128BigIntRoundTrip(t *testing.T) {
	assert := assert.New(t)

	// Strategy: for a range of 20 values on either side of the powers of
	// two up to 2**128, check the round trip for each value and for its
	// negative.
	for pow := uint(0); pow < 128; pow++ {
		x := big.NewInt(1)
		x.Lsh(x, pow)
		i := int64(-10)
		for {
			posV := big.NewInt(0).Set(x)
			posV.Add(posV, big.NewInt(i))
			got := ToBigInt(FromBigInt(posV))
			assert.True(posV.Cmp(got) == 0, "expected %s got %s", posV, got)

			negV := big.NewInt(0).Set(posV)
			negV.Neg(negV)
			got = ToBigInt(FromBigInt(negV))
			assert.True(negV.Cmp(got) == 0, "expected %s got %s", negV, got)

			i++
			if i == 10 {
				break
			}
		}
	}
}

// It's the test case of the overflow for FromUInt64
// TODO fix the overflow
func TestOverflowFromUInt64(t *testing.T) {

	assert := assert.New(t)

	// The normal case
	// The longest LEB128 encoded sequence is 10-bytes long (9 0xff's and 1 0x7f)
	input := []byte{0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0x01} // 10 bytes
	got := ToUInt64(input)
	assert.Equal(^uint64(0), got) // the uint64 max -> 8 0xff's
	output := FromUInt64(^uint64(0))
	assert.Equal(output, input[:])
	fmt.Printf("%x = %x \n", output, input)

	// case 1 : incorrect result
	input = []byte{0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF,
		0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF,
		0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0x01}
	assert.True(len(input) > 20, "len=%d", len(input))
	got = ToUInt64(input)
	assert.Equal(^uint64(0), got) // not correct, wrong input
	output = FromUInt64(^uint64(0))
	assert.NotEqual(output, input[:])
	fmt.Printf("%x != %x \n", FromUInt64(^uint64(0)), input) //max

	// case 2 : panic when bad input
	bad, _ := hex.DecodeString("ffffffffffffffffffff")
	bad = append([]byte(nil), bad...)
	assert.Panics(func() { ToUInt64(bad) }, "should overflow with bad input :%s", bad)
}
