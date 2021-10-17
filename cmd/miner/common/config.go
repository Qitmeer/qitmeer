// Copyright (c) 2019 The qitmeer developers
// Use of this source code is governed by an ISC
// license that can be found in the LICENSE file.

package common

import (
	"fmt"
	"github.com/Qitmeer/qitmeer/cmd/miner/common/go-flags"
	"github.com/Qitmeer/qitmeer/core/address"
	l "github.com/Qitmeer/qitmeer/log"
	"github.com/Qitmeer/qitmeer/params"
	"log"
	"net"
	"os"
	"path/filepath"
	"runtime"
	"strings"
)

const (
	defaultConfigFilename = "qitmeer.conf"
)

var CurrentHeight = uint64(0)
var JobID = ""

var (
	minerHomeDir       = GetCurrentDir()
	defaultRPCServer   = "127.0.0.1"
	defaultPow         = "cuckaroo"
	defaultSymbol      = "PMEER"
	defaultTimeout     = 60
	defaultMaxTxCount  = 100000000
	defaultMaxSigCount = 400000000
	defaultStatsServer = ""
)

type CommandConfig struct {
	ListDevices bool `short:"l" long:"listdevices" description:"List number of devices."`
	Version     bool `short:"v" long:"version" description:"show the version of miner"`
}

type FileConfig struct {
	ConfigFile string `short:"C" long:"configfile" description:"Path to configuration file"`
	// Debugging options
	MinerLogFile string `long:"minerlog" description:"Write miner log file"`
}

type OptionalConfig struct {
	// Config / log options
	CPUMiner      bool   `long:"cpuminer" description:"CPUMiner" default-mask:"false"`
	CpuWorkers    int    `long:"cpuworkers" description:"CPUWorkers" default-mask:"1"`
	Proxy         string `long:"proxy" description:"Connect via SOCKS5 proxy (eg. 127.0.0.1:9050)"`
	ProxyUser     string `long:"proxyuser" description:"Username for proxy server"`
	ProxyPass     string `long:"proxypass" default-mask:"-" description:"Password for proxy server"`
	Timeout       int    `long:"timeout" description:"rpc timeout." default-mask:"60"`
	UseDevices    string `long:"use_devices" description:"all gpu devices,you can use ./qitmeer-miner -l to see. examples:0,1 use the #0 device and #1 device"`
	MaxTxCount    int    `long:"max_tx_count" description:"max pack tx count" default-mask:"10000000"`
	MaxSigCount   int    `long:"max_sig_count" description:"max sign tx count" default-mask:"400000000"`
	LogLevel      string `long:"log_level" description:"info|debug|error|warn|trace" default-mask:"info"`
	StatsServer   string `long:"stats_server" description:"stats web server" default-mask:"127.0.0.1:1235"`
	Restart       int    ` description:"restart server" default-mask:"0"`
	Accept        int    ` description:"Accept count" default-mask:"0"`
	Reject        int    ` description:"Reject count" default-mask:"0"`
	Stale         int    ` description:"Stale count" default-mask:"0"`
	Target        string ` description:"Target"`
	NumOfChips    int    `long:"num_of_chips" description:"num of chips" default-mask:"1"`
	TaskInterval  int    `long:"task_interval" description:"get blocktemplate interval" default-mask:"5000"`
	TaskForceStop bool   `long:"task_force_stop" description:"force stop the current task when miner fail to get blocktemplate from the qitmeer full node." default-mask:"true"`
	ForceSolo     bool   `long:"force_solo" description:"force solo mining" default-mask:"false"`
	Freqs         string `long:"freqs" description:"freq settings" default-mask:"1000,200|"`
	BaseDiff      uint   `long:"base_diff" description:"base_diff settings,default 4G" default-mask:"224"`
	UartPath      string `long:"uart_path" description:"uarts path split with ," default-mask:"/dev/ttyS1"`
}

type PoolConfig struct {
	// Pool related options
	Pool         string `short:"o" long:"pool" description:"Pool to connect to (e.g.stratum+tcp://pool:port)"`
	PoolUser     string `short:"m" long:"pooluser" description:"Pool username"`
	PoolPassword string `short:"n" long:"poolpass" default-mask:"-" description:"Pool password"`
}

type SoloConfig struct {
	// RPC connection options
	MinerAddr        string `short:"M" long:"mineraddress" description:"Miner Address" default-mask:""`
	RPCServer        string `short:"s" long:"rpcserver" description:"RPC server to connect to"`
	RPCUser          string `short:"u" long:"rpcuser" description:"RPC username"`
	RPCPassword      string `short:"p" long:"rpcpass" default-mask:"-" description:"RPC password"`
	RandStr          string `long:"randstr" description:"Rand String,Your Unique Marking." default-mask:""`
	NoTLS            bool   `long:"notls" description:"Do not verify tls certificates" default-mask:"true"`
	RPCCert          string `long:"rpccert" description:"RPC server certificate chain for validation"`
	ConfirmHeight    uint64 `long:"confirmheight" description:"min confirm height" default-mask:"3"`
	NotConfirmHeight uint64 `long:"notconfirmheight" description:"min confirm height" default-mask:"50000000"`
}

