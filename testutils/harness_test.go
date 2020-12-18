// Copyright (c) 2020 The qitmeer developers
// Use of this source code is governed by an ISC
// license that can be found in the LICENSE file.

package testutils

import (
	"github.com/Qitmeer/qitmeer/params"
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
	defer func() {
		if err := h.Teardown(); err != nil {
			t.Errorf("tear down test harness instance failed %v", err)
		}
	}()
	_, err = NewHarness(t, params.PrivNetParam.Params, nil)
	if numOfHarnessInstances != 2 {
		t.Errorf("harness num is wrong, expect %d , but got %d", 2, numOfHarnessInstances)
	}
}
