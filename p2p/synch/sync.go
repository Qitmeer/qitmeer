/*
 * Copyright (c) 2017-2020 The qitmeer developers
 */

package synch

import (
	"context"
	"fmt"
	"github.com/Qitmeer/qitmeer/p2p/common"
	"github.com/Qitmeer/qitmeer/p2p/encoder"
	"github.com/Qitmeer/qitmeer/p2p/peers"
	pb "github.com/Qitmeer/qitmeer/p2p/proto/v1"
	libp2pcore "github.com/libp2p/go-libp2p-core"
	"github.com/libp2p/go-libp2p-core/network"
	"github.com/libp2p/go-libp2p-core/peer"
	"github.com/libp2p/go-libp2p-core/protocol"
	"reflect"
	"strings"
	"time"
)

const (

	// RPCGoodByeTopic defines the topic for the goodbye rpc method.
	RPCGoodByeTopic = "/qitmeer/req/goodbye/1"
	// RPCPingTopic defines the topic for the ping rpc method.
	RPCPingTopic = "/qitmeer/req/ping/1"
	// RPCMetaDataTopic defines the topic for the metadata rpc method.
	RPCMetaDataTopic = "/qitmeer/req/metadata/1"
	// RPCChainState defines the topic for the chain state rpc method.
	RPCChainState = "/qitmeer/req/chainstate/1"
	// RPCGetBlocks defines the topic for the get blocks rpc method.
	RPCGetBlocks = "/qitmeer/req/getblocks/1"
	// RPCGetBlockDatas defines the topic for the get blocks rpc method.
	RPCGetBlockDatas = "/qitmeer/req/getblockdatas/1"
	// RPCGetBlocks defines the topic for the get blocks rpc method.
	RPCSyncDAG = "/qitmeer/req/syncdag/1"
	// RPCTransaction defines the topic for the transaction rpc method.
	RPCTransaction = "/qitmeer/req/transaction/1"
	// RPCInventory defines the topic for the inventory rpc method.
	RPCInventory = "/qitmeer/req/inventory/1"
	// RPCGraphState defines the topic for the graphstate rpc method.
	RPCGraphState = "/qitmeer/req/graphstate/1"
)

// Time to first byte timeout. The maximum time to wait for first byte of
// request response (time-to-first-byte). The client is expected to give up if
// they don't receive the first byte within 5 seconds.
const ttfbTimeout = 5 * time.Second

// rpcHandler is responsible for handling and responding to any incoming message.
// This method may return an error to internal monitoring, but the error will
// not be relayed to the peer.
type rpcHandler func(context.Context, interface{}, libp2pcore.Stream) error

// RespTimeout is the maximum time for complete response transfer.
const RespTimeout = 10 * time.Second

// ReqTimeout is the maximum time for complete request transfer.
const ReqTimeout = 10 * time.Second

// HandleTimeout is the maximum time for complete handler.
const HandleTimeout = 5 * time.Second

type Sync struct {
	peers    *peers.Status
	peerSync *PeerSync
	p2p      common.P2P
}

func (s *Sync) Start() error {
	s.registerHandlers()

	s.AddConnectionHandler()
	s.AddDisconnectionHandler()

	s.maintainPeerStatuses()

	return s.peerSync.Start()
}

func (s *Sync) Stop() error {
	return s.peerSync.Stop()
}

func (s *Sync) registerHandlers() {
	s.registerRPCHandlers()
	//s.registerSubscribers()
}

// registerRPCHandlers for p2p RPC.
func (s *Sync) registerRPCHandlers() {

	s.registerRPC(
		RPCGoodByeTopic,
		new(uint64),
		s.goodbyeRPCHandler,
	)

	s.registerRPC(
		RPCPingTopic,
		new(uint64),
		s.pingHandler,
	)

	s.registerRPC(
		RPCMetaDataTopic,
		new(interface{}),
		s.metaDataHandler,
	)

	s.registerRPC(
		RPCChainState,
		&pb.ChainState{},
		s.chainStateHandler,
	)

	s.registerRPC(
		RPCGetBlocks,
		&pb.GetBlocks{},
		s.getBlocksHandler,
	)

	s.registerRPC(
		RPCGetBlockDatas,
		&pb.GetBlockDatas{},
		s.getBlockDataHandler,
	)

	s.registerRPC(
		RPCSyncDAG,
		&pb.SyncDAG{},
		s.syncDAGHandler,
	)

	s.registerRPC(
		RPCTransaction,
		&pb.Hash{},
		s.txHandler,
	)

	s.registerRPC(
		RPCInventory,
		&pb.Inventory{},
		s.inventoryHandler,
	)

	s.registerRPC(
		RPCGraphState,
		&pb.GraphState{},
		s.graphStateHandler,
	)
}

