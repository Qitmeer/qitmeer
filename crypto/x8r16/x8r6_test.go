package x8r16_test

import (
	"encoding/hex"
	"fmt"
	"github.com/Qitmeer/qitmeer/crypto/x8r16"
)

func ExampleNew() {
	b := []byte("helloworldhelloworldhelloworldhelloworldhelloworldhelloworldhelloworldhelloworldhelloworldhelloworldhelloworldhel")
	h := x8r16.New()
	out := make([]byte, 32)
	h.Hash(b, out)
	fmt.Printf(hex.EncodeToString(out))
	// output:
	// 52ac0c51e33f308f838998528d492cb135162a90f235121a65033f143c214a16
}
