package siphash

import (
	"encoding/binary"
)

type SipHash struct {
	k0, k1 uint64    // two parts of key
	V      [4]uint64 // v is the current internal state.
}

/*
siphashKey[:] = [196 107 38 219 80 75 209 213 243 49 219 252 101 35 20 105]
siphashKey[8:] = [243 49 219 252 101 35 20 105]
k0 = 15407178610857372612
k1 = 7571715794457539059
before s.V = [0 0 0 0]
after s.V = [12015082867820662449 971459224712208030 13377991302290020773 2121554997101417600]
*/
func Newsip(siphashKey []byte) *SipHash {
	s := &SipHash{
		k0: binary.LittleEndian.Uint64(siphashKey[:]),  // 17624113405423192833
		k1: binary.LittleEndian.Uint64(siphashKey[8:]), // 16633022555945065022
	}

	//fmt.Printf("before s.V = %v\n",s.V)
	s.V[0] = s.k0 ^ 0x736f6d6570736575 //9798144632717078132, 0x87fa00a173006e74
	s.V[1] = s.k1 ^ 0x646f72616e646f6d //9420175285032990035, 0x82bb2f82f24e3553
	s.V[2] = s.k0 ^ 0x6c7967656e657261 //11019194076704962912,0x98ec0aa16d167960
	s.V[3] = s.k1 ^ 0x7465646279746573 //10570293030476988237,0x92b13981e55e3f4d

	return s
}

func Siphash(k0, k1, b uint64) uint64 {
	// Initialization.
	var v [4]uint64
	v[0] = k0 ^ 0x736f6d6570736575
	v[1] = k1 ^ 0x646f72616e646f6d
	v[2] = k0 ^ 0x6c7967656e657261
	v[3] = k1 ^ 0x7465646279746573
	return SiphashPRF(&v, b)
}

func SiphashPRF(v *[4]uint64, b uint64) uint64 {
	v0 := v[0]
	v1 := v[1]
	v2 := v[2]
	v3 := v[3]
	// Initialization.
	// Compression.
	v3 ^= b

	// Round 1.
	v0 += v1
	v1 = v1<<13 | v1>>(64-13)
	v1 ^= v0
	v0 = v0<<32 | v0>>(64-32)

	v2 += v3
	v3 = v3<<16 | v3>>(64-16)
	v3 ^= v2

	v0 += v3
	v3 = v3<<21 | v3>>(64-21)
	v3 ^= v0

	v2 += v1
	v1 = v1<<17 | v1>>(64-17)
	v1 ^= v2
	v2 = v2<<32 | v2>>(64-32)

	// Round 2.
	v0 += v1
	v1 = v1<<13 | v1>>(64-13)
	v1 ^= v0
	v0 = v0<<32 | v0>>(64-32)

	v2 += v3
	v3 = v3<<16 | v3>>(64-16)
	v3 ^= v2

	v0 += v3
	v3 = v3<<21 | v3>>(64-21)
	v3 ^= v0

	v2 += v1
	v1 = v1<<17 | v1>>(64-17)
	v1 ^= v2
	v2 = v2<<32 | v2>>(64-32)

	v0 ^= b

	// Finalization.
	v2 ^= 0xff

	// Round 1.
	v0 += v1
	v1 = v1<<13 | v1>>(64-13)
	v1 ^= v0
	v0 = v0<<32 | v0>>(64-32)

	v2 += v3
	v3 = v3<<16 | v3>>(64-16)
	v3 ^= v2

	v0 += v3
	v3 = v3<<21 | v3>>(64-21)
	v3 ^= v0

	v2 += v1
	v1 = v1<<17 | v1>>(64-17)
	v1 ^= v2
	v2 = v2<<32 | v2>>(64-32)

	// Round 2.
	v0 += v1
	v1 = v1<<13 | v1>>(64-13)
	v1 ^= v0
	v0 = v0<<32 | v0>>(64-32)

	v2 += v3
	v3 = v3<<16 | v3>>(64-16)
	v3 ^= v2

	v0 += v3
	v3 = v3<<21 | v3>>(64-21)
	v3 ^= v0

	v2 += v1
	v1 = v1<<17 | v1>>(64-17)
	v1 ^= v2
	v2 = v2<<32 | v2>>(64-32)

	// Round 3.
	v0 += v1
	v1 = v1<<13 | v1>>(64-13)
	v1 ^= v0
	v0 = v0<<32 | v0>>(64-32)

	v2 += v3
	v3 = v3<<16 | v3>>(64-16)
	v3 ^= v2

	v0 += v3
	v3 = v3<<21 | v3>>(64-21)
	v3 ^= v0

	v2 += v1
	v1 = v1<<17 | v1>>(64-17)
	v1 ^= v2
	v2 = v2<<32 | v2>>(64-32)

	// Round 4.
	v0 += v1
	v1 = v1<<13 | v1>>(64-13)
	v1 ^= v0
	v0 = v0<<32 | v0>>(64-32)

	v2 += v3
	v3 = v3<<16 | v3>>(64-16)
	v3 ^= v2

	v0 += v3
	v3 = v3<<21 | v3>>(64-21)
	v3 ^= v0

	v2 += v1
	v1 = v1<<17 | v1>>(64-17)
	v1 ^= v2
	v2 = v2<<32 | v2>>(64-32)

	return v0 ^ v1 ^ v2 ^ v3
}

