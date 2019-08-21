package blockdag

import (
	"fmt"
	"github.com/Qitmeer/qitmeer-lib/common/hash"
	"github.com/Qitmeer/qitmeer-lib/core/dag"
	"strconv"
	"testing"
)

func Test_GetFutureSet(t *testing.T) {
	ibd, tbMap := InitBlockDAG(phantom,"PH_fig2-blocks")
	if ibd==nil {
		t.FailNow()
	}

	//ph:=ibd.(*Phantom)
	anBlock := bd.GetBlock(tbMap[testData.PH_GetFutureSet.Input])
	bset := dag.NewHashSet()
	bd.GetFutureSet(bset,anBlock)
	fmt.Printf("Get %s future set：\n", testData.PH_GetFutureSet.Input)
	printBlockSetTag(bset,tbMap)
	//
	if !processResult(bset,changeToHashList(testData.PH_GetFutureSet.Output, tbMap)) {
		t.FailNow()
	}
}

func Test_GetAnticone(t *testing.T) {
	ibd, tbMap := InitBlockDAG(phantom,"PH_fig2-blocks")
	if ibd==nil {
		t.FailNow()
	}
	ph:=ibd.(*Phantom)
	//
	anBlock := bd.GetBlock(tbMap[testData.PH_GetAnticone.Input])

	////////////
	bset := ph.bd.GetAnticone(anBlock, nil)
	fmt.Printf("Get %s anticone set：\n", testData.PH_GetAnticone.Input)
	printBlockSetTag(bset,tbMap)
	//
	if !processResult(bset,changeToHashList(testData.PH_GetAnticone.Output, tbMap)) {
		t.FailNow()
	}

}


func Test_BlueSetFig2(t *testing.T) {
	ibd, tbMap := InitBlockDAG(phantom,"PH_fig2-blocks")
	if ibd==nil {
		t.FailNow()
	}
	ph:=ibd.(*Phantom)
	//
	blueSet := ph.GetDiffBlueSet()
	fmt.Println("Fig2 diff blue set：")
	printBlockSetTag(blueSet,tbMap)
	if !processResult(blueSet, changeToHashList(testData.PH_BlueSetFig2.Output, tbMap)) {
		t.FailNow()
	}
}

func Test_BlueSetFig4(t *testing.T) {
	ibd, tbMap := InitBlockDAG(phantom,"PH_fig4-blocks")
	if ibd==nil {
		t.FailNow()
	}
	ph:=ibd.(*Phantom)
	//
	blueSet := ph.GetDiffBlueSet()
	fmt.Println("Fig4 diff blue set：")
	printBlockSetTag(blueSet,tbMap)
	if !processResult(blueSet, changeToHashList(testData.PH_BlueSetFig4.Output, tbMap)) {
		t.FailNow()
	}
}

func Test_OrderFig2(t *testing.T) {
	ibd, tbMap := InitBlockDAG(phantom,"PH_fig2-blocks")
	if ibd==nil {
		t.FailNow()
	}
	//ph:=ibd.(*Phantom)
	order:=[]*hash.Hash{}
	var i uint
	for i=0;i<bd.GetBlockTotal() ;i++  {
		order=append(order,bd.GetBlockByOrder(i))
	}
	fmt.Printf("The Fig.2 Order: ")
	printBlockChainTag(order,tbMap)

	if !processResult(order, changeToHashList(testData.PH_OrderFig2.Output, tbMap)) {
		t.FailNow()
	}
}

func Test_OrderFig4(t *testing.T) {
	ibd, tbMap := InitBlockDAG(phantom,"PH_fig4-blocks")
	if ibd==nil {
		t.FailNow()
	}
	//ph:=ibd.(*Phantom)
	order:=[]*hash.Hash{}
	var i uint
	for i=0;i<bd.GetBlockTotal() ;i++  {
		order=append(order,bd.GetBlockByOrder(i))
	}
	fmt.Printf("The Fig.4 Order: ")
	printBlockChainTag(order,tbMap)

	if !processResult(order, changeToHashList(testData.PH_OrderFig4.Output, tbMap)) {
		t.FailNow()
	}
}

func Test_GetLayer(t *testing.T) {
	ibd, _ := InitBlockDAG(phantom,"PH_fig2-blocks")
	if ibd==nil {
		t.FailNow()
	}
	var result string=""
	var i uint
	for i=0;i<bd.GetBlockTotal() ;i++  {
		l:=bd.GetLayer(bd.GetBlockByOrder(i))
		result=fmt.Sprintf("%s%d",result,l)
	}
	if result != testData.PH_GetLayer.Output[0] {
		t.FailNow()
	}
}

func Test_IsOnMainChain(t *testing.T) {
	ibd, tbMap := InitBlockDAG(phantom,"PH_fig2-blocks")
	if ibd==nil {
		t.FailNow()
	}
	if strconv.FormatBool(bd.IsOnMainChain(tbMap[testData.PH_IsOnMainChain.Input]))!=testData.PH_IsOnMainChain.Output[0] {
		t.FailNow()
	}
}

func Test_LocateBlocks(t *testing.T) {
	ibd, tbMap := InitBlockDAG(phantom,"PH_fig2-blocks")
	if ibd==nil {
		t.FailNow()
	}
	gs:=dag.NewGraphState()
	gs.GetTips().Add(bd.GetGenesisHash())
	gs.SetTotal(1)
	gs.SetLayer(0)
	lb:=bd.LocateBlocks(gs,100)
	lbhs:=dag.NewHashSet()
	lbhs.AddList(lb)

	if !processResult(lbhs,changeToHashList(testData.PH_LocateBlocks.Output, tbMap)) {
		t.FailNow()
	}
}

func Test_LocateMaxBlocks(t *testing.T) {
	ibd, tbMap := InitBlockDAG(phantom,"PH_fig2-blocks")
	if ibd==nil {
		t.FailNow()
	}
	gs:=dag.NewGraphState()
	gs.GetTips().Add(bd.GetGenesisHash())
	gs.GetTips().Add(tbMap["C"])
	gs.GetTips().Add(tbMap["D"])
	gs.SetTotal(3)
	gs.SetLayer(1)
	lb:=bd.LocateBlocks(gs,4)
	//printBlockChainTag(lb,tbMap)
	if !processResult(lb,changeToHashList(testData.PH_LocateMaxBlocks.Output, tbMap)) {
		t.FailNow()
	}
}

