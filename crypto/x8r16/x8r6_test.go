package x8r16

import (
	"encoding/hex"
	"fmt"
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

func TestHashX8(t *testing.T) {
	b, _ := hex.DecodeString("0c0000003ebbe1de1524c5d8450759652bfa0b1502f7b4b878b320dfcb56a3325b6307a3ec63ff9b780e9f6c854d954445ab866b81e5906c831f9908bd8b18c2bfea81d40000000000000000000000000000000000000000000000000000000000000000ffff1f1ce164b55e0100000005")
	x8r16 := New()
	out := make([]byte, 32)
	x8r16.Hash(b, out)
	fmt.Println(hex.EncodeToString(out))
}
