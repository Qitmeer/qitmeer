/*
 * Copyright (c) 2017-2020 The qitmeer developers
 */

package runutil

import (
	"context"
	"sync"
	"testing"
	"time"
)

func TestEveryRuns(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())

	lock := sync.RWMutex{}
	i := 0

	RunEvery(ctx, 100*time.Millisecond, func() {
		lock.Lock()
		defer lock.Unlock()
		i++
	})

	// Sleep for a bit and ensure the value has increased.
	time.Sleep(200 * time.Millisecond)

	lock.RLock()
	if i == 0 {
		t.Error("Counter failed to increment with ticker")
	}
	lock.RUnlock()

	cancel()

	// Sleep for a bit to let the cancel take place.
	time.Sleep(100 * time.Millisecond)

	lock.RLock()
	last := i
	lock.RUnlock()

	// Sleep for a bit and ensure the value has not increased.
	time.Sleep(200 * time.Millisecond)

	lock.RLock()
	if i != last {
		t.Error("Counter incremented after stop")
	}
	lock.RUnlock()
}
