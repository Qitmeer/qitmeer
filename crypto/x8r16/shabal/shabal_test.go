package shabal

import (
	"encoding/binary"
	"github.com/stretchr/testify/assert"
	"testing"
)

var input = [...]uint32{
	0x02000000, 0x8d870b41, 0x404883ac, 0x195d9920, 0x1225a41d, 0xd77969a6, 0x8374e68e, 0xc8ee7500,
	0x00000000, 0xa2123af0, 0x394e7606, 0xb5fec3cb, 0x96ddeea4, 0xd1d376ac, 0xc0daeb20, 0x2c5fc670,
	0x6c5bb067, 0xc7044a53, 0xe3e6001c, 0x00104d49}

func TestShabal_512_process(t *testing.T) {
	buf := make([]uint8, 80)
	hash := make([]uint8, 64)
	var hash0 uint32

	for i := 0; i < 20; i++ {
		binary.LittleEndian.PutUint32(buf[i*4:i*4+4], input[i])
	}

	Shabal_512_process(buf[:], hash, 80)

	hash0 = binary.LittleEndian.Uint32(hash)
	assert.Equal(t, hash0, uint32(0x7d52bae6))

	Shabal_512_process(buf, hash, 64)

	hash0 = binary.LittleEndian.Uint32(hash)
	assert.Equal(t, hash0, uint32(0x757a0334))
}
