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
