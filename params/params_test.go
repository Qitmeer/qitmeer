package params

import (
	"github.com/Qitmeer/qitmeer/core/types/pow"
	"github.com/stretchr/testify/assert"
	"math/big"
	"testing"
)

//test blake2bd percent params
func TestPercent(t *testing.T) {
	types := []pow.PowType{pow.BLAKE2BD, pow.CUCKAROO, pow.CUCKATOO, pow.CUCKAROOM}
	for _, powType := range types {
		instance := pow.GetInstance(powType, 0, []byte{})
		instance.SetParams(PrivNetParam.PowConfig)
		percent := new(big.Int)
		for _, p := range PrivNetParam.PowConfig.Percent {
			instance.SetMainHeight(p.MainHeight + 1)
			switch powType {
			case pow.BLAKE2BD:
				percent.SetInt64(int64(p.Blake2bDPercent))
			case pow.CUCKAROO:
				percent.SetInt64(int64(p.CuckarooPercent))
			case pow.CUCKATOO:
				percent.SetInt64(int64(p.CuckatooPercent))
			case pow.CUCKAROOM:
				percent.SetInt64(int64(p.CuckaroomPercent))
			}
			percent.Lsh(percent, 32)
			assert.Equal(t, percent.Uint64(), instance.PowPercent().Uint64())
		}
	}

}
