package blockdag

// IdSlice is used to sort id list
type IdSlice []uint

func (sh IdSlice) Len() int {
	return len(sh)
}

func (sh IdSlice) Less(i, j int) bool {
	return sh[i] < sh[j]
}

func (sh IdSlice) Swap(i, j int) {
	sh[i], sh[j] = sh[j], sh[i]
}

func (sh IdSlice) Has(id uint) bool {
	for _, v := range sh {
		if v == id {
			return true
		}
	}
	return false
}
