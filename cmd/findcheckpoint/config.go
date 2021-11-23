package main

import (
	"fmt"
	"github.com/Qitmeer/qng-core/common/util"
	"github.com/Qitmeer/qng-core/params"
	"github.com/jessevdk/go-flags"
	"os"
	"path/filepath"
	"strings"
)

const (
	minCandidates        = 1
	maxCandidates        = 20
	defaultNumCandidates = 5
	defaultDataDirname   = "data"
)

var (
	defaultHomeDir = util.AppDataDir("qitmeerd", false)
	defaultDataDir = filepath.Join(defaultHomeDir, defaultDataDirname)
	defaultDbType  = "ffldb"
	defaultDAGType = "phantom"
)

type Config struct {
	HomeDir       string `short:"A" long:"appdata" description:"Path to application home directory"`
	DataDir       string `short:"b" long:"datadir" description:"Directory to store data"`
	TestNet       bool   `long:"testnet" description:"Use the test network"`
	MixNet        bool   `long:"mixnet" description:"Use the test mix pow network"`
	PrivNet       bool   `long:"privnet" description:"Use the private network"`
	DbType        string `long:"dbtype" description:"Database backend to use for the Block Chain"`
	DAGType       string `short:"G" long:"dagtype" description:"DAG type {phantom,conflux,spectre} "`
	NumCandidates int    `short:"n" long:"numcandidates" description:"Max num of checkpoint candidates to show {1-20}"`
	UseGoOutput   bool   `short:"g" long:"gooutput" description:"Display the candidates using Go syntax that is ready to insert into the qitmeer checkpoint list"`
	IsCheckPoint  string `short:"I" long:"ischeckpoint" description:"Determine if it's a check point"`
}

// loadConfig initializes and parses the config using a config file and command
// line options.
func LoadConfig() (*Config, []string, error) {

	// Default config.
	cfg := Config{
		HomeDir:       defaultHomeDir,
		DataDir:       defaultDataDir,
		DbType:        defaultDbType,
		DAGType:       defaultDAGType,
		TestNet:       true,
		NumCandidates: defaultNumCandidates,
	}

	// Pre-parse the command line options to see if an alternative config
	// file or the version flag was specified.  Any errors aside from the
	// help message error can be ignored here since they will be caught by
	// the final parse below.
	preCfg := cfg
	preParser := flags.NewParser(&preCfg, flags.HelpFlag)
	remainingArgs, err := preParser.Parse()
	if err != nil {
		if e, ok := err.(*flags.Error); ok && e.Type != flags.ErrHelp {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		} else if ok && e.Type == flags.ErrHelp {
			fmt.Fprintln(os.Stdout, err)
			os.Exit(0)
		}
	}
	// Show the version and exit if the version flag was specified.
	appName := filepath.Base(os.Args[0])
	appName = strings.TrimSuffix(appName, filepath.Ext(appName))
	usageMessage := fmt.Sprintf("Use %s -h to show usage", appName)

	// Update the home directory for qitmeerd if specified. Since the home
	// directory is updated, other variables need to be updated to
	// reflect the new changes.
	if preCfg.HomeDir != "" {
		cfg.HomeDir, _ = filepath.Abs(preCfg.HomeDir)

		if preCfg.DataDir == defaultDataDir {
			cfg.DataDir = filepath.Join(cfg.HomeDir, defaultDataDirname)
		} else {
			cfg.DataDir = preCfg.DataDir
		}
		cfg.IsCheckPoint = preCfg.IsCheckPoint
	}

	// Create the home directory if it doesn't already exist.
	funcName := "loadConfig"
	err = os.MkdirAll(cfg.HomeDir, 0700)
	if err != nil {
		// Show a nicer error message if it's because a symlink is
		// linked to a directory that does not exist (probably because
		// it's not mounted).
		if e, ok := err.(*os.PathError); ok && os.IsExist(err) {
			if link, lerr := os.Readlink(e.Path); lerr == nil {
				str := "is symlink %s -> %s mounted?"
				err = fmt.Errorf(str, e.Path, link)
			}
		}
		str := "%s: failed to create home directory: %v"
		err := fmt.Errorf(str, funcName, err)
		fmt.Fprintln(os.Stderr, err)
		return nil, nil, err
	}

	// assign active network params while we're at it
	numNets := 0
	if cfg.TestNet {
		numNets++
		params.ActiveNetParams = &params.TestNetParam
	}
	if cfg.PrivNet {
		numNets++
		// Also disable dns seeding on the private test network.
		params.ActiveNetParams = &params.PrivNetParam
	}
	if cfg.MixNet {
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
		err := fmt.Errorf(str, funcName)
		fmt.Fprintln(os.Stderr, err)
		fmt.Fprintln(os.Stderr, usageMessage)
		return nil, nil, err
	}

	if err := params.ActiveNetParams.PowConfig.Check(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		return nil, nil, err
	}

	cfg.DataDir = util.CleanAndExpandPath(cfg.DataDir)
	cfg.DataDir = filepath.Join(cfg.DataDir, params.ActiveNetParams.Name)

	// Validate the number of candidates.
	if cfg.NumCandidates < minCandidates || cfg.NumCandidates > maxCandidates {
		str := "%s: The specified number of candidates is out of " +
			"range -- parsed [%v]"
		err = fmt.Errorf(str, "loadConfig", cfg.NumCandidates)
		fmt.Fprintln(os.Stderr, err)
		preParser.WriteHelp(os.Stderr)
		return nil, nil, err
	}
	return &cfg, remainingArgs, nil
}
