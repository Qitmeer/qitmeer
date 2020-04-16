// +build !zmq

package zmq

import (
	"fmt"
	"github.com/Qitmeer/qitmeer/config"
	"github.com/Qitmeer/qitmeer/core/types"
)

type ZMQNotification struct {
	cfg             *config.Config
	publishNotifier *ZMQPublishNotifier
}

func (zn *ZMQNotification) Init(cfg *config.Config) {
	log.Info("ZMQ:Supported")
	zn.cfg = cfg

	zn.publishNotifier = NewZMQPublishNotifier(cfg)
}

func (zn *ZMQNotification) IsEnable() bool {
	return true
}

// block accepted
func (zn *ZMQNotification) BlockAccepted(block *types.SerializedBlock) {
	log.Info(fmt.Sprintf("BlockAccepted:%s", block.Hash().String()))
}

// block connected
func (zn *ZMQNotification) BlockConnected(block *types.SerializedBlock) {
	log.Info(fmt.Sprintf("BlockConnected:%s", block.Hash().String()))
}

// block connected
func (zn *ZMQNotification) BlockDisconnected(block *types.SerializedBlock) {
	log.Info(fmt.Sprintf("BlockDisconnected:%s", block.Hash().String()))
}
