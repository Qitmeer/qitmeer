package shabal

import (
	"encoding/binary"
	"log"
	//	"unsafe"
	//	"bytes"
	//	"reflect"
	//	"strconv"
	//	"fmt"
)

var A_init_512 = [...]uint32{
	(0x20728DFD), (0x46C0BD53), (0xE782B699), (0x55304632),
	(0x71B4EF90), (0x0EA9E82C), (0xDBB930F1), (0xFAD06B8B),
	(0xBE0CAE40), (0x8BD14410), (0x76D2ADAC), (0x28ACAB7F),
}

var B_init_512 = [...]uint32{
	(0xC1099CB7), (0x07B385F3), (0xE7442C26), (0xCC8AD640),
	(0xEB6F56C7), (0x1EA81AA9), (0x73B9D314), (0x1DE85D08),
	(0x48910A5A), (0x893B22DB), (0xC5A0DF44), (0xBBC4324E),
	(0x72D2F240), (0x75941D99), (0x6D8BDE82), (0xA1A7502B),
}

var C_init_512 = [...]uint32{
	(0xD9BF68D1), (0x58BAD750), (0x56028CB2), (0x8134F359),
	(0xB5D469D8), (0x941A8CC2), (0x418B2A6E), (0x04052780),
	(0x7F07D787), (0x5194358F), (0x3C60D665), (0xBE97D79A),
	(0x950C3434), (0xAED9A06D), (0x2537DC8D), (0x7CDB5969),
}

var buf []uint8
var C, M, A, B []uint32
var Whigh, Wlow uint32
var ptr int

func PERM_ELT(xa0, xa1, xb0, xb1, xb2, xb3, xc, xm *uint32) {
	*xa0 = (((*xa0) ^ ((((*xa1) << 15) | ((*xa1) >> 17)) * 5) ^ (*xc)) * 3) ^ (*xb1) ^ ((*xb2) & (^(*xb3))) ^ (*xm)
	*xb0 = (^((((*xb0) << 1) | ((*xb0) >> 31)) ^ (*xa0)))
}

func shabal_cal_DECODE() {
	for i := 0; i < 16; i++ {
		M[i] = binary.LittleEndian.Uint32(buf[i*4:])
		B[i] = (B[i] + M[i])
	}
}

