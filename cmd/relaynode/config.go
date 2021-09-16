package main

import (
	"fmt"
	"github.com/Qitmeer/qitmeer/common/util"
	l "github.com/Qitmeer/qitmeer/log"
	"github.com/Qitmeer/qitmeer/params"
	"github.com/Qitmeer/qitmeer/services/common"
	"github.com/urfave/cli/v2"
	"os"
	"path/filepath"
)

const (
	defaultDataDirname   = "relay"
	defaultPort          = "2001"
	defaultIP            = "0.0.0.0"
	defaultLogDirname    = "logs"
	defaultLogFilename   = "relaynode.log"
	defaultRPCKeyFile    = "rpc.key"
	defaultRPCCertFile   = "rpc.cert"
	defaultMaxRPCClients = 10
	defaultRPCListener   = "127.0.0.1:2002"
	defaultMaxPeers      = 1000
)

var (
	defaultHomeDir     = util.AppDataDir(".", false)
	defaultDataDir     = filepath.Join(defaultHomeDir, defaultDataDirname)
	defaultRPCCertPath = filepath.Join(defaultDataDir, defaultRPCCertFile)
	defaultRPCKeyPath  = filepath.Join(defaultDataDir, defaultRPCKeyFile)

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

	NoFileLogging = &cli.BoolFlag{
		Name:        "nofilelogging",
		Aliases:     []string{"l"},
		Usage:       "Disable file logging.",
		Value:       false,
		Destination: &conf.NoFileLogging,
	}

	DebugLevel = &cli.StringFlag{
		Name:        "debuglevel",
		Aliases:     []string{"dl"},
		Usage:       "Logging level {trace, debug, info, warn, error, critical} ",
		Value:       "info",
		Destination: &conf.DebugLevel,
	}

	DisableRPC = &cli.BoolFlag{
		Name:        "norpc",
		Aliases:     []string{"nr"},
		Usage:       "Disable built-in RPC server -- NOTE: The RPC server is disabled by default if no rpcuser/rpcpass",
		Value:       true,
		Destination: &conf.DisableRPC,
	}

	RPCListeners = &cli.StringSliceFlag{
		Name:        "rpclisten",
		Aliases:     []string{"rl"},
		Usage:       "Add an interface/port to listen for RPC connections",
		Destination: &conf.RPCListeners,
	}

	RPCUser = &cli.StringFlag{
		Name:        "rpcuser",
		Aliases:     []string{"ru"},
		Usage:       "Username for RPC connections",
		Value:       "test",
		Destination: &conf.RPCUser,
	}

	RPCPass = &cli.StringFlag{
		Name:        "rpcpass",
		Aliases:     []string{"rp"},
		Usage:       "Password for RPC connections",
		Value:       "test",
		Destination: &conf.RPCPass,
	}

	RPCCert = &cli.StringFlag{
		Name:        "rpccert",
		Aliases:     []string{"rc"},
		Usage:       "File containing the certificate file",
		Value:       defaultRPCCertPath,
		Destination: &conf.RPCCert,
	}

	RPCKey = &cli.StringFlag{
		Name:        "rpckey",
		Aliases:     []string{"rk"},
		Usage:       "File containing the certificate key",
		Value:       defaultRPCKeyPath,
		Destination: &conf.RPCKey,
	}

	RPCMaxClients = &cli.IntFlag{
		Name:        "rpcmaxclients",
		Aliases:     []string{"rmc"},
		Usage:       "Max number of RPC clients for standard connections",
		Value:       defaultMaxRPCClients,
		Destination: &conf.RPCMaxClients,
	}

	DisableTLS = &cli.BoolFlag{
		Name:        "notls",
		Aliases:     []string{"nt"},
		Usage:       "Disable TLS for the RPC server -- NOTE: This is only allowed if the RPC server is bound to localhost",
		Value:       false,
		Destination: &conf.DisableTLS,
	}

	EnableRelay = &cli.BoolFlag{
		Name:        "relay",
		Aliases:     []string{"re"},
		Usage:       "Enable relay service for node",
		Value:       false,
		Destination: &conf.EnableRelay,
	}

	MaxPeers = &cli.IntFlag{
		Name:        "maxpeers",
		Aliases:     []string{"mp"},
		Usage:       "Max number of inbound and outbound peers",
		Value:       defaultMaxPeers,
		Destination: &conf.MaxPeers,
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
		NoFileLogging,
		DebugLevel,
		DisableRPC,
		RPCListeners,
		RPCUser,
		RPCPass,
		RPCCert,
		RPCKey,
		RPCMaxClients,
		DisableTLS,
		EnableRelay,
		MaxPeers,
	}
)

type Config struct {
	HomeDir       string
	DataDir       string
	PrivateKey    string
	ExternalIP    string
	Port          string
	EnableNoise   bool
	Network       string
	HostDNS       string
	UsePeerStore  bool
	NoFileLogging bool
	DebugLevel    string

	DisableRPC    bool
	RPCListeners  cli.StringSlice
	RPCUser       string
	RPCPass       string
	RPCCert       string
	RPCKey        string
	RPCMaxClients int
	DisableTLS    bool
	EnableRelay   bool
	MaxPeers      int
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

	// Set logging file if presented
	if !c.NoFileLogging {
		logDir := filepath.Join(c.DataDir, defaultLogDirname, params.ActiveNetParams.Name)

		l.InitLogRotator(filepath.Join(logDir, defaultLogFilename))
	}
	err = common.ParseAndSetDebugLevels(c.DebugLevel)
	if err != nil {
		return err
	}

	if c.RPCCert == defaultRPCCertPath {
		c.RPCCert = filepath.Join(c.DataDir, defaultRPCCertFile)
	}

	if c.RPCKey == defaultRPCKeyPath {
		c.RPCKey = filepath.Join(c.DataDir, defaultRPCKeyFile)
	}

	if len(c.RPCListeners.Value()) <= 0 {
		c.RPCListeners = *cli.NewStringSlice(defaultRPCListener)
	}
	return nil
}
