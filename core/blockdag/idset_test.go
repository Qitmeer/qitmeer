/*
 * Copyright (c) 2020.
 * Project:qitmeer
 * File:idset_test.go
 * Date:3/29/20 9:11 PM
 * Author:Jin
 * Email:lochjin@gmail.com
 */

package blockdag

import (
	"fmt"
	"github.com/Qitmeer/qng-core/common/hash"
	"testing"
)

func Test_AddId(t *testing.T) {
	hs := NewIdSet()
	hs.Add(1)

	if !hs.Has(1) {
		t.FailNow()
	}
}

func Test_AddSetId(t *testing.T) {
	hs := NewIdSet()
	other := NewIdSet()
	other.Add(1)

	hs.AddSet(other)
	if !hs.Has(1) {
		t.FailNow()
	}
}

func Test_AddPairId(t *testing.T) {
	var intData int = 123
	hs := NewIdSet()
	hs.AddPair(1, int(intData))

	if !hs.Has(1) || hs.Get(1).(int) != intData {
		t.FailNow()
	}
}

func Test_RemoveId(t *testing.T) {
	hs := NewIdSet()
	hs.Add(1)
	hs.Remove(1)

	if hs.Has(1) {
		t.FailNow()
	}
}

func Test_RemoveSetId(t *testing.T) {
	hs := NewIdSet()
	other := NewIdSet()
	other.Add(1)

	hs.AddSet(other)
	hs.RemoveSet(other)

	if hs.Has(1) {
		t.FailNow()
	}
}

func Test_SortListId(t *testing.T) {
	hs := NewIdSet()
	hl := IdSlice{}
	var hashNum uint = 5
	for i := uint(0); i < hashNum; i++ {
		hs.Add(i)
		hl = append(hl, i)
	}
	shs := hs.SortList(false)

	for i := uint(0); i < hashNum; i++ {
		if hl[i] != shs[i] {
			t.FailNow()
		}
	}
	rshs := hs.SortList(true)

	for i := uint(0); i < hashNum; i++ {
		if hl[i] != rshs[hashNum-i-1] {
			t.FailNow()
		}
	}
}

func Test_SortListHash(t *testing.T) {
	hs := NewIdSet()
	hl := BlockHashSlice{}
	var hashNum uint = 5
	for i := uint(0); i < hashNum; i++ {
		hashStr := fmt.Sprintf("%d", i)
		h := hash.MustHexToDecodedHash(hashStr)
		block := &Block{id: i, hash: h}
		hs.AddPair(block.GetID(), block)
		hl = append(hl, block)
	}
	shs := hs.SortHashList(false)

	for i := uint(0); i < hashNum; i++ {
		if hl[i].GetID() != shs[i] {
			t.FailNow()
		}
	}
	rshs := hs.SortHashList(true)

	for i := uint(0); i < hashNum; i++ {
		if hl[i].GetID() != rshs[hashNum-i-1] {
			t.FailNow()
		}
	}
}

func Test_ForId(t *testing.T) {
	hs := NewIdSet()
	var hashNum uint = 5
	for i := uint(0); i < hashNum; i++ {
		hs.AddPair(i, i)
	}
	for k, v := range hs.GetMap() {
		fmt.Printf("%d - %d\n", v, k)
	}
}

func Test_SortListPriority(t *testing.T) {
	hs := NewIdSet()
	hl := BlockPrioritySlice{}
	var hashNum uint = 5
	for i := uint(0); i < hashNum; i++ {
		hashStr := fmt.Sprintf("%d", i)
		h := hash.MustHexToDecodedHash(hashStr)
		block := &PhantomBlock{Block: &Block{id: i, hash: h, data: &TestBlock{}}, blueNum: i}
		hs.AddPair(block.GetID(), block)
		hl = append(hl, block)
	}

	shs := hs.SortPriorityList(false)

	for i := uint(0); i < hashNum; i++ {
		if hl[i].GetID() != shs[i] {
			t.FailNow()
		}
	}
	rshs := hs.SortPriorityList(true)

	for i := uint(0); i < hashNum; i++ {
		if hl[i].GetID() != rshs[hashNum-i-1] {
			t.FailNow()
		}
	}
}
