package vm

import (
	"context"
	"testing"
)

func Test_VM(t *testing.T) {
	_, err := NewVM(context.Background(), "", "")
	if err != nil {
		//t.Fatal(err)
	}
}
