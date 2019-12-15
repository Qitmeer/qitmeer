// Copyright 2017-2018 The qitmeer developers
// Use of this source code is governed by an ISC
// license that can be found in the LICENSE file.
package qx

import (
	`errors`
	"fmt"
	`github.com/Qitmeer/qitmeer/core/types/pow`
	`math/big`
	"strconv"
)

func CompactToUint64(input string) {
	u32,err := strconv.ParseUint(input,10,32)
	if err != nil {
		ErrExit(err)
	}
	diffBig := pow.CompactToBig(uint32(u32))
	fmt.Printf("%d\n", diffBig.Uint64())
}

func Uint64ToCompact(input string) {
	u64,err := strconv.ParseUint(input,10,64)
	if err != nil {
		ErrExit(err)
	}
	diffBig := &big.Int{}
	diffBig.SetUint64(u64)
	diffCompact := pow.BigToCompact(diffBig)
	fmt.Printf("%d\n", diffCompact)
}

func CompactToGPS(compact string,edgeBits string,blockTime string) {
	u32,err := strconv.ParseUint(compact,10,32)
	if err != nil {
		ErrExit(err)
	}
	diffBig := pow.CompactToBig(uint32(u32))
	edgeBitsU32,err := strconv.ParseUint(edgeBits,10,32)
	if err != nil {
		ErrExit(err)
	}
	scale := pow.GraphWeight(uint32(edgeBitsU32))
	if scale <= 0{
		ErrExit(errors.New("edgeBits must between 24-32"))
	}
	blockTimeU32,err := strconv.ParseUint(blockTime,10,32)
	if err != nil {
		ErrExit(err)
	}
	if blockTimeU32 <= 0{
		ErrExit(errors.New("blockTime must bigger than 0"))
	}
	needGPS := float64(diffBig.Uint64()) / float64(scale) * 50.00 / float64(blockTimeU32)
	fmt.Printf("The difficulty at least need hashrate :%f GPS\n", needGPS)
}