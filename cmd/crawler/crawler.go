package main

import (
	"github.com/Qitmeer/qitmeer/cmd/crawler/config"
	"github.com/Qitmeer/qitmeer/cmd/crawler/log"
	"github.com/Qitmeer/qitmeer/cmd/crawler/node"
	"github.com/Qitmeer/qitmeer/common/roughtime"
	_ "github.com/Qitmeer/qitmeer/database/ffldb"
	_ "github.com/Qitmeer/qitmeer/services/common"
	"github.com/urfave/cli/v2"
	"os"
)

func main() {
	if err := crawlerNode(); err != nil {
		log.Log.Error(err.Error())
		os.Exit(1)
	}
}

func crawlerNode() error {
	n := &node.Node{}
	app := &cli.App{
		Name:     "CrawlerNode",
		Version:  "V0.0.1",
		Compiled: roughtime.Now(),
		Authors: []*cli.Author{
			&cli.Author{
				Name: "Qitmeer",
			},
		},
		Copyright:            "(c) 2020 Qitmeer",
		Usage:                "Crawler Node",
		Flags:                config.AppFlags,
		EnableBashCompletion: true,
		Before: func(c *cli.Context) error {
			return n.Init(config.Conf)
		},
		After: func(c *cli.Context) error {
			return n.Exit()
		},
		Action: func(c *cli.Context) error {
			return n.Run()
		},
	}

	return app.Run(os.Args)
}
