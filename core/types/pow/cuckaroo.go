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
func (this *Cuckaroo) Verify(headerWithoutProofData []byte,targetDiffBits uint32,powConfig *PowConfig) error{
    targetDiff := CompactToBig(targetDiffBits).Uint64()
    if !this.CheckAvailable(this.PowPercent(powConfig)){
        str := fmt.Sprintf("cuckaroo is not supported")
        return errors.New(str)
    }
    h := hash.HashH(headerWithoutProofData)
    nonces := this.GetCircleNonces()
    edgeBits := this.GetEdgeBits()
    if edgeBits < MIN_CUCKAROOEDGEBITS{
        return fmt.Errorf("edge bits:%d is too short! less than %d",edgeBits,MIN_CUCKAROOEDGEBITS)
    }
    if edgeBits > MAX_CUCKAROOEDGEBITS{
        return fmt.Errorf("edge bits:%d is too large! more than %d",edgeBits,MAX_CUCKAROOEDGEBITS)
    }
    err := cuckoo.VerifyCuckaroo(h[:],nonces[:],uint(edgeBits))
    if err != nil{
        log.Debug("Verify Error!",err)
        return err
    }
    //The target difficulty must be more than the min diff.
    if targetDiff < uint64(powConfig.CuckarooMinDifficulty) {
        str := fmt.Sprintf("block target difficulty of %d is "+
            "less than min diff :%d", targetDiff, powConfig.CuckarooMinDifficulty)
        return errors.New(str)
    }
    if CalcCuckooDiff(this.GetScale(),this.GetBlockHash([]byte{})) < targetDiff{
        return errors.New("difficulty is too easy!")
    }
    return nil
}

func (this *Cuckaroo) GetNextDiffBig(weightedSumDiv *big.Int,oldDiffBig *big.Int,currentPowPercent *big.Int,param *PowConfig) *big.Int{
    nextDiffBig := oldDiffBig.Div(oldDiffBig, weightedSumDiv)
    targetPercent := this.PowPercent(param)
    if currentPowPercent.Cmp(targetPercent) > 0{
        currentPowPercent.Div(currentPowPercent,targetPercent)
        nextDiffBig.Mul(nextDiffBig,currentPowPercent)
    }
    return nextDiffBig
}

func (this *Cuckaroo) PowPercent(param *PowConfig) *big.Int{
    targetPercent := big.NewInt(int64(param.CuckarooPercent))
    targetPercent.Lsh(targetPercent,32)
    return targetPercent
}

func (this *Cuckaroo) GetSafeDiff(param *PowConfig,cur_reduce_diff uint64) uint64{
    minDiff := uint64(param.CuckarooMinDifficulty)
    if cur_reduce_diff <= 0{
        return minDiff
    }
    if cur_reduce_diff > minDiff{
        return cur_reduce_diff
    }
    return minDiff
}