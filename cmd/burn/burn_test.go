// Copyright 2021 The qitmeer developers
// Use of this source code is governed by an ISC
// license that can be found in the LICENSE file.
package main

import (
	"fmt"
	"github.com/Qitmeer/qitmeer/params"
	"strings"
	"testing"
)

func TestBurnAddrss(t *testing.T) {
	tests := []struct{
		param *params.Params
		addr string
	}{
		{ &params.MainNetParams,"MmQitmeerMainnetBurnAddressXXSFKLc1",},
		{ &params.TestNetParams,"TmQitmeerTestnetBurnAddressXXaDBvN7"},
		{ &params.MixNetParams, "XmQitmeerMixnetBurnAddressXXXWkhgxQ"},
		{ &params.PrivNetParams, "RmQitmeerPrivnetBurnAddressXXVVcD5m"},
	}
	for _,test := range tests {
		addr := testGetAddr(test.param, false, t)
		if addr != test.addr {
			t.Fatalf("failed test gen default burn address, expect=%s, but got=%s", test.addr, addr)
		}
		naddr := testGetAddr(test.param, true, t)
		fmt.Printf("test %s burn addr ok! [%s,%s]\n", test.param.Name, addr, naddr)
	}
}

func testGetAddr(p *params.Params, genNew bool, t *testing.T) string{
	template := genTemplateByParams(p)
	addr, err := getAddr(template,p,genNew)
	if err != nil {
		t.Fatalf("Failed to genAddr: %v", err)
	}
	if !strings.HasPrefix(string(addr),template) {
		t.Fatalf("Incorrect %s burn addr %s not prefixed by %s \n",p.Name, addr, template)
	}
	return string(addr)
}