/*
 * Copyright (c) 2017-2020 The qitmeer developers
 */

package vm

import (
	"context"
	"github.com/Qitmeer/qitmeer/engine/vm/proto"
	"github.com/hashicorp/go-plugin"
	"google.golang.org/grpc"
	"google.golang.org/protobuf/types/known/emptypb"
)

type VMClient struct {
	*ChainState
	client proto.VMClient
	broker *plugin.GRPCBroker
	proc   *plugin.Client

	conns []*grpc.ClientConn

	ctx context.Context
}

func (vm *VMClient) SetProcess(proc *plugin.Client) {
	vm.proc = proc
}

func (vm *VMClient) Shutdown() error {
	var ret error
	_, err := vm.client.Shutdown(context.Background(), &emptypb.Empty{})
	if err != nil {
		log.Error(err.Error())
		ret = err
	}
	for _, conn := range vm.conns {
		err := conn.Close()
		if err != nil {
			log.Error(err.Error())
			ret = err
		}
	}

	vm.proc.Kill()
	return ret
}

func NewVMClient(client proto.VMClient, broker *plugin.GRPCBroker) *VMClient {
	return &VMClient{
		client: client,
		broker: broker,
	}
}
