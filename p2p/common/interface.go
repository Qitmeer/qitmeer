package common

import (
	"context"
	"github.com/Qitmeer/qitmeer/common/hash"
	"github.com/Qitmeer/qitmeer/core/blockchain"
	"github.com/Qitmeer/qitmeer/node/notify"
	"github.com/Qitmeer/qitmeer/p2p/encoder"
	pb "github.com/Qitmeer/qitmeer/p2p/proto/v1"
	"github.com/Qitmeer/qitmeer/p2p/qnode"
	"github.com/Qitmeer/qitmeer/services/mempool"
	"github.com/libp2p/go-libp2p-core/host"
	"github.com/libp2p/go-libp2p-core/peer"
)

type P2P interface {
	GetGenesisHash() *hash.Hash
	BlockChain() *blockchain.BlockChain
	Host() host.Host
	Disconnect(pid peer.ID) error
	Context() context.Context
	Encoding() encoder.NetworkEncoding
	Config() *Config
	TxMemPool() *mempool.TxPool
	Metadata() *pb.MetaData
	MetadataSeq() uint64
	TimeSource() blockchain.MedianTimeSource
	Notify() notify.Notify
	ConnectTo(node *qnode.Node)
	Resolve(n *qnode.Node) *qnode.Node
	Node() *qnode.Node
	RelayNodeInfo() *peer.AddrInfo
	IncreaseBytesSent(pid peer.ID, size int)
	IncreaseBytesRecv(pid peer.ID, size int)
}

type P2PRPC interface {
	Host() host.Host
	Context() context.Context
	Encoding() encoder.NetworkEncoding
	Disconnect(pid peer.ID) error
	IncreaseBytesSent(pid peer.ID, size int)
	IncreaseBytesRecv(pid peer.ID, size int)
}
