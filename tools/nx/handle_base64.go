// Copyright 2017-2018 The nox developers
// Use of this source code is governed by an ISC
// license that can be found in the LICENSE file.
package main

import (
	"encoding/base64"
	"encoding/hex"
	"fmt"
)

func base64Encode(input string){
	data, err := hex.DecodeString(input)
	if err!=nil {
		errExit(err)
	}
	encoded := base64.StdEncoding.EncodeToString(data)
	fmt.Printf("%s\n",encoded)
}

func base64Decode(input string){
	data, err := base64.StdEncoding.DecodeString(input)
	if err != nil {
		errExit(err)
	}
	fmt.Printf("%x\n", data)
}
