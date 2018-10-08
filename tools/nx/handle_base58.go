// Copyright 2017-2018 The nox developers
// Use of this source code is governed by an ISC
// license that can be found in the LICENSE file.
package main

import (
	"encoding/hex"
	"fmt"
	"github.com/noxproject/nox/common/encode/base58"
	"github.com/noxproject/nox/common/hash"
)

func base58CheckEncode(version string, mode string,hasher string, cksumSize int, input string){
	if hasher != "" && mode != "nox" {
		errExit(fmt.Errorf("invaid flag -a %s with -m %s",hasher,mode))
	}
	data, err := hex.DecodeString(input)
	if err!=nil {
		errExit(err)
	}
	ver, err := hex.DecodeString(version)
	if err !=nil {
		errExit(err)
	}
	var encoded string

	if hasher != "" {
		var cksumfunc func([]byte) []byte
		switch (hasher) {
		case "sha256":
			cksumfunc = base58.SingleHashChecksumFunc(hash.GetHasher(hash.SHA256), cksumSize)
		case "dsha256":
			cksumfunc = base58.DoubleHashChecksumFunc(hash.GetHasher(hash.SHA256), cksumSize)
		case "blake2b256":
			cksumfunc = base58.SingleHashChecksumFunc(hash.GetHasher(hash.Blake2b_256), cksumSize)
		case "dblake2b256":
			cksumfunc = base58.DoubleHashChecksumFunc(hash.GetHasher(hash.Blake2b_256), cksumSize)
		case "blake2b512":
			cksumfunc = base58.SingleHashChecksumFunc(hash.GetHasher(hash.Blake2b_512), cksumSize)
		default:
			err = fmt.Errorf("unknown hasher %s", hasher)
		}
		encoded = base58.CheckEncode(data, ver, cksumSize, cksumfunc)
	}else {
		switch mode {
		case "nox":
			if len(ver) != 2 {
				errExit(fmt.Errorf("invaid version byte size"))
			}
			var versionByte [2]byte
			versionByte[0] = ver[0]
			versionByte[1] = ver[1]
			encoded = base58.NoxCheckEncode(data, versionByte[:])
		case "btc":
			if len(ver) > 1 {
				errExit(fmt.Errorf("invaid version size for btc base58check encode"))
			}
			encoded = base58.BtcCheckEncode(data, ver[0])
		case "ss":
			encoded = base58.CheckEncode(data, ver[:], 2, base58.SingleHashChecksumFunc(hash.GetHasher(hash.Blake2b_512), 2))
		default:
			errExit(fmt.Errorf("unknown encode mode %s", mode))
		}
	}
	// Show the encoded data.
	//fmt.Printf("Encoded Data ver[%v] : %s\n",ver, encoded)
	fmt.Printf("%s\n",encoded)
}

func base58CheckDecode(mode, hasher string, versionSize, cksumSize int, input string) {
	var err error
	var data []byte
	var version []byte
	if hasher != "" && mode != "nox" {
		errExit(fmt.Errorf("invaid flag -a %s with -m %s",hasher,mode))
	}
	if hasher != "" {
		var v []byte
		switch hasher {
		case "sha256":
			data, v, err = base58.CheckDecode(input, versionSize, cksumSize, base58.SingleHashChecksumFunc(hash.GetHasher(hash.SHA256), cksumSize))
		case "dsha256":
			data, v, err = base58.CheckDecode(input, versionSize, cksumSize, base58.DoubleHashChecksumFunc(hash.GetHasher(hash.SHA256), cksumSize))
		case "blake2b256":
			data, v, err = base58.CheckDecode(input, versionSize, cksumSize, base58.SingleHashChecksumFunc(hash.GetHasher(hash.Blake2b_256), cksumSize))
		case "dblake2b256":
			data, v, err = base58.CheckDecode(input, versionSize, cksumSize, base58.DoubleHashChecksumFunc(hash.GetHasher(hash.Blake2b_256), cksumSize))
		case "blake2b512":
			data, v, err = base58.CheckDecode(input, versionSize, cksumSize, base58.SingleHashChecksumFunc(hash.GetHasher(hash.Blake2b_512), cksumSize))
		default:
			err = fmt.Errorf("unknown hasher %s",hasher)
		}
		if err!=nil {
			errExit(err)
		}
		version = v
	}else {
		switch mode {
		case "btc":
			v := byte(0)
			data, v, err = base58.BtcCheckDecode(input)
			if err != nil {
				errExit(err)
			}
			version = []byte{0x0, v}
		case "nox":
			v := [2]byte{}
			data, v, err = base58.NoxCheckDecode(input)
			if err != nil {
				errExit(err)
			}
			version = []byte{v[0], v[1]}
		case "ss":
			var v []byte
			data, v, err = base58.CheckDecode(input, 1, 2, base58.SingleHashChecksumFunc(hash.GetHasher(hash.Blake2b_512), 2))
			if err != nil {
				errExit(err)
			}
			version = v
		default:
			errExit(fmt.Errorf("unknown mode %s", mode))
		}
	}
	if showDecodeDetails {
		decoded := base58.Decode(input)
		if hasher!="" {
			fmt.Printf("hasher  : %s\n", hasher)
		}else {
			fmt.Printf("mode    : %s\n", mode)
		}
		fmt.Printf("version : %x\n", version)
		fmt.Printf("payload : %x\n", data)
		fmt.Printf("checksum: %x\n", decoded[len(decoded)-cksumSize:])
	} else {
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

