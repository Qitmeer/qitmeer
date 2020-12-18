// Copyright (c) 2020 The qitmeer developers
// Use of this source code is governed by an ISC
// license that can be found in the LICENSE file.

package testutils_test

import (
	"github.com/Qitmeer/qitmeer/params"
	. "github.com/Qitmeer/qitmeer/testutils"
	"sync"
	"testing"
)

func TestHarness(t *testing.T) {
	h, err := NewHarness(t, params.PrivNetParam.Params, nil)
	if err != nil {
		t.Errorf("create new test harness instance failed %v", err)
	}
	if err := h.Setup(); err != nil {
		t.Errorf("setup test harness instance failed %v", err)
	}

	h2, err := NewHarness(t, params.PrivNetParam.Params, nil)
	defer func() {

		if err := h.Teardown(); err != nil {
			t.Errorf("tear down test harness instance failed %v", err)
		}
		numOfHarnessInstances := len(AllHarnesses())
		if numOfHarnessInstances != 10 {
			t.Errorf("harness num is wrong, expect %d , but got %d", 0, numOfHarnessInstances)
			for _, h := range AllHarnesses() {
				t.Errorf("%v\n", h.Id())
			}
		}

		TearDownAll()
		numOfHarnessInstances = len(AllHarnesses())
		if numOfHarnessInstances != 0 {
			t.Errorf("harness num is wrong, expect %d , but got %d", 0, numOfHarnessInstances)
			for _, h := range AllHarnesses() {
				t.Errorf("%v\n", h.Id())
			}
		}

	}()
	numOfHarnessInstances := len(AllHarnesses())
	if numOfHarnessInstances != 2 {
		t.Errorf("harness num is wrong, expect %d , but got %d", 2, numOfHarnessInstances)
	}
	h2.Teardown()
	numOfHarnessInstances = len(AllHarnesses())
	if numOfHarnessInstances != 1 {
		t.Errorf("harness num is wrong, expect %d , but got %d", 1, numOfHarnessInstances)
	}
	var wg sync.WaitGroup
	for i:=0 ; i<10 ; i++ {
		wg.Add(1)
		go func() {
			NewHarness(t, params.PrivNetParam.Params, nil)
			wg.Done()
		}()
	}
	wg.Wait()
}
