// Copyright 2017-2018 The nox developers
// Use of this source code is governed by an ISC
// license that can be found in the LICENSE file.
package main

import (
	"bytes"
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"github.com/noxproject/nox/common/encode/base58"
)

func base58CheckEncode(version string, input string){
	data, err := hex.DecodeString(input)
	if err!=nil {
		errExit(err)
	}
	ver, err := hex.DecodeString(version)
	if err !=nil {
		errExit(err)
	}
	/*
	if len(ver) != 2 {
		errExit(fmt.Errorf("invaid version byte"))
	}
	var versionByte [2]byte
	versionByte[0] = ver[0]
	versionByte[1] = ver[1]
	*/
	encoded := base58.CheckEncode(data, ver[:])
	// Show the encoded data.
	//fmt.Printf("Encoded Data ver[%v] : %s\n",ver, encoded)
	fmt.Printf("%s\n",encoded)
}

func base58CheckDecode(mode string, input string) {
	var err error
	var data []byte
	var version []byte
	switch mode {
	case "btc" :
		v := byte(0)
		data, v, err = base58.BtcCheckDecode(input)
		if err != nil {
			errExit(err)
		}
		version = []byte{0x0,v}
	default:
		v := [2]byte{}
		data, v, err = base58.CheckDecode(input)
		if err != nil {
			errExit(err)
		}
		version = []byte{v[0],v[1]}
	}

	if showDecodeDetails {
		cksum, err := base58.CheckInput(mode,input)
		if err != nil {
			errExit(err)
		}
		fmt.Printf("mode    : %s\n", mode)
		fmt.Printf("payload : %x\n", data)
		var dec_l uint32
		var dec_b uint32
		// the default parse string use bigEndian
		// dec,err := strconv.ParseUint(hex.EncodeToString(cksum), 16, 64)
		buff :=  bytes.NewReader(cksum)
		err = binary.Read(buff, binary.LittleEndian, &dec_l)
		if err!=nil {
			errExit(err)
		}
		buff =  bytes.NewReader(cksum)
		err = binary.Read(buff, binary.BigEndian, &dec_b)
		if err!=nil {
			errExit(err)
		}
		fmt.Printf("checksum: %d (le) %d (be) %x (hex)\n",dec_l, dec_b, cksum)
		fmt.Printf("version : %x\n",version)
	}else {
		fmt.Printf("%x\n", data)
	}
}


func base58Encode(input string){
	data, err := hex.DecodeString(input)
	if err!=nil {
		errExit(err)
	}
	encoded := base58.Encode(data)
	fmt.Printf("%s\n",encoded)
}

func base58Decode(input string){
	data := base58.Decode(input)
	fmt.Printf("%x\n", data)
}

