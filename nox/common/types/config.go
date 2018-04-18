// Copyright 2017-2018 The nox developers

package types

import (
	"math/big"
	"encoding/json"
	"errors"
)

var (
	ErrConfigIdRange = errors.New("field Config.Id < 1")
)

type Config struct {
	Id  *big.Int
}

type configJSON struct {
	Id  *UInt256              `json:"Id"`
}

func (c Config) MarshalJSON() ([]byte, error) {
	var enc configJSON
	enc.Id = (*UInt256)(c.Id)
	return json.Marshal(&enc)
}

func (c *Config) UnmarshalJSON(input []byte) error {
	var dec configJSON
	if err := json.Unmarshal(input, &dec); err != nil {
		return err
	}
	if dec.Id != nil {
		c.Id = (*big.Int)(dec.Id)
		if c.Id.Cmp(big.NewInt(0)) <= 0 {
			return ErrConfigIdRange
		}
	}
	return nil
}
