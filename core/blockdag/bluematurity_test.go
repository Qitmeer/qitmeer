package blockdag

import "testing"

func Test_CheckBlueAndMature(t *testing.T) {
	ibd := InitBlockDAG(phantom, "PH_fig2-blocks")
	if ibd == nil {
		t.FailNow()
	}
	err := bd.CheckBlueAndMature([]uint{tbMap["D"].GetID()}, []uint{tbMap["I"].GetID()}, 2)
	if err != nil {
		t.Fatal(err)
	}
}

func Test_CheckBlueAndMatureMT(t *testing.T) {
	ibd := InitBlockDAG(phantom, "PH_fig2-blocks")
	if ibd == nil {
		t.FailNow()
	}
	err := bd.CheckBlueAndMatureMT([]uint{tbMap["D"].GetID()}, []uint{tbMap["I"].GetID()}, 2)
	if err != nil {
		t.Fatal(err)
	}
}
