package main

import (
	"github.com/Qitmeer/qitmeer/common/util"
	"github.com/urfave/cli/v2"
	"os"
	"path/filepath"
)

const (
	defaultDataDirname = "relay"
	defaultPort        = "1001"
	defaultIP          = "0.0.0.0"
)

var (
	defaultHomeDir = util.AppDataDir(".", false)
	defaultDataDir = filepath.Join(defaultHomeDir, defaultDataDirname)

	conf = &Config{}

	HomeDir = &cli.StringFlag{
		Name:        "appdata",
		Aliases:     []string{"A"},
		Usage:       "Path to application home directory",
		Value:       defaultHomeDir,
		Destination: &conf.HomeDir,
	}
	DataDir = &cli.StringFlag{
		Name:        "datadir",
		Aliases:     []string{"b"},
		Usage:       "Directory to store data",
		Value:       defaultDataDir,
		Destination: &conf.DataDir,
	}

	PrivateKey = &cli.StringFlag{
		Name:        "privatekey",
		Aliases:     []string{"p"},
		Usage:       "private key",
		Destination: &conf.PrivateKey,
	}

	ExternalIP = &cli.StringFlag{
		Name:        "externalip",
		Aliases:     []string{"i"},
		Usage:       "listen external ip",
		Destination: &conf.ExternalIP,
	}

	Port = &cli.StringFlag{
		Name:        "port",
		Aliases:     []string{"o"},
		Usage:       "listen port",
		Value:       defaultPort,
		Destination: &conf.Port,
	}

	AppFlags = []cli.Flag{
		HomeDir,
		DataDir,
		PrivateKey,
		ExternalIP,
		Port,
	}
)

type Config struct {
	HomeDir    string
	DataDir    string
	PrivateKey string
	ExternalIP string
	Port       string
}

func (c *Config) load() error {
	var err error
	if c.HomeDir != defaultHomeDir {
		c.HomeDir, err = filepath.Abs(c.HomeDir)
		if err != nil {
			return err
		}
		if c.DataDir == defaultDataDir {
			c.DataDir = filepath.Join(c.HomeDir, defaultDataDirname)
		}
	}
	_, err = os.Stat(c.DataDir)
	if err != nil && !os.IsExist(err) {
		err = os.MkdirAll(c.DataDir, 0700)
		if err != nil {
			return err
		}
	}
	return nil
}
