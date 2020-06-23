package blockdag

import (
	"fmt"
	"strconv"
	"testing"
)

func Test_GetFutureSet(t *testing.T) {
	ibd := InitBlockDAG(phantom, "PH_fig2-blocks")
	if ibd == nil {
		t.FailNow()
	}

	//ph:=ibd.(*Phantom)
	anBlock := tbMap[testData.PH_GetFutureSet.Input]
	bset := NewIdSet()
	bd.getFutureSet(bset, anBlock)
	fmt.Printf("Get %s future set：\n", testData.PH_GetFutureSet.Input)
	printBlockSetTag(bset)
	//
	if !processResult(bset, changeToIDList(testData.PH_GetFutureSet.Output)) {
		t.FailNow()
	}
}

func Test_GetAnticone(t *testing.T) {
	ibd := InitBlockDAG(phantom, "PH_fig2-blocks")
	if ibd == nil {
		t.FailNow()
	}
	ph := ibd.(*Phantom)
	//
	anBlock := tbMap[testData.PH_GetAnticone.Input]

	////////////
	bset := ph.bd.getAnticone(anBlock, nil)
	fmt.Printf("Get %s anticone set：\n", testData.PH_GetAnticone.Input)
	printBlockSetTag(bset)
	//
	if !processResult(bset, changeToIDList(testData.PH_GetAnticone.Output)) {
		t.FailNow()
	}

}

func Test_BlueSetFig2(t *testing.T) {
	ibd := InitBlockDAG(phantom, "PH_fig2-blocks")
	if ibd == nil {
		t.FailNow()
	}
	ph := ibd.(*Phantom)
	//
	blueSet := ph.GetDiffBlueSet()
	fmt.Println("Fig2 diff blue set：")
	printBlockSetTag(blueSet)
	if !processResult(blueSet, changeToIDList(testData.PH_BlueSetFig2.Output)) {
		t.FailNow()
	}
}

func Test_BlueSetFig4(t *testing.T) {
	ibd := InitBlockDAG(phantom, "PH_fig4-blocks")
	if ibd == nil {
		t.FailNow()
	}
	ph := ibd.(*Phantom)
	//
	blueSet := ph.GetDiffBlueSet()
	fmt.Println("Fig4 diff blue set：")
	printBlockSetTag(blueSet)
	if !processResult(blueSet, changeToIDList(testData.PH_BlueSetFig4.Output)) {
		t.FailNow()
	}
}

func Test_OrderFig2(t *testing.T) {
	ibd := InitBlockDAG(phantom, "PH_fig2-blocks")
	if ibd == nil {
		t.FailNow()
	}
	ph := ibd.(*Phantom)
	order := []uint{}
	var i uint
	ph.UpdateVirtualBlockOrder()
	for i = 0; i < bd.GetBlockTotal(); i++ {
		order = append(order, bd.order[i])
	}
	fmt.Printf("The Fig.2 Order: ")
	printBlockChainTag(order)

	if !processResult(order, changeToIDList(testData.PH_OrderFig2.Output)) {
		t.FailNow()
	}

	//
	da := ph.GetDiffAnticone()
	fmt.Printf("The diffanticoner: ")
	printBlockSetTag(da)
}

func Test_OrderFig4(t *testing.T) {
	ibd := InitBlockDAG(phantom, "PH_fig4-blocks")
	if ibd == nil {
		t.FailNow()
	}
	ph := ibd.(*Phantom)
	order := []uint{}
	var i uint
	ph.UpdateVirtualBlockOrder()
	for i = 0; i < bd.GetBlockTotal(); i++ {
		order = append(order, bd.order[i])
	}
	fmt.Printf("The Fig.4 Order: ")
	printBlockChainTag(order)

	if !processResult(order, changeToIDList(testData.PH_OrderFig4.Output)) {
		t.FailNow()
	}

	//
	da := ph.GetDiffAnticone()
	fmt.Printf("The diffanticoner: ")
	printBlockSetTag(da)
}

func Test_GetLayer(t *testing.T) {
	ibd := InitBlockDAG(phantom, "PH_fig2-blocks")
	if ibd == nil {
		t.FailNow()
	}
	var result string = ""
	var i uint
	ph := ibd.(*Phantom)
	ph.UpdateVirtualBlockOrder()
	for i = 0; i < bd.GetBlockTotal(); i++ {
		l := bd.GetLayer(bd.order[i])
		result = fmt.Sprintf("%s%d", result, l)
	}
	if result != testData.PH_GetLayer.Output[0] {
		t.FailNow()
	}
}

