package blockdag

import (
	"github.com/Qitmeer/qng-core/common/hash"
	"testing"
)

func Test_Has(t *testing.T) {
	hl := HashSlice{}
	hl = append(hl, &hash.ZeroHash)

	if !hl.Has(&hash.ZeroHash) {
		t.FailNow()
	}
}
