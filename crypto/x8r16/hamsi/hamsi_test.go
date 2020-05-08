// Copyright (c) 2020 The qitmeer developers
// Use of this source code is governed by an ISC
// license that can be found in the LICENSE file.

package hamsi

import (
	"encoding/hex"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestSph_hamsi512_process(t *testing.T) {
	test_hamsi512(t, "", "5cd7436a91e27fc809d7015c3407540633dab391127113ce6ba360f0c1e35f404510834a551610d6e871e75651ea381a8ba628af1dcf2b2be13af2eb6247290f")
	test_hamsi512_hexInput(t, "cc", "7da1be62a813a8e24d200671cffb1d0be79d2bc176ff0b163b11eded2414ef66261ff52c745383442bc7f1884d5166f26f41d335fc2d2fdb2f93b24b8d079265")
}

func test_hamsi512(t *testing.T, input string, wantOutput string) {
	var in, out []byte
	in = []byte(input)
	out = make([]byte, 64)
	Sph_hamsi512_process(in[:], out[:], uint(len(in)))
	want, _ := hex.DecodeString(wantOutput)
	//fmt.Printf(" inTxt: %s\noutput: %x \n", in, out)
	assert.Equal(t, want, out)
}

func test_hamsi512_hexInput(t *testing.T, hexInput string, wantOutput string) {
	var in, out []byte
	in, _ = hex.DecodeString(hexInput)
	out = make([]byte, 64)
	Sph_hamsi512_process(in[:], out[:], uint(len(in)))
	want, _ := hex.DecodeString(wantOutput)
	//fmt.Printf(" inHex: %x\noutput: %x \n", in, out)
	assert.Equal(t, want, out)
}