// registerRPC for a given topic with an expected protobuf message type.
func (s *Sync) registerRPC(topic string, base interface{}, handle rpcHandler) {
	topic += s.Encoding().ProtocolSuffix()
	s.SetStreamHandler(topic, func(stream network.Stream) {
		ctx, cancel := context.WithTimeout(context.Background(), ttfbTimeout)
		defer cancel()
		defer func() {
			if err := stream.Close(); err != nil {
				log.Error(fmt.Sprintf("topic:%s Failed to close stream:%v", topic, err))
			}
		}()
		if err := stream.SetReadDeadline(time.Now().Add(ttfbTimeout)); err != nil {
			log.Error(fmt.Sprintf("topic:%s peer:%s Could not set stream read deadline:%v",
				topic, stream.Conn().RemotePeer().Pretty(), err))
			return
		}

		// since metadata requests do not have any data in the payload, we
		// do not decode anything.
		if strings.Contains(topic, RPCMetaDataTopic) {
			if err := handle(ctx, new(interface{}), stream); err != nil {
				log.Warn(fmt.Sprintf("Failed to handle p2p RPC:%v", err))
			}
			return
		}

		// Given we have an input argument that can be pointer or [][32]byte, this gives us
		// a way to check for its reflect.Kind and based on the result, we can decode
		// accordingly.
		t := reflect.TypeOf(base)
		var ty reflect.Type
		if t.Kind() == reflect.Ptr {
			ty = t.Elem()
		} else {
			ty = t
		}
		msg := reflect.New(ty)
		if err := s.Encoding().DecodeWithMaxLength(stream, msg.Interface()); err != nil {
			// Debug logs for goodbye errors
			if strings.Contains(topic, RPCGoodByeTopic) {
				log.Debug(fmt.Sprintf("Failed to decode goodbye stream message:%v", err))
				return
			}
			log.Warn(fmt.Sprintf("Failed to decode stream message:%v", err))
			return
		}
		if err := handle(ctx, msg.Interface(), stream); err != nil {
			log.Warn(fmt.Sprintf("Failed to handle p2p RPC:%v", err))
		}
	})
}

// Send a message to a specific peer. The returned stream may be used for reading, but has been
// closed for writing.
func (s *Sync) Send(ctx context.Context, message interface{}, baseTopic string, pid peer.ID) (network.Stream, error) {
	topic := baseTopic + s.Encoding().ProtocolSuffix()

	var deadline = ttfbTimeout + RespTimeout
	ctx, cancel := context.WithTimeout(ctx, deadline)
	defer cancel()

	stream, err := s.p2p.Host().NewStream(ctx, pid, protocol.ID(topic))
	if err != nil {
		return nil, err
	}
	if err := stream.SetReadDeadline(time.Now().Add(deadline)); err != nil {
		return nil, err
	}
	if err := stream.SetWriteDeadline(time.Now().Add(deadline)); err != nil {
		return nil, err
	}
	// do not encode anything if we are sending a metadata request
	if baseTopic == RPCMetaDataTopic {
		return stream, nil
	}

	if _, err := s.Encoding().EncodeWithMaxLength(stream, message); err != nil {
		return nil, err
	}

	// Close stream for writing.
	if err := stream.Close(); err != nil {
		return nil, err
	}

	return stream, nil
}

func (s *Sync) PeerSync() *PeerSync {
	return s.peerSync
}

// Peers returns the peer status interface.
func (s *Sync) Peers() *peers.Status {
	return s.peers
}

func (s *Sync) Encoding() encoder.NetworkEncoding {
	return s.p2p.Encoding()
}

// SetStreamHandler sets the protocol handler on the p2p host multiplexer.
// This method is a pass through to libp2pcore.Host.SetStreamHandler.
func (s *Sync) SetStreamHandler(topic string, handler network.StreamHandler) {
	s.p2p.Host().SetStreamHandler(protocol.ID(topic), handler)
}

func NewSync(p2p common.P2P) *Sync {
	sy := &Sync{p2p: p2p, peers: peers.NewStatus(p2p)}
	sy.peerSync = NewPeerSync(sy)

	return sy
}
