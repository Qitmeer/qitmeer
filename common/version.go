// Copyright (c) 2017-2019 The Qitmeer developers
//
// Use of this source code is governed by an ISC
// license that can be found in the LICENSE file.

// The parts code inspired & originated from
// http://github.com/ethereum/go-ethereum/common/size.go

package common

import (
	"encoding/binary"
)

func GetQitmeerBlockVersion(blockVersion uint32) uint32 {
	b := make([]byte, 4)
	newVersionData := make([]byte, 4)
	binary.LittleEndian.PutUint32(b, blockVersion)
	copy(newVersionData[:2], b[:2])
	return binary.LittleEndian.Uint32(newVersionData)
}
