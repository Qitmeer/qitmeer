// +build !zmq

package zmq

import (
	"fmt"
	"github.com/Qitmeer/qitmeer/config"
	"github.com/zeromq/goczmq"
)

const (
	BlockHash = "BlockHash"
	BlockRaw  = "BlockRaw"
	TxHash    = "TxHash"
	TxRaw     = "TxRaw"
)

type IZMQPublishNotifier interface {
	Init(cfg *config.Config) error
	NotifyBlock()
	NotifyTransaction()
	Shutdown()
}

type ZMQPublishNotifier struct {
	name      string
	endpoints string
	pub       *goczmq.Sock
}

func (zp *ZMQPublishNotifier) initialization(name string, endpoints string) error {
	zp.name = name
	zp.endpoints = endpoints
	log.Info(fmt.Sprintf("Initialize ZMQ public notifier:%s %s", name, endpoints))

	pub, err := goczmq.NewPub(endpoints)
	if err != nil {
		return err
	}
	zp.pub = pub
	return nil
}

func (zp *ZMQPublishNotifier) sendMessage() {

}

func (zp *ZMQPublishNotifier) shutdown() {
	log.Info("Shutdown ZMQ:%s")
	if zp.pub != nil {
		zp.pub.Destroy()
	}
}

// New ZMQ Publish Notifier
func NewZMQPublishNotifier(cfg *config.Config, notifierType string) IZMQPublishNotifier {

	var zmq IZMQPublishNotifier
	switch notifierType {
	case BlockHash:
		zmq = &ZMQBlockHashPublishNotifier{}
	case BlockRaw:
		zmq = &ZMQBlockRawPublishNotifier{}
	case TxHash:
		zmq = &ZMQTxHashPublishNotifier{}
	case TxRaw:
		zmq = &ZMQTxRawPublishNotifier{}
	}
	if zmq == nil {
		return nil
	}

	err := zmq.Init(cfg)
	if err != nil {
		log.Info(fmt.Sprintf("ZMQPublishNotifier can't init:%s", err))
		return nil
	}
	return zmq
}
