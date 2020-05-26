// Copyright (c) 2017-2020 The qitmeer developers
// license that can be found in the LICENSE file.
// Reference resources of rust bitVector
package pow

import (
	"errors"
	"fmt"
	"github.com/Qitmeer/qitmeer/common/hash"
	"github.com/Qitmeer/qitmeer/crypto/cuckoo"
	"github.com/Qitmeer/qitmeer/log"
	"math/big"
)

type Cuckaroo struct {
	Cuckoo
}

const MIN_CUCKAROOEDGEBITS = 24
const MAX_CUCKAROOEDGEBITS = 32

func (this *Cuckaroo) Verify(headerData []byte, blockHash hash.Hash, targetDiffBits uint32) error {
	targetDiff := CompactToBig(targetDiffBits)
	baseDiff := CompactToBig(this.params.CuckarooMinDifficulty)
	h := this.GetSipHash(headerData)
	nonces := this.GetCircleNonces()
	edgeBits := this.GetEdgeBits()
	if edgeBits < MIN_CUCKAROOEDGEBITS {
		return fmt.Errorf("edge bits:%d is too short! less than %d", edgeBits, MIN_CUCKAROOEDGEBITS)
	}
	if edgeBits > MAX_CUCKAROOEDGEBITS {
		return fmt.Errorf("edge bits:%d is too large! more than %d", edgeBits, MAX_CUCKAROOEDGEBITS)
	}
	err := cuckoo.VerifyCuckaroo(h[:], nonces[:], uint(edgeBits))
	if err != nil {
		log.Debug("Verify Error!", err)
		return err
	}

	//The target difficulty must be more than the min diff.
	if targetDiff.Cmp(baseDiff) < 0 {
		str := fmt.Sprintf("block target difficulty of %d is "+
			"less than min diff :%d", targetDiff, this.params.CuckarooMinDifficulty)
		return errors.New(str)
	}
	if CalcCuckooDiff(this.GraphWeight(), blockHash).Cmp(targetDiff) < 0 {
		return errors.New("difficulty is too easy!")
	}
	return nil
}

func (this *Cuckaroo) GetNextDiffBig(weightedSumDiv *big.Int, oldDiffBig *big.Int, currentPowPercent *big.Int) *big.Int {
	oldDiffBig.Lsh(oldDiffBig, 32)
	nextDiffBig := oldDiffBig.Div(oldDiffBig, weightedSumDiv)
	targetPercent := this.PowPercent()
	if targetPercent.Cmp(big.NewInt(0)) <= 0 {
		return nextDiffBig
	}
	currentPowPercent.Mul(currentPowPercent, big.NewInt(100))
	nextDiffBig.Mul(nextDiffBig, currentPowPercent)
	nextDiffBig.Div(nextDiffBig, targetPercent)
	return nextDiffBig
}

func (this *Cuckaroo) PowPercent() *big.Int {
	targetPercent := big.NewInt(int64(this.params.GetPercentByHeight(this.mainHeight).CuckarooPercent))
	targetPercent.Lsh(targetPercent, 32)
	return targetPercent
}

func (this *Cuckaroo) GetSafeDiff(cur_reduce_diff uint64) *big.Int {
	minDiffBig := CompactToBig(this.params.CuckarooMinDifficulty)
	if cur_reduce_diff <= 0 {
		return minDiffBig
	}
	newTarget := &big.Int{}
	newTarget = newTarget.SetUint64(cur_reduce_diff)
	// Limit new value to the proof of work limit.
	if newTarget.Cmp(minDiffBig) < 0 {
		newTarget.Set(minDiffBig)
	}
	return newTarget
}

//check pow is available
func (this *Cuckaroo) CheckAvailable() bool {
	return this.params.GetPercentByHeight(this.mainHeight).CuckarooPercent > 0
}

//calc scale
//the edge_bits is bigger ,then scale is bigger
//Reference resources https://eprint.iacr.org/2014/059.pdf 9. Difficulty control page 6
//while the average number of cycles found increases slowly with size; from 2 at 2^20 to 3 at 2^30
//Less times of hash calculation with the same difficulty
// 24 => 48 25 => 100 26 => 208 27 => 432 28 => 896 29 => 1856 30 => 3840 31 => 7936
//assume init difficulty is 1000
//24 target is 0c49ba5e353f7ced000000000000000000000000000000000000000000000000
//（The meaning of difficulty needs to be found 1000/48 * 50 ≈ 1000 times in edge_bits 24, and the answer may be obtained once.）
// why * 50 , because the when edge_count/nodes = 1/2,to find 42 cycles the probality is 2.2%
//29 target is db22d0e560418937000000000000000000000000000000000000000000000000
//（The difficulty needs to be found 1000/1856 * 50 ≈ 26 times in edge_bits 29, and the answer may be obtained once.）
//so In order to ensure the fairness of different edge indexes, the mining difficulty is different.
func (this *Cuckaroo) GraphWeight() uint64 {
	//45 days
	bigScale := this.params.AdjustmentStartMainHeight - this.mainHeight
	bigScale = bigScale * 40
	if bigScale <= 0 || int(this.GetEdgeBits()) == MIN_CUCKAROOEDGEBITS {
		bigScale = 1
	}

	scale := (2 << (this.GetEdgeBits() - MIN_CUCKAROOEDGEBITS)) * uint64(this.GetEdgeBits()) / uint64(bigScale)
	if scale <= 0 {
		scale = 1
	}
	return scale
}
