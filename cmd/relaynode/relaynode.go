package main

import (
	"github.com/Qitmeer/qng-core/common/roughtime"
	_ "github.com/Qitmeer/qng-core/database/ffldb"
	_ "github.com/Qitmeer/qitmeer/services/common"
	"github.com/Qitmeer/qitmeer/version"
	"github.com/urfave/cli/v2"
	"os"
	"runtime"
	"runtime/debug"
)

func main() {
	runtime.GOMAXPROCS(runtime.NumCPU())
	debug.SetGCPercent(20)
	if err := relayNode(); err != nil {
		log.Error(err.Error())
		os.Exit(1)
	}
}

func relayNode() error {
	node := &Node{}
	app := &cli.App{
		Name:     "RelayNode",
		Version:  version.String(),
		Compiled: roughtime.Now(),
		Authors: []*cli.Author{
			&cli.Author{
				Name: "Qitmeer",
			},
		},
		Copyright:            "(c) 2020 Qitmeer",
		Usage:                "Relay Node",
		Flags:                AppFlags,
		EnableBashCompletion: true,
		Before: func(c *cli.Context) error {
			return node.init(conf)
		},
		After: func(c *cli.Context) error {
			return node.Stop()
		},
		Action: func(c *cli.Context) error {
			return node.Start()
		},
	}

	return app.Run(os.Args)
}
