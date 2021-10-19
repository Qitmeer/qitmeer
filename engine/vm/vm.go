package vm

import (
	"context"
	"fmt"
	"github.com/Qitmeer/qitmeer/engine/vm/proto"
	"github.com/hashicorp/go-hclog"
	"github.com/hashicorp/go-plugin"
	"google.golang.org/grpc"
	"os"
	"os/exec"
)

func NewVM(ctx context.Context, path string, arg ...string) (interface{}, error) {
	logger := hclog.New(&hclog.LoggerOptions{
		Name:   "VM",
		Output: os.Stdout,
		Level:  hclog.Debug,
	})

	config := &plugin.ClientConfig{
		HandshakeConfig: Handshake,
		Plugins:         PluginMap,
		Cmd:             exec.Command(path, arg...),
		AllowedProtocols: []plugin.Protocol{
			plugin.ProtocolNetRPC,
			plugin.ProtocolGRPC,
		},
		Managed: true,
		Logger:  logger,
	}

	client := plugin.NewClient(config)

	rpcClient, err := client.Client()
	if err != nil {
		client.Kill()
		return nil, err
	}

	raw, err := rpcClient.Dispense("vm")
	if err != nil {
		client.Kill()
		return nil, err
	}

	vm, ok := raw.(*VMClient)
	if !ok {
		client.Kill()
		return nil, fmt.Errorf("wrong vm type")
	}

	vm.SetProcess(client)
	vm.ctx = ctx
	return vm, nil
}

var Handshake = plugin.HandshakeConfig{
	ProtocolVersion:  1,
	MagicCookieKey:   "VM_PLUGIN",
	MagicCookieValue: "dynamic",
}

var PluginMap = map[string]plugin.Plugin{
	"vm": &Plugin{},
}

type Plugin struct {
	plugin.NetRPCUnsupportedPlugin
	vm ChainVM
}

func New(vm ChainVM) *Plugin { return &Plugin{vm: vm} }

// GRPCServer registers a new GRPC server.
func (p *Plugin) GRPCServer(broker *plugin.GRPCBroker, s *grpc.Server) error {
	proto.RegisterVMServer(s, NewServer(p.vm, broker))
	return nil
}

// GRPCClient returns a new GRPC client
func (p *Plugin) GRPCClient(ctx context.Context, broker *plugin.GRPCBroker, c *grpc.ClientConn) (interface{}, error) {
	return NewVMClient(proto.NewVMClient(c), broker), nil
}
