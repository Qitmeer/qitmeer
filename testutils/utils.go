// Copyright (c) 2020 The qitmeer developers
// Use of this source code is governed by an ISC
// license that can be found in the LICENSE file.

package testutils

import "testing"

// GenerateBlock will generate a number of blocks by the input number for
// the appointed test harness.
// It will return the hashes of the generated blocks or an error
func GenerateBlock(t *testing.T, h *Harness, num uint64) {
	if blocks, err := h.Client.Generate(num); err != nil {
		t.Errorf("generate block failed : %v", err)
	} else {
		for _, b := range blocks {
			t.Logf("node [%v] generate block [%v] ok", h.Node.Id(), b)
		}
	}
}

// AssertBlockOrderAndHeight will verify the current block order, total block number
// and current main-chain height of the appointed test harness and assert it ok or
// cause the test failed.
func AssertBlockOrderAndHeight(t *testing.T, h *Harness, order, total, height uint64) {
	// order
	if c, err := h.Client.BlockCount(); err != nil {
		t.Errorf("test failed : %v", err)
	} else {
		expect := order
		if c != expect {
			t.Errorf("test failed, expect %v , but got %v", expect, c)
		}
	}
	// total block
	if tal, err := h.Client.BlockTotal(); err != nil {
		t.Errorf("test failed : %v", err)
	} else {
		expect := total
		if tal != expect {
			t.Errorf("test failed, expect %v , but got %v", expect, tal)
		}
	}
	// main height
	if h, err := h.Client.MainHeight(); err != nil {
		t.Errorf("test failed : %v", err)
	} else {
		expect := height
		if h != expect {
			t.Errorf("test failed, expect %v , but got %v", expect, h)
		}
	}
}

func PayAndSend(t *testing.T, h *Harness) {
	addr, err := h.Wallet.newAddress()
	if err != nil {
		t.Fatalf("failed to generate new address for test wallet: %v", err)
	}
	t.Logf("test wallet generated new address %v ok", addr.Encode())

}