type NecessaryConfig struct {
	// Config / log options
	Pow     string `short:"P" long:"pow" description:"blake2bd|cuckaroo|cuckatoo"`
	Symbol  string `short:"S" long:"symbol" description:"Symbol" default-mask:"PMEER"`
	NetWork string `short:"N" long:"network" description:"network privnet|testnet|mainnet" default-mask:"testnet"`
	Param   *params.Params
}

type GlobalConfig struct {
	OptionConfig    OptionalConfig
	LogConfig       FileConfig
	DeviceConfig    CommandConfig
	SoloConfig      SoloConfig
	PoolConfig      PoolConfig
	NecessaryConfig NecessaryConfig
}

// removeDuplicateAddresses returns a new slice with all duplicate entries in
// addrs removed.
func removeDuplicateAddresses(addrs []string) []string {
	result := make([]string, 0, len(addrs))
	seen := map[string]struct{}{}
	for _, val := range addrs {
		if _, ok := seen[val]; !ok {
			result = append(result, val)
			seen[val] = struct{}{}
		}
	}
	return result
}

// normalizeAddress returns addr with the passed default port appended if
// there is not already a port specified.
func normalizeAddress(addr string, defaultPort string) string {
	_, _, err := net.SplitHostPort(addr)
	if err != nil {
		return net.JoinHostPort(addr, defaultPort)
	}
	return addr
}

// normalizeAddresses returns a new slice with all the passed peer addresses
// normalized with the given default port, and all duplicates removed.
func normalizeAddresses(addrs []string, defaultPort string) []string {
	for i, addr := range addrs {
		addrs[i] = normalizeAddress(addr, defaultPort)
	}

	return removeDuplicateAddresses(addrs)
}

// cleanAndExpandPath expands environement variables and leading ~ in the
// passed path, cleans the result, and returns it.
func cleanAndExpandPath(path string) string {
	// Expand initial ~ to OS specific home directory.
	if strings.HasPrefix(path, "~") {
		homeDir := filepath.Dir(minerHomeDir)
		path = strings.Replace(path, "~", homeDir, 1)
	}

	// NOTE: The os.ExpandEnv doesn't work with Windows-style %VARIABLE%,
	// but they variables can still be expanded via POSIX-style $VARIABLE.
	return filepath.Clean(os.ExpandEnv(path))
}

