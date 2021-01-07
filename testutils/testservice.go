// Copyright (c) 2020 The qitmeer developers
// Use of this source code is governed by an ISC
// license that can be found in the LICENSE file.

package testutils

import "strings"

// A dummy service for integrated-test of the rpc client with the test Harness
type TestService struct{}

type Complex struct {
	X, Y int
	Name string
}

type EchoResult struct {
	String  string
	Int     int
	Complex *Complex
}

// The Echo service will echo the input to the uppercase and return back
func (s *TestService) Echo(str string, i int, comp *Complex) EchoResult {
	comp.Name = strings.ToUpper(comp.Name)
	return EchoResult{strings.ToUpper(str), i, comp}
}
