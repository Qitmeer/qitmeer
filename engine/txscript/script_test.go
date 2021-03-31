// Copyright (c) 2017-2020 The qitmeer developers
// Copyright (c) 2013-2016 The btcsuite developers
// Copyright (c) 2015-2018 The Decred developers
// Use of this source code is governed by an ISC
// license that can be found in the LICENSE file.
package txscript

import (
	"testing"
)

func TestPushedData(t *testing.T) {
	tests := []struct {
		num int;
		valid bool;
		script []byte}{
		{0, true, []byte{OP_TRUE}},
		{1, true, []byte{OP_DATA_1, OP_TRUE}},
		{2, false, []byte{OP_DATA_2, OP_TRUE}},
	}
	for _,test :=range tests {
		_, err := PushedData(test.script)
		if test.valid && err != nil {
			t.Errorf("TestPushedData failed, valid script got error. test #%d: %v\n", test.num, err)
			continue
		}
		if !test.valid && err == nil {
			t.Errorf("TestPushedData failed, invalid script no error. test #%d: %v\n", test.num, test.script)
			continue
		}
	}
}

func TestIsPushOnlyScript(t *testing.T) {
	tests := []struct{
		num int;
		script []byte}{
		{0,[]byte{OP_TRUE}},
		{1, []byte{OP_DATA_1, OP_TRUE}},
		{2, []byte{OP_DATA_2, OP_TRUE, OP_FALSE}},
	}
	for _,test :=range tests {
		if !IsPushOnlyScript(test.script) {
			t.Errorf("IsPushOnlyScript: test %d failed: %x\n", test.num, test.script)
			continue
		}
	}
}