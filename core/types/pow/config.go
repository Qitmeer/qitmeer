package pow

import (
	"errors"
	"math/big"
	"sort"
)

var PowConfigInstance *PowConfig

type Percent struct {
	//percent of every pow sum of them must be 100
	CuckarooPercent int
	CuckatooPercent int
	Blake2bDPercent int
	MainHeight      int64
	X16rv3Percent   int
}

type PowConfig struct {
	// PowLimit defines the highest allowed proof of work value for a block
	// as a uint256.
	Blake2bdPowLimit *big.Int
	// PowLimitBits defines the highest allowed proof of work value for a
	// block in compact form.
	// highest value is mean min difficulty
	Blake2bdPowLimitBits uint32

	X16rv3PowLimit     *big.Int
	X16rv3PowLimitBits uint32

	// cuckoo difficulty calc params  min difficulty
	CuckarooMinDifficulty uint32
	CuckatooMinDifficulty uint32

	Percent []Percent

	AdjustmentStartMainHeight int64

	//is init
	init bool
}

//global cache
func GetPowConfig() *PowConfig {
	if PowConfigInstance != nil {
		return PowConfigInstance
	}
	PowConfigInstance = &PowConfig{}
	PowConfigInstance.init = false
	return PowConfigInstance
}

// set config
// GetPowConfig().Set(params.PowConfig)
func (this *PowConfig) Set(p *PowConfig) *PowConfig {
	if !this.init {
		this = p
		this.init = true
	}
	return this
}

// get Percent By height
func (this *PowConfig) GetPercentByHeight(h int64) (res Percent) {
	//sort by main height asc
	sort.Slice(this.Percent, func(i, j int) bool {
		return this.Percent[i].MainHeight < this.Percent[j].MainHeight
	})
	// get best match percent
	for i := 0; i < len(this.Percent); i++ {
		if h >= this.Percent[i].MainHeight {
			res = this.Percent[i]
		}
	}
	return
}

// check percent
func (this *PowConfig) Check() error {
	allPercent := 0
	heightArr := map[int64]int{}
	for _, p := range this.Percent {
		if p.MainHeight < 0 {
			return errors.New("pow config error, must greater than or equal to 0!")
		}
		if _, ok := heightArr[p.MainHeight]; ok {
			return errors.New("pow config error, mainHeight set repeat!")
		}
		heightArr[p.MainHeight] = 1
		if p.CuckarooPercent < 0 || p.Blake2bDPercent < 0 || p.CuckatooPercent < 0 {
			return errors.New("pow config error, all percent must greater than or equal to 0!")
		}
		allPercent = p.CuckarooPercent + p.Blake2bDPercent + p.CuckatooPercent + p.X16rv3Percent
		if allPercent != 100 {
			return errors.New("pow config error, all pow not equal 100%!")
		}
	}
	return nil
}
