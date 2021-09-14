/*
 * Copyright (c) 2017-2020 The qitmeer developers
 */

package blockdag

import "fmt"

type BlueInfo struct {
	num    uint
	rate   int64
	weight int64
}

func (bi *BlueInfo) GetNum() uint {
	return bi.num
}

func (bi *BlueInfo) GetRate() int64 {
	return bi.rate
}

func (bi *BlueInfo) GetWeight() int64 {
	return bi.weight
}

func (bi *BlueInfo) String() string {
	return fmt.Sprintf("Blue Info:num=%d rate=%d", bi.num, bi.rate)
}

func NewBlueInfo(num uint, rate int64, weight int64) *BlueInfo {
	return &BlueInfo{num: num, rate: rate, weight: weight}
}
