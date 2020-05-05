package blockdag

import (
	"fmt"
	"testing"
)

func Test_GetMainChain(t *testing.T) {
	ibd := InitBlockDAG(conflux, "CO_Blocks")
	if ibd == nil {
		t.FailNow()
	}
	con := ibd.(*Conflux)
	fmt.Println("Conflux main chain ：")
	mainChain := con.GetMainChain()
	mainChain = reverseBlockList(mainChain)
	printBlockChainTag(mainChain)
	if !processResult(mainChain, changeToIDList(testData.CO_GetMainChain.Output)) {
		t.FailNow()
	}
}

func Test_GetOrder(t *testing.T) {
	ibd := InitBlockDAG(conflux, "CO_Blocks")
	if ibd == nil {
		t.FailNow()
	}
	//con:=ibd.(*Conflux)
	fmt.Println("Conflux order ：")
	order := []uint{}
	var i uint
	for i = 0; i < bd.GetBlockTotal(); i++ {
		order = append(order, bd.order[i])
	}
	printBlockChainTag(order)
	if !processResult(order, changeToIDList(testData.CO_GetOrder.Output)) {
		t.FailNow()
	}
}
