// +build !zmq

package zmq

import (
	"fmt"
	"github.com/Qitmeer/qitmeer/config"
	"github.com/Qitmeer/qitmeer/core/types"
)

type ZMQNotification struct {
	cfg              *config.Config
	publishNotifiers []IZMQPublishNotifier
}

func (zn *ZMQNotification) Init(cfg *config.Config) {
	log.Info("ZMQ:Supported")
	zn.cfg = cfg

	zn.publishNotifiers = []IZMQPublishNotifier{}
	notiTypeArr := []string{BlockHash, BlockRaw, TxHash, TxRaw}
	for _, notiType := range notiTypeArr {
		publishNotifier := NewZMQPublishNotifier(cfg, notiType)
		if publishNotifier != nil {
			zn.publishNotifiers = append(zn.publishNotifiers, publishNotifier)
		}
	}
}

func (zn *ZMQNotification) IsEnable() bool {
	return true
}

// block accepted
func (zn *ZMQNotification) BlockAccepted(block *types.SerializedBlock) {
	log.Info(fmt.Sprintf("BlockAccepted:%s", block.Hash().String()))

	for _, notifier := range zn.publishNotifiers {
		notifier.NotifyBlock()
	}
}

// block connected
func (zn *ZMQNotification) BlockConnected(block *types.SerializedBlock) {
	log.Info(fmt.Sprintf("BlockConnected:%s", block.Hash().String()))
}

// block connected
func (zn *ZMQNotification) BlockDisconnected(block *types.SerializedBlock) {
	log.Info(fmt.Sprintf("BlockDisconnected:%s", block.Hash().String()))
}
