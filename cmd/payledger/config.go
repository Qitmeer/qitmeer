/*
 * Copyright (c) 2020.
 * Project:qitmeer
 * File:config.go
 * Date:5/12/20 9:55 PM
 * Author:Jin
 * Email:lochjin@gmail.com
 */
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
	defaultDataDirname    = "data"
	defaultSrcDataDirname = "srcdata"
)

var (
	defaultHomeDir    = util.AppDataDir(".", false)
	defaultDataDir    = filepath.Join(defaultHomeDir, defaultDataDirname)
	defaultSrcDataDir = filepath.Join(defaultHomeDir, defaultSrcDataDirname)
	defaultDbType     = "ffldb"
	defaultDAGType    = "phantom"
)

type Config struct {
	HomeDir string `short:"A" long:"appdata" description:"Path to application home directory"`
	DataDir string `short:"b" long:"datadir" description:"Directory to store data"`
	TestNet bool   `long:"testnet" description:"Use the test network"`
	MixNet  bool   `long:"mixnet" description:"Use the test mix pow network"`
	PrivNet bool   `long:"privnet" description:"Use the private network"`
	DbType  string `long:"dbtype" description:"Database backend to use for the Block Chain"`
	DAGType string `short:"G" long:"dagtype" description:"DAG type {phantom,conflux,spectre} "`

	SrcDataDir      string `long:"srcdatadir" description:"Original directory to store data"`
	EndPoint        string `long:"endpoint" description:"The end point block hash when building ledger"`
	CheckEndPoint   string `long:"checkendpoint" description:"Check the end point"`
	ShowEndPoints   int    `long:"showendpoints" description:"Recommend some end blocks from main chain tip to genesis."`
	EndPointSkips   int    `long:"endpointskips" description:"Recommend some end blocks and skip some main chain blocks."`
	SavePayoutsFile bool   `long:"savefile"  description:"save result to the payouts file."`
	DisableBar      bool   `long:"disablebar"  description:"Hide progress bar."`
	DebugAddress    string `long:"debugaddress"  description:"Debug address."`
	DebugAddrUTXO   bool   `long:"debugaddrutxo"  description:"Print only utxo about the address."`
	DebugAddrValid  bool   `long:"debugaddrvalid"  description:"Print only valid data about the address."`
	Last            bool   `long:"last"  description:"Show ledger by last building data."`
	BlocksInfo      bool   `long:"blocksinfo"  description:"Show all blocks information."`

	UnlocksPerHeight int `long:"unlocksperheight"  description:"How many will be unlocked at each DAG main height."`
}

func LoadConfig() (*Config, []string, error) {
	// Default config.
	cfg := Config{
		HomeDir:          defaultHomeDir,
		DataDir:          defaultDataDir,
		DbType:           defaultDbType,
		DAGType:          defaultDAGType,
		SrcDataDir:       defaultSrcDataDir,
		SavePayoutsFile:  false,
		DisableBar:       false,
		DebugAddrUTXO:    false,
		DebugAddrValid:   false,
		Last:             false,
		BlocksInfo:       false,
		UnlocksPerHeight: 0,
	}

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
	}

	cfg.DbType = preCfg.DbType
	cfg.DAGType = preCfg.DAGType
	cfg.EndPoint = preCfg.EndPoint
	cfg.ShowEndPoints = preCfg.ShowEndPoints
	cfg.CheckEndPoint = preCfg.CheckEndPoint
	cfg.EndPointSkips = preCfg.EndPointSkips
	cfg.SavePayoutsFile = preCfg.SavePayoutsFile
	cfg.DisableBar = preCfg.DisableBar
	cfg.DebugAddress = preCfg.DebugAddress
	cfg.DebugAddrUTXO = preCfg.DebugAddrUTXO
	cfg.DebugAddrValid = preCfg.DebugAddrValid
	cfg.Last = preCfg.Last
	cfg.BlocksInfo = preCfg.BlocksInfo
	cfg.MixNet = preCfg.MixNet
	cfg.TestNet = preCfg.TestNet
	cfg.PrivNet = preCfg.PrivNet
	cfg.UnlocksPerHeight = preCfg.UnlocksPerHeight

	if len(preCfg.SrcDataDir) > 0 {
		cfg.SrcDataDir = preCfg.SrcDataDir
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

	cfg.SrcDataDir = util.CleanAndExpandPath(cfg.SrcDataDir)
	cfg.SrcDataDir = filepath.Join(cfg.SrcDataDir, params.ActiveNetParams.Name)

	if len(cfg.EndPoint) == 0 &&
		len(cfg.CheckEndPoint) == 0 &&
		cfg.ShowEndPoints == 0 &&
		len(cfg.DebugAddress) == 0 &&
		!cfg.Last &&
		!cfg.BlocksInfo {
		err := fmt.Errorf("No Command")
		fmt.Fprintln(os.Stderr, err)
		fmt.Fprintln(os.Stderr, usageMessage)
		return nil, nil, err
	}
	return &cfg, remainingArgs, nil
}