// loadConfig initializes and parses the config using a config file and command
// line options.
//
// The configuration proceeds as follows:
// 	1) Start with a default config with sane settings
// 	2) Pre-parse the command line to check for an alternative config file
// 	3) Load configuration file overwriting defaults with any specified options
// 	4) Parse CLI options and overwrite/add any specified options
//
// The above results in btcd functioning properly without any config settings
// while still allowing the user to override settings with config files and
// command line options.  Command line options always take precedence.
func LoadConfig() (*GlobalConfig, []string, error) {
	// Default config.
	soloCfg := SoloConfig{
		RPCServer: defaultRPCServer,
		NoTLS:     true,
	}
	poolCfg := PoolConfig{}
	// Default config.
	deviceCfg := CommandConfig{}
	// Default config.
	fileCfg := FileConfig{
		//ConfigFile:defaultConfigFile,
	}
	necessaryCfg := NecessaryConfig{
		Pow:     defaultPow,
		Symbol:  defaultSymbol,
		NetWork: "testnet",
	}
	optionalCfg := OptionalConfig{
		CPUMiner:      false,
		Timeout:       defaultTimeout,
		UseDevices:    "",
		MaxTxCount:    defaultMaxTxCount,
		MaxSigCount:   defaultMaxSigCount,
		StatsServer:   defaultStatsServer,
		TaskInterval:  5000,
		TaskForceStop: true,
		ForceSolo:     false,
		NumOfChips:    14,
	}

	// Create the home directory if it doesn't already exist.
	err := os.MkdirAll(minerHomeDir, 0700)
	if err != nil {
		return nil, []string{}, err
	}
	appName := filepath.Base(os.Args[0])
	appName = strings.TrimSuffix(appName, filepath.Ext(appName))
	// Pre-parse the command line options to see if an alternative config
	// file or the version flag was specified.
	preParser := flags.NewNamedParser(appName, flags.HelpFlag)

	_, err = preParser.AddGroup("Debug Command", "The Miner Debug tools", &deviceCfg)
	if err != nil {
		return nil, []string{}, err
	}

	_, err = preParser.AddGroup("The Config File Options", "The Config File Options", &fileCfg)
	if err != nil {
		return nil, []string{}, err
	}
	_, err = preParser.AddGroup("The Necessary Config Options", "The Necessary Config Options", &necessaryCfg)
	if err != nil {
		return nil, []string{}, err
	}
	_, err = preParser.AddGroup("The Solo Config Option", "The Solo Config Option", &soloCfg)
	if err != nil {
		return nil, []string{}, err
	}
	_, err = preParser.AddGroup("The pool Config Option", "The pool Config Option", &poolCfg)
	if err != nil {
		return nil, []string{}, err
	}
	_, err = preParser.AddGroup("The Optional Config Option", "The Optional Config Option", &optionalCfg)
	if err != nil {
		return nil, []string{}, err
	}
	_, err = preParser.Parse()
	if err != nil {
		_, _ = fmt.Fprintln(os.Stderr, err)
		os.Exit(0)
	}
	if deviceCfg.ListDevices {
		MinerLoger.Info("[GPU Devices List]:")
		os.Exit(0)
	}

	if deviceCfg.Version {
		fmt.Printf(GetVersion())
		os.Exit(0)
	}

	if fileCfg.ConfigFile == "" {
		MinerLoger.Warn("Don't have config file.")
	} else {
		err = flags.NewIniParser(preParser).ParseFile(fileCfg.ConfigFile)
		if err != nil {
			if _, ok := err.(*os.PathError); !ok {
				_, _ = fmt.Fprintln(os.Stderr, err)
				return nil, nil, err
			}
		}
	}

	remainingArgs, err := preParser.Parse()
	if err != nil {
		if _, ok := err.(*flags.Error); !ok {
			_, _ = fmt.Fprintln(os.Stderr, err)
			return nil, nil, err
		}
		preParser.WriteHelp(os.Stderr)
		os.Exit(0)
	}
	if fileCfg.MinerLogFile != "" {
		l.InitLogRotator(fileCfg.MinerLogFile)
	}
	l.Glogger().Verbosity(ConvertLogLevel(optionalCfg.LogLevel))
	if poolCfg.Pool == "" && soloCfg.MinerAddr == "" {
		MinerLoger.Error("Solo need address -M , pool need -o pool address")
		preParser.WriteHelp(os.Stderr)
		os.Exit(0)
	}
	necessaryCfg.Param = InitNet(necessaryCfg.NetWork, necessaryCfg.Param)
	if necessaryCfg.Param == nil {
		os.Exit(0)
	}
	if poolCfg.Pool == "" && !CheckBase58Addr(soloCfg.MinerAddr, necessaryCfg.NetWork, necessaryCfg.Param) {
		os.Exit(0)
	}
	if optionalCfg.ForceSolo {
		//solo
		poolCfg.Pool = ""
	}
	if poolCfg.Pool != "" && !strings.Contains(poolCfg.Pool, "stratum+tcp://") {
		//solo
		soloCfg.RPCServer = poolCfg.Pool
		soloCfg.RPCUser = poolCfg.PoolUser
		soloCfg.RPCPassword = poolCfg.PoolPassword
		poolCfg.Pool = ""
	}

	// Show the version and exit if the version flag was specified.

	// Handle environment variable expansion in the RPC certificate path.
	soloCfg.RPCCert = cleanAndExpandPath(soloCfg.RPCCert)
	return &GlobalConfig{
		optionalCfg,
		fileCfg,
		deviceCfg,
		soloCfg,
		poolCfg,
		necessaryCfg,
	}, remainingArgs, nil
}

func CheckBase58Addr(addr, network string, p *params.Params) bool {
	_, err := address.DecodeAddress(addr)
	if err != nil {
		log.Fatalln(network, "qitmeer address error!", err, addr)
		return false
	}
	networkChar := addr[0:1]
	if p.NetworkAddressPrefix != networkChar {
		log.Fatalln(network, "qitmeer address not match the network,please check your config network param!", p.NetworkAddressPrefix, networkChar)
		return false
	}
	return true
}

func InitNet(network string, p *params.Params) *params.Params {
	switch network {
	case params.MainNetParams.Name:
		p = &params.MainNetParams
	case params.TestNetParams.Name:
		p = &params.TestNetParams
	case params.PrivNetParams.Name:
		p = &params.PrivNetParams
	case params.MixNetParams.Name:
		p = &params.MixNetParams
	default:
		log.Fatalln(network, "Please define the network parameter for qitmeer!")
		return nil
	}
	return p
}

func GetVersion() string {
	appName := filepath.Base(os.Args[0])
	appName = strings.TrimSuffix(appName, filepath.Ext(appName))
	return fmt.Sprintf("%s version %s (Go version %s)\n", appName, String(), runtime.Version())
}

func ConvertLogLevel(level string) l.Lvl {
	switch level {
	case "warn":
		return l.LvlWarn
	case "info":
		return l.LvlInfo
	case "debug":
		return l.LvlDebug
	case "error":
		return l.LvlError
	case "trace":
		return l.LvlTrace
	default:
		return l.LvlDebug
	}
}
