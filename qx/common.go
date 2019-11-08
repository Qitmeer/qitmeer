package qx

import (
	"fmt"
	"os"
)

func ErrExit(err error) {
	fmt.Fprintf(os.Stderr, "Qx Error : %q\n", err)
	os.Exit(1)
}
