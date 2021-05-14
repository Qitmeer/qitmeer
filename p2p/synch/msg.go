/*
 * Copyright (c) 2017-2020 The qitmeer developers
 */

package synch

import (
	"github.com/Qitmeer/qitmeer/common/hash"
	"github.com/Qitmeer/qitmeer/core/types"
	"github.com/Qitmeer/qitmeer/p2p/peers"
	pb "github.com/Qitmeer/qitmeer/p2p/proto/v1"
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

type GetDatasMsg struct {
	pe   *peers.Peer
	data *pb.Inventory
}

type OnFilterAddMsg struct {
	pe   *peers.Peer
	data *types.MsgFilterAdd
}

type OnFilterClearMsg struct {
	pe   *peers.Peer
	data *types.MsgFilterClear
}

type OnFilterLoadMsg struct {
	pe   *peers.Peer
	data *types.MsgFilterLoad
}

type OnMsgMemPool struct {
	pe   *peers.Peer
	data *MsgMemPool
}

type UpdateGraphStateMsg struct {
	pe *peers.Peer
}

type syncDAGBlocksMsg struct {
	pe *peers.Peer
}

type PeerUpdateMsg struct {
	pe     *peers.Peer
	orphan bool
}

type getTxsMsg struct {
	pe  *peers.Peer
	txs []*hash.Hash
}

type SyncQNRMsg struct {
	pe  *peers.Peer
	qnr string
}
