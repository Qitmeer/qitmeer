/*
 * Copyright (c) 2017-2020 The qitmeer developers
 */

package runutil

import (
	"context"
	"fmt"
	"reflect"
	"runtime"
	"time"
)

// RunEvery runs the provided command periodically.
// It runs in a goroutine, and can be cancelled by finishing the supplied context.
func RunEvery(ctx context.Context, period time.Duration, f func()) {
	funcName := runtime.FuncForPC(reflect.ValueOf(f).Pointer()).Name()
	ticker := time.NewTicker(period)
	go func() {
		for {
			select {
			case <-ticker.C:
				//log.Trace("running")
				f()
			case <-ctx.Done():
				log.Debug(fmt.Sprintf("context is closed, exiting, function:%s", funcName))
				ticker.Stop()
				return
			}
		}
	}()
}
