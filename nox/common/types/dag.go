// Copyright 2017-2018 The nox developers

package types

type BlockDag struct {
	genesis       Block
	blocks        map[Hash]Block
	tips          BlockSet
	blueSet       BlockSet
	parentsIndex  map[Hash]BlockSet
	childIndex    map[Hash]BlockSet
	anticoneIndex map[Hash]BlockSet
	scoreMap      map[Hash]uint64
}
