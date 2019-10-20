// Copyright (c) 2017-2020 The qitmeer developers
// license that can be found in the LICENSE file.
// Reference resources of rust bitVector
package pow

import (
	"encoding/binary"
	"errors"
	"fmt"
	"github.com/Qitmeer/qitmeer/common/hash"
	"math/big"
)

type Blake2bd struct {
	Pow
}

func (this *Blake2bd) Verify(headerData []byte, blockHash hash.Hash, targetDiffBits uint32, powConfig *PowConfig) error {
	if !this.CheckAvailable(this.PowPercent(powConfig)) {
		str := fmt.Sprintf("blake2bd is not supported")
		return errors.New(str)
	}
	target := CompactToBig(targetDiffBits)
	if target.Sign() <= 0 {
		str := fmt.Sprintf("block target difficulty of %064x is too "+
			"low", target)
		return errors.New(str)
	}

	//The target difficulty must be less than the maximum allowed.
	if target.Cmp(powConfig.Blake2bdPowLimit) > 0 {
		str := fmt.Sprintf("block target difficulty of %064x is "+
			"higher than max of %064x", target, powConfig.Blake2bdPowLimit)
		return errors.New(str)
	}
	hashNum := HashToBig(&blockHash)
	if hashNum.Cmp(target) > 0 {
		str := fmt.Sprintf("block hash of %064x is higher than"+
			" expected max of %064x", hashNum, target)
		return errors.New(str)
	}
	return nil
}

func (this *Blake2bd) GetNextDiffBig(weightedSumDiv *big.Int, oldDiffBig *big.Int, currentPowPercent *big.Int, param *PowConfig) *big.Int {
	nextDiffBig := weightedSumDiv.Mul(weightedSumDiv, oldDiffBig)
	defer func() {
		nextDiffBig = nextDiffBig.Rsh(nextDiffBig, 32)

	}()
	targetPercent := this.PowPercent(param)
	if targetPercent.Cmp(big.NewInt(0)) <= 0 {
		return nextDiffBig
	}
	currentPowPercent.Mul(currentPowPercent, big.NewInt(100))
	if currentPowPercent.Cmp(targetPercent) > 0 {
		nextDiffBig.Mul(nextDiffBig, currentPowPercent)
		nextDiffBig.Div(nextDiffBig, targetPercent)
	} else {
		nextDiffBig.Mul(nextDiffBig, targetPercent)
		nextDiffBig.Div(nextDiffBig, currentPowPercent)
	}
	return nextDiffBig
}

func (this *Blake2bd) PowPercent(param *PowConfig) *big.Int {
	targetPercent := big.NewInt(int64(param.Blake2bDPercent))
	targetPercent.Lsh(targetPercent, 32)
	return targetPercent
}

func (this *Blake2bd) GetSafeDiff(param *PowConfig, cur_reduce_diff uint64) *big.Int {
	limitBits := param.Blake2bdPowLimitBits
	limitBitsBig := CompactToBig(limitBits)
	if cur_reduce_diff <= 0 {
		return limitBitsBig
	}
	newTarget := &big.Int{}
	newTarget = newTarget.SetUint64(cur_reduce_diff)
	// Limit new value to the proof of work limit.
	if newTarget.Cmp(param.Blake2bdPowLimit) > 0 {
		newTarget.Set(param.Blake2bdPowLimit)
	}
	return newTarget
}

// compare the target
// wether target match the target diff
func (this *Blake2bd) CompareDiff(newTarget *big.Int, target *big.Int) bool {
	return newTarget.Cmp(target) <= 0
}

// pow proof data
func (this *Blake2bd) Bytes() PowBytes {
	r := make(PowBytes, 0)
	//write nonce 4 bytes
	n := make([]byte, 4)
	binary.LittleEndian.PutUint32(n, this.Nonce)
	r = append(r, n...)

	t := make([]byte, 1)
	//write pow type 1 byte
	t[0] = uint8(this.PowType)
	r = append(r, t...)
	return PowBytes(r)
}
