// +build zmq

package zmq

import (
	"github.com/Qitmeer/qitmeer/config"
	"github.com/Qitmeer/qitmeer/core/types"
)

type ZMQNotification struct {
}

func (zn *ZMQNotification) Init(cfg *config.Config) {
	log.Info("ZMQ:Not Supported")
}

func (zn *ZMQNotification) IsEnable() bool {
	return false
}

// block accepted
func (zn *ZMQNotification) BlockAccepted(block *types.SerializedBlock) {

}

// block connected
func (zn *ZMQNotification) BlockConnected(block *types.SerializedBlock) {

}

// block connected
func (zn *ZMQNotification) BlockDisconnected(block *types.SerializedBlock) {

}

// Shutdown
func (zn *ZMQNotification) Shutdown() {

}
