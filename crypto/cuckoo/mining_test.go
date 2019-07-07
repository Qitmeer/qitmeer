package cuckoo

import (
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"github.com/HalalChain/qitmeer-lib/common/hash"
	"log"
	"math/big"
	"runtime"
	"strconv"
	"testing"
)

const targetBits = 1

var txRoot = ""

func TestCuckooMining(t *testing.T) {

	// Use all processor cores.
	runtime.GOMAXPROCS(runtime.NumCPU())
	var targetDifficulty = big.NewInt(2)
	targetDifficulty.Lsh(targetDifficulty, uint(255-targetBits))
	fmt.Printf("Target:%064x\n", targetDifficulty.Bytes())
	c := NewCuckoo() // 构建 Cuckoo 结构体

	//rand.Seed(time.Now().UnixNano())

	var cycleNonces []uint32
	var isFound bool
	var nonce int64
	var cycleNoncesHashInt big.Int
	var siphashKey []byte
	test := "helloworld"

	for nonce = 0; ; nonce++ {
		test += fmt.Sprintf("%d",nonce)
		siphashKey = hash.DoubleHashB([]byte(test))
		fmt.Println("header nonce = ", nonce)
		cycleNonces, isFound = c.PoW(siphashKey[:16])
		if !isFound {
			continue
		}
		if err := Verify(siphashKey[:16], cycleNonces); err != nil {
			continue
		}
		log.Println("Found 42 Cycles:")
		for _, v := range cycleNonces {
			fmt.Printf("%d,", v)
		}
		cycleNoncesHash := hash.DoubleHashB(Uint32ToBytes(cycleNonces))
		cycleNoncesHashInt.SetBytes(cycleNoncesHash[:])

		// The block is solved when the new block hash is less
		// than the target difficulty.  Yay!
		if cycleNoncesHashInt.Cmp(targetDifficulty) <= 0 {
			fmt.Println("\nsuccess!",hex.EncodeToString(cycleNoncesHash))
			break
		} else {
			txRoot = strconv.FormatInt(nonce, 10)
			fmt.Println("txRoot=", txRoot)
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