func shabal_cal_APPLY_P() {
	for i := 0; i < 16; i++ {
		B[i] = (B[i] << 17) | (B[i] >> 15)
	}

	PERM_ELT(&A[0], &A[11], &B[0], &B[13], &B[9], &B[6], &C[8], &M[0])
	PERM_ELT(&A[1], &A[0], &B[1], &B[14], &B[10], &B[7], &C[7], &M[1])
	PERM_ELT(&A[2], &A[1], &B[2], &B[15], &B[11], &B[8], &C[6], &M[2])
	PERM_ELT(&A[3], &A[2], &B[3], &B[0], &B[12], &B[9], &C[5], &M[3])
	PERM_ELT(&A[4], &A[3], &B[4], &B[1], &B[13], &B[10], &C[4], &M[4])
	PERM_ELT(&A[5], &A[4], &B[5], &B[2], &B[14], &B[11], &C[3], &M[5])
	PERM_ELT(&A[6], &A[5], &B[6], &B[3], &B[15], &B[12], &C[2], &M[6])
	PERM_ELT(&A[7], &A[6], &B[7], &B[4], &B[0], &B[13], &C[1], &M[7])
	PERM_ELT(&A[8], &A[7], &B[8], &B[5], &B[1], &B[14], &C[0], &M[8])
	PERM_ELT(&A[9], &A[8], &B[9], &B[6], &B[2], &B[15], &C[15], &M[9])
	PERM_ELT(&A[10], &A[9], &B[10], &B[7], &B[3], &B[0], &C[14], &M[10])
	PERM_ELT(&A[11], &A[10], &B[11], &B[8], &B[4], &B[1], &C[13], &M[11])
	PERM_ELT(&A[0], &A[11], &B[12], &B[9], &B[5], &B[2], &C[12], &M[12])
	PERM_ELT(&A[1], &A[0], &B[13], &B[10], &B[6], &B[3], &C[11], &M[13])
	PERM_ELT(&A[2], &A[1], &B[14], &B[11], &B[7], &B[4], &C[10], &M[14])
	PERM_ELT(&A[3], &A[2], &B[15], &B[12], &B[8], &B[5], &C[9], &M[15])
	PERM_ELT(&A[4], &A[3], &B[0], &B[13], &B[9], &B[6], &C[8], &M[0])
	PERM_ELT(&A[5], &A[4], &B[1], &B[14], &B[10], &B[7], &C[7], &M[1])
	PERM_ELT(&A[6], &A[5], &B[2], &B[15], &B[11], &B[8], &C[6], &M[2])
	PERM_ELT(&A[7], &A[6], &B[3], &B[0], &B[12], &B[9], &C[5], &M[3])
	PERM_ELT(&A[8], &A[7], &B[4], &B[1], &B[13], &B[10], &C[4], &M[4])
	PERM_ELT(&A[9], &A[8], &B[5], &B[2], &B[14], &B[11], &C[3], &M[5])
	PERM_ELT(&A[10], &A[9], &B[6], &B[3], &B[15], &B[12], &C[2], &M[6])
	PERM_ELT(&A[11], &A[10], &B[7], &B[4], &B[0], &B[13], &C[1], &M[7])
	PERM_ELT(&A[0], &A[11], &B[8], &B[5], &B[1], &B[14], &C[0], &M[8])
	PERM_ELT(&A[1], &A[0], &B[9], &B[6], &B[2], &B[15], &C[15], &M[9])
	PERM_ELT(&A[2], &A[1], &B[10], &B[7], &B[3], &B[0], &C[14], &M[10])
	PERM_ELT(&A[3], &A[2], &B[11], &B[8], &B[4], &B[1], &C[13], &M[11])
	PERM_ELT(&A[4], &A[3], &B[12], &B[9], &B[5], &B[2], &C[12], &M[12])
	PERM_ELT(&A[5], &A[4], &B[13], &B[10], &B[6], &B[3], &C[11], &M[13])
	PERM_ELT(&A[6], &A[5], &B[14], &B[11], &B[7], &B[4], &C[10], &M[14])
	PERM_ELT(&A[7], &A[6], &B[15], &B[12], &B[8], &B[5], &C[9], &M[15])
	PERM_ELT(&A[8], &A[7], &B[0], &B[13], &B[9], &B[6], &C[8], &M[0])
	PERM_ELT(&A[9], &A[8], &B[1], &B[14], &B[10], &B[7], &C[7], &M[1])
	PERM_ELT(&A[10], &A[9], &B[2], &B[15], &B[11], &B[8], &C[6], &M[2])
	PERM_ELT(&A[11], &A[10], &B[3], &B[0], &B[12], &B[9], &C[5], &M[3])
	PERM_ELT(&A[0], &A[11], &B[4], &B[1], &B[13], &B[10], &C[4], &M[4])
	PERM_ELT(&A[1], &A[0], &B[5], &B[2], &B[14], &B[11], &C[3], &M[5])
	PERM_ELT(&A[2], &A[1], &B[6], &B[3], &B[15], &B[12], &C[2], &M[6])
	PERM_ELT(&A[3], &A[2], &B[7], &B[4], &B[0], &B[13], &C[1], &M[7])
	PERM_ELT(&A[4], &A[3], &B[8], &B[5], &B[1], &B[14], &C[0], &M[8])
	PERM_ELT(&A[5], &A[4], &B[9], &B[6], &B[2], &B[15], &C[15], &M[9])
	PERM_ELT(&A[6], &A[5], &B[10], &B[7], &B[3], &B[0], &C[14], &M[10])
	PERM_ELT(&A[7], &A[6], &B[11], &B[8], &B[4], &B[1], &C[13], &M[11])
	PERM_ELT(&A[8], &A[7], &B[12], &B[9], &B[5], &B[2], &C[12], &M[12])
	PERM_ELT(&A[9], &A[8], &B[13], &B[10], &B[6], &B[3], &C[11], &M[13])
	PERM_ELT(&A[10], &A[9], &B[14], &B[11], &B[7], &B[4], &C[10], &M[14])
	PERM_ELT(&A[11], &A[10], &B[15], &B[12], &B[8], &B[5], &C[9], &M[15])
	A[11] = (A[11] + C[6])
	A[10] = (A[10] + C[5])
	A[9] = (A[9] + C[4])
	A[8] = (A[8] + C[3])
	A[7] = (A[7] + C[2])
	A[6] = (A[6] + C[1])
	A[5] = (A[5] + C[0])
	A[4] = (A[4] + C[15])
	A[3] = (A[3] + C[14])
	A[2] = (A[2] + C[13])
	A[1] = (A[1] + C[12])
	A[0] = (A[0] + C[11])
	A[11] = (A[11] + C[10])
	A[10] = (A[10] + C[9])
	A[9] = (A[9] + C[8])
	A[8] = (A[8] + C[7])
	A[7] = (A[7] + C[6])
	A[6] = (A[6] + C[5])
	A[5] = (A[5] + C[4])
	A[4] = (A[4] + C[3])
	A[3] = (A[3] + C[2])
	A[2] = (A[2] + C[1])
	A[1] = (A[1] + C[0])
	A[0] = (A[0] + C[15])
	A[11] = (A[11] + C[14])
	A[10] = (A[10] + C[13])
	A[9] = (A[9] + C[12])
	A[8] = (A[8] + C[11])
	A[7] = (A[7] + C[10])
	A[6] = (A[6] + C[9])
	A[5] = (A[5] + C[8])
	A[4] = (A[4] + C[7])
	A[3] = (A[3] + C[6])
	A[2] = (A[2] + C[5])
	A[1] = (A[1] + C[4])
	A[0] = (A[0] + C[3])
}

