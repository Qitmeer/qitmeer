// Copyright 2017-2018 The qitmeer developers
// Use of this source code is governed by an ISC
// license that can be found in the LICENSE file.
package qx

import (
	"errors"
	"fmt"
	"github.com/Qitmeer/qitmeer/core/types/pow"
	"math/big"
	"strconv"
	"strings"
)

func CompactToUint64(input string) {
	u32, err := strconv.ParseUint(input, 10, 32)
	if err != nil {
		ErrExit(err)
	}
	diffBig := pow.CompactToBig(uint32(u32))
	fmt.Printf("%d\n", diffBig.Uint64())
}

func HashCompactToDiff(input string) {
	u32, err := strconv.ParseUint(input, 10, 32)
	if err != nil {
		ErrExit(err)
	}
	diffBig := pow.CompactToBig(uint32(u32))
	maxBig, _ := new(big.Int).SetString("ffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff", 16)
	maxBig.Div(maxBig, diffBig)
	fmt.Printf("%d\n", maxBig)
}

func Uint64ToCompact(input string) {
	u64, err := strconv.ParseUint(input, 10, 64)
	if err != nil {
		ErrExit(err)
	}
	diffBig := &big.Int{}
	diffBig.SetUint64(u64)
	diffCompact := pow.BigToCompact(diffBig)
	fmt.Printf("%d\n", diffCompact)
}

func CompactToGPS(diff string, blockTime, scale int) {
	u64, err := strconv.ParseUint(diff, 10, 64)
	if err != nil {
		ErrExit(err)
	}

	if u64 <= 0 {
		ErrExit(errors.New("diff must bigger than 0"))
	}
	if scale <= 0 {
		ErrExit(errors.New("edgeBits must between 24-32"))
	}
	if blockTime <= 0 {
		ErrExit(errors.New("blockTime must bigger than 0"))
	}
	needGPS := float64(u64) / float64(scale) * 50.00 / float64(blockTime)
	fmt.Printf("%f\n", needGPS)
}

func CompactToTarget(diffCompact string) {
	u64, err := strconv.ParseUint(diffCompact, 10, 64)
	if err != nil {
		ErrExit(err)
	}
	diffBig := pow.CompactToBig(uint32(u64))
	fmt.Printf("0x%064x\n", diffBig)
}

func TargetToCompact(target string) {
	target = strings.TrimPrefix(target, "0x")
	bigT, ok := new(big.Int).SetString(target, 16)
	if !ok {
		fmt.Println("target error")
		return
	}
	difftarget := pow.BigToCompact(bigT)
	fmt.Printf("%d\n", difftarget)
}

func CompactToHashrate(diffCompact string, blocktime int) {
	u64, err := strconv.ParseUint(diffCompact, 10, 64)
	if err != nil {
		ErrExit(err)
	}
	diffBig := pow.CompactToBig(uint32(u64))
	maxBig, _ := new(big.Int).SetString("ffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff", 16)
	needAtleasthashrate := maxBig.Div(maxBig, diffBig)
	needAtleasthashrate = needAtleasthashrate.Div(needAtleasthashrate, big.NewInt(int64(blocktime)))
	fmt.Printf("%s\n", GetHashrate(needAtleasthashrate))
}

func HashrateToCompact(hashrate string) {
	hashrateBig, _ := new(big.Int).SetString(hashrate, 10)
	maxBig, _ := new(big.Int).SetString("ffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff", 16)
	maxBig.Div(maxBig, hashrateBig)
	compact := pow.BigToCompact(maxBig)
	fmt.Printf("%d\n", compact)
}

func GPSToDiff(gps string, blockTime, scale int) {
	f64, err := strconv.ParseFloat(gps, 64)
	if err != nil {
		ErrExit(err)
	}
	if scale <= 0 {
		ErrExit(errors.New("edgeBits must between 24-32"))
	}
	diff := f64 * float64(blockTime) / 50 * float64(scale)
	fmt.Printf("%d\n", uint64(diff))
}

func GetHashrate(hashBig *big.Int) string {
	if hashBig.Cmp(big.NewInt(1000)) <= 0 {
		return fmt.Sprintf("%d H/s", hashBig.Uint64())
	}
	if hashBig.Cmp(big.NewInt(1000000)) <= 0 {
		f := new(big.Float).SetInt(hashBig)
		f.Quo(f, big.NewFloat(1000))
		return fmt.Sprintf("%s KH/s", f.String())
	}
	if hashBig.Cmp(big.NewInt(1000000000)) <= 0 {
		f := new(big.Float).SetInt(hashBig)
		f.Quo(f, big.NewFloat(1000000))
		return fmt.Sprintf("%s MH/s", f.String())
	}
	if hashBig.Cmp(big.NewInt(1000000000000)) <= 0 {
		f := new(big.Float).SetInt(hashBig)
		f.Quo(f, big.NewFloat(1000000000))
		return fmt.Sprintf("%s GH/s", f.String())
	}
	base, _ := new(big.Int).SetString("1000000000000000", 10)
	if hashBig.Cmp(base) <= 0 {
		f := new(big.Float).SetInt(hashBig)
		base1, _ := new(big.Int).SetString("1000000000000", 10)
		f.Quo(f, new(big.Float).SetInt(base1))
		return fmt.Sprintf("%s TH/s", f.String())
	}
	base, _ = new(big.Int).SetString("1000000000000000000", 10)
	if hashBig.Cmp(base) <= 0 {
		f := new(big.Float).SetInt(hashBig)
		base1, _ := new(big.Int).SetString("1000000000000000", 10)
		f.Quo(f, new(big.Float).SetInt(base1))
		return fmt.Sprintf("%s PH/s", f.String())
	}
	f := new(big.Float).SetInt(hashBig)
	f.Quo(f, new(big.Float).SetInt(base))
	return fmt.Sprintf("%s EH/s", f.String())
}
