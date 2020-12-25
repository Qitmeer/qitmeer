// Copyright (c) 2020 The qitmeer developers
// Use of this source code is governed by an ISC
// license that can be found in the LICENSE file.

package testutils

import "strings"

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

func (s *TestService) Echo(str string, i int, comp *Complex) EchoResult {
	comp.Name = strings.ToUpper(comp.Name)
	return EchoResult{strings.ToUpper(str), i, comp}
}
