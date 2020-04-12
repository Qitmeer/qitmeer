package cuckoo

import (
	"encoding/binary"
	"fmt"
	"github.com/magiconair/properties/assert"
	"golang.org/x/crypto/blake2b"
	"testing"
)

var SipHashKeys = map[int][4]uint64{
	19: {
		0xdb7896f799c76dab,
		0x352e8bf25df7a723,
		0xf0aa29cbb1150ea6,
		0x3206c2759f41cbd5,
	},
	29: {
		0xe4b4a751f2eac47d,
		0x3115d47edfb69267,
		0x87de84146d9d609e,
		0x7deb20eab6d976a1,
	},
}

// Copyright 2020 BlockCypher
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
var v1_19SolCuckaroom = []uint32{
	0x0413c, 0x05121, 0x0546e, 0x1293a, 0x1dd27, 0x1e13e, 0x1e1d2, 0x22870, 0x24642, 0x24833,
	0x29190, 0x2a732, 0x2ccf6, 0x302cf, 0x32d9a, 0x33700, 0x33a20, 0x351d9, 0x3554b, 0x35a70,
	0x376c1, 0x398c6, 0x3f404, 0x3ff0c, 0x48b26, 0x49a03, 0x4c555, 0x4dcda, 0x4dfcd, 0x4fbb6,
	0x50275, 0x584a8, 0x5da0d, 0x5dbf1, 0x6038f, 0x66540, 0x72bbd, 0x77323, 0x77424, 0x77a14,
	0x77dc9, 0x7d9dc,
}

var v2_29SolCuckaroom = []uint32{
	0x04acd28, 0x29ccf71, 0x2a5572b, 0x2f31c2c, 0x2f60c37, 0x317fe1d, 0x32f6d4c, 0x3f51227,
	0x45ee1dc, 0x535eeb8, 0x5e135d5, 0x6184e3d, 0x6b1b8e0, 0x6f857a9, 0x8916a0f, 0x9beb5f8,
	0xa3c8dc9, 0xa886d94, 0xaab6a57, 0xd6df8f8, 0xe4d630f, 0xe6ae422, 0xea2d658, 0xf7f369b,
	0x10c465d8, 0x1130471e, 0x12049efb, 0x12f43bc5, 0x15b493a6, 0x16899354, 0x1915dfca,
	0x195c3dac, 0x19b09ab6, 0x1a1a8ed7, 0x1bba748f, 0x1bdbf777, 0x1c806542, 0x1d201b53,
	0x1d9e6af7, 0x1e99885e, 0x1f255834, 0x1f9c383b,
}

func TestCuckaroom19(t *testing.T) {
	err := VerifyCuckaroom(SipHashKeys[19], v1_19SolCuckaroom, 19)
	assert.Equal(t, err, nil)
}

func TestCuckaroom29(t *testing.T) {
	err := VerifyCuckaroom(SipHashKeys[29], v2_29SolCuckaroom, 29)
	assert.Equal(t, err, nil)
}

func TestPow19(t *testing.T) {
	header := make([]byte, 80)
	b := []byte("123")
	copy(header[:len([]byte(b))], b[:])
	nonce := uint32(103)
	nonceBytes := make([]byte, 4)
	binary.LittleEndian.PutUint32(nonceBytes, nonce)
	copy(header[76:], nonceBytes[:])
	h := blake2b.Sum256(header)
	sipHashKeys := SipHashKey(h[:])
	for k := 0; k < 4; k++ {
		fmt.Printf("0x%x\n", sipHashKeys[k])
	}
	nonces := []uint32{0x2e7, 0x93b5, 0x1112b, 0x11e3c, 0x12fe0, 0x14d25, 0x16fd1, 0x171d2, 0x19156, 0x1ff58,
		0x2565b, 0x2729e, 0x29552, 0x2de76, 0x34181, 0x349d9, 0x36e7f, 0x3cc36, 0x3f213, 0x405cd, 0x41f27, 0x49a18,
		0x4bdb7, 0x4d7f9, 0x4f116, 0x51761, 0x54ef8, 0x560f4, 0x5a7fa, 0x5efbc, 0x65629, 0x667bc, 0x6819d, 0x6d117,
		0x6d9a8, 0x6fa59, 0x723b7, 0x73c8c, 0x7a478, 0x7a592, 0x7b869, 0x7d030}
	err := VerifyCuckaroom(sipHashKeys, nonces, 19)
	assert.Equal(t, err, nil)
}
