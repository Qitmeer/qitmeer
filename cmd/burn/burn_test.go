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

func TestBurnAddrForNetworks(t *testing.T) {
	testNetworkPrefix("Mm", &params.MainNetParams, t)
	testNetworkPrefix("Tm", &params.TestNetParams, t)
	testNetworkPrefix("Xm", &params.MixNetParams, t)
	testNetworkPrefix("Rm", &params.PrivNetParams, t)
}

func testNetworkPrefix(prefix string , p *params.Params, t *testing.T) {
	var sb strings.Builder
	sb.WriteString(prefix)
	sb.WriteString("Qitmeer")
	sb.WriteString(strings.Title(p.Name))
	sb.WriteString("BurnAddress")
	prefixed := sb.String()
	addr, err := getAddr(prefixed, p)
	if err != nil {
		t.Fail();
	}
	if !strings.HasPrefix(string(addr),prefixed) {
	    t.Errorf("Incorrect %s burn addr %s not prefixed by %s \n",p.Name, addr, prefixed)
	}
	fmt.Printf("%s burn addr: %s tested ok!\n",p.Name, addr)
}