// Copyright (c) 2019 Aidos Developer

// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:

// The above copyright notice and this permission notice shall be included in
// all copies or substantial portions of the Software.

// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
// THE SOFTWARE.

package siphash

import (
	"math/rand"
	"testing"
)

func TestSiphash(t *testing.T) {
	var k0 uint64 = 0x0011223344556677
	var k1 uint64 = 0x8899aabbccddeeff
	var b0 uint64 = 0x7766554433221100
	var b1 uint64 = 0xffeeddccbbaa9988

	var r0 uint64 = 12289717139560654282
	var r1 uint64 = 9875031879028705471

	h0 := Siphash(k0, k1, b0)
	h1 := Siphash(k0, k1, b1)
	if h0 != r0 || h1 != r1 {
		t.Error("incorrect resut0\n")
	}
	var v [4]uint64
	v[0] = k0 ^ 0x736f6d6570736575
	v[1] = k1 ^ 0x646f72616e646f6d
	v[2] = k0 ^ 0x6c7967656e657261
	v[3] = k1 ^ 0x7465646279746573
	h0 = SiphashPRF(&v, b0)
	h1 = SiphashPRF(&v, b1)
	if h0 != r0 || h1 != r1 {
		t.Error("incorrect resut\n")
	}
}

func TestSiphash16(t *testing.T) {
	var k0 uint64 = 0x0011223344556677
	var k1 uint64 = 0x8899aabbccddeeff
	var v [4]uint64
	v[0] = k0 ^ 0x736f6d6570736575
	v[1] = k1 ^ 0x646f72616e646f6d
	v[2] = k0 ^ 0x6c7967656e657261
	v[3] = k1 ^ 0x7465646279746573
	var nonce [8192]uint64
	for i := range nonce {
		nonce[i] = uint64(rand.Int63())
	}
	var uorv uint64 = 1
	res := make([]uint64, 8192)
	for i := range res {
		b := (nonce[i] << 1) | uorv
		res[i] = Siphash(k0, k1, b)
	}
	var ts [8192]uint64
	SiphashPRF8192(&v, &nonce, uorv, &ts)
	for i := range ts {
		if ts[i] != res[i] {
			t.Error("invalid siphash16 at", i)
		}
	}
}

func TestSiphash16Seq(t *testing.T) {
	var k0 uint64 = 0x0011223344556677
	var k1 uint64 = 0x8899aabbccddeeff
	var v [4]uint64
	v[0] = k0 ^ 0x736f6d6570736575
	v[1] = k1 ^ 0x646f72616e646f6d
	v[2] = k0 ^ 0x6c7967656e657261
	v[3] = k1 ^ 0x7465646279746573
	nonce := uint64(rand.Int63())
	var uorv uint64 = 1
	res := make([]uint64, 8192)
	for i := range res {
		b := ((nonce + uint64(i)) << 1) | uorv
		res[i] = Siphash(k0, k1, b)
	}
	var ts [8192]uint64
	SiphashPRF8192Seq(&v, nonce, uorv, &ts)
	for i := range ts {
		if ts[i] != res[i] {
			t.Error("invalid siphash16 at", i)
		}
	}
}

func BenchmarkSiphash(b *testing.B) {
	var k0 uint64 = 0x0011223344556677
	var k1 uint64 = 0x8899aabbccddeeff
	var b0 uint64 = 0x7766554433221100

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = Siphash(k0, k1, b0)
	}
}

func BenchmarkSiphash16(b *testing.B) {
	var k0 uint64 = 0x0011223344556677
	var k1 uint64 = 0x8899aabbccddeeff
	var v [4]uint64
	v[0] = k0 ^ 0x736f6d6570736575
	v[1] = k1 ^ 0x646f72616e646f6d
	v[2] = k0 ^ 0x6c7967656e657261
	v[3] = k1 ^ 0x7465646279746573
	var nonce [8192]uint64
	for i := range nonce {
		nonce[i] = uint64(rand.Int63())
	}
	var uorv uint64 = 1
	b.ResetTimer()
	var ts [8192]uint64
	for i := 0; i < b.N; i++ {
		SiphashPRF8192(&v, &nonce, uorv, &ts)
	}
}
