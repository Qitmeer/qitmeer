/*
 * Copyright (c) 2017-2020 The qitmeer developers
 */

package synch

import (
	"github.com/Qitmeer/qitmeer/common/hash"
	"github.com/Qitmeer/qitmeer/p2p/peers"
	"github.com/libp2p/go-libp2p-core/network"
	"github.com/libp2p/go-libp2p-core/peer"
)

type pauseMsg struct {
	unpause <-chan struct{}
}

type ConnectedMsg struct {
	ID   peer.ID
	Conn network.Conn
}

type DisconnectedMsg struct {
	ID   peer.ID
	Conn network.Conn
}

type GetBlocksMsg struct {
	pe     *peers.Peer
	blocks []*hash.Hash
}

type GetBlockDatasMsg struct {
	pe     *peers.Peer
	blocks []*hash.Hash
}

type UpdateGraphStateMsg struct {
	pe *peers.Peer
}

type syncDAGBlocksMsg struct {
	pe *peers.Peer
}

type PeerUpdateMsg struct {
	pe *peers.Peer
}

type getTxsMsg struct {
	pe  *peers.Peer
	txs []*hash.Hash
}
