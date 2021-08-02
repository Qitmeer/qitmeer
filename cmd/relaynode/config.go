package main

import (
	"fmt"
	"github.com/Qitmeer/qitmeer/common/util"
	"github.com/Qitmeer/qitmeer/params"
	"github.com/urfave/cli/v2"
	"os"
	"path/filepath"
)

const (
	defaultDataDirname = "relay"
	defaultPort        = "2001"
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

	EnableNoise = &cli.BoolFlag{
		Name:        "noise",
		Aliases:     []string{"n"},
		Usage:       "noise",
		Value:       false,
		Destination: &conf.EnableNoise,
	}

	Network = &cli.StringFlag{
		Name:        "network",
		Aliases:     []string{"e"},
		Usage:       "Network {mainnet,mixnet,privnet,testnet}",
		Value:       params.MixNetParam.Name,
		Destination: &conf.Network,
	}

	HostDNS = &cli.StringFlag{
		Name:        "externaldns",
		Aliases:     []string{"s"},
		Usage:       "The DNS address advertised by libp2p. This may be used to advertise an external DNS.",
		Value:       "",
		Destination: &conf.HostDNS,
	}

	UsePeerStore = &cli.BoolFlag{
		Name:        "pstore",
		Aliases:     []string{"r"},
		Usage:       "P2P Peer store",
		Value:       true,
		Destination: &conf.UsePeerStore,
	}

	AppFlags = []cli.Flag{
		HomeDir,
		DataDir,
		PrivateKey,
		ExternalIP,
		Port,
		EnableNoise,
		Network,
		HostDNS,
		UsePeerStore,
	}
)

type Config struct {
	HomeDir      string
	DataDir      string
	PrivateKey   string
	ExternalIP   string
	Port         string
	EnableNoise  bool
	Network      string
	HostDNS      string
	UsePeerStore bool
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

	// assign active network params while we're at it
	numNets := 0
	if c.Network == params.TestNetParam.Name {
		numNets++
		params.ActiveNetParams = &params.TestNetParam
	}
	if c.Network == params.PrivNetParam.Name {
		numNets++
		// Also disable dns seeding on the private test network.
		params.ActiveNetParams = &params.PrivNetParam
	}
	if c.Network == params.MixNetParam.Name {
		numNets++
		params.ActiveNetParams = &params.MixNetParam
	}

	if numNets == 0 {
		numNets++
		params.ActiveNetParams = &params.MainNetParam
	}

	// Multiple networks can't be selected simultaneously.
	if numNets > 1 {
		str := "%s: the testnet and simnet params can't be " +
			"used together -- choose one of the three"
		return fmt.Errorf("%s", str)
	}

	if err := params.ActiveNetParams.PowConfig.Check(); err != nil {
		return err
	}
	return nil
}
