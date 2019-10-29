package pow

import (
	`errors`
	`math/big`
)

var PowConfigInstance *PowConfig

type Percent struct {
	//percent of every pow sum of them must be 100
	CuckarooPercent int
	CuckatooPercent int
	Blake2bDPercent int
	MainHeight      int64
}

type PowConfig struct {
	// PowLimit defines the highest allowed proof of work value for a block
	// as a uint256.
	Blake2bdPowLimit *big.Int
	// PowLimitBits defines the highest allowed proof of work value for a
	// block in compact form.
	// highest value is mean min difficulty
	Blake2bdPowLimitBits uint32

	// cuckoo difficulty calc params  min difficulty
	CuckarooMinDifficulty uint32
	CuckatooMinDifficulty uint32

	Percent []Percent

	//is init
	init bool
}
//global cache
func GetPowConfig() *PowConfig{
	if PowConfigInstance != nil{
		return PowConfigInstance
	}
	PowConfigInstance = &PowConfig{}
	PowConfigInstance.init = false
	return PowConfigInstance
}

// set config
// GetPowConfig().Set(params.PowConfig)
func (this *PowConfig) Set(p *PowConfig) *PowConfig{
	if !this.init{
		this = p
		this.init = true
	}
	return this
}

// get Percent By height
func (this *PowConfig) GetPercentByHeight(h int64) (res Percent){
	for _,p := range this.Percent{
		if h >= p.MainHeight {
			res = p
		}
	}
	return
}

// check percent
func (this *PowConfig) Check() error{
	allPercent := 0
	heightArr := map[int64]int{}
	for _,p := range this.Percent{
		if _,ok := heightArr[p.MainHeight];ok{
			return errors.New("pow config error, mainHeight set repeat!")
		}
		heightArr[p.MainHeight] = 1
		allPercent = p.CuckarooPercent + p.Blake2bDPercent + p.CuckatooPercent
		if allPercent != 100{
			return errors.New("pow config error, all pow not equal 100%!")
		}
	}
	return nil
}