const sipHashBlockBits uint64 = 6
const sipHashBlockSize uint64 = 1 << sipHashBlockBits
const sipHashBlockMask uint64 = sipHashBlockSize - 1

// SipHashBlock builds a block of siphash values by repeatedly hashing from the
// nonce truncated to its closest block start, up to the end of the block.
// Returns the resulting hash at the nonce's position.
func SipHashBlock(v [4]uint64, nonce uint64, rotE uint8, xorAll bool) uint64 {
	// beginning of the block of hashes
	nonce0 := nonce & ^sipHashBlockMask
	nonceI := nonce & sipHashBlockMask
	nonceHash := make([]uint64, sipHashBlockSize)
	// repeated hashing over the whole block
	s := new(sipHash24)
	siphash := s.new(v)
	var i uint64
	for i = 0; i < sipHashBlockSize; i++ {
		siphash.hash(nonce0+i, rotE)
		nonceHash[i] = siphash.digest()
	}
	// xor the hash at nonce_i < SIPHASH_BLOCK_MASK with some or all later hashes to force hashing the whole block
	var xor uint64 = nonceHash[nonceI]
	var xorFrom uint64
	if xorAll || nonceI == sipHashBlockMask {
		xorFrom = nonceI + 1
	} else {
		xorFrom = sipHashBlockMask
	}

	for i := xorFrom; i < sipHashBlockSize; i++ {
		xor ^= nonceHash[i]
	}
	return xor
}

type sipHash24 struct {
	v0, v1, v2, v3 uint64
}

func (s *sipHash24) new(v [4]uint64) sipHash24 {
	return sipHash24{v[0], v[1], v[2], v[3]}
}

// One siphash24 hashing, consisting of 2 and then 4 rounds
func (s *sipHash24) hash(nonce uint64, rotE uint8) {
	s.v3 ^= nonce
	s.round(rotE)
	s.round(rotE)

	s.v0 ^= nonce
	s.v2 ^= 0xff

	for i := 0; i < 4; i++ {
		s.round(rotE)
	}
}

// Resulting hash digest
func (s *sipHash24) digest() uint64 {
	return (s.v0 ^ s.v1) ^ (s.v2 ^ s.v3)
}

func (s *sipHash24) round(rotE uint8) {
	s.v0 = s.v0 + s.v1
	s.v2 = s.v2 + s.v3
	s.v1 = rotl(s.v1, 13)
	s.v3 = rotl(s.v3, 16)
	s.v1 ^= s.v0
	s.v3 ^= s.v2
	s.v0 = rotl(s.v0, 32)
	s.v2 = s.v2 + s.v1
	s.v0 = s.v0 + s.v3
	s.v1 = rotl(s.v1, 17)
	s.v3 = rotl(s.v3, rotE)
	s.v1 ^= s.v2
	s.v3 ^= s.v0
	s.v2 = rotl(s.v2, 32)
}

func rotl(val uint64, shift uint8) uint64 {
	num := (val << shift) | (val >> (64 - shift))
	return num
}
