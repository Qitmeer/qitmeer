/*
 * Copyright (c) 2017-2020 The qitmeer developers
 */

package rpc

import (
	"fmt"
	"github.com/Qitmeer/qitmeer/core/blockchain"
	"github.com/Qitmeer/qitmeer/core/event"
	"github.com/Qitmeer/qitmeer/core/types"
	"github.com/Qitmeer/qitmeer/rpc/websocket"
	"time"
)

var timeZeroVal time.Time

func (s *RpcServer) subscribe(events *event.Feed) {
	ch := make(chan *event.Event)
	sub := events.Subscribe(ch)
	go func() {
		defer sub.Unsubscribe()
		for {
			select {
			case ev := <-ch:
				if ev.Data != nil {
					switch value := ev.Data.(type) {
					case *blockchain.Notification:
						s.handleNotifyMsg(value)
					}
				}
				if ev.Ack != nil {
					ev.Ack <- struct{}{}
				}
			case <-s.quit:
				log.Info("Close RpcServer Event Subscribe")
				return
			}
		}
	}()
}

func (s *RpcServer) handleNotifyMsg(notification *blockchain.Notification) {
	switch notification.Type {
	case blockchain.BlockAccepted:
		// TODO ACCEPTED
		_, ok := notification.Data.(*blockchain.BlockAcceptedNotifyData)
		if !ok {
			log.Warn("Chain accepted notification is not " +
				"BlockAcceptedNotifyData.")
			break
		}

	case blockchain.BlockConnected:
		blockSlice, ok := notification.Data.([]*types.SerializedBlock)
		if !ok {
			log.Warn("Chain connected notification is not a block slice.")
			break
		}

		if len(blockSlice) != 1 {
			log.Warn("Chain connected notification is wrong size slice.")
			break
		}
		s.ntfnMgr.NotifyBlockConnected(blockSlice[0].Block())

	case blockchain.BlockDisconnected:
		block, ok := notification.Data.(*types.SerializedBlock)
		if !ok {
			log.Warn("Chain disconnected notification is not a block slice.")
			break
		}
		s.ntfnMgr.NotifyBlockDisconnected(block.Block())
	}
}

func (s *RpcServer) WebsocketHandler(conn *websocket.Conn, remoteAddr string, isAdmin bool) {
	// Clear the read deadline that was set before the websocket hijacked
	// the connection.
	conn.SetReadDeadline(timeZeroVal)
	// Limit max number of websocket clients.
	log.Info(fmt.Sprintf("New websocket client %s", remoteAddr))
	if s.ntfnMgr.NumClients()+1 > s.config.RPCMaxWebsockets {
		log.Info(fmt.Sprintf("Max websocket clients exceeded [%d] - disconnecting client %s", s.config.RPCMaxWebsockets,
			remoteAddr))
		conn.Close()
		return
	}

	// Create a new websocket client to handle the new websocket connection
	// and wait for it to shutdown.  Once it has shutdown (and hence
	// disconnected), remove it and any notifications it registered for.
	client, err := newWebsocketClient(s, conn, remoteAddr, isAdmin)
	if err != nil {
		log.Error(fmt.Sprintf("Failed to serve client %s: %v", remoteAddr, err))
		conn.Close()
		return
	}
	s.ntfnMgr.AddClient(client)
	client.Start()
	client.WaitForShutdown()
	s.ntfnMgr.RemoveClient(client)
	log.Info(fmt.Sprintf("Disconnected websocket client %s", remoteAddr))
}
