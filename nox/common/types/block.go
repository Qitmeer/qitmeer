// Copyright 2017-2018 The nox developers

package types

type BlockHeader struct {
	parents []Hash
}

type Block struct {
	header  BlockHeader
	payload []byte
}

type BlockLink struct{
	from Block
	to   Block
}

type BlockSet map[Hash]bool
