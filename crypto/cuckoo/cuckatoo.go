// Copyright (c) 2017-2018 The qitmeer developers
package cuckoo

import (
	"errors"
	"fmt"
	"github.com/Qitmeer/qitmeer/crypto/cuckoo/siphash"
	"log"
)

func Sipnode(h *siphash.SipHash, edge, uorv uint64, shift bool, edgemask uint64) uint64 {
	sipHash := siphash.SiphashPRF(&h.V, uint64(2*edge)+uorv)
	masked := sipHash & edgemask
	if shift {
		masked <<= 1
		masked |= uint64(uorv)
	}
	return masked
}

//Verify cuckoo nonces.
func VerifyCuckatoo(sipkey []byte, nonces []uint32, edgeBits uint) error {
	nedge := (1 << edgeBits)             //number of edges：
	nnode := 2 * nedge                   //
	edgemask := nedge - 1                // used to mask siphash output
	easiness := uint32(nnode * 50 / 100) //
	sip := siphash.Newsip(sipkey)
	var uvs [2 * ProofSize]uint64
	var xor0, xor1 uint64
	xor0 = (ProofSize / 2) & 1
	xor1 = xor0
	if len(nonces) != ProofSize {
		return errors.New("length of nonce is not correct")
	}

	if nonces[ProofSize-1] > easiness {
		return errors.New("nonce is too big")
	}

	for n := 0; n < ProofSize; n++ {
		if n > 0 && nonces[n] <= nonces[n-1] {
			fmt.Printf("n=%d\n", n)
			return errors.New("nonces are not in order")
		}
		uvs[2*n] = Sipnode(sip, uint64(nonces[n]), uint64(0), false, uint64(edgemask))
		uvs[2*n+1] = Sipnode(sip, uint64(nonces[n]), uint64(1), false, uint64(edgemask))
		xor0 ^= uvs[2*n]
		xor1 ^= uvs[2*n+1]
	}
	if xor0 != 0 {
		log.Println(xor0)
		return errors.New("U endpoinsts don't match")
	}
	if xor1 != 0 {
		log.Println(xor1)
		return errors.New("V endpoinsts don't match")
	}
	n := 0
	for i := 0; ; {
		another := i
		k := another
		for {
			k = (k + 2) % (2 * ProofSize)
			if k == i {
				break
			}
			if (uvs[k] >> 1) == (uvs[i] >> 1) {
				if another != i {
					return errors.New("there are branches in nonce")
				}
				another = k
			}
		}
		if another == i || uvs[another] == uvs[i] {
			return errors.New("dead end in nonce")
		}
		i = another ^ 1
		n++
		if i == 0 {
			break
		}
	}
	if n != ProofSize {
		return fmt.Errorf("%d cycle is too short", n)
	}
	return nil
}
