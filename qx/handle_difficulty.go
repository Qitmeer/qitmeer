// Copyright 2017-2018 The qitmeer developers
// Use of this source code is governed by an ISC
// license that can be found in the LICENSE file.
package qx

import (
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