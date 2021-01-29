package types

import "fmt"

const (
	// Greater than or equal to
	FloorFeeType = 0

	// Strict equality
	EqualFreeType = 1
)

type FeeType byte

type CoinConfig struct {
	Id    CoinID
	Type  FeeType
	Value int64
}

type CoinConfigs []*CoinConfig

func (cc *CoinConfigs) CheckFees(fees AmountMap) error {
	for coinid, fee := range fees {
		cfg := cc.getConfig(coinid)
		if cfg == nil {
			continue
		}
		if cfg.Type == FloorFeeType {
			if fee < cfg.Value {
				return fmt.Errorf("The fee must be greater than or equal to %d, but actually it is %d", cfg.Value, fee)
			}
		} else if cfg.Type == EqualFreeType {
			if fee != cfg.Value {
				return fmt.Errorf("The fee must be equal to %d, but actually it is %d", cfg.Value, fee)
			}
		} else {
			return fmt.Errorf("unknown fee type")
		}
	}
	return nil
}

func (cc *CoinConfigs) getConfig(id CoinID) *CoinConfig {
	if len(*cc) <= 0 {
		return nil
	}
	for _, v := range *cc {
		if v.Id == id {
			return v
		}
	}
	return nil
}
