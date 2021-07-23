// Copyright 2021 The qitmeer developers
// Use of this source code is governed by an ISC
// license that can be found in the LICENSE file.
package main

import (
	"fmt"
	"strings"
	"testing"
)

func TestBurnAddrss(t *testing.T) {
	tests := []struct{
		network string
		addr string
	}{
		{ "mainnet", "MmQitmeerMainnetBurnAddressXXSFKLc1",},
		{ "0.9testnet", "TmQitmeerTestnetBurnAddressXXaDBvN7"},
		{ "testnet","TnQitmeerTestnetBurnAddressXXd8arKJ"},
		{ "mixnet", "XmQitmeerMixnetBurnAddressXXXWkhgxQ"},
		{ "privnet","RmQitmeerPrivnetBurnAddressXXVVcD5m"},
	}
	for _,test := range tests {
		p, err :=getParams(test.network)
		if err != nil {
			t.Fatalf("Failed to getParams: %v", err)
		}
		addr := testGetAddr(p, test.network, false, t)
		if addr != test.addr {
			t.Fatalf("failed test gen default burn address, expect=%s, but got=%s", test.addr, addr)
		}
		naddr := testGetAddr(p, test.network, true, t)
		fmt.Printf("test %s burn addr ok! default=%s, new=%s\n",test.network, addr, naddr)
	}
}

func testGetAddr(p *NetParams, network string, genNew bool, t *testing.T) string{
	template := genTemplateByParams(p,network)
	addr, err := getAddr(template,p,genNew)
	if err != nil {
		t.Fatalf("Failed to genAddr: %v", err)
	}
	if !strings.HasPrefix(string(addr),template) {
		t.Fatalf("Incorrect %s burn addr %s not prefixed by %s \n",p.Name, addr, template)
	}
	return string(addr)
}