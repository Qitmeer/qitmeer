// Copyright (c) 2017-2018 The qitmeer developers
package cuckoo

import (
	"encoding/hex"
	"fmt"
	"github.com/Qitmeer/qitmeer/common/hash"
	"log"
	"math/big"
	"os"
	"runtime"
	"testing"
)

const targetBits = 1

func TestCuckooMining(t *testing.T) {
	if os.Getenv("TEST_CUCKOO") == ""  {
		t.Skip("skipping the long test by default. use 'TEST_CUCKOO=true go test -v' to run the test.")
	}
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
		headerBytes := []byte(fmt.Sprintf("%s%d", header, nonce))
		siphashKey = hash.DoubleHashB(headerBytes)
		cycleNonces, isFound = c.PoW(siphashKey[:16])
		if !isFound {
			continue
		}
		if err := VerifyCuckaroo(siphashKey[:16], cycleNonces, Edgebits); err != nil {
			fmt.Println(err)
			continue
		}
		cycleNoncesHash := hash.DoubleHashB(Uint32ToBytes(cycleNonces))
		cycleNoncesHashInt.SetBytes(cycleNoncesHash[:])

		log.Println(fmt.Sprintf("Current Nonce:%d", nonce))
		log.Println(fmt.Sprintf("Found %d Cycles Nonces:", ProofSize), cycleNonces)
		fmt.Println("Found Hash ", hex.EncodeToString(cycleNoncesHash))
		break
	}
}
