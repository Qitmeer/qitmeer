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
	// RPCGetMerkleBlocks defines the topic for the get merkle blocks rpc method.
	RPCGetMerkleBlocks = "/qitmeer/req/getmerkles/1"
	// RPCFilterAdd defines the topic for the filter add rpc method.
	RPCFilterAdd = "/qitmeer/req/filteradd/1"
	// RPCFilterClear defines the topic for the filter add rpc method.
	RPCFilterClear = "/qitmeer/req/filterclear/1"
	// RPCFilterLoad defines the topic for the filter add rpc method.
	RPCFilterLoad = "/qitmeer/req/filterload/1"
	// RPCMemPool defines the topic for the mempool rpc method.
	RPCMemPool = "/qitmeer/req/mempool/1"
	// RPCMemPool defines the topic for the getdata rpc method.
	RPCGetData = "/qitmeer/req/getdata/1"
)

// Time to first byte timeout. The maximum time to wait for first byte of
// request response (time-to-first-byte). The client is expected to give up if
// they don't receive the first byte within 6 seconds.
const TtfbTimeout = 6 * time.Second

// rpcHandler is responsible for handling and responding to any incoming message.
// This method may return an error to internal monitoring, but the error will
// not be relayed to the peer.
type rpcHandler func(context.Context, interface{}, libp2pcore.Stream) *common.Error

// RespTimeout is the maximum time for complete response transfer.
const RespTimeout = 10 * time.Second

// ReqTimeout is the maximum time for complete request transfer.
const ReqTimeout = 10 * time.Second

// HandleTimeout is the maximum time for complete handler.
const HandleTimeout = 6 * time.Second

type Sync struct {
	peers        *peers.Status
	peerSync     *PeerSync
	p2p          common.P2P
	PeerInterval time.Duration
	LANPeers     map[peer.ID]struct{}

	disconnectionNotify *network.NotifyBundle
	connectionNotify    *network.NotifyBundle
}

func (s *Sync) Start() error {
	s.registerHandlers()

	s.AddConnectionHandler()
	s.AddDisconnectionHandler()

	s.maintainPeerStatuses()

	return s.peerSync.Start()
}

func (s *Sync) Stop() error {
	if s.connectionNotify != nil {
		s.p2p.Host().Network().StopNotify(s.connectionNotify)
	}
	if s.disconnectionNotify != nil {
		s.p2p.Host().Network().StopNotify(s.disconnectionNotify)
	}
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

	s.registerRPC(
		RPCGetMerkleBlocks,
		&pb.MerkleBlockRequest{},
		s.getMerkleBlockDataHandler,
	)

	s.registerRPC(
		RPCFilterAdd,
		&pb.FilterAddRequest{},
		s.HandlerFilterMsgAdd,
	)

	s.registerRPC(
		RPCFilterClear,
		&pb.FilterClearRequest{},
		s.HandlerFilterMsgClear,
	)

	s.registerRPC(
		RPCFilterLoad,
		&pb.FilterLoadRequest{},
		s.HandlerFilterMsgLoad,
	)

	s.registerRPC(
		RPCMemPool,
		&pb.MemPoolRequest{},
		s.HandlerMemPool,
	)

	s.registerRPC(
		RPCGetData,
		&pb.Inventory{},
		s.GetDataHandler,
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
	return EncodeResponseMsg(s.p2p, stream, msg, common.ErrNone)
}

func (s *Sync) EncodeResponseMsgPro(stream libp2pcore.Stream, msg interface{}, retCode common.ErrorCode) *common.Error {
	return EncodeResponseMsg(s.p2p, stream, msg, retCode)
}

func NewSync(p2p common.P2P) *Sync {
	sy := &Sync{p2p: p2p, peers: peers.NewStatus(p2p),
		PeerInterval: params.ActiveNetParams.TargetTimePerBlock * 2,
		LANPeers:     map[peer.ID]struct{}{}}
	sy.peerSync = NewPeerSync(sy)

	for _, pid := range p2p.Config().LANPeers {
		peid, err := peer.Decode(pid)
		if err != nil {
			log.Warn(fmt.Sprintf("LANPeers configuration error:%s", pid))
			continue
		}
		sy.LANPeers[peid] = struct{}{}
	}
	return sy
}

// registerRPC for a given topic with an expected protobuf message type.
func RegisterRPC(rpc common.P2PRPC, basetopic string, base interface{}, handle rpcHandler) {
	topic := getTopic(basetopic) + rpc.Encoding().ProtocolSuffix()

	rpc.Host().SetStreamHandler(protocol.ID(topic), func(stream network.Stream) {
		var e *common.Error
		ctx, cancel := context.WithTimeout(rpc.Context(), TtfbTimeout)
		defer func() {
			processError(e, stream, rpc)
			time.Sleep(time.Second)
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
			size := rpc.Encoding().GetSize(msg)
			rpc.IncreaseBytesRecv(stream.Conn().RemotePeer(), size)
		}

		SetRPCStreamDeadlines(stream)
		e = handle(ctx, msg, stream)
	})
}

func processError(e *common.Error, stream network.Stream, rpc common.P2PRPC) {
	if e == nil {
		return
	}
	peInfo := ""
	if stream != nil {
		peInfo = stream.ID()
		if stream.Conn() != nil {
			peInfo += " "
			peInfo += stream.Conn().RemotePeer().String()
		}
	}
	resp, err := generateErrorResponse(e, rpc.Encoding())
	if err != nil {
		log.Warn(fmt.Sprintf("Failed to generate a response error:%v %s", err, peInfo))
	} else {
		if _, err := stream.Write(resp); err != nil {
			log.Debug(fmt.Sprintf("Failed to write to stream:%v", err))
		}
	}
	if e.Code != common.ErrDAGConsensus {
		log.Debug(fmt.Sprintf("Process error (%s):%s %s", e.Code.String(), e.Error.Error(), peInfo))
	}
}

// Send a message to a specific peer. The returned stream may be used for reading, but has been
// closed for writing.
func Send(ctx context.Context, rpc common.P2PRPC, message interface{}, baseTopic string, pid peer.ID) (network.Stream, error) {
	curState := rpc.Host().Network().Connectedness(pid)
	if curState != network.Connected {
		return nil, fmt.Errorf("%s is %s", pid, curState)
	}

	topic := getTopic(baseTopic) + rpc.Encoding().ProtocolSuffix()

	var deadline = TtfbTimeout + RespTimeout
	ctx, cancel := context.WithTimeout(ctx, deadline)
	defer cancel()

	stream, err := rpc.Host().NewStream(ctx, pid, protocol.ID(topic))
	if err != nil {
		log.Trace(fmt.Sprintf("open stream on topic %v failed", topic))
		return nil, err
	}
	if err := stream.SetReadDeadline(time.Now().Add(deadline)); err != nil {
		log.Trace(fmt.Sprintf("set stream read dealine %v failed", deadline))
		return nil, err
	}
	if err := stream.SetWriteDeadline(time.Now().Add(deadline)); err != nil {
		log.Trace(fmt.Sprintf("set stream write dealine %v failed", deadline))
		return nil, err
	}
	// do not encode anything if we are sending a metadata request
	if baseTopic == RPCMetaDataTopic {
		return stream, nil
	}
	size, err := rpc.Encoding().EncodeWithMaxLength(stream, message)
	if err != nil {
		log.Trace(fmt.Sprintf("encocde rpc message %v to stream failed", message))
		return nil, err
	}
	rpc.IncreaseBytesSent(pid, size)
	// Close stream for writing.
	if err := stream.Close(); err != nil {
		log.Trace(fmt.Sprintf("close stream failed: %v ", err))
		return nil, err
	}

	return stream, nil
}

func EncodeResponseMsg(rpc common.P2PRPC, stream libp2pcore.Stream, msg interface{}, retCode common.ErrorCode) *common.Error {
	_, err := stream.Write([]byte{byte(retCode)})
	if err != nil {
		return common.NewError(common.ErrStreamWrite, err)
	}
	if msg != nil {
		size, err := rpc.Encoding().EncodeWithMaxLength(stream, msg)
		if err != nil {
			return common.NewError(common.ErrStreamWrite, err)
		}
		rpc.IncreaseBytesSent(stream.Conn().RemotePeer(), size)
	}
	return nil
}

func getTopic(baseTopic string) string {
	if baseTopic == RPCChainState || baseTopic == RPCGoodByeTopic {
		return baseTopic
	}
	return baseTopic + "/" + params.ActiveNetParams.Name
}
