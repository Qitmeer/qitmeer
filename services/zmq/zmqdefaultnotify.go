// +build !zmq

package zmq

import (
	"github.com/Qitmeer/qng-core/config"
	"github.com/Qitmeer/qng-core/core/types"
)

// This ZeroMQ notification is default for Qitmeer
// If you want to enable ZMQ for Qitmeer, you must use 'zmq' tags when go building
type ZMQNotification struct {
}

// Initialization notification
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
