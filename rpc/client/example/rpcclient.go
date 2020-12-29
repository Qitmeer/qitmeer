/*
 * Copyright (c) 2017-2020 The qitmeer developers
 */

// Example connect to local qitmeer RPC server using http.

package main

import (
	"github.com/Qitmeer/qitmeer/rpc/client"
	"log"
)

func main() {
	connCfg := &client.ConnConfig{
		Host:         "localhost:1234",
		User:         "test",
		Pass:         "test",
		DisableTLS:   true,
		HTTPPostMode: true,
	}
	client, err := client.New(connCfg, nil)
	if err != nil {
		log.Fatal(err)
	}
	// call RPC: getNodeInfo
	nodeInfo, err := client.GetNodeInfo()
	if err != nil {
		log.Fatal(err)
	}
	log.Printf("Node info: %v", nodeInfo)
}
