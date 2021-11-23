/*
 * Copyright (c) 2020.
 * Project:qitmeer
 * File:bar.go
 * Date:5/13/20 12:01 PM
 * Author:Jin
 * Email:lochjin@gmail.com
 */

package main

import (
	"fmt"
	"github.com/Qitmeer/qng-core/common/roughtime"
	"os"
	"time"
)

type ProgressBar struct {
	width     int
	max       int
	cur       int
	startTime time.Time
	name      string
}

func (bar *ProgressBar) init(name string) {
	bar.width = 100
	bar.max = 100
	bar.cur = 0
	bar.name = name
}

func (bar *ProgressBar) reset(max int) {
	bar.cur = 0
	bar.max = max

	bar.startTime = roughtime.Now()
}

func (bar *ProgressBar) add() {
	bar.cur++
	if bar.cur > bar.max {
		bar.cur = bar.max
	}
	bar.refresh()
}

func (bar *ProgressBar) setMax() {
	bar.cur = bar.max
	bar.refresh()
}

func (bar *ProgressBar) refresh() {
	cur := float64(bar.cur*100) / float64(bar.max)
	cost := roughtime.Now().Sub(bar.startTime)
	fmt.Fprintf(os.Stdout, "%s %d%% [%s] %s\r", bar.name, int(cur), bar.getProgress(), cost.String())
}

func (bar *ProgressBar) getProgress() string {
	result := make([]byte, bar.width)
	hasCursor := false
	for i := 0; i < bar.width; i++ {
		curI := int(float64(i) / float64(bar.width) * float64(bar.max))
		if curI >= bar.max {
			curI = bar.max - 1
		}
		if !hasCursor {
			if curI >= bar.cur {
				result[i] = []byte(">")[0]
				hasCursor = true
				continue
			}
		}
		if curI < bar.cur {
			result[i] = []byte("=")[0]
		} else {
			result[i] = []byte("-")[0]
		}
	}
	return string(result)
}
