package blockdag

import (
	"qitmeer/common/hash"
	"testing"
)

func Test_Has(t *testing.T) {
	hl:=HashList{}
	hl=append(hl,&hash.ZeroHash)

	if !hl.Has(&hash.ZeroHash) {
		t.FailNow()
	}
}