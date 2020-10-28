/*
 * Copyright (c) 2020.
 * Project:qitmeer
 * File:every_test.go
 * Date:7/15/20 4:55 PM
 * Author:Jin
 * Email:lochjin@gmail.com
 */

package runutil

import (
	"context"
	"testing"
	"time"
)

func TestEveryRuns(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())

	i := 0
	RunEvery(ctx, 100*time.Millisecond, func() {
		i++
	})

	// Sleep for a bit and ensure the value has increased.
	time.Sleep(200 * time.Millisecond)

	if i == 0 {
		t.Error("Counter failed to increment with ticker")
	}

	cancel()

	// Sleep for a bit to let the cancel take place.
	time.Sleep(100 * time.Millisecond)

	last := i

	// Sleep for a bit and ensure the value has not increased.
	time.Sleep(200 * time.Millisecond)

	if i != last {
		t.Error("Counter incremented after stop")
	}
}
