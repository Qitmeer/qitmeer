// Copyright (c) 2017-2018 The qitmeer developers
package cuckoo

import (
	"github.com/Qitmeer/qitmeer/crypto/cuckoo/siphash"
	"github.com/pkg/errors"
)

//Verify cuckaroo nonces.
func VerifyCuckaroo(sipkey []byte, nonces []uint32,edgeBits uint) error {
	nedge     := (1 << edgeBits)    //number of edges：
	nnode     := 2 * nedge        //
	easiness  := uint32(nnode * 50 / 100) //
	edgemask  := uint64(nedge - 1)
	sip := siphash.Newsip(sipkey)
	var uvs [2 * ProofSize]uint32
	var xor0, xor1 uint32

	if len(nonces) != ProofSize {
		return errors.New("length of nonce is not correct")
	}

	if nonces[ProofSize-1] > easiness {
		return errors.New("nonce is too big")
	}

	for n := 0; n < ProofSize; n++ {
		if n > 0 && nonces[n] <= nonces[n-1] {
			return errors.New("nonces are not in order")
		}
		u00 := siphash.SiphashPRF(&sip.V, uint64(nonces[n]<<1))
		v00 := siphash.SiphashPRF(&sip.V, (uint64(nonces[n])<<1)|1)
		u0 := uint32(u00&edgemask) << 1
		xor0 ^= u0
		uvs[2*n] = u0
		v0 := (uint32(v00&edgemask) << 1) | 1
		xor1 ^= v0
		uvs[2*n+1] = v0
	}
	if xor0 != 0 {
		return errors.New("U endpoinsts don't match")
	}
	if xor1 != 0 {
		return errors.New("V endpoinsts don't match")
	}

	n := 0
	for i := 0; ; {
		another := i
		for k := (i + 2) % (2 * ProofSize); k != i; k = (k + 2) % (2 * ProofSize) {
			if uvs[k] == uvs[i] {
				if another != i {
					return errors.New("there are branches in nonce")
				}
				another = k
			}
		}
		if another == i {
			return errors.New("dead end in nonce")
		}
		i = another ^ 1
		n++
		if i == 0 {
			break
		}
	}
	if n != ProofSize {
		return errors.New("cycle is too short")
	}
	return nil
}