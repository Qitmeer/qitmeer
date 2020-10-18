package params

import (
	"fmt"
	"github.com/Qitmeer/qitmeer/core/types/pow"
	"github.com/stretchr/testify/assert"
	"math/big"
	"testing"
)

func TestPowLimitToBits(t *testing.T) {
	compact := pow.BigToCompact(testMixNetPowLimit)
	assert.Equal(t, fmt.Sprintf("0x%064x", testMixNetPowLimit), "0x03ffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff")
	assert.Equal(t, fmt.Sprintf("0x%x", compact), "0x2003ffff")
}

//test blake2bd percent params
func TestPercent(t *testing.T) {
	types := []pow.PowType{pow.BLAKE2BD, pow.CUCKAROO, pow.CUCKATOO, pow.CUCKAROOM}
	for _, powType := range types {
		instance := pow.GetInstance(powType, 0, []byte{})
		instance.SetParams(PrivNetParam.PowConfig)
		percent := new(big.Int)
		for mheight, pi := range PrivNetParam.PowConfig.Percent {
			instance.SetMainHeight(int64(mheight + 1))
			percent.SetInt64(int64(pi[powType]))

			percent.Lsh(percent, 32)
			assert.Equal(t, percent.Uint64(), instance.PowPercent().Uint64())
		}
	}

}
