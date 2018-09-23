// Copyright 2017-2018 The nox developers
// Use of this source code is governed by an ISC
// license that can be found in the LICENSE file.
package main

import (
	"fmt"
	"github.com/noxproject/nox/crypto/seed"
)

func newSeed(size uint) {
	s,err :=seed.GenerateSeed(uint16(size))
	if err!=nil {
		errExit(err)
	}
	fmt.Printf("%x\n",s)
}
