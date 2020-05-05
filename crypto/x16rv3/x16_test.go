package x16rv3

import (
	"encoding/binary"
	"fmt"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestHash(t *testing.T) {
	b := []byte("helloworldhelloworldhelloworldhelloworldhelloworldhelloworldhelloworldhelloworldhelloworldhelloworldhelloworldhel")
	x16rv3 := New()
	out := make([]byte, 32)
	x16rv3.Hash(b, out)
	s := ""
	for i := 0; i < 8; i++ {
		a := binary.LittleEndian.Uint32(out[i*4 : 4*(i+1)])
		s += fmt.Sprintf("%x", a)
	}
	assert.Equal(t, s, "c768288d52785bf19a2f1e6da377d020eff11d3593de3aaf3869e583d3095816")
}
