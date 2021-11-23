package blockdag

import "github.com/Qitmeer/qng-core/common/hash"

// HashSlice is used to sort hash list
type HashSlice []*hash.Hash

func (sh HashSlice) Len() int {
	return len(sh)
}

func (sh HashSlice) Less(i, j int) bool {
	return sh[i].String() < sh[j].String()
}

func (sh HashSlice) Swap(i, j int) {
	sh[i], sh[j] = sh[j], sh[i]
}

func (sh HashSlice) Has(h *hash.Hash) bool {
	for _, v := range sh {
		if v.IsEqual(h) {
			return true
		}
	}
	return false
}
