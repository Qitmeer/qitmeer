package blockdag

import (
	"testing"
)

func Test_HasId(t *testing.T) {
	hl := IdSlice{}
	hl = append(hl, 0)

	if !hl.Has(0) {
		t.FailNow()
	}
}
