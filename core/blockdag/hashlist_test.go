package blockdag

import (
	"github.com/HalalChain/qitmeer-lib/common/hash"
	"testing"
)

func Test_Has(t *testing.T) {
	hl:=HashList{}
	hl=append(hl,&hash.ZeroHash)

	if !hl.Has(&hash.ZeroHash) {
		t.FailNow()
	}
}