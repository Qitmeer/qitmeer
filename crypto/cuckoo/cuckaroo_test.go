package cuckoo

import (
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"github.com/Qitmeer/qitmeer/common/hash"
	"log"
	"math/big"
	"runtime"
	"testing"
)

const targetBits = 1

func TestCuckooMining(t *testing.T) {

	// Use all processor cores.
	runtime.GOMAXPROCS(runtime.NumCPU())
	var targetDifficulty = big.NewInt(2)
	targetDifficulty.Lsh(targetDifficulty, uint(255-targetBits))
	fmt.Printf("Target Diff:%064x\n", targetDifficulty.Bytes())
	c := NewCuckoo()

	var cycleNonces []uint32
	var isFound bool
	var nonce int64
	var cycleNoncesHashInt big.Int
	var siphashKey []byte
	header := "helloworld"

	for nonce = 0; ; nonce++ {
		headerBytes := []byte(fmt.Sprintf("%s%d",header,nonce))
		siphashKey = hash.DoubleHashB(headerBytes)
		cycleNonces, isFound = c.PoW(siphashKey[:16])
		if !isFound {
			continue
		}
		if err := Verify(siphashKey[:16], cycleNonces); err != nil {
			continue
		}
		cycleNoncesHash := hash.DoubleHashB(Uint32ToBytes(cycleNonces))
		cycleNoncesHashInt.SetBytes(cycleNoncesHash[:])

		// The block is solved when the new block hash is less
		// than the target difficulty.  Yay!
		if cycleNoncesHashInt.Cmp(targetDifficulty) <= 0 {
			log.Println(fmt.Sprintf("Current Nonce:%d",nonce))
			log.Println(fmt.Sprintf("Found %d Cycles Nonces:",ProofSize),cycleNonces)
			fmt.Println("【Found Hash】",hex.EncodeToString(cycleNoncesHash))
			break
		}
	}
}

// HashToBig converts a hash.Hash into a big.Int that can be used to
// perform math comparisons.
func Uint32ToBytes(v []uint32) []byte {
	var buf = make([]byte, 4*len(v))
	for i, x := range v {
		binary.LittleEndian.PutUint32(buf[4*i:], x)
	}
	return buf
}
