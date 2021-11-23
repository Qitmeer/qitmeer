// +build zmq

package zmq

import (
	"fmt"
	"github.com/Qitmeer/qng-core/config"
	"github.com/Qitmeer/qng-core/core/types"
)

// This ZeroMQ notification is for Qitmeer
// If you want to enable ZMQ for Qitmeer, you must use 'zmq' tags when go building
type ZMQNotification struct {
	cfg              *config.Config
	publishNotifiers []IZMQPublishNotifier
}

// Initialization notification
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

// return enable for ZMQ
func (zn *ZMQNotification) IsEnable() bool {
	return true
}

// block accepted
func (zn *ZMQNotification) BlockAccepted(block *types.SerializedBlock) {
	log.Debug(fmt.Sprintf("BlockAccepted:%s", block.Hash().String()))

	for i := 0; i < len(zn.publishNotifiers); {
		err := zn.publishNotifiers[i].NotifyBlock(block)
		if err != nil {
			zn.publishNotifiers[i].Shutdown()
			zn.publishNotifiers = append(zn.publishNotifiers[:i], zn.publishNotifiers[i+1:]...)
		} else {
			i++
		}
	}
}

// block connected
func (zn *ZMQNotification) BlockConnected(block *types.SerializedBlock) {
	log.Debug(fmt.Sprintf("BlockConnected:%s", block.Hash().String()))
	for i := 0; i < len(zn.publishNotifiers); {
		err := zn.publishNotifiers[i].NotifyTransaction(block.Transactions())
		if err != nil {
			zn.publishNotifiers[i].Shutdown()
			zn.publishNotifiers = append(zn.publishNotifiers[:i], zn.publishNotifiers[i+1:]...)
		} else {
			i++
		}
	}
}

// block connected
func (zn *ZMQNotification) BlockDisconnected(block *types.SerializedBlock) {
	log.Debug(fmt.Sprintf("BlockDisconnected:%s", block.Hash().String()))
	for i := 0; i < len(zn.publishNotifiers); {
		err := zn.publishNotifiers[i].NotifyTransaction(block.Transactions())
		if err != nil {
			zn.publishNotifiers[i].Shutdown()
			zn.publishNotifiers = append(zn.publishNotifiers[:i], zn.publishNotifiers[i+1:]...)
		} else {
			i++
		}
	}
}

// Shutdown
func (zn *ZMQNotification) Shutdown() {
	log.Info("ZMQ: Shutdown...")
	for _, notifier := range zn.publishNotifiers {
		notifier.Shutdown()
	}
	zn.publishNotifiers = []IZMQPublishNotifier{}
}
