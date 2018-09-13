// Copyright 2017-2018 The nox developers

package types

import "github.com/noxproject/nox/common/hash"

type BlockSet struct {
	m map[hash.Hash]bool
}

type BlockDag struct {
	genesis       Block
	blocks        map[hash.Hash]Block
	tips          BlockSet
	blueSet       BlockSet
	parentsIndex  map[hash.Hash]BlockSet
	childIndex    map[hash.Hash]BlockSet
	anticoneIndex map[hash.Hash]BlockSet
	scoreMap      map[hash.Hash]uint64
}
