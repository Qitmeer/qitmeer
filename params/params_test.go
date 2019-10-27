package params

import (
	`github.com/Qitmeer/qitmeer/core/types/pow`
	`github.com/stretchr/testify/assert`
	`math/big`
	`testing`
)

//test blake2bd percent params
func TestBlake2bdPercent(t *testing.T)  {
	instance := pow.GetInstance(pow.BLAKE2BD,0,[]byte{})
	instance.SetHeight(1)
	instance.SetParams(PrivNetParam.PowConfig)
	//height 0 - 49 10%
	percent := big.NewInt(10)
	percent.Lsh(percent,32)
	assert.Equal(t,percent,instance.PowPercent())
	//height 50 - 99 30%
	instance.SetHeight(50)
	percent = big.NewInt(30)
	percent.Lsh(percent,32)
	assert.Equal(t,percent,instance.PowPercent())
	//height 100 -  80%
	instance.SetHeight(100)
	percent = big.NewInt(80)
	percent.Lsh(percent,32)
	assert.Equal(t,percent,instance.PowPercent())
}

//test cuckaroo percent params
func TestCuckarooPercent(t *testing.T)  {
	instance := pow.GetInstance(pow.CUCKAROO,0,[]byte{})
	instance.SetHeight(1)
	instance.SetParams(PrivNetParam.PowConfig)
	//height 0 - 49 70%
	percent := big.NewInt(70)
	percent.Lsh(percent,32)
	assert.Equal(t,percent,instance.PowPercent())
	//height 50 - 99 30%
	instance.SetHeight(50)
	percent = big.NewInt(30)
	percent.Lsh(percent,32)
	assert.Equal(t,percent,instance.PowPercent())
	//height 100 -  10%
	instance.SetHeight(100)
	percent = big.NewInt(10)
	percent.Lsh(percent,32)
	assert.Equal(t,percent,instance.PowPercent())
}

//test cuckatoo percent params
func TestCuckatooPercent(t *testing.T)  {
	instance := pow.GetInstance(pow.CUCKATOO,0,[]byte{})
	instance.SetHeight(1)
	instance.SetParams(PrivNetParam.PowConfig)
	//height 0 - 49 20%
	percent := big.NewInt(20)
	percent.Lsh(percent,32)
	assert.Equal(t,percent,instance.PowPercent())
	//height 50 - 99 40%
	instance.SetHeight(50)
	percent = big.NewInt(40)
	percent.Lsh(percent,32)
	assert.Equal(t,percent,instance.PowPercent())
	//height 100 -  10%
	instance.SetHeight(100)
	percent = big.NewInt(10)
	percent.Lsh(percent,32)
	assert.Equal(t,percent,instance.PowPercent())
}
