// +build !zmq

package zmq

import "github.com/Qitmeer/qitmeer/config"

type ZMQPublishNotifier struct {
}

func (zp *ZMQPublishNotifier) Init(cfg *config.Config) {
	log.Info("Initialize ZMQ public notifier.")
}

func (zp *ZMQPublishNotifier) SendMessage() {

}

func (zp *ZMQPublishNotifier) Shutdown() {
	log.Info("Shutdown ZMQ")
}

// New ZMQ Publish Notifier
func NewZMQPublishNotifier(cfg *config.Config) *ZMQPublishNotifier {
	zmq := &ZMQPublishNotifier{}
	zmq.Init(cfg)
	return zmq
}
