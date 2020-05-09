package x8r16

import (
	"encoding/hex"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestHash(t *testing.T) {
	b := []byte("helloworldhelloworldhelloworldhelloworldhelloworldhelloworldhelloworldhelloworldhelloworldhelloworldhelloworldhel")
	x8r16 := New()
	out := make([]byte, 32)
	x8r16.Hash(b, out)
	assert.Equal(t, hex.EncodeToString(out), "52ac0c51e33f308f838998528d492cb135162a90f235121a65033f143c214a16")
}
