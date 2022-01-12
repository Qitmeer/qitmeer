package main

import (
	"github.com/Qitmeer/qitmeer/common/roughtime"
	_ "github.com/Qitmeer/qitmeer/database/ffldb"
	_ "github.com/Qitmeer/qitmeer/services/common"
	"github.com/urfave/cli/v2"
	"os"
	"runtime"
	"runtime/debug"
	ver "github.com/Qitmeer/qitmeer/version"
)

func main() {
	runtime.GOMAXPROCS(runtime.NumCPU())
	debug.SetGCPercent(20)
	if err := fastIBD(); err != nil {
		log.Error(err.Error())
		os.Exit(1)
	}
}

func fastIBD() error {
	cfg := &Config{}
	node := &Node{}
	aidnode := &AidNode{}

	app := &cli.App{
		Name:     "FastIBD",
		Version:  ver.String(),
		Compiled: roughtime.Now(),
		Authors: []*cli.Author{
			&cli.Author{
				Name: "Qitmeer",
			},
		},
		Copyright: "(c) 2020 Qitmeer",
		Usage:     "Fast Initial Block Download",
		Commands: []*cli.Command{
			&cli.Command{
				Name:        "export",
				Aliases:     []string{"e"},
				Category:    "IBD",
				Usage:       "Export all blocks from database",
				Description: "Export all blocks from database",
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:        "path",
						Aliases:     []string{"p"},
						Usage:       "Path to output data",
						Value:       defaultHomeDir,
						Destination: &cfg.OutputPath,
					},
					&cli.StringFlag{
						Name:        "endpoint",
						Aliases:     []string{"e"},
						Usage:       "End point for output data",
						Destination: &cfg.EndPoint,
					},
					&cli.BoolFlag{
						Name:        "byid",
						Aliases:     []string{"i"},
						Usage:       "Export by block id",
						Destination: &cfg.ByID,
					},
				},
				Before: func(c *cli.Context) error {
					return node.init(cfg)
				},
				After: func(c *cli.Context) error {
					return node.exit()
				},
				Action: func(c *cli.Context) error {
					return node.Export()
				},
			},
			&cli.Command{
				Name:        "import",
				Aliases:     []string{"i"},
				Category:    "IBD",
				Usage:       "Import all blocks from database",
				Description: "Import all blocks from database",
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:        "path",
						Aliases:     []string{"p"},
						Usage:       "Path to input data",
						Value:       defaultHomeDir,
						Destination: &cfg.InputPath,
					},
				},
				Before: func(c *cli.Context) error {
					return node.init(cfg)
				},
				After: func(c *cli.Context) error {
					return node.exit()
				},
				Action: func(c *cli.Context) error {
					return node.Import()
				},
			},
			&cli.Command{
				Name:        "upgrade",
				Aliases:     []string{"u"},
				Category:    "IBD",
				Usage:       "Upgrade all blocks from database for Qitmeer",
				Description: "Upgrade all blocks from database for Qitmeer",
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:        "path",
						Aliases:     []string{"p"},
						Usage:       "Path to input data",
						Value:       defaultHomeDir,
						Destination: &cfg.InputPath,
					},
					&cli.BoolFlag{
						Name:        "aidmode",
						Aliases:     []string{"ai"},
						Usage:       "Export by block id",
						Value:       false,
						Destination: &cfg.AidMode,
					},
				},
				Before: func(c *cli.Context) error {
					if cfg.AidMode {
						return aidnode.init(cfg)
					}
					return node.init(cfg)
				},
				After: func(c *cli.Context) error {
					if cfg.AidMode {
						return aidnode.exit()
					}
					return node.exit()
				},
				Action: func(c *cli.Context) error {
					if cfg.AidMode {
						return aidnode.Upgrade()
					}
					return node.Upgrade()
				},
			},
		},
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:        "appdata",
				Aliases:     []string{"A"},
				Usage:       "Path to application home directory",
				Value:       defaultHomeDir,
				Destination: &cfg.HomeDir,
			},
			&cli.StringFlag{
				Name:        "datadir",
				Aliases:     []string{"b"},
				Usage:       "Directory to store data",
				Value:       defaultDataDir,
				Destination: &cfg.DataDir,
			},
			&cli.BoolFlag{
				Name:        "testnet",
				Usage:       "Use the test network",
				Value:       false,
				Destination: &cfg.TestNet,
			},
			&cli.BoolFlag{
				Name:        "mixnet",
				Usage:       "Use the test mix pow network",
				Value:       false,
				Destination: &cfg.MixNet,
			},
			&cli.BoolFlag{
				Name:        "privnet",
				Usage:       "Use the private network",
				Value:       false,
				Destination: &cfg.PrivNet,
			},
			&cli.StringFlag{
				Name:        "dbtype",
				Usage:       "Database backend to use for the Block Chain",
				Value:       defaultDbType,
				Destination: &cfg.DbType,
			},
			&cli.StringFlag{
				Name:        "dagtype",
				Aliases:     []string{"G"},
				Usage:       "DAG type {phantom,spectre}",
				Value:       defaultDAGType,
				Destination: &cfg.DAGType,
			},
			&cli.BoolFlag{
				Name:        "disablebar",
				Usage:       "Hide progress bar",
				Value:       false,
				Destination: &cfg.DisableBar,
			},
			&cli.IntFlag{
				Name:        "cpunum",
				Usage:       "Use the num of cpu",
				Value:       runtime.NumCPU(),
				Destination: &cfg.CPUNum,
			},
		},
		EnableBashCompletion: true,
		Action: func(c *cli.Context) error {
			return cli.ShowAppHelp(c)
		},
	}

	return app.Run(os.Args)
}
