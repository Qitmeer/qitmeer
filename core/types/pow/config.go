package pow

import (
	"errors"
	"fmt"
	"math/big"
	"sort"
)

var PowConfigInstance *PowConfig

type Percent struct {
	//percent of every pow sum of them must be 100
	CuckarooPercent  int
	CuckatooPercent  int
	Blake2bDPercent  int
	CuckaroomPercent int
	Keccak256Percent int
	X16rv3Percent    int
	X8r16Percent     int
	MainHeight       int64
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

	X8r16PowLimit     *big.Int
	X8r16PowLimitBits uint32

	Keccak256PowLimit     *big.Int
	Keccak256PowLimitBits uint32

	// cuckoo difficulty calc params  min difficulty
	CuckarooMinDifficulty  uint32
	CuckaroomMinDifficulty uint32
	CuckatooMinDifficulty  uint32

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
		if p.CuckarooPercent < 0 || p.Blake2bDPercent < 0 ||
			p.CuckatooPercent < 0 || p.CuckaroomPercent < 0 ||
			p.Keccak256Percent < 0 ||
			p.X16rv3Percent < 0 || p.X8r16Percent < 0 {
			return errors.New("pow config error, all percent must greater than or equal to 0!")
		}
		allPercent = p.CuckarooPercent + p.Blake2bDPercent +
			p.CuckatooPercent + p.CuckaroomPercent + p.X16rv3Percent + p.X8r16Percent + p.Keccak256Percent
		if allPercent != 100 {
			return errors.New("pow config error, all pow not equal 100%!actual is " + fmt.Sprintf("%d", allPercent))
		}
	}
	return nil
}
