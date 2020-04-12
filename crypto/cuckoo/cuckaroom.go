package cuckoo

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

import (
	"encoding/binary"
	"errors"
	"github.com/Qitmeer/qitmeer/crypto/cuckoo/siphash"
)

func SipHashKey(sipkey []byte) [4]uint64 {
	var s [4]uint64
	s[0] = binary.LittleEndian.Uint64(sipkey[0:8])
	s[1] = binary.LittleEndian.Uint64(sipkey[8:16])
	s[2] = binary.LittleEndian.Uint64(sipkey[16:24])
	s[3] = binary.LittleEndian.Uint64(sipkey[24:32])
	return s
}

// Verify verifies the Cuckaroom context.
func VerifyCuckaroom(sipHashKeys [4]uint64, nonces []uint32, edgeBits uint) error {
	nedge := (1 << edgeBits) //number of edges：
	edgemask := uint64(nedge - 1)
	if len(nonces) != ProofSize {
		return errors.New("length of nonce is not correct")
	}
	from := make([]uint32, len(nonces))
	to := make([]uint32, len(nonces))
	var xorFrom uint32 = 0
	var xorTo uint32 = 0

	nodemask := edgemask >> 1

	for n := 0; n < ProofSize; n++ {
		if uint64(nonces[n]) > edgemask {
			return errors.New("edge too big")
		}
		if n > 0 && nonces[n] <= nonces[n-1] {
			return errors.New("edges not ascending")
		}
		edge := siphash.SipHashBlock(sipHashKeys, uint64(nonces[n]), 21, true)
		from[n] = uint32(edge & nodemask)
		xorFrom ^= from[n]
		to[n] = uint32((edge >> 32) & nodemask)
		xorTo ^= to[n]
	}
	if xorFrom != xorTo {
		return errors.New("endpoints don't match up")
	}
	visited := make([]bool, ProofSize)
	n := 0
	i := 0
	for {
		// follow cycle
		if visited[i] {
			return errors.New("branch in cycle")
		}
		visited[i] = true
		nexti := 0
		for from[nexti] != to[i] {
			nexti++
			if nexti == ProofSize {
				return errors.New("cycle dead ends")
			}
		}
		i = nexti
		n++
		if i == 0 {
			// must cycle back to start or find branch
			break
		}
	}
	if n == ProofSize {
		return nil
	}
	return errors.New("cycle too short")
}
