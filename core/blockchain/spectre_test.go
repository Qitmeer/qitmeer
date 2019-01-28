package blockchain

import (
	"testing"
	"github.com/noxproject/nox/common/hash"
	"fmt"
	"sort"
)

type SpectreBlockHelp struct {
	hash      hash.Hash
	parents   []*hash.Hash
	timeStamp int64
}

func (cb *SpectreBlockHelp) GetHash() *hash.Hash {
	return &cb.hash
}

// Get all parents set,the dag block has more than one parent
func (cb *SpectreBlockHelp) GetParents() []*hash.Hash {
	return cb.parents
}

func (cb *SpectreBlockHelp) GetTimestamp() int64 {
	return cb.timeStamp
}

//
var sp *Spectre

func InitSpectre() (bool, map[string]hash.Hash) {
	bd:=&BlockDAG{}
	sp=bd.Init(spectre).(*Spectre)
	tbMap := map[string]hash.Hash{}

	buildSpectreBlock("Gen", []*hash.Hash{},&tbMap)

	b1 := buildSpectreBlock("b1",[]*hash.Hash{bd.GetGenesisHash()},&tbMap)

	b2 := buildSpectreBlock("b2",[]*hash.Hash{bd.GetGenesisHash()},&tbMap)

	b3 := buildSpectreBlock("b3",[]*hash.Hash{bd.GetGenesisHash()},&tbMap)

	b4 := buildSpectreBlock("b4",[]*hash.Hash{b1.GetHash(), b2.GetHash(), b3.GetHash()},&tbMap)

	b5 := buildSpectreBlock("b5",[]*hash.Hash{b3.GetHash()},&tbMap)

	b6 := buildSpectreBlock("b6",[]*hash.Hash{b2.GetHash(), b3.GetHash()},&tbMap)

	b7 := buildSpectreBlock("b7",[]*hash.Hash{b4.GetHash()},&tbMap)

	b8 := buildSpectreBlock("b8",[]*hash.Hash{b4.GetHash(), b5.GetHash()},&tbMap)

	b9 := buildSpectreBlock("b9",[]*hash.Hash{b5.GetHash(), b6.GetHash()},&tbMap)

	b10 := buildSpectreBlock("b10",[]*hash.Hash{b6.GetHash()},&tbMap)

	b11 := buildSpectreBlock("b11",[]*hash.Hash{b7.GetHash(), b8.GetHash()},&tbMap)

	b12 := buildSpectreBlock("b12",[]*hash.Hash{b9.GetHash(), b10.GetHash()},&tbMap)

	b13 := buildSpectreBlock("b13",[]*hash.Hash{b9.GetHash(), b11.GetHash()},&tbMap)

	fmt.Printf("tips: %v %v\n", b12, b13)

	return true, tbMap
}

func buildSpectreBlock(tag string, parents []*hash.Hash, tbMap *map[string]hash.Hash) *SpectreBlockHelp {
	tempHash++
	hashStr:=fmt.Sprintf("%d",tempHash)
	h:=hash.MustHexToDecodedHash(hashStr)
	spBlock := &SpectreBlockHelp{hash: h}
	spBlock.parents = parents
	sp.bd.AddBlock(spBlock)
	(*tbMap)[tag] = *spBlock.GetHash()
	//
	return spBlock
}

func Log() {
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

func TestSpectre1(t *testing.T) {
	ret, tbMap := InitSpectre()
	if !ret {
		t.FailNow()
	}
	b4h:=tbMap["b4"]
	b6h:=tbMap["b6"]
	b4, b6 := sp.bd.GetBlock(&b4h), sp.bd.GetBlock(&b6h)

	if ret, err := sp.Vote(b4, b6); err != nil {
		t.Error(err.Error())
	} else {
		Log()
		if !ret {
			t.Error(ret)
		}
	}
}
