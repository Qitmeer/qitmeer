// Copyright (c) 2017-2020 The qitmeer developers
// license that can be found in the LICENSE file.
// Reference resources of rust bitVector
package pow

import (
    "encoding/binary"
    "github.com/Qitmeer/qitmeer/common/hash"
    "github.com/Qitmeer/qitmeer/common/util"
    "github.com/Qitmeer/qitmeer/crypto/cuckoo"
    `math/big`
    "sort"
)

type Cuckoo struct {
    Pow
}

const (
    PROOF_DATA_EDGE_BITS_START = 0
    PROOF_DATA_EDGE_BITS_END = 1
    PROOF_DATA_CIRCLE_NONCE_END = 169
)

// set edge bits
func (this *Cuckoo) SetEdgeBits (edge_bits uint8) {
    copy(this.ProofData[PROOF_DATA_EDGE_BITS_START:PROOF_DATA_EDGE_BITS_END] , []byte{edge_bits})
}

// get edge bits
func (this *Cuckoo) GetEdgeBits () uint8 {
    return uint8(this.ProofData[PROOF_DATA_EDGE_BITS_START:PROOF_DATA_EDGE_BITS_END][0])
}

// set edge circles
func (this *Cuckoo) SetCircleEdges (edges []uint32) {
    for i:=0 ;i<len(edges);i++{
        b := make([]byte,4)
        binary.LittleEndian.PutUint32(b,edges[i])
        copy(this.ProofData[(i*4)+PROOF_DATA_EDGE_BITS_END:(i*4)+PROOF_DATA_EDGE_BITS_END+4],b)
    }
}

func (this *Cuckoo) GetCircleNonces () (nonces [cuckoo.ProofSize]uint32) {
    nonces = [cuckoo.ProofSize]uint32{}
    j := 0
    for i :=PROOF_DATA_EDGE_BITS_END;i<PROOF_DATA_CIRCLE_NONCE_END;i+=4{
        nonceBytes := this.ProofData[i:i+4]
        nonces[j] = binary.LittleEndian.Uint32(nonceBytes)
        j++
    }
    return
}

// set scale ,the diff scale of circle
func (this *Cuckoo) SetScale (scale uint32) {

}

//get scale ,the diff scale of circle
func (this *Cuckoo) GetScale () int64 {
    return 1856
}

//get cuckoo block hash bitarray with 42 circle nonces
//then blake2b
func (this *Cuckoo)GetBlockHash (data []byte) hash.Hash {
    circlNonces := [cuckoo.ProofSize]uint64{}
    nonces := this.GetCircleNonces()
    for i:=0;i<len(nonces);i++{
        circlNonces[i] = uint64(nonces[i])
    }
    return this.CuckooHash(circlNonces[:],int(this.GetEdgeBits()))
}

//calc cuckoo hash
func (this *Cuckoo)CuckooHash(nonces []uint64,nonce_bits int) hash.Hash {
    sort.Slice(nonces, func(i, j int) bool {
        return nonces[i] < nonces[j]
    })
    bitvec,_ := util.New(nonce_bits*cuckoo.ProofSize)
    for i:=41;i>=0;i--{
        n := i
        nonce := nonces[i]
        for bit:= 0;bit < nonce_bits;bit++{
            if nonce & (1 << uint(bit)) != 0 {
                bitvec.SetBitAt(n * nonce_bits + bit)
            }
        }
    }
    h := hash.HashH(bitvec.Bytes())
    util.ReverseBytes(h[:])
    return h
}


func (this *Cuckoo)Bytes() PowBytes {
    r := make(PowBytes,0)
    //write pow type 1 byte
    r = append(r,[]byte{byte(this.PowType)}...)
    //write nonce 4 bytes
    n := make([]byte,4)
    binary.LittleEndian.PutUint32(n,this.Nonce)
    r = append(r,n...)
    //write ProofData 169 bytes
    r = append(r,this.ProofData[:]...)
    return PowBytes(r)
}

// compare the target
// wether target match the target diff
func (this *Cuckoo) CompareDiff(newTarget *big.Int,target *big.Int) bool{
    if newTarget.Cmp(target) < 0{
        return false
    }
    return true
}