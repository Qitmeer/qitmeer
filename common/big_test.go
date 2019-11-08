// Copyright 2017-2018 The qitmeer developers

package common

import (
	"bytes"
	"encoding/hex"
	"github.com/stretchr/testify/assert"
	"math/big"
	"testing"
)

func TestBigPow(t *testing.T) {
	a256 := new(big.Int).Lsh(big.NewInt(1), 256)
	b256 := BigPow(2, 256)
	assert.Equal(t, a256, b256)
	assert.Equal(t, "115792089237316195423570985008687907853269984665640564039457584007913129639936", a256.String())

	c256m1, _ := new(big.Int).SetString("ffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff", 16)
	assert.Equal(t, tt256m1, c256m1)
}

// bad 1281 ns/op
func BenchmarkBigPow(b *testing.B) {
	for i := 0; i < b.N; i++ {
		BigPow(2, 255)
	}
}

// best 334 ns/op
func BenchmarkBigPow2(b *testing.B) {
	for i := 0; i < b.N; i++ {
		new(big.Int).Lsh(big.NewInt(1), 256)
	}
}

// worst 1860 ns/op
func BenchmarkBigPow3(b *testing.B) {
	for i := 0; i < b.N; i++ {
		new(big.Int).SetString("ffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff", 16)
	}
}

func TestReadBits(t *testing.T) {
	check := func(input string) {
		want, _ := hex.DecodeString(input)
		int, _ := new(big.Int).SetString(input, 16)
		buf := make([]byte, len(want))
		ReadBits(int, buf)
		if !bytes.Equal(buf, want) {
			t.Errorf("have: %x\nwant: %x", buf, want)
		}
	}
	check("000000000000000000000000000000000000000000000000000000FEFCF3F8F0")
	check("0000000000012345000000000000000000000000000000000000FEFCF3F8F0")
	check("18F8F8F1000111000110011100222004330052300000000000000000FEFCF3F8F0")
}
