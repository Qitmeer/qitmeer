package blockdag

import (
	"fmt"
	"sort"
)

func Log(sp *Spectre) {
	votes1, votes2 := make([]string, 0), make([]string, 0)
	for h, v := range sp.Votes() {
		if v {
			votes1 = append(votes1, h.String())
		} else {
			votes2 = append(votes2, h.String())
		}
	}
	sort.Strings(votes1)
	sort.Strings(votes2)
	fmt.Println(votes1)
	fmt.Println(votes2)
}

/*
func TestSpectre1(t *testing.T) {
	ibd, tbMap := InitBlockDAG(spectre, "SP_Blocks")
	if ibd == nil {
		t.FailNow()
	}
	sp := ibd.(*Spectre)
	b4h := tbMap["b4"]
	b6h := tbMap["b6"]
	b4, b6 := sp.bd.GetBlock(b4h), sp.bd.GetBlock(b6h)

	if ret, err := sp.Vote(b4, b6); err != nil {
		t.Error(err.Error())
	} else {
		Log(sp)
		if !ret {
			t.Error(ret)
		}
	}
}*/
