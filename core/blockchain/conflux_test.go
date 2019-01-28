package blockchain

import (
	"fmt"
	"testing"
	"github.com/noxproject/nox/common/hash"
)

type ConfluxBlock struct {
	hash      hash.Hash
	parents   []*hash.Hash
	timeStamp int64
}

func (cb *ConfluxBlock) GetHash() *hash.Hash {
	return &cb.hash
}

// Get all parents set,the dag block has more than one parent
func (cb *ConfluxBlock) GetParents() []*hash.Hash {
	return cb.parents
}

func (cb *ConfluxBlock) GetTimestamp() int64 {
	return cb.timeStamp
}

// This is the interface for DAG,can use to call public function.
var con *Conflux
var tempHash int

func InitConflux() (bool, map[hash.Hash]string) {

	bd:=&BlockDAG{}
	con=bd.Init(conflux).(*Conflux)
	tbMap := map[hash.Hash]string{}

	Gen := buildConfluxBlock("Gen", []*hash.Hash{}, &tbMap)

	A := buildConfluxBlock("A", []*hash.Hash{Gen.GetHash()}, &tbMap)
	B := buildConfluxBlock("B", []*hash.Hash{Gen.GetHash()}, &tbMap)
	C := buildConfluxBlock("C", []*hash.Hash{A.GetHash(), B.GetHash()}, &tbMap)
	D := buildConfluxBlock("D", []*hash.Hash{A.GetHash()}, &tbMap)
	F := buildConfluxBlock("F", []*hash.Hash{B.GetHash()}, &tbMap)
	E := buildConfluxBlock("E", []*hash.Hash{C.GetHash(), D.GetHash(), F.GetHash()}, &tbMap)
	G := buildConfluxBlock("G", []*hash.Hash{A.GetHash()}, &tbMap)
	J := buildConfluxBlock("J", []*hash.Hash{F.GetHash()}, &tbMap)
	I := buildConfluxBlock("I", []*hash.Hash{J.GetHash(), C.GetHash()}, &tbMap)

	buildConfluxBlock("H", []*hash.Hash{E.GetHash(), G.GetHash(), I.GetHash()}, &tbMap)
	buildConfluxBlock("K", []*hash.Hash{I.GetHash()}, &tbMap)

	return true, tbMap
}

func buildConfluxBlock(tag string, parents []*hash.Hash, tbMap *map[hash.Hash]string) *ConfluxBlock {
	tempHash++
	hashStr:=fmt.Sprintf("%d",tempHash)
	h:=hash.MustHexToDecodedHash(hashStr)
	conBlock := &ConfluxBlock{hash: h}
	conBlock.parents = parents
	con.bd.AddBlock(conBlock)
	(*tbMap)[*conBlock.GetHash()] = tag
	//
	return conBlock
}

func Test_GetMainChain(t *testing.T) {
	ret, tbMap := InitConflux()
	if !ret {
		t.FailNow()
	}
	var result string
	fmt.Println("Conflux main chain ：")
	mainChain := con.GetMainChain()
	for i := len(mainChain) - 1; i >= 0; i-- {
		name := tbMap[*mainChain[i]]
		if i == 0 {
			result += fmt.Sprintf("%s", name)
		} else {
			result += fmt.Sprintf("%s-->", name)
		}

	}
	fmt.Println(result)

	if result == "Gen-->A-->C-->E-->H" {
		fmt.Println("Success!")
	} else {
		t.FailNow()
	}

	fmt.Println()
	fmt.Println()
}

func Test_Order(t *testing.T) {
	ret, tbMap := InitConflux()
	if !ret {
		t.FailNow()
	}
	var result string
	fmt.Println("Conflux order ：")
	order := con.bd.GetOrder()
	for i := 0; i < len(order); i++ {
		name := tbMap[*order[i]]
		if i == len(order)-1 {
			result += fmt.Sprintf("%s", name)
		} else {
			result += fmt.Sprintf("%s-->", name)
		}

	}
	fmt.Println(result)
	if result == "Gen-->A-->B-->C-->D-->F-->E-->G-->J-->I-->H-->K" {
		fmt.Println("Success!")
	} else {
		t.FailNow()
	}

	fmt.Println()
	fmt.Println()
}
