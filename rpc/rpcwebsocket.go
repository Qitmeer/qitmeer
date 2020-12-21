/*
 * Copyright (c) 2017-2020 The qitmeer developers
 */

package rpc

import (
	"fmt"
	"github.com/Qitmeer/qitmeer/rpc/websocket"
	"github.com/prometheus/common/log"
	"time"
)

var timeZeroVal time.Time

func (s *RpcServer) WebsocketHandler(conn *websocket.Conn, remoteAddr string, isAdmin bool) {
	// Clear the read deadline that was set before the websocket hijacked
	// the connection.
	conn.SetReadDeadline(timeZeroVal)
	// Limit max number of websocket clients.
	log.Info(fmt.Sprintf("New websocket client %s", remoteAddr))
	if s.ntfnMgr.NumClients()+1 > cfg.RPCMaxWebsockets {
		rpcsLog.Infof("Max websocket clients exceeded [%d] - "+
			"disconnecting client %s", cfg.RPCMaxWebsockets,
			remoteAddr)
		conn.Close()
		return
	}

	// Create a new websocket client to handle the new websocket connection
	// and wait for it to shutdown.  Once it has shutdown (and hence
	// disconnected), remove it and any notifications it registered for.
	client, err := newWebsocketClient(s, conn, remoteAddr, authenticated, isAdmin)
	if err != nil {
		rpcsLog.Errorf("Failed to serve client %s: %v", remoteAddr, err)
		conn.Close()
		return
	}
	s.ntfnMgr.AddClient(client)
	client.Start()
	client.WaitForShutdown()
	s.ntfnMgr.RemoveClient(client)
	rpcsLog.Infof("Disconnected websocket client %s", remoteAddr)
}
