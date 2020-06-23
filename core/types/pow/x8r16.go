// Copyright (c) 2017-2020 The qitmeer developers
// license that can be found in the LICENSE file.
// Reference resources of rust bitVector
package pow

import (
	"encoding/binary"
	"errors"
	"fmt"
	"github.com/Qitmeer/qitmeer/common/hash"
	"github.com/Qitmeer/qitmeer/core/json"
	"math/big"
)

type X8r16 struct {
	Pow
}

func (this *X8r16) GetPowResult() json.PowResult {
	return json.PowResult{
		PowName:   PowMapString[this.GetPowType()].(string),
		PowType:   uint8(this.GetPowType()),
		Nonce:     this.GetNonce(),
		ProofData: nil,
	}
}

func (this *X8r16) Verify(headerData []byte, blockHash hash.Hash, targetDiffBits uint32) error {
	target := CompactToBig(targetDiffBits)
	if target.Sign() <= 0 {
		str := fmt.Sprintf("block target difficulty of %064x is too "+
			"low", target)
		return errors.New(str)
	}

	//The target difficulty must be less than the maximum allowed.
	if target.Cmp(this.params.X8r16PowLimit) > 0 {
		str := fmt.Sprintf("block target difficulty of %064x is "+
			"higher than max of %064x", target, this.params.X8r16PowLimit)
		return errors.New(str)
	}
	h := hash.HashX8r16(headerData)
	hashNum := HashToBig(&h)
	if hashNum.Cmp(target) > 0 {
		str := fmt.Sprintf("block hash of %064x is higher than"+
			" expected max of %064x", hashNum, target)
		return errors.New(str)
	}
	return nil
}

func (this *X8r16) GetNextDiffBig(weightedSumDiv *big.Int, oldDiffBig *big.Int, currentPowPercent *big.Int) *big.Int {
	nextDiffBig := weightedSumDiv.Mul(weightedSumDiv, oldDiffBig)
	defer func() {
		nextDiffBig = nextDiffBig.Rsh(nextDiffBig, 32)

	}()
	targetPercent := this.PowPercent()
	if targetPercent.Cmp(big.NewInt(0)) <= 0 {
		return nextDiffBig
	}
	currentPowPercent.Mul(currentPowPercent, big.NewInt(100))
	nextDiffBig.Mul(nextDiffBig, targetPercent)
	nextDiffBig.Div(nextDiffBig, currentPowPercent)
	return nextDiffBig
}

func (this *X8r16) PowPercent() *big.Int {
	targetPercent := big.NewInt(int64(this.params.GetPercentByHeight(this.mainHeight).X8r16Percent))
	targetPercent.Lsh(targetPercent, 32)
	return targetPercent
}

func (this *X8r16) GetSafeDiff(cur_reduce_diff uint64) *big.Int {
	limitBits := this.params.X8r16PowLimitBits
	limitBitsBig := CompactToBig(limitBits)
	if cur_reduce_diff <= 0 {
		return limitBitsBig
	}
	newTarget := &big.Int{}
	newTarget = newTarget.SetUint64(cur_reduce_diff)
	// Limit new value to the proof of work limit.
	if newTarget.Cmp(this.params.X8r16PowLimit) > 0 {
		newTarget.Set(this.params.X8r16PowLimit)
	}
	return newTarget
}

// compare the target
// wether target match the target diff
func (this *X8r16) CompareDiff(newTarget *big.Int, target *big.Int) bool {
	return newTarget.Cmp(target) <= 0
}

// pow proof data
func (this *X8r16) Bytes() PowBytes {
	r := make(PowBytes, 0)
	//write nonce 4 bytes
	n := make([]byte, 4)
	binary.LittleEndian.PutUint32(n, this.Nonce)
	r = append(r, n...)

	t := make([]byte, 1)
	//write pow type 1 byte
	t[0] = uint8(this.PowType)
	r = append(r, t...)
	//write ProofData 169 bytes
	r = append(r, this.ProofData[:]...)
	return PowBytes(r)
}

// pow proof data
func (this *X8r16) BlockData() PowBytes {
	l := len(this.Bytes())
	return PowBytes(this.Bytes()[:l-PROOFDATA_LENGTH])
}

//check pow is available
func (this *X8r16) CheckAvailable() bool {
	return this.params.GetPercentByHeight(this.mainHeight).X8r16Percent > 0
}
