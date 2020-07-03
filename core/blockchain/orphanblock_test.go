/*
 * Copyright (c) 2020.
 * Project:qitmeer
 * File:orphanblock_test.go
 * Date:7/3/20 1:05 PM
 * Author:Jin
 * Email:lochjin@gmail.com
 */

package blockchain

import (
	"math/rand"
	"sort"
	"testing"
)

func Test_SortOrphanBlockSlice(t *testing.T) {
	obs := orphanBlockSlice{}

	for i := uint(0); i < 5; i++ {
		obs = append(obs, &orphanBlock{height: uint64(rand.Intn(100))})
	}
	if len(obs) >= 2 {
		sort.Sort(obs)
	}

}
