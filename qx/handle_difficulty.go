// Copyright 2017-2018 The qitmeer developers
// Use of this source code is governed by an ISC
// license that can be found in the LICENSE file.
package qx

import (
	"errors"
	"fmt"
	"github.com/Qitmeer/qng-core/common/hash"
	"github.com/Qitmeer/qng-core/core/types/pow"
	"math/big"
	"strconv"
	"strings"
)

func CompactToTarget(input, powtype string) {
	u32, err := strconv.ParseUint(input, 10, 32)
	if err != nil {
		ErrExit(err)
	}
	diffBig := pow.CompactToBig(uint32(u32))
	switch powtype {
	case "hash":
		fmt.Printf("0x%064x\n", diffBig)
	case "cuckoo24":
		target := pow.CuckooDiffToTarget(48, diffBig)
		fmt.Printf("0x%s\n", target)
	case "cuckoo29":
		target := pow.CuckooDiffToTarget(1856, diffBig)
		fmt.Printf("0x%s\n", target)
	default:
		ErrExit(errors.New("mode error!"))
	}
}

func TargetToCompact(input, powtype string) {
	input = strings.TrimPrefix(input, "0x")
	bigT, ok := new(big.Int).SetString(input, 16)
	if !ok {
		fmt.Println("error : invalid input, the target should be a hex string. ")
		return
	}
	switch powtype {
	case "hash":
		compact := pow.BigToCompact(bigT)
		fmt.Printf("%d\n", compact)
	case "cuckoo24":
		b := [32]byte{}
		copy(b[:], bigT.Bytes()[:])
		diffBig := pow.CalcCuckooDiff(48, hash.Hash(b))
		compact := pow.BigToCompact(diffBig)
		fmt.Printf("%d\n", compact)
	case "cuckoo29":
		b := [32]byte{}
		copy(b[:], bigT.Bytes()[:])
		diffBig := pow.CalcCuckooDiff(1856, hash.Hash(b))
		compact := pow.BigToCompact(diffBig)
		fmt.Printf("%d\n", compact)
	default:
		ErrExit(errors.New("mode error!"))
	}
}

func CompactToHashrate(input, unit string, printDetail bool, blocktime int) {
	u32, err := strconv.ParseUint(input, 10, 32)
	if err != nil {
		ErrExit(err)
	}
	diffBig := pow.CompactToBig(uint32(u32))
	maxBig, _ := new(big.Int).SetString("ffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff", 16)
	maxBig.Div(maxBig, diffBig)
	maxBig.Div(maxBig, big.NewInt(int64(blocktime)))
	val, u := GetHashrate(maxBig, unit)
	fmt.Printf("%s", val)
	if printDetail {
		fmt.Printf("%s", u)
	}
	fmt.Printf("\n")
}

func HashrateToCompact(difficulty string, blocktime int) {
	maxBig, _ := new(big.Int).SetString("ffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff", 16)
	diffBig, ok := new(big.Int).SetString(difficulty, 10)
	if !ok {
		fmt.Printf("error : invalid input %s, the hashrate should be a integer. \n", difficulty)
		return
	}
	maxBig.Div(maxBig, diffBig)
	maxBig.Div(maxBig, new(big.Int).SetInt64(int64(blocktime)))
	compact := pow.BigToCompact(maxBig)
	fmt.Printf("%d\n", compact)
}

func CompactToGPS(compactS string, blockTime, scale int, printDetail bool) {
	compact, err := strconv.ParseUint(compactS, 10, 32)
	if err != nil {
		ErrExit(err)
	}
	u64Big := pow.CompactToBig(uint32(compact))
	if u64Big.Uint64() <= 0 {
		ErrExit(errors.New("compact must bigger than 0"))
	}
	if scale <= 0 {
		ErrExit(errors.New("edgeBits must between 24-32"))
	}
	if blockTime <= 0 {
		ErrExit(errors.New("blockTime must bigger than 0"))
	}
	//2.2% graph found rate
	needGPS := float64(u64Big.Uint64()) / float64(scale) * 50.00 / float64(blockTime)
	fmt.Printf("%f", needGPS)
	if printDetail {
		fmt.Printf(" GPS")
	}
	fmt.Printf("\n")
}

func GPSToCompact(gps string, blockTime, scale int) {
	needGPS, err := strconv.ParseFloat(gps, 64)
	if err != nil {
		ErrExit(err)
	}
	if needGPS <= 0 {
		ErrExit(errors.New("gps must bigger than 0"))
	}
	if scale <= 0 {
		ErrExit(errors.New("edgeBits must between 24-32"))
	}
	if blockTime <= 0 {
		ErrExit(errors.New("blockTime must bigger than 0"))
	}
	//2.2% graph found rate
	f64 := needGPS * float64(scale) * float64(blockTime) / 50.00
	u64s := fmt.Sprintf("%.0f", f64)
	u64, _ := strconv.Atoi(u64s)
	u64Big := new(big.Int).SetInt64(int64(u64))
	compact := pow.BigToCompact(u64Big)
	fmt.Printf("%d\n", compact)
}

func GetHashrate(hashBig *big.Int, unit string) (string, string) {
	if unit == "H" {
		return fmt.Sprintf("%d", hashBig.Uint64()), " H/s"
	}
	if unit == "K" {
		f := new(big.Float).SetInt(hashBig)
		f.Quo(f, big.NewFloat(1000))
		return f.String(), " KH/s"
	}
	if unit == "M" {
		f := new(big.Float).SetInt(hashBig)
		f.Quo(f, big.NewFloat(1000000))
		return f.String(), " MH/s"
	}
	if unit == "G" {
		f := new(big.Float).SetInt(hashBig)
		f.Quo(f, big.NewFloat(1000000000))
		return f.String(), " GH/s"
	}
	if unit == "T" {
		f := new(big.Float).SetInt(hashBig)
		base1, _ := new(big.Int).SetString("1000000000000", 10)
		f.Quo(f, new(big.Float).SetInt(base1))
		return f.String(), " TH/s"
	}
	base, _ := new(big.Int).SetString("1000000000000000000", 10)
	if unit == "P" {
		f := new(big.Float).SetInt(hashBig)
		base1, _ := new(big.Int).SetString("1000000000000000", 10)
		f.Quo(f, new(big.Float).SetInt(base1))
		return f.String(), " PH/s"
	}
	f := new(big.Float).SetInt(hashBig)
	f.Quo(f, new(big.Float).SetInt(base))
	return f.String(), " EH/s"
}
