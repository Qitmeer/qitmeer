package zmq

import (
	"github.com/Qitmeer/qitmeer/config"
	"github.com/Qitmeer/qitmeer/core/types"
	_ "log"
)

type IZMQNotification interface {
	// init
	Init(cfg *config.Config)

	// is enable
	IsEnable() bool

	// block accepted
	BlockAccepted(block *types.SerializedBlock)

	// block connected
	BlockConnected(block *types.SerializedBlock)

	// block connected
	BlockDisconnected(block *types.SerializedBlock)
}

// New ZMQ Notification
func NewZMQNotification(cfg *config.Config) IZMQNotification {
	zmq := &ZMQNotification{}
	zmq.Init(cfg)
	return zmq
}
