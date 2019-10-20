// Copyright (c) 2017-2020 The qitmeer developers
// license that can be found in the LICENSE file.
// Reference resources of rust bitVector
package pow

import (
	"encoding/binary"
	"github.com/Qitmeer/qitmeer/common/hash"
	"github.com/Qitmeer/qitmeer/crypto/cuckoo"
	"math/big"
)

type Cuckoo struct {
	Pow
}

const (
	PROOF_DATA_EDGE_BITS_START  = 0
	PROOF_DATA_EDGE_BITS_END    = 1
	PROOF_DATA_CIRCLE_NONCE_END = 169
)

// set edge bits
func (this *Cuckoo) SetEdgeBits(edge_bits uint8) {
	copy(this.ProofData[PROOF_DATA_EDGE_BITS_START:PROOF_DATA_EDGE_BITS_END], []byte{edge_bits})
}

// get edge bits
func (this *Cuckoo) GetEdgeBits() uint8 {
	return uint8(this.ProofData[PROOF_DATA_EDGE_BITS_START:PROOF_DATA_EDGE_BITS_END][0])
}

// set edge circles
func (this *Cuckoo) SetCircleEdges(edges []uint32) {
	for i := 0; i < len(edges); i++ {
		b := make([]byte, 4)
		binary.LittleEndian.PutUint32(b, edges[i])
		copy(this.ProofData[(i*4)+PROOF_DATA_EDGE_BITS_END:(i*4)+PROOF_DATA_EDGE_BITS_END+4], b)
	}
}

func (this *Cuckoo) GetCircleNonces() (nonces [cuckoo.ProofSize]uint32) {
	arr := ConvertBytesToUint32Array(this.ProofData[PROOF_DATA_EDGE_BITS_END:PROOF_DATA_CIRCLE_NONCE_END])
	copy(nonces[:cuckoo.ProofSize], arr[:cuckoo.ProofSize])
	return
}

func ConvertBytesToUint32Array(data []byte) []uint32 {
	nonces := make([]uint32, 0)
	j := 0
	l := len(data)
	for i := 0; i < l; i += 4 {
		nonceBytes := data[i : i+4]
		nonces = append(nonces, binary.LittleEndian.Uint32(nonceBytes))
		j++
	}
	return nonces
}

//get sip hash
//first header data 113 bytes hash
func (this *Cuckoo) GetSipHash(headerData []byte) hash.Hash {
	return hash.HashH(headerData[:len(headerData)-PROOF_DATA_CIRCLE_NONCE_END])
}

//cuckoo pow proof data
func (this *Cuckoo) Bytes() PowBytes {
	r := make(PowBytes, 0)

	//write nonce 4 bytes
	n := make([]byte, 4)
	binary.LittleEndian.PutUint32(n, this.Nonce)
	r = append(r, n...)

	//write pow type 1 byte
	r = append(r, []byte{byte(this.PowType)}...)

	//write ProofData 169 bytes
	r = append(r, this.ProofData[:]...)
	return PowBytes(r)
}

// compare the target
// wether target match the target diff
func (this *Cuckoo) CompareDiff(newTarget *big.Int, target *big.Int) bool {
	return newTarget.Cmp(target) >= 0
}
