/*
 * Copyright (c) 2017-2020 The qitmeer developers
 */

package token

import "github.com/Qitmeer/qng-core/core/types"

type TokenFeeConfig struct {
	Type  types.FeeType
	Value int64
}

func (tf *TokenFeeConfig) GetData() int64 {
	return int64(tf.Type) | (tf.Value << 56)
}

func NewTokenFeeConfig(data int64) *TokenFeeConfig {
	return &TokenFeeConfig{Type: types.FeeType(data >> 56), Value: int64(0x00ffffffffffffff & data)}
}
