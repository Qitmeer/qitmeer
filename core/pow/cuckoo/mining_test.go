package cuckoo

import (
	"encoding/binary"
	"fmt"
	"golang.org/x/crypto/blake2b"
	"math/big"
	"runtime"
	"strconv"
	"testing"
	"time"
)

const targetBits = 1

type Hash [32]byte

var txRoot = ""

func TestCuckooMining(t *testing.T) {

	// Use all processor cores.
	runtime.GOMAXPROCS(runtime.NumCPU())

	var targetDifficulty = big.NewInt(2)
	targetDifficulty.Lsh(targetDifficulty, uint(256-targetBits))
	fmt.Printf("%064x\n", targetDifficulty.Bytes())
	fmt.Printf("%0x\n", targetDifficulty.Bytes())

	c := NewCuckoo()

	var cycleNonces []uint32
	var cycleNoncesArr [20]uint32
	var isFound bool
	var nonce int64
	var cycleNoncesHashInt big.Int
	var siphashKey []byte

	for {
		for nonce = 0; ; nonce++ {
			fmt.Println("header nonce = ", nonce)
			siphashKey = getBlockHeaderHash(nonce) // len(siphashKey) = 16
			cycleNonces, isFound = c.PoW(siphashKey)
			for i, v := range cycleNonces {
				cycleNoncesArr[i] = v
			}

			if !isFound {
				continue
			}

			if err := Verify(siphashKey, cycleNoncesArr); err != nil {
				continue
			}

			for _, v := range cycleNonces {
				fmt.Printf("%#x,", v)
			}
			println()
			break
		}

		cycleNoncesHash := DoubleHashH(Uint32ToBytes(cycleNonces))
		cycleNoncesHashInt.SetBytes(cycleNoncesHash[:])

		// The block is solved when the new block hash is less
		// than the target difficulty.  Yay!
		if cycleNoncesHashInt.Cmp(targetDifficulty) == -1 {
			fmt.Println("success!")
			break
		} else {
			txRoot = strconv.FormatInt(nonce, 10)
			println("txRoot=", txRoot)
		}

	}
}

// DoubleHashH calculates hash(hash(b)) and returns the resulting bytes as a
// Hash.
func DoubleHashH(b []byte) Hash {
	first := blake2b.Sum256(b)
	return Hash(blake2b.Sum256(first[:]))
}

func getBlockHeaderHash(nonce int64) []byte {
	timestamp := strconv.FormatInt(time.Now().Unix(), 10)
	blockHeader := "version+parentRoot+stateRoot+height" + txRoot + timestamp + strconv.FormatInt(nonce, 10)
	doubleHashB := DoubleHashH([]byte(blockHeader))
	fmt.Printf("%x\n", doubleHashB)
	b := doubleHashB[:16]
	//fmt.Printf("%x\n",b)
	return b
}

func Uint32ToBytes(v []uint32) []byte {
	var buf = make([]byte, 4*len(v))
	for i, x := range v {
		binary.LittleEndian.PutUint32(buf[4*i:], x)
	}
	return buf
}
