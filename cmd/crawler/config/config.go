package config

import (
	"fmt"
	"github.com/Qitmeer/qng-core/common/util"
	"github.com/Qitmeer/qng-core/params"
	"github.com/urfave/cli/v2"
	"net"
	"os"
	"path/filepath"
)

const (
	DefaultDataDirname = "crawler-data"
	DefaultPort        = "1024"
	DefaultRpcPort     = "2024"
	DefaultIP          = "0.0.0.0"
	DefaultDB          = DefaultDataDirname + "/db"
)

var (
	defaultHomeDir = util.AppDataDir(".", false)
	defaultDataDir = filepath.Join(defaultHomeDir, DefaultDataDirname)

	Conf = &Config{}

	HomeDir = &cli.StringFlag{
		Name:        "appdata",
		Aliases:     []string{"A"},
		Usage:       "Path to application home directory",
		Value:       defaultHomeDir,
		Destination: &Conf.HomeDir,
	}
	DataDir = &cli.StringFlag{
		Name:        "datadir",
		Aliases:     []string{"b"},
		Usage:       "Directory to store data",
		Value:       defaultDataDir,
		Destination: &Conf.DataDir,
	}

	PrivateKey = &cli.StringFlag{
		Name:        "privatekey",
		Aliases:     []string{"p"},
		Usage:       "private key",
		Destination: &Conf.PrivateKey,
	}

	ExternalIP = &cli.StringFlag{
		Name:        "externalip",
		Aliases:     []string{"i"},
		Usage:       "listen external ip",
		Destination: &Conf.ExternalIP,
	}

	Port = &cli.StringFlag{
		Name:        "port",
		Aliases:     []string{"o"},
		Usage:       "listen port",
		Value:       DefaultPort,
		Destination: &Conf.Port,
	}

	EnableNoise = &cli.BoolFlag{
		Name:        "noise",
		Aliases:     []string{"n"},
		Usage:       "noise",
		Value:       false,
		Destination: &Conf.EnableNoise,
	}

	Network = &cli.StringFlag{
		Name:        "network",
		Aliases:     []string{"e"},
		Usage:       "Network {mainnet,mixnet,privnet,testnet}",
		Value:       params.MixNetParam.Name,
		Destination: &Conf.Network,
	}

	RpcUser = &cli.StringFlag{
		Name:        "rpcuser",
		Aliases:     []string{"u"},
		Usage:       "rpc user",
		Value:       "admin",
		Destination: &Conf.RPCUser,
	}

	RpcPass = &cli.StringFlag{
		Name:        "rpcpass",
		Aliases:     []string{"s"},
		Usage:       "rpc pass",
		Value:       "123",
		Destination: &Conf.RPCPass,
	}

	RPCMaxClients = &cli.IntFlag{
		Name:        "rpcmaxclients",
		Aliases:     []string{"m"},
		Usage:       "rpc max clients",
		Value:       100,
		Destination: &Conf.RPCMaxClients,
	}

	AppFlags = []cli.Flag{
		HomeDir,
		DataDir,
		PrivateKey,
		ExternalIP,
		Port,
		EnableNoise,
		Network,
		RpcUser,
		RpcPass,
		RPCMaxClients,
	}
)

type Config struct {
	HomeDir           string
	DataDir           string
	PrivateKey        string
	ExternalIP        string
	Port              string
	EnableNoise       bool
	Network           string
	StaticPeers       []string
	BootstrapNodeAddr []string
	RPCUser           string
	RPCPass           string
	RPCListeners      []string
	RPCMaxClients     int
}

func (c *Config) Load() error {
	var err error
	if c.HomeDir != defaultHomeDir {
		c.HomeDir, err = filepath.Abs(c.HomeDir)
		if err != nil {
			return err
		}
		if c.DataDir == defaultDataDir {
			c.DataDir = filepath.Join(c.HomeDir, DefaultDataDirname)
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

	if len(c.BootstrapNodeAddr) == 0 {
		c.BootstrapNodeAddr = params.ActiveNetParams.Bootstrap
	}

	if len(c.RPCListeners) == 0 {
		c.RPCListeners = []string{
			net.JoinHostPort("", DefaultRpcPort),
		}
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
