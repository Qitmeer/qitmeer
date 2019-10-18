package pow

import (
    `github.com/Qitmeer/qitmeer/common`
    `github.com/stretchr/testify/assert`
    `math/big`
    `testing`
)

func TestCalcScale(t *testing.T) {
    assert.Equal(t,uint64(48),GraphWeight(24))
    assert.Equal(t,uint64(100),GraphWeight(25))
    assert.Equal(t,uint64(208),GraphWeight(26))
    assert.Equal(t,uint64(7936),GraphWeight(31))
}

func TestScaleToTarget(t *testing.T) {
    diff := uint64(1000)
    diffBig := &big.Int{}
    diffBig.SetUint64(diff)
    assert.Equal(t,"0c49ba5e353f7ced000000000000000000000000000000000000000000000000",CuckooDiffToTarget(GraphWeight(24),diffBig))
    assert.Equal(t,"db22d0e560418937000000000000000000000000000000000000000000000000",CuckooDiffToTarget(GraphWeight(29),diffBig))
}

// scale * 2^ 64 / diff is target
//edge bits 24 scale is 48
func TestBigToCompact(t *testing.T) {
    diff := 48
    diffBig := &big.Int{}
    diffBig.SetUint64(uint64(diff))
    assert.Equal(t,uint32(0x1300000),BigToCompact(diffBig))
}

// scale * 2^ 64 / diff is target
//edge bits 24 scale is 48
func TestCalcNextDiff(t *testing.T) {
    p := &PowConfig{
        Blake2bdPowLimit:      new(big.Int).Sub(new(big.Int).Lsh(common.Big1, 232), common.Big1),
        Blake2bdPowLimitBits:  0x1e00ffff,
        Blake2bDPercent:       34,
        CuckarooPercent:       33,
        CuckatooPercent:       33,
        //hash ffffffffffffffff000000000000000000000000000000000000000000000000 corresponding difficulty is 48 for edge bits 24
        // Uniform field type uint64 value is 48 . bigToCompact the uint32 value
        // 24 edge_bits only need hash 1*4 times use for privnet if GPS is 2. need 50 /2 * 2 ≈ 1min find once
        CuckarooMinDifficulty:     0x1300000 * 2,
        CuckatooMinDifficulty:     0x1300000 * 2,
    }
    blakeObj := &Blake2bd{}
    oldDiff := int64(10000)
    oldDiffBig := big.NewInt(oldDiff)
    // actual time 2s  target time 5s
    // current pow count 4 all count 100
    weightBig := big.NewInt(2)
    weightBig.Lsh(weightBig,32)
    weightBig.Div(weightBig,big.NewInt(5))
    currentPowPercent := big.NewInt(4)
    currentPowPercent.Lsh(currentPowPercent,32)
    currentPowPercent.Div(currentPowPercent,big.NewInt(100))
    nextDiffBig := blakeObj.GetNextDiffBig(weightBig,oldDiffBig,currentPowPercent,p)
    //10000 * ( 2 / 5 ) / (4 / 34)
    assert.Equal(t,uint64(10000 * 2 * 34 / 5 / 4),nextDiffBig.Uint64())
}

// scale * 2^ 64 / diff is target
//edge bits 24 scale is 48
func TestCalcCuckarooNextDiff(t *testing.T) {
    p := &PowConfig{
        Blake2bdPowLimit:      new(big.Int).Sub(new(big.Int).Lsh(common.Big1, 232), common.Big1),
        Blake2bdPowLimitBits:  0x1e00ffff,
        Blake2bDPercent:       34,
        CuckarooPercent:       33,
        CuckatooPercent:       33,
        //hash ffffffffffffffff000000000000000000000000000000000000000000000000 corresponding difficulty is 48 for edge bits 24
        // Uniform field type uint64 value is 48 . bigToCompact the uint32 value
        // 24 edge_bits only need hash 1*4 times use for privnet if GPS is 2. need 50 /2 * 2 ≈ 1min find once
        CuckarooMinDifficulty:     0x1300000 * 2,
        CuckatooMinDifficulty:     0x1300000 * 2,
    }
    oldDiff := int64(10000)
    oldDiffBig := big.NewInt(oldDiff)
    // actual time 2s  target time 5s
    // current pow count 4 all count 100
    weightBig := big.NewInt(2)
    weightBig.Lsh(weightBig,32)
    weightBig.Div(weightBig,big.NewInt(5))
    //cuckaroo diff ajustment
    cuckarooObj := &Cuckaroo{}
    // actual time 2s  target time 5s
    // current pow count 4 all count 100
    currentPowPercent := big.NewInt(4)
    currentPowPercent.Lsh(currentPowPercent,32)
    currentPowPercent.Div(currentPowPercent,big.NewInt(100))
    nextDiffBig := cuckarooObj.GetNextDiffBig(weightBig,oldDiffBig,currentPowPercent,p)
    //10000 / ( 2 / 5 ) * (4 / 33)
    assert.Equal(t,uint64(10000 * 5 * 4 / 2 / 33),nextDiffBig.Uint64())
}

// scale * 2^ 64 / diff is target
//edge bits 24 scale is 48
func TestCalcCuckatooNextDiff(t *testing.T) {
    p := &PowConfig{
        Blake2bdPowLimit:      new(big.Int).Sub(new(big.Int).Lsh(common.Big1, 232), common.Big1),
        Blake2bdPowLimitBits:  0x1e00ffff,
        Blake2bDPercent:       34,
        CuckarooPercent:       33,
        CuckatooPercent:       33,
        //hash ffffffffffffffff000000000000000000000000000000000000000000000000 corresponding difficulty is 48 for edge bits 24
        // Uniform field type uint64 value is 48 . bigToCompact the uint32 value
        // 24 edge_bits only need hash 1*4 times use for privnet if GPS is 2. need 50 /2 * 2 ≈ 1min find once
        CuckarooMinDifficulty:     0x1300000 * 2,
        CuckatooMinDifficulty:     0x1300000 * 2,
    }
    oldDiff := int64(10000)
    oldDiffBig := big.NewInt(oldDiff)
    // actual time 2s  target time 5s
    // current pow count 4 all count 100
    weightBig := big.NewInt(2)
    weightBig.Lsh(weightBig,32)
    weightBig.Div(weightBig,big.NewInt(5))
    //cuckaroo diff ajustment
    cuckarooObj := &Cuckatoo{}
    // actual time 2s  target time 5s
    // current pow count 4 all count 100
    currentPowPercent := big.NewInt(4)
    currentPowPercent.Lsh(currentPowPercent,32)
    currentPowPercent.Div(currentPowPercent,big.NewInt(100))
    nextDiffBig := cuckarooObj.GetNextDiffBig(weightBig,oldDiffBig,currentPowPercent,p)
    //10000 / ( 2 / 5 ) * (4 / 33)
    assert.Equal(t,uint64(10000 * 5 * 4 / 2 / 33),nextDiffBig.Uint64())
}