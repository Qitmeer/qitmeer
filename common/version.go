// Copyright (c) 2017-2019 The Qitmeer developers
//
package common

import (
	"encoding/binary"
)

//To maintain version
//use the first 2 bytes for version
//last 2 bytes use for rand nonce
func GetQitmeerBlockVersion(blockVersion uint32) uint32 {
	b := make([]byte, 4)
	newVersionData := make([]byte, 4)
	binary.LittleEndian.PutUint32(b, blockVersion)
	copy(newVersionData[:2], b[:2])
	return binary.LittleEndian.Uint32(newVersionData)
}
