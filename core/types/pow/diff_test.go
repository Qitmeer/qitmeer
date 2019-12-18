package pow

import (
	"encoding/hex"
	"fmt"
	"github.com/Qitmeer/qitmeer/common"
	"github.com/Qitmeer/qitmeer/common/hash"
	"github.com/Qitmeer/qitmeer/common/util"
	`github.com/Qitmeer/qitmeer/params`
	"github.com/stretchr/testify/assert"
	"math/big"
	"testing"
)

func TestCalcScale(t *testing.T) {
	assert.Equal(t, uint64(48), GraphWeight(24,2,params.TestNetParams.PowConfig.BigGraphStartHeight,CUCKAROO))
	assert.Equal(t, uint64(100), GraphWeight(25,params.TestNetParams.PowConfig.BigGraphStartHeight,params.TestNetParams.PowConfig.BigGraphStartHeight,CUCKAROO))
	assert.Equal(t, uint64(208), GraphWeight(26,params.TestNetParams.PowConfig.BigGraphStartHeight,params.TestNetParams.PowConfig.BigGraphStartHeight,CUCKAROO))
	assert.Equal(t, uint64(1), GraphWeight(29,2,params.TestNetParams.PowConfig.BigGraphStartHeight,CUCKAROO))
	assert.Equal(t, uint64(1856), GraphWeight(29,params.TestNetParams.PowConfig.BigGraphStartHeight,params.TestNetParams.PowConfig.BigGraphStartHeight,CUCKAROO))
	assert.Equal(t, uint64(7936), GraphWeight(31,params.TestNetParams.PowConfig.BigGraphStartHeight,params.TestNetParams.PowConfig.BigGraphStartHeight,CUCKAROO))
}

func TestScaleToTarget(t *testing.T) {
	diff := uint64(1000)
	diffBig := &big.Int{}
	diffBig.SetUint64(diff)
	assert.Equal(t, "0c49ba5e353f7ced916872b020c49ba5e353f7ced916872b020c49ba5e353f7c", CuckooDiffToTarget(GraphWeight(24,params.TestNetParams.PowConfig.BigGraphStartHeight,params.TestNetParams.PowConfig.BigGraphStartHeight,CUCKAROO), diffBig))
	assert.Equal(t, "004189374bc6a7ef9db22d0e5604189374bc6a7ef9db22d0e5604189374bc6a7", CuckooDiffToTarget(GraphWeight(29,2,params.TestNetParams.PowConfig.BigGraphStartHeight,CUCKAROO), diffBig))
	assert.Equal(t, "ffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff", CuckooDiffToTarget(GraphWeight(29,params.TestNetParams.PowConfig.BigGraphStartHeight,params.TestNetParams.PowConfig.BigGraphStartHeight,CUCKAROO), diffBig))

	assert.Equal(t, "ffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff", CuckooDiffToTarget(GraphWeight(24,params.TestNetParams.PowConfig.BigGraphStartHeight,params.TestNetParams.PowConfig.BigGraphStartHeight,CUCKAROO), big.NewInt(48)))
	assert.Equal(t, "8000000000000000000000000000000000000000000000000000000000000000", CuckooDiffToTarget(GraphWeight(24,params.TestNetParams.PowConfig.BigGraphStartHeight,params.TestNetParams.PowConfig.BigGraphStartHeight,CUCKAROO), big.NewInt(96)))
	assert.Equal(t, "000017b5dbd6151319c5e8a604ddc87e903df63f7e7512ea5a30f9dab794f2be", CuckooDiffToTarget(GraphWeight(24,params.TestNetParams.PowConfig.BigGraphStartHeight,params.TestNetParams.PowConfig.BigGraphStartHeight,CUCKAROO), big.NewInt(33964288)))
}

func TestDiffCompare(t *testing.T) {
	str := "ffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff"
	b, _ := hex.DecodeString(str)
	h := hash.Hash{}
	util.ReverseBytes(b)
	copy(h[:], b)
	fmt.Println(h)
	fmt.Println(CalcCuckooDiff(1856, h))
	str = "000000ffffffffffffffffffffffffffffffffffffffffffffffffffffffffff"
	b, _ = hex.DecodeString(str)
	util.ReverseBytes(b)
	copy(h[:], b)
	fmt.Println(h)
	fmt.Println(CalcCuckooDiff(1856, h))
	str = "0000000000000000000000ffffffffffffffffffffffffffffffffffffffffff"
	b, _ = hex.DecodeString(str)
	util.ReverseBytes(b)
	copy(h[:], b)
	fmt.Println(h)
	fmt.Println(CalcCuckooDiff(1856, h))
	str = "0000000000000000000000000000000000000000000000000000000000000000"
	b, _ = hex.DecodeString(str)
	util.ReverseBytes(b)
	copy(h[:], b)
	fmt.Println(h)
	fmt.Println(CalcCuckooDiff(1856, h))
}

// scale * 2^ 64 / diff is target
//edge bits 24 scale is 48
func TestBigToCompact(t *testing.T) {
	diff := 48
	diffBig := &big.Int{}
	diffBig.SetUint64(uint64(diff))
	assert.Equal(t, uint32(0x1300000), BigToCompact(diffBig))
}

