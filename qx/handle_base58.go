// Copyright 2017-2018 The qitmeer developers
// Use of this source code is governed by an ISC
// license that can be found in the LICENSE file.
package qx

import (
	"encoding/hex"
	"fmt"
	"github.com/Qitmeer/qng-core/common/encode/base58"
	"github.com/Qitmeer/qng-core/common/hash"
	"github.com/Qitmeer/qng-core/common/util"
	"strconv"
)

func Base58CheckEncode(version []byte, mode string, hasher string, cksumSize int, input string) {
	if hasher != "" && mode != "qitmeer" {
		ErrExit(fmt.Errorf("invaid flag -a %s with -m %s", hasher, mode))
	}
	data, err := hex.DecodeString(input)
	if err != nil {
		ErrExit(err)
	}
	var encoded []byte

	if hasher != "" {
		var cksumfunc func([]byte) []byte
		switch hasher {
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
		if err != nil {
			ErrExit(err)
		}
		encoded, err = base58.CheckEncode(data, version, cksumSize, cksumfunc)
	} else {
		switch mode {
		case "qitmeer":
			if len(version) != 2 {
				ErrExit(fmt.Errorf("invaid version byte size for qitmeer base58 check encode. input = %x (len = %d, required 2)", version, len(version)))
			}
			encoded, err = base58.QitmeerCheckEncode(data, version[:])
		case "btc":
			if len(version) > 1 {
				ErrExit(fmt.Errorf("invaid version size for btc base58check encode"))
			}
			encoded, err = base58.BtcCheckEncode(data, version[0])
		case "ss":
			encoded, err = base58.CheckEncode(data, version[:], 2, base58.SingleHashChecksumFunc(hash.GetHasher(hash.Blake2b_512), 2))
		default:
			ErrExit(fmt.Errorf("unknown encode mode %s", mode))
		}
	}
	// Show the encoded data.
	//fmt.Printf("Encoded Data ver[%v] : %s\n",ver, encoded)
	if err != nil {
		ErrExit(err)
	} else {
		fmt.Printf("%s\n", encoded)
	}
}

func Base58CheckDecode(mode, hasher string, versionSize, cksumSize int, input string, showDetails bool) {
	var err error
	var data []byte
	var version []byte
	if hasher != "" && mode != "qitmeer" {
		ErrExit(fmt.Errorf("invaid flag -a %s with -m %s", hasher, mode))
	}
	if hasher != "" {
		var v []byte
		switch hasher {
		case "sha256":
			data, v, err = base58.CheckDecode([]byte(input), versionSize, cksumSize, base58.SingleHashChecksumFunc(hash.GetHasher(hash.SHA256), cksumSize))
		case "dsha256":
			data, v, err = base58.CheckDecode([]byte(input), versionSize, cksumSize, base58.DoubleHashChecksumFunc(hash.GetHasher(hash.SHA256), cksumSize))
		case "blake2b256":
			data, v, err = base58.CheckDecode([]byte(input), versionSize, cksumSize, base58.SingleHashChecksumFunc(hash.GetHasher(hash.Blake2b_256), cksumSize))
		case "dblake2b256":
			data, v, err = base58.CheckDecode([]byte(input), versionSize, cksumSize, base58.DoubleHashChecksumFunc(hash.GetHasher(hash.Blake2b_256), cksumSize))
		case "blake2b512":
			data, v, err = base58.CheckDecode([]byte(input), versionSize, cksumSize, base58.SingleHashChecksumFunc(hash.GetHasher(hash.Blake2b_512), cksumSize))
		default:
			err = fmt.Errorf("unknown hasher %s", hasher)
		}
		if err != nil {
			ErrExit(err)
		}
		version = v
	} else {
		switch mode {
		case "btc":
			v := byte(0)
			data, v, err = base58.BtcCheckDecode(input)
			if err != nil {
				ErrExit(err)
			}
			version = []byte{0x0, v}
		case "qitmeer":
			v := [2]byte{}
			data, v, err = base58.QitmeerCheckDecode(input)
			if err != nil {
				ErrExit(err)
			}
			version = []byte{v[0], v[1]}
		case "ss":
			var v []byte
			data, v, err = base58.CheckDecode([]byte(input), 1, 2, base58.SingleHashChecksumFunc(hash.GetHasher(hash.Blake2b_512), 2))
			if err != nil {
				ErrExit(err)
			}
			version = v
		default:
			ErrExit(fmt.Errorf("unknown mode %s", mode))
		}
	}
	if showDetails {
		decoded := base58.Decode([]byte(input))
		if hasher != "" {
			fmt.Printf("hasher  : %s\n", hasher)
		} else {
			fmt.Printf("mode    : %s\n", mode)
		}
		version_d, err := strconv.ParseUint(fmt.Sprintf("%x", version[:]), 16, 64)
		if err != nil {
			ErrExit(err)
		}
		version_r := util.CopyBytes(version[:])
		util.ReverseBytes(version_r)
		version_d2, err := strconv.ParseUint(fmt.Sprintf("%x", version_r[:]), 16, 64)
		if err != nil {
			ErrExit(fmt.Errorf("convert version %x error : %w", version[:], err))
		}
		fmt.Printf("version : %x (hex) %v (BE) %v (LE)\n", version, version_d, version_d2)
		fmt.Printf("payload : %x\n", data)
		cksum := decoded[len(decoded)-cksumSize:]
		cksum_d, err := strconv.ParseUint(fmt.Sprintf("%x", cksum[:]), 16, 64)
		if err != nil {
			ErrExit(fmt.Errorf("convert version %x error : %w", cksum[:], err))
		}
		//convere to  little endian
		cksum_r := util.CopyBytes(cksum[:])
		util.ReverseBytes(cksum_r)
		cksum_d2, err := strconv.ParseUint(fmt.Sprintf("%x", cksum_r[:]), 16, 64)
		if err != nil {
			ErrExit(err)
		}
		fmt.Printf("checksum: %x (hex) %v (BE) %v (LE)\n", cksum, cksum_d, cksum_d2)

	} else {
		fmt.Printf("%x\n", data)
	}
}

func Base58Encode(input string) {
	data, err := hex.DecodeString(input)
	if err != nil {
		ErrExit(err)
	}
	encoded, _ := base58.Encode(data)
	fmt.Printf("%s\n", encoded)
}

func Base58Decode(input string) {
	data := base58.Decode([]byte(input))
	fmt.Printf("%x\n", data)
}
