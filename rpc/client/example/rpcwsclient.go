/*
 * Copyright (c) 2017-2020 The qitmeer developers
 */

// Example connect to local qitmeer RPC server using websockets.

package main

import (
	"fmt"
	"github.com/Qitmeer/qitmeer/common/hash"
	"github.com/Qitmeer/qitmeer/common/util"
	"github.com/Qitmeer/qitmeer/rpc/client"
	"io/ioutil"
	"log"
	"path/filepath"
	"time"
)

func main() {
	ntfnHandlers := client.NotificationHandlers{
		OnBlockConnected: func(hash *hash.Hash, order int64, t time.Time) {
			fmt.Println("OnBlockConnected", hash, order)
		},
		OnBlockDisconnected: func(hash *hash.Hash, order int64, t time.Time) {
			fmt.Println("OnBlockDisconnected", hash, order)
		},
	}

	connCfg := &client.ConnConfig{
		Host:       "localhost:1234",
		Endpoint:   "ws",
		User:       "test",
		Pass:       "test",
		DisableTLS: false,
	}
	if !connCfg.DisableTLS {
		homeDir := util.AppDataDir("qitmeerd", false)
		certs, err := ioutil.ReadFile(filepath.Join(homeDir, "rpc.cert"))
		if err != nil {
			log.Fatal(err)
		}
		connCfg.Certificates = certs
	}

	client, err := client.New(connCfg, &ntfnHandlers)
	if err != nil {
		log.Fatal(err)
	}

	// Register for block connect and disconnect notifications.
	if err := client.NotifyBlocks(); err != nil {
		log.Fatal(err)
	}
	log.Println("NotifyBlocks: Registration Complete")

	// call RPC: getNodeInfo
	nodeInfo, err := client.GetNodeInfo()
	if err != nil {
		log.Fatal(err)
	}
	log.Printf("Node info: %v", nodeInfo)

	waitTime := time.Second * 15
	log.Printf("Client shutdown in %s...\n", waitTime.String())
	time.AfterFunc(waitTime, func() {
		log.Println("Client shutting down...")
		client.Shutdown()
		log.Println("Client shutdown complete.")
	})
	client.WaitForShutdown()
}
