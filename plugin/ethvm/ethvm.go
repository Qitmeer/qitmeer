/*
 * Copyright (c) 2017-2020 The qitmeer developers
 */

package ethvm

import (
	"github.com/Qitmeer/qitmeer/engine/vm"
	"github.com/Qitmeer/qitmeer/plugin/ethvm/evm"
	"github.com/Qitmeer/qitmeer/version"
	"github.com/hashicorp/go-plugin"
	"runtime"
)

func main() {
	log.Info("System info", "ETH VM Version", version.String(), "Go version", runtime.Version())

	plugin.Serve(&plugin.ServeConfig{
		HandshakeConfig: vm.Handshake,
		Plugins: map[string]plugin.Plugin{
			"vm": vm.New(&evm.VM{}),
		},

		GRPCServer: plugin.DefaultGRPCServer,
	})
}
