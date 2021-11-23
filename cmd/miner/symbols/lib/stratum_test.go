// Copyright (c) 2019 The qitmeer developers
// Use of this source code is governed by an ISC
// license that can be found in the LICENSE file.
package lib

import (
	"encoding/hex"
	"fmt"
	qitmeer "github.com/Qitmeer/qng-core/common/hash"
	"testing"
)

func TestMerkle(t *testing.T) {
	coinbaseTx, _ := hex.DecodeString("56407afb2f3dcaa084d367d6f4c6c5f396f0d683a1ad9df0f3c44ae4e06c9f08")
	merkle_root := string(coinbaseTx)
	branches := []string{"661b5b82f7c456b844b2db610f64ce8a5629942eb2c87c412ca2d280e84140c4",
		"9eeba4ac0fa36c3459897e4fe898f25d81762f4f4590e81aacda0606d37e6e7c",
	}
	for _, h := range branches {
		d, _ := hex.DecodeString(h)
		bs := merkle_root + string(d)
		merkle_root = string(qitmeer.DoubleHashB([]byte(bs)))
	}
	merkleRootStr := hex.EncodeToString([]byte(merkle_root))
	fmt.Println("merkleRootStr", merkleRootStr)
	ddd, _ := hex.DecodeString(merkleRootStr)
	fmt.Println("ddd", ddd)
}
