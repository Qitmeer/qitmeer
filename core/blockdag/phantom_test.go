package blockdag

import (
	"fmt"
	"github.com/Qitmeer/qitmeer/common/hash"
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
	bset := NewHashSet()
	bd.getFutureSet(bset,anBlock)
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
	bset := ph.bd.getAnticone(anBlock, nil)
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
	ph:=ibd.(*Phantom)
	order:=[]*hash.Hash{}
	var i uint
	ph.UpdateVirtualBlockOrder()
	for i=0;i<bd.GetBlockTotal() ;i++  {
		order=append(order,bd.order[i])
	}
	fmt.Printf("The Fig.2 Order: ")
	printBlockChainTag(order,tbMap)

	if !processResult(order, changeToHashList(testData.PH_OrderFig2.Output, tbMap)) {
		t.FailNow()
	}

	//
	da:=ph.GetDiffAnticone()
	fmt.Printf("The diffanticoner: ")
	printBlockSetTag(da,tbMap)
}

func Test_OrderFig4(t *testing.T) {
	ibd, tbMap := InitBlockDAG(phantom,"PH_fig4-blocks")
	if ibd==nil {
		t.FailNow()
	}
	ph:=ibd.(*Phantom)
	order:=[]*hash.Hash{}
	var i uint
	ph.UpdateVirtualBlockOrder()
	for i=0;i<bd.GetBlockTotal() ;i++  {
		order=append(order,bd.order[i])
	}
	fmt.Printf("The Fig.4 Order: ")
	printBlockChainTag(order,tbMap)

	if !processResult(order, changeToHashList(testData.PH_OrderFig4.Output, tbMap)) {
		t.FailNow()
	}

	//
	da:=ph.GetDiffAnticone()
	fmt.Printf("The diffanticoner: ")
	printBlockSetTag(da,tbMap)
}

func Test_GetLayer(t *testing.T) {
	ibd, _ := InitBlockDAG(phantom,"PH_fig2-blocks")
	if ibd==nil {
		t.FailNow()
	}
	var result string=""
	var i uint
	ph:=ibd.(*Phantom)
	ph.UpdateVirtualBlockOrder()
	for i=0;i<bd.GetBlockTotal() ;i++  {
		l:=bd.GetLayer(bd.order[i])
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
	gs:=NewGraphState()
	gs.GetTips().Add(bd.GetGenesisHash())
	gs.SetTotal(1)
	gs.SetLayer(0)
	lb:=bd.LocateBlocks(gs,100)
	lbhs:=NewHashSet()
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
	gs:=NewGraphState()
	gs.GetTips().Add(bd.GetGenesisHash())
	gs.GetTips().Add(tbMap["G"])
	gs.SetTotal(4)
	gs.SetLayer(2)
	lb:=bd.LocateBlocks(gs,4)
	//printBlockChainTag(lb,tbMap)
	if !processResult(lb,changeToHashList(testData.PH_LocateMaxBlocks.Output, tbMap)) {
		t.FailNow()
	}
}

func Test_Confirmations(t *testing.T) {
	ibd, tbMap := InitBlockDAG(phantom,"PH_fig2-blocks")
	if ibd==nil {
		t.FailNow()
	}
	mainTip:=bd.GetMainChainTip()
	mainChain:=[]*hash.Hash{}
	for cur:=mainTip; cur != nil; cur = bd.GetBlock(cur.GetMainParent()) {
		mainChain=append(mainChain,cur.GetHash())
	}
	printBlockChainTag(reverseBlockList(mainChain),tbMap)

	ph:=ibd.(*Phantom)
	ph.UpdateVirtualBlockOrder()
	for i:=uint(0);i<bd.GetBlockTotal();i++ {
		blockHash:=bd.order[i]
		fmt.Printf("%s : %d\n",getBlockTag(blockHash,tbMap),bd.GetConfirmations(blockHash))
	}
}