// +build zmq

package zmq

import (
	"fmt"
	"github.com/Qitmeer/qng-core/config"
	"github.com/Qitmeer/qng-core/core/types"
	"github.com/zeromq/goczmq"
)

const (
	BlockHash = "BlockHash"
	BlockRaw  = "BlockRaw"
	TxHash    = "TxHash"
	TxRaw     = "TxRaw"

	defaultBlockHashEndpoint = "tcp://*:8230"
	defaultBlockRawEndpoint  = "tcp://*:8231"
	defaultTxHashEndpoint    = "tcp://*:8232"
	defaultTxRawEndpoint     = "tcp://*:8233"
)

type IZMQPublishNotifier interface {
	Init(cfg *config.Config) error
	NotifyBlock(block *types.SerializedBlock) error
	NotifyTransaction(transaction []*types.Tx) error
	Shutdown()
}

type ZMQPublishNotifier struct {
	name      string
	endpoints string
	pub       *goczmq.Sock
}

func (zp *ZMQPublishNotifier) initialization(endpoints string) error {
	zp.endpoints = endpoints
	log.Info(fmt.Sprintf("Initialize ZMQ public notifier:%s %s", zp.name, endpoints))

	pub, err := goczmq.NewPub(endpoints)
	if err != nil {
		log.Error(fmt.Sprintf("%s ZMQ Publish Notifier can't initialization:%v", zp.name, err))
		return err
	}
	zp.pub = pub
	return nil
}

func (zp *ZMQPublishNotifier) sendMessage(data []byte, more bool) error {
	if zp.pub == nil {
		return fmt.Errorf("No pub")
	}
	flags := goczmq.FlagNone
	if more {
		flags = goczmq.FlagMore
	}
	err := zp.pub.SendFrame(data, flags)
	if err != nil {
		fmt.Errorf("Send message error:%v", err)
		return err
	}
	return nil
}

func (zp *ZMQPublishNotifier) shutdown() {
	log.Info(fmt.Sprintf("Shutdown:ZMQPublishNotifier [%s ---> %s]", zp.name, zp.endpoints))
	if zp.pub != nil {
		zp.pub.Destroy()
	}
}

// New ZMQ Publish Notifier
func NewZMQPublishNotifier(cfg *config.Config, notifierType string) IZMQPublishNotifier {

	var zmq IZMQPublishNotifier
	switch notifierType {
	case BlockHash:
		zmq = &ZMQBlockHashPublishNotifier{&ZMQPublishNotifier{name: notifierType}}
	case BlockRaw:
		zmq = &ZMQBlockRawPublishNotifier{&ZMQPublishNotifier{name: notifierType}}
	case TxHash:
		zmq = &ZMQTxHashPublishNotifier{&ZMQPublishNotifier{name: notifierType}}
	case TxRaw:
		zmq = &ZMQTxRawPublishNotifier{&ZMQPublishNotifier{name: notifierType}}
	}
	if zmq == nil {
		return nil
	}

	err := zmq.Init(cfg)
	if err != nil {
		return nil
	}
	return zmq
}
