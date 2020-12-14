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
	"github.com/Qitmeer/qitmeer/params"
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
	// RPCSyncQNR defines the topic for the syncqnr rpc method.
	RPCSyncQNR = "/qitmeer/req/syncqnr/1"
)

// Time to first byte timeout. The maximum time to wait for first byte of
// request response (time-to-first-byte). The client is expected to give up if
// they don't receive the first byte within 5 seconds.
const TtfbTimeout = 5 * time.Second

// rpcHandler is responsible for handling and responding to any incoming message.
// This method may return an error to internal monitoring, but the error will
// not be relayed to the peer.
type rpcHandler func(context.Context, interface{}, libp2pcore.Stream) *common.Error

// RespTimeout is the maximum time for complete response transfer.
const RespTimeout = 10 * time.Second

// ReqTimeout is the maximum time for complete request transfer.
const ReqTimeout = 10 * time.Second

// HandleTimeout is the maximum time for complete handler.
const HandleTimeout = 5 * time.Second

type Sync struct {
	peers        *peers.Status
	peerSync     *PeerSync
	p2p          common.P2P
	PeerInterval time.Duration
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
		nil,
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

	s.registerRPC(
		RPCSyncQNR,
		&pb.SyncQNR{},
		s.QNRHandler,
	)
}

// registerRPC for a given topic with an expected protobuf message type.
func (s *Sync) registerRPC(topic string, base interface{}, handle rpcHandler) {
	RegisterRPC(s.p2p, topic, base, handle)
}

// Send a message to a specific peer. The returned stream may be used for reading, but has been
// closed for writing.
func (s *Sync) Send(ctx context.Context, message interface{}, baseTopic string, pid peer.ID) (network.Stream, error) {
	return Send(ctx, s.p2p, message, baseTopic, pid)
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

func (s *Sync) EncodeResponseMsg(stream libp2pcore.Stream, msg interface{}) *common.Error {
	return EncodeResponseMsg(s.p2p, stream, msg)
}

func NewSync(p2p common.P2P) *Sync {
	sy := &Sync{p2p: p2p, peers: peers.NewStatus(p2p),
		PeerInterval: params.ActiveNetParams.TargetTimePerBlock * 2}
	sy.peerSync = NewPeerSync(sy)

	return sy
}

// registerRPC for a given topic with an expected protobuf message type.
func RegisterRPC(rpc common.P2PRPC, topic string, base interface{}, handle rpcHandler) {
	topic += rpc.Encoding().ProtocolSuffix()
	rpc.Host().SetStreamHandler(protocol.ID(topic), func(stream network.Stream) {
		var e *common.Error
		ctx, cancel := context.WithTimeout(rpc.Context(), TtfbTimeout)
		defer func() {
			processError(e, stream, rpc)
			cancel()
			closeSteam(stream)
		}()
		if err := stream.SetReadDeadline(time.Now().Add(TtfbTimeout)); err != nil {
			log.Error(fmt.Sprintf("topic:%s peer:%s Could not set stream read deadline:%v",
				topic, stream.Conn().RemotePeer().Pretty(), err))
			e = common.NewError(common.ErrStreamBase, err)
			return
		}

		// Given we have an input argument that can be pointer or [][32]byte, this gives us
		// a way to check for its reflect.Kind and based on the result, we can decode
		// accordingly.
		var msg interface{}
		if base != nil {
			t := reflect.TypeOf(base)
			var ty reflect.Type
			if t.Kind() == reflect.Ptr {
				ty = t.Elem()
			} else {
				ty = t
			}
			msgT := reflect.New(ty)
			msg = msgT.Interface()
			if err := rpc.Encoding().DecodeWithMaxLength(stream, msg); err != nil {
				e = common.NewError(common.ErrStreamRead, err)
				// Debug logs for goodbye errors
				if strings.Contains(topic, RPCGoodByeTopic) {
					log.Debug(fmt.Sprintf("Failed to decode goodbye stream message:%v", err))
					return
				}
				log.Warn(fmt.Sprintf("Failed to decode stream message:%v", err))
				return
			}
		}

		SetRPCStreamDeadlines(stream)
		if e = handle(ctx, msg, stream); e != nil {
			log.Warn(fmt.Sprintf("Failed to handle p2p RPC:%v", e.Error.Error()))
		}
	})
}

func processError(e *common.Error, stream network.Stream, rpc common.P2PRPC) {
	if e == nil {
		return
	}
	resp, err := generateErrorResponse(e, rpc.Encoding())
	if err != nil {
		log.Error(fmt.Sprintf("Failed to generate a response error:%v", err))
	} else {
		if _, err := stream.Write(resp); err != nil {
			log.Debug(fmt.Sprintf("Failed to write to stream:%v", err))
		}
	}
	if e.Code == common.ErrDAGConsensus {
		if err := sendGoodByeAndDisconnect(rpc.Context(), common.ErrDAGConsensus, stream.Conn().RemotePeer(), rpc); err != nil {
			log.Error(err.Error())
			return
		}
	} else {
		log.Warn(fmt.Sprintf("Process error (%s):%s", e.Code.String(), e.Error.Error()))
	}
}

// Send a message to a specific peer. The returned stream may be used for reading, but has been
// closed for writing.
func Send(ctx context.Context, rpc common.P2PRPC, message interface{}, baseTopic string, pid peer.ID) (network.Stream, error) {
	topic := baseTopic + rpc.Encoding().ProtocolSuffix()

	var deadline = TtfbTimeout + RespTimeout
	ctx, cancel := context.WithTimeout(ctx, deadline)
	defer cancel()

	stream, err := rpc.Host().NewStream(ctx, pid, protocol.ID(topic))
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

	if _, err := rpc.Encoding().EncodeWithMaxLength(stream, message); err != nil {
		return nil, err
	}

	// Close stream for writing.
	if err := stream.Close(); err != nil {
		return nil, err
	}

	return stream, nil
}

func EncodeResponseMsg(rpc common.P2PRPC, stream libp2pcore.Stream, msg interface{}) *common.Error {
	_, err := stream.Write([]byte{byte(common.ErrNone)})
	if err != nil {
		return common.NewError(common.ErrStreamWrite, err)
	}
	if msg != nil {
		_, err = rpc.Encoding().EncodeWithMaxLength(stream, msg)
		if err != nil {
			return common.NewError(common.ErrStreamWrite, err)
		}
	}
	return nil
}
