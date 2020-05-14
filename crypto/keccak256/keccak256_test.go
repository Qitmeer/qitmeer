package keccak256

import (
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"log"
	"testing"
)

var input = [...]uint32{
	0x02000000, 0x8d870b41, 0x404883ac, 0x195d9920, 0x1225a41d, 0xd77969a6, 0x8374e68e, 0xc8ee7500,
	0x00000000, 0xa2123af0, 0x394e7606, 0xb5fec3cb, 0x96ddeea4, 0xd1d376ac, 0xc0daeb20, 0x2c5fc670,
	0x394e7606, 0xa2123af0, 0x394e7606, 0xb5fec3cb, 0x96ddeea4, 0xd1d376ac, 0xc0daeb20, 0x2c5fc670,
	0xa2123af0, 0xa2123af0, 0x394e7606, 0xb5fec3cb, 0x96ddeea4, 0xd1d376ac, 0xc0daeb20, 0x2c5fc670,
	0x6c5bb067, 0xc7044a53, 0xe3e6001c, 0x00104d49}

func TestKeccak256(t *testing.T) {
	buf := make([]uint8, 200)
	hash := make([]uint8, 64)
	var hash0 uint32

	for i := 0; i < 32; i++ {
		binary.LittleEndian.PutUint32(buf[i*4:i*4+4], input[i])
	}

	log.Println("do keccak test 0 ...")
	Sph_keccak256_process(buf[:], hash, 113)
	//log.Printf("-hash = %x",hash)

	hash0 = binary.LittleEndian.Uint32(hash)

	if hash0 == 0x8fc8bb17 {
		log.Println(" keccak test OK!!! ")
	} else {
		log.Printf(" keccak test ERROR: %x!!! ", hash0)
	}

	log.Println(" do keccak test 1 ...")

	Sph_keccak256_process(buf[4:], hash, 113)

	hash0 = binary.LittleEndian.Uint32(hash)
	if hash0 == 0xc43d87ad {
		log.Println(" keccak test OK!!! ")
	} else {
		log.Printf(" keccak test ERROR: %x!!! ", hash0)
	}

	log.Println("==============================================")

}

func TestHeader(t *testing.T) {
	hash := make([]byte, 32)
	b := []byte("helloworldhelloworldhelloworldhelloworldhelloworldhelloworldhelloworldhelloworldhelloworldhelloworldhelloworldhel")
	Sph_keccak256_process(b, hash, 113)
	fmt.Println(hex.EncodeToString(hash[:]))
	h := Sum256(b)
	fmt.Println(hex.EncodeToString(h[:]))
}
