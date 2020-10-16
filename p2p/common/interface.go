package common

import (
	"context"
	"github.com/Qitmeer/qitmeer/common/hash"
	"github.com/Qitmeer/qitmeer/core/blockchain"
	"github.com/Qitmeer/qitmeer/p2p/encoder"
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
}
