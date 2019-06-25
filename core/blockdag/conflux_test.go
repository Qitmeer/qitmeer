package blockdag

import (
	"fmt"
	"github.com/HalalChain/qitmeer-lib/common/hash"
	"testing"
)

func Test_GetMainChain(t *testing.T) {
	ibd, tbMap := InitBlockDAG(conflux,"CO_Blocks")
	if ibd==nil {
		t.FailNow()
	}
	con:=ibd.(*Conflux)
	fmt.Println("Conflux main chain ：")
	mainChain := con.GetMainChain()
	mainChain=reverseBlockList(mainChain)
	printBlockChainTag(mainChain,tbMap)
	if !processResult(mainChain,changeToHashList(testData.CO_GetMainChain.Output, tbMap)) {
		t.FailNow()
	}
}

func Test_GetOrder(t *testing.T) {
	ibd, tbMap := InitBlockDAG(conflux,"CO_Blocks")
	if ibd==nil {
		t.FailNow()
	}
	//con:=ibd.(*Conflux)
	fmt.Println("Conflux order ：")
	order:=[]*hash.Hash{}
	var i uint
	for i=0;i<bd.GetBlockTotal() ;i++  {
		order=append(order,bd.GetBlockByOrder(i))
	}
	printBlockChainTag(order,tbMap)
	if !processResult(order,changeToHashList(testData.CO_GetOrder.Output, tbMap)) {
		t.FailNow()
	}
}
