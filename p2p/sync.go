package p2p

import (
	"context"
	"fmt"
	pb "github.com/Qitmeer/qitmeer/p2p/proto/v1"
	libp2pcore "github.com/libp2p/go-libp2p-core"
	"github.com/libp2p/go-libp2p-core/network"
	"github.com/libp2p/go-libp2p-core/peer"
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
	// RPCGetBlocks defines the topic for the get blocks rpc method.
	RPCSyncDAG = "/qitmeer/req/syncdag/1"
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

func (s *Service) registerHandlers() {
	s.registerRPCHandlers()
	//s.registerSubscribers()
}

func (s *Service) startSync() {
	s.peerSync = NewPeerSync(s)
	s.peerSync.Start()

	s.AddConnectionHandler(s.reValidatePeer, s.sendGenericGoodbyeMessage)
	s.AddDisconnectionHandler(func(_ context.Context, _ peer.ID) error {
		// TODO
		return nil
	})

	s.AddPingMethod(s.sendPingRequest)

	s.maintainPeerStatuses()

}

// registerRPCHandlers for p2p RPC.
func (s *Service) registerRPCHandlers() {

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
		&pb.Hash{},
		s.getBlocksHandler,
	)

	s.registerRPC(
		RPCSyncDAG,
		&pb.SyncDAG{},
		s.syncDAGHandler,
	)
}

// registerRPC for a given topic with an expected protobuf message type.
func (s *Service) registerRPC(topic string, base interface{}, handle rpcHandler) {
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

func closeSteam(stream libp2pcore.Stream) {
	if err := stream.Close(); err != nil {
		log.Error(fmt.Sprintf("Failed to close stream:%v", err))
	}
}

// sync
