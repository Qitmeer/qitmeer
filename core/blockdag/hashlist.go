package blockdag

import "github.com/HalalChain/qitmeer/common/hash"

// HashList is used to sort hash list
type HashList []*hash.Hash

func (sh HashList) Len() int {
	return len(sh)
}

func (sh HashList) Less(i, j int) bool {
	return sh[i].String() < sh[j].String()
}

func (sh HashList) Swap(i, j int) {
	sh[i], sh[j] = sh[j], sh[i]
}

func (sh HashList) Has(h *hash.Hash) bool {
	for _,v:=range sh{
		if v.IsEqual(h) {
			return true
		}
	}
	return false
}