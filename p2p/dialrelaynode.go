/*
 * Copyright (c) 2020.
 * Project:qitmeer
 * File:dialrelaynode.go
 * Date:7/15/20 7:05 PM
 * Author:Jin
 * Email:lochjin@gmail.com
 */

package p2p

import (
	"context"
	"github.com/libp2p/go-libp2p-core/host"
	"github.com/libp2p/go-libp2p-core/peer"
	"go.opencensus.io/trace"
)

// MakePeer from multiaddress string.
func MakePeer(addr string) (*peer.AddrInfo, error) {
	maddr, err := multiAddrFromString(addr)
	if err != nil {
		return nil, err
	}
	return peer.AddrInfoFromP2pAddr(maddr)
}

func dialRelayNode(ctx context.Context, h host.Host, relayAddr string) error {
	ctx, span := trace.StartSpan(ctx, "p2p_dialRelayNode")
	defer span.End()

	p, err := MakePeer(relayAddr)
	if err != nil {
		return err
	}

	return h.Connect(ctx, *p)
}
