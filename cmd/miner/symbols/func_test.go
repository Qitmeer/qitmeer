package symbols

import (
	"encoding/hex"
	"fmt"
	"testing"
)

func TestA(t *testing.T) {
	s := "00000020d78e072fa1a4b7a8a062b6b556d1144fc3ff2cff593ec60ee686e7df5afacb11d5cee577979c2bd629f3aefb32e1f4636353ac43a7ab268b17637977dda312120000000000000000000000000000000000000000000000000000000000000000ffff7f2004666a61080100000000000000"
	b, err := hex.DecodeString(s)
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println(b)
}
