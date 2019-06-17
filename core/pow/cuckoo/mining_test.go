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
	//fmt.Printf("%064x\n", targetDifficulty.Bytes())
	targetDifficulty.Lsh(targetDifficulty, uint(256-targetBits))
	fmt.Printf("%064x\n", targetDifficulty.Bytes())
	fmt.Printf("%0x\n", targetDifficulty.Bytes())

	// var targetDifficulty = big.NewInt(2)
	//0000000000000000000000000000000000000000000000000000000000000002
	//010000000000000000000000000000000000000000000000000000000000000000

	// c.cuckoo = [0 0 0...0 0 0],长度为131073 = 2^17 + 1
	// c.ncpu =  8
	// c.us =  []，创建了[]uint32
	// c.vs =  []
	// c.matrixs = [8][32][32]uint64, 创建这样一个多维数组，但没有初始化，[[[[] [] [].32.].32.].8.]
	// c.m2tmp = [1][8][32]uint64, 创建这样一个多维切片，但没有初始化，[[[[] [] [].32.].8.].1.]
	// c.sip =  <nil>
	// c.m2 = [1][8][32]uint64, 创建这样一个多维切片，但没有初始化，[[[[] [] [].32.].8.].1.]
	c := NewCuckoo() // 构建 Cuckoo 结构体

	//rand.Seed(time.Now().UnixNano())

	var cycleNonces []uint32
	var isFound bool
	var nonce int64
	var cycleNoncesHashInt big.Int
	var siphashKey []byte

	//defer func() {
	//	fmt.Print("defer:")
	//	for _, v := range cycleNonces {
	//		fmt.Printf("%#x,", v)
	//	}
	//	println()
	//}()

	for {
		for nonce = 0; ; nonce++ {
			fmt.Println("header nonce = ", nonce)
			siphashKey = getBlockHeaderHash(nonce) // len(siphashKey) = 16
			cycleNonces, isFound = c.PoW(siphashKey)

			if !isFound {
				continue
			}

			if err := Verify(siphashKey, cycleNonces); err != nil {
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

// HashToBig converts a hash.Hash into a big.Int that can be used to
// perform math comparisons.
//func HashToBig(hash *Hash) *big.Int {
//	// A Hash is in little-endian, but the big package wants the bytes in
//	// big-endian, so reverse them.
//	buf := *hash
//	blen := len(buf)
//	for i := 0; i < blen/2; i++ {
//		buf[i], buf[blen-1-i] = buf[blen-1-i], buf[i]
//	}
//	return new(big.Int).SetBytes(buf[:])
//}

func Uint32ToBytes(v []uint32) []byte {
	var buf = make([]byte, 4*len(v))
	for i, x := range v {
		binary.LittleEndian.PutUint32(buf[4*i:], x)
	}
	return buf
}
