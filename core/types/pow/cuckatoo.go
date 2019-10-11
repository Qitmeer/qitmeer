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

type Cuckatoo struct {
    Cuckoo
}

const MIN_CUCKATOOEDGEBITS = 29
const MAX_CUCKATOOEDGEBITS = 32

func (this *Cuckatoo) Verify(headerWithoutProofData []byte,targetDiffBits uint32,powConfig * PowConfig) error{
    targetDiff := CompactToBig(targetDiffBits).Uint64()
    if !this.CheckAvailable(this.PowPercent(powConfig)){
        str := fmt.Sprintf("cuckatoo is not supported")
        return errors.New(str)
    }
    h := hash.HashH(headerWithoutProofData)
    nonces := this.GetCircleNonces()
    edgeBits := this.GetEdgeBits()
    if edgeBits < MIN_CUCKATOOEDGEBITS{
        return fmt.Errorf("edge bits:%d is too short!less than %d",edgeBits,MIN_CUCKATOOEDGEBITS)
    }
    if edgeBits > MAX_CUCKATOOEDGEBITS{
        return fmt.Errorf("edge bits:%d is too large! more than %d",edgeBits,MAX_CUCKATOOEDGEBITS)
    }
    err := cuckoo.VerifyCuckatoo(h[:],nonces[:],uint(edgeBits))
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

func (this *Cuckatoo) GetNextDiffBig(weightedSumDiv *big.Int,oldDiffBig *big.Int,currentPowPercent *big.Int,param *PowConfig) *big.Int{
    nextDiffBig := oldDiffBig.Div(oldDiffBig, weightedSumDiv)
    targetPercent := this.PowPercent(param)
    if currentPowPercent.Cmp(targetPercent) > 0{
        currentPowPercent.Div(currentPowPercent,targetPercent)
        nextDiffBig.Mul(nextDiffBig,currentPowPercent)
    }
    return nextDiffBig
}
func (this *Cuckatoo) PowPercent(param *PowConfig) *big.Int{
    targetPercent := big.NewInt(int64(param.CuckatooPercent))
    targetPercent.Lsh(targetPercent,32)
    return targetPercent
}

func (this *Cuckatoo) GetSafeDiff(param *PowConfig,cur_reduce_diff uint64) uint64{
    minDiff := uint64(param.CuckatooMinDifficulty)
    if cur_reduce_diff <= 0{
        return minDiff
    }
    if cur_reduce_diff > minDiff{
        return cur_reduce_diff
    }
    return minDiff
}