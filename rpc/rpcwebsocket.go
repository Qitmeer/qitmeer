/*
 * Copyright (c) 2017-2020 The qitmeer developers
 */

package rpc

import (
	"fmt"
	"github.com/Qitmeer/qitmeer/rpc/websocket"
	"github.com/prometheus/common/log"
)

func (s *RpcServer) WebsocketHandler(conn *websocket.Conn, remoteAddr string, isAdmin bool) {

	// Limit max number of websocket clients.
	log.Info(fmt.Sprintf("New websocket client %s", remoteAddr))
}