// scale * 2^ 64 / diff is target
//edge bits 24 scale is 48
func TestCalcNextDiff(t *testing.T) {
	p := &PowConfig{
		Blake2bdPowLimit:     new(big.Int).Sub(new(big.Int).Lsh(common.Big1, 232), common.Big1),
		Blake2bdPowLimitBits: 0x1e00ffff,
		Percent: []Percent{
			{
				Blake2bDPercent: 34,
				CuckarooPercent: 33,
				CuckatooPercent: 33,
				MainHeight:      0,
			},
		},
		//hash ffffffffffffffff000000000000000000000000000000000000000000000000 corresponding difficulty is 48 for edge bits 24
		// Uniform field type uint64 value is 48 . bigToCompact the uint32 value
		// 24 edge_bits only need hash 1*4 times use for privnet if GPS is 2. need 50 /2 * 2 ≈ 1min find once
		CuckarooMinDifficulty: 0x1300000 * 2,
		CuckatooMinDifficulty: 0x1300000 * 2,
	}
	blakeObj := &Blake2bd{}
	oldDiff := int64(10000)
	oldDiffBig := big.NewInt(oldDiff)
	blakeObj.SetMainHeight(1)
	blakeObj.SetParams(p)
	// actual time 2s  target time 5s
	// current pow count 4 all count 100
	weightBig := big.NewInt(2)
	weightBig.Lsh(weightBig, 32)
	weightBig.Div(weightBig, big.NewInt(5))
	currentPowPercent := big.NewInt(4)
	currentPowPercent.Lsh(currentPowPercent, 32)
	currentPowPercent.Div(currentPowPercent, big.NewInt(100))
	nextDiffBig := blakeObj.GetNextDiffBig(weightBig, oldDiffBig, currentPowPercent)
	//10000 * ( 2 / 5 ) / (4 / 34)
	assert.Equal(t, uint64(10000*2*34/5/4), nextDiffBig.Uint64())
}

// scale * 2^ 64 / diff is target
//edge bits 24 scale is 48
func TestCalcCuckarooNextDiff(t *testing.T) {
	p := &PowConfig{
		Blake2bdPowLimit:     new(big.Int).Sub(new(big.Int).Lsh(common.Big1, 232), common.Big1),
		Blake2bdPowLimitBits: 0x1e00ffff,
		Percent: []Percent{
			{
				Blake2bDPercent: 34,
				CuckarooPercent: 33,
				CuckatooPercent: 33,
				MainHeight:      0,
			},
		},
		//hash ffffffffffffffff000000000000000000000000000000000000000000000000 corresponding difficulty is 48 for edge bits 24
		// Uniform field type uint64 value is 48 . bigToCompact the uint32 value
		// 24 edge_bits only need hash 1*4 times use for privnet if GPS is 2. need 50 /2 * 2 ≈ 1min find once
		CuckarooMinDifficulty: 0x1300000 * 2,
		CuckatooMinDifficulty: 0x1300000 * 2,
	}

	oldDiff := int64(10000)
	oldDiffBig := big.NewInt(oldDiff)
	// actual time 2s  target time 5s
	// current pow count 4 all count 100
	weightBig := big.NewInt(2)
	weightBig.Lsh(weightBig, 32)
	weightBig.Div(weightBig, big.NewInt(5))
	//cuckaroo diff ajustment
	cuckarooObj := &Cuckaroo{}
	cuckarooObj.SetMainHeight(1)
	cuckarooObj.SetParams(p)
	// actual time 2s  target time 5s
	// current pow count 4 all count 100
	currentPowPercent := big.NewInt(4)
	currentPowPercent.Lsh(currentPowPercent, 32)
	currentPowPercent.Div(currentPowPercent, big.NewInt(100))
	nextDiffBig := cuckarooObj.GetNextDiffBig(weightBig, oldDiffBig, currentPowPercent)
	//10000 / ( 2 / 5 ) * (4 / 33)
	assert.Equal(t, uint64(10000*5*4/2/33), nextDiffBig.Uint64())
}

// scale * 2^ 64 / diff is target
//edge bits 24 scale is 48
func TestCalcCuckatooNextDiff(t *testing.T) {
	p := &PowConfig{
		Blake2bdPowLimit:     new(big.Int).Sub(new(big.Int).Lsh(common.Big1, 232), common.Big1),
		Blake2bdPowLimitBits: 0x1e00ffff,
		Percent: []Percent{
			{
				Blake2bDPercent: 34,
				CuckarooPercent: 33,
				CuckatooPercent: 33,
				MainHeight:      0,
			},
		},
		//hash ffffffffffffffff000000000000000000000000000000000000000000000000 corresponding difficulty is 48 for edge bits 24
		// Uniform field type uint64 value is 48 . bigToCompact the uint32 value
		// 24 edge_bits only need hash 1*4 times use for privnet if GPS is 2. need 50 /2 * 2 ≈ 1min find once
		CuckarooMinDifficulty: 0x1300000 * 2,
		CuckatooMinDifficulty: 0x1300000 * 2,
	}
	oldDiff := int64(10000)
	oldDiffBig := big.NewInt(oldDiff)
	// actual time 2s  target time 5s
	// current pow count 4 all count 100
	weightBig := big.NewInt(2)
	weightBig.Lsh(weightBig, 32)
	weightBig.Div(weightBig, big.NewInt(5))
	//cuckaroo diff ajustment
	cuckatooObj := &Cuckatoo{}
	cuckatooObj.SetMainHeight(1)
	cuckatooObj.SetParams(p)
	// actual time 2s  target time 5s
	// current pow count 4 all count 100
	currentPowPercent := big.NewInt(4)
	currentPowPercent.Lsh(currentPowPercent, 32)
	currentPowPercent.Div(currentPowPercent, big.NewInt(100))
	nextDiffBig := cuckatooObj.GetNextDiffBig(weightBig, oldDiffBig, currentPowPercent)
	//10000 / ( 2 / 5 ) * (4 / 33)
	assert.Equal(t, uint64(10000*5*4/2/33), nextDiffBig.Uint64())
}
