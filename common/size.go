// Copyright (c) 2017-2019 The Qitmeer developers
//
// Use of this source code is governed by an ISC
// license that can be found in the LICENSE file.

// The parts code inspired & originated from
// http://github.com/ethereum/go-ethereum/common/size.go

package common

import (
	"fmt"
)

// StorageSize is a wrapper around a float value that supports user friendly
// formatting.
type StorageSize float64

// String implements the stringer interface.
func (s StorageSize) String() string {
	if s > 1000000 {
		return fmt.Sprintf("%.2f mB", s/1000000)
	} else if s > 1000 {
		return fmt.Sprintf("%.2f kB", s/1000)
	} else {
		return fmt.Sprintf("%.2f B", s)
	}
}

// TerminalString implements log.TerminalStringer, formatting a string for console
// output during logging.
func (s StorageSize) TerminalString() string {
	if s > 1000000 {
		return fmt.Sprintf("%.2fmB", s/1000000)
	} else if s > 1000 {
		return fmt.Sprintf("%.2fkB", s/1000)
	} else {
		return fmt.Sprintf("%.2fB", s)
	}
}