func shabal_cal_SUB() {
	for i := 0; i < 16; i++ {
		C[i] = (C[i] - M[i])
	}
}

func shabal_cal_SWAP() {
	for i := 0; i < 16; i++ {
		A := B[i]
		B[i] = C[i]
		C[i] = A
	}
}

func Shabal_512_process(data []uint8, dst []uint8, length int) {
	buf = make([]uint8, 64)
	C = make([]uint32, 16)
	M = make([]uint32, 16)
	A = make([]uint32, 12)
	B = make([]uint32, 16)

	copy(A, A_init_512[:])
	copy(B, B_init_512[:])
	copy(C, C_init_512[:])

	Wlow = 1
	Whigh = 0
	ptr = 0

	//var clen int;
	for length > 0 {

		clen := 64 - ptr
		if clen > length {
			clen = length
		}
		//memcpy(buf + ptr, data, clen);
		for i := 0; i < clen; i++ {
			buf[ptr+i] = data[i]
		}
		ptr += clen
		//data = (uint8 *)data + clen;
		data = data[clen:]
		length -= clen
		if ptr == 64 {
			shabal_cal_DECODE()

			A[0] ^= Wlow
			A[1] ^= Whigh

			shabal_cal_APPLY_P()

			shabal_cal_SUB()

			shabal_cal_SWAP()

			Wlow = (Wlow + 1)

			ptr = 0
		}
	}

	buf[ptr] = 0x80
	for i := 0; i < 64-(ptr+1); i++ {
		buf[ptr+1+i] = 0
	}
	//memset(buf + ptr + 1, 0, 64 - (ptr + 1));

	shabal_cal_DECODE()
	A[0] ^= Wlow
	A[1] ^= Whigh
	shabal_cal_APPLY_P()
	for i := 0; i < 3; i++ {

		shabal_cal_SWAP()

		A[0] ^= Wlow
		A[1] ^= Whigh

		shabal_cal_APPLY_P()

	}

	for i := 0; i < 16; i++ {
		binary.LittleEndian.PutUint32(dst[i*4:i*4+4], B[i])
	}
}

var input = [...]uint32{
	0x02000000, 0x8d870b41, 0x404883ac, 0x195d9920, 0x1225a41d, 0xd77969a6, 0x8374e68e, 0xc8ee7500,
	0x00000000, 0xa2123af0, 0x394e7606, 0xb5fec3cb, 0x96ddeea4, 0xd1d376ac, 0xc0daeb20, 0x2c5fc670,
	0x6c5bb067, 0xc7044a53, 0xe3e6001c, 0x00104d49}

func main() {
	buf := make([]uint8, 80)
	hash := make([]uint8, 64)
	var hash0 uint32

	for i := 0; i < 20; i++ {
		binary.LittleEndian.PutUint32(buf[i*4:i*4+4], input[i])
	}

	log.Println("\n do shabal test 80bytes...\n")
	Shabal_512_process(buf[:], hash, 80)
	//log.Printf("-hash = %x",hash)

	hash0 = binary.LittleEndian.Uint32(hash)

	if hash0 == 0x7d52bae6 {
		log.Println("\n shabal test OK!!! \n")
	} else {
		log.Printf("\n shabal test ERROR: %x!!! \n", hash0)
	}

	log.Println("\n do shabal test 64bytes...\n")

	Shabal_512_process(buf, hash, 64)
	//log.Printf("-hash = %x",hash)

	hash0 = binary.LittleEndian.Uint32(hash)
	if hash0 == 0x757a0334 {
		log.Println("\n shabal test OK!!! \n")
	} else {
		log.Printf("\n shabal test ERROR: %x!!! \n", hash0)
	}

	log.Println("==============================================\n")

}
