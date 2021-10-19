package vm

import (
	"context"
	"github.com/Qitmeer/qitmeer/engine/vm/proto"
	"github.com/hashicorp/go-plugin"
)

type VMServer struct {
	proto.UnimplementedVMServer
	vm     ChainVM
	broker *plugin.GRPCBroker

	ctx    *context.Context
	closed chan struct{}
}

func NewServer(vm ChainVM, broker *plugin.GRPCBroker) *VMServer {
	return &VMServer{
		vm:     vm,
		broker: broker,
	}
}
