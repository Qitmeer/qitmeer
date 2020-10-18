package pow

import (
	"errors"
	"fmt"
	"math/big"
	"sort"
)

var PowConfigInstance *PowConfig

type MainHeight uint32
type PercentValue uint32
type Percent map[MainHeight]PercentItem
type PercentItem map[PowType]PercentValue
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

	QitmeerKeccak256PowLimit     *big.Int
	QitmeerKeccak256PowLimitBits uint32

	// cuckoo difficulty calc params  min difficulty
	CuckarooMinDifficulty  uint32
	CuckaroomMinDifficulty uint32
	CuckatooMinDifficulty  uint32

	Percent map[MainHeight]PercentItem

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
func (this *PowConfig) GetPercentByHeightAndType(h int64, powType PowType) PercentValue {
	//sort by main height asc
	var keys []int
	for k := range this.Percent {
		keys = append(keys, int(k))
	}
	sort.Ints(keys)
	currentPercent := map[PowType]PercentValue{}
	// get best match percent
	for _, k := range keys {
		if h >= int64(k) {
			currentPercent = this.Percent[MainHeight(k)]
		}
	}
	val, ok := currentPercent[powType]
	if !ok {
		return 0
	}
	return val
}

// check percent
func (this *PowConfig) Check() error {
	allPercent := 0
	heightArr := map[MainHeight]int{}
	for mHeight, p := range this.Percent {
		if _, ok := heightArr[mHeight]; ok {
			return errors.New("pow config error, mainHeight set repeat!")
		}
		heightArr[mHeight] = 1
		for pty, val := range p {
			powName := GetPowName(pty)
			if powName == "" {
				return errors.New(fmt.Sprintf("Pow Type %d Not Config Name in IPow.go!", pty))
			}
			if val < 0 {
				return errors.New(powName + " percent config error, all percent must greater than or equal to 0!")
			}
			allPercent += int(val)
		}
		if allPercent != 100 {
			return errors.New("pow config error, all pow not equal 100%!actual is " + fmt.Sprintf("%d", allPercent))
		}
	}
	return nil
}