func Test_IsOnMainChain(t *testing.T) {
	ibd := InitBlockDAG(phantom, "PH_fig2-blocks")
	if ibd == nil {
		t.FailNow()
	}
	if strconv.FormatBool(bd.IsOnMainChain(tbMap[testData.PH_IsOnMainChain.Input].GetID())) != testData.PH_IsOnMainChain.Output[0] {
		t.FailNow()
	}
}

func Test_LocateBlocks(t *testing.T) {
	ibd := InitBlockDAG(phantom, "PH_fig2-blocks")
	if ibd == nil {
		t.FailNow()
	}
	gs := NewGraphState()
	gs.GetTips().Add(bd.GetGenesisHash())
	gs.SetTotal(1)
	gs.SetLayer(0)
	lb := bd.locateBlocks(gs, 100)
	lbhs := NewHashSet()
	lbhs.AddList(lb)
	if !processResult(lbhs, changeToIDList(testData.PH_LocateBlocks.Output)) {
		t.FailNow()
	}
}

func Test_LocateMaxBlocks(t *testing.T) {
	ibd := InitBlockDAG(phantom, "PH_fig2-blocks")
	if ibd == nil {
		t.FailNow()
	}
	gs := NewGraphState()
	gs.GetTips().Add(bd.GetGenesisHash())
	gs.GetTips().Add(tbMap["G"].GetHash())
	gs.SetTotal(4)
	gs.SetLayer(2)
	lb := bd.locateBlocks(gs, 4)
	//printBlockChainTag(lb,tbMap)
	if !processResult(lb, changeToIDList(testData.PH_LocateMaxBlocks.Output)) {
		t.FailNow()
	}
}

func Test_Confirmations(t *testing.T) {
	ibd := InitBlockDAG(phantom, "PH_fig2-blocks")
	if ibd == nil {
		t.FailNow()
	}
	mainTip := bd.GetMainChainTip()
	mainChain := []uint{}
	for cur := mainTip; cur != nil; cur = bd.GetBlockById(cur.GetMainParent()) {
		mainChain = append(mainChain, cur.GetID())
	}
	printBlockChainTag(reverseBlockList(mainChain))

	ph := ibd.(*Phantom)
	ph.UpdateVirtualBlockOrder()
	for i := uint(0); i < bd.GetBlockTotal(); i++ {
		blockHash := bd.order[i]
		fmt.Printf("%s : %d\n", getBlockTag(blockHash), bd.GetConfirmations(blockHash))
	}
}

func Test_IsDAG(t *testing.T) {
	ibd := InitBlockDAG(phantom, "PH_fig2-blocks")
	if ibd == nil {
		t.FailNow()
	}
	//ph:=ibd.(*Phantom)
	//
	parentsTag := []string{"I", "G"}
	parents := NewIdSet()
	for _, parent := range parentsTag {
		parents.Add(tbMap[parent].GetID())
	}
	block := buildBlock(parents)
	l, ib := bd.AddBlock(block)
	if l != nil && l.Len() > 0 {
		tbMap["L"] = ib
	} else {
		t.Fatalf("Error:%d  L\n", tempHash)
	}

}

func Test_IsHourglass(t *testing.T) {
	ibd := InitBlockDAG(phantom, "CP_Blocks")
	if ibd == nil {
		t.FailNow()
	}
	if !bd.IsHourglass(tbMap["J"].GetID()) {
		t.Fatal()
	}
}

func Test_GetMaturity(t *testing.T) {
	ibd := InitBlockDAG(phantom, "PH_fig2-blocks")
	if ibd == nil {
		t.FailNow()
	}
	if bd.GetMaturity(tbMap["D"].GetID(), []uint{tbMap["I"].GetID()}) != 2 {
		t.Fatal()
	}
}

func Test_GetMainParentConcurrency(t *testing.T) {
	ibd := InitBlockDAG(phantom, "PH_fig2-blocks")
	if ibd == nil {
		t.FailNow()
	}

	//ph:=ibd.(*Phantom)
	anBlock := bd.GetBlock(tbMap[testData.PH_MPConcurrency.Input].GetHash())
	//fmt.Println(bd.GetMainParentConcurrency(anBlock))
	if bd.GetMainParentConcurrency(anBlock) != testData.PH_MPConcurrency.Output {
		t.Fatal()
	}
}

func Test_GetBlockConcurrency(t *testing.T) {
	ibd := InitBlockDAG(phantom, "PH_fig2-blocks")
	if ibd == nil {
		t.FailNow()
	}

	//ph:=ibd.(*Phantom)
	blueNum, err := bd.GetBlockConcurrency(tbMap[testData.PH_MPConcurrency.Input].GetHash())
	if err != nil {
		t.Fatal(err)
	}
	if blueNum != uint(testData.PH_BConcurrency.Output) {
		t.Fatal()
	}
}
