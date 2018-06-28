// Copyright (c) 2017-2018 The nox developers
// Copyright (c) 2015-2016 The Decred developers
// Copyright (c) 2013-2016 The btcsuite developers

package main

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"github.com/jessevdk/go-flags"
	"github.com/noxproject/nox/common/util"
	"github.com/noxproject/nox/log"
	"github.com/noxproject/nox/config"
)

const (
	defaultConfigFilename        = "noxd.conf"
	defaultDataDirname           = "data"
	defaultLogLevel              = "info"
	defaultDebugPrintOrigins     = false
	defaultLogDirname            = "logs"
	defaultLogFilename           = "noxd.log"
	defaultGenerate              = false
	defaultMaxRPCClients         = 10

)

var (
	defaultHomeDir     = util.AppDataDir("noxd", false)
	defaultConfigFile  = filepath.Join(defaultHomeDir, defaultConfigFilename)
	defaultDataDir     = filepath.Join(defaultHomeDir, defaultDataDirname)
	defaultDbType      = "ffldb"
	defaultLogDir      = filepath.Join(defaultHomeDir, defaultLogDirname)
	defaultRPCKeyFile  = filepath.Join(defaultHomeDir, "rpc.key")
	defaultRPCCertFile = filepath.Join(defaultHomeDir, "rpc.cert")
)



// loadConfig initializes and parses the config using a config file and command
// line options.
func loadConfig() (*config.Config, []string, error) {

	// Default config.
	cfg := config.Config{
		HomeDir:              defaultHomeDir,
		ConfigFile:           defaultConfigFile,
		DebugLevel:           defaultLogLevel,
		DebugPrintOrigins:    defaultDebugPrintOrigins,
		DataDir:              defaultDataDir,
		LogDir:               defaultLogDir,
		DbType:               defaultDbType,
		RPCKey:               defaultRPCKeyFile,
		RPCCert:              defaultRPCCertFile,
		RPCMaxClients:        defaultMaxRPCClients,
		Generate:             defaultGenerate,
	}

	// Pre-parse the command line options to see if an alternative config
	// file or the version flag was specified.  Any errors aside from the
	// help message error can be ignored here since they will be caught by
	// the final parse below.
	preCfg := cfg
	preParser := newConfigParser(&preCfg, flags.Default)
	_, err := preParser.Parse()
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
	if preCfg.ShowVersion {
		fmt.Printf("%s version %s (Go version %s)\n", appName, version(), runtime.Version())
		os.Exit(0)
	}

	usageMessage := fmt.Sprintf("Use %s -h to show usage", appName)

	// TODO
	// Perform service command and exit if specified.  Invalid service
	// commands show an appropriate error.  Only runs on Windows since
	// the runServiceCommand function will be nil when not on Windows.
	// TODO

	// Update the home directory for noxd if specified. Since the home
	// directory is updated, other variables need to be updated to
	// reflect the new changes.
	if preCfg.HomeDir != "" {
		cfg.HomeDir, _ = filepath.Abs(preCfg.HomeDir)

		if preCfg.ConfigFile == defaultConfigFile {
			defaultConfigFile = filepath.Join(cfg.HomeDir,
				defaultConfigFilename)
			preCfg.ConfigFile = defaultConfigFile
			cfg.ConfigFile = defaultConfigFile
		} else {
			cfg.ConfigFile = preCfg.ConfigFile
		}
		if preCfg.DataDir == defaultDataDir {
			cfg.DataDir = filepath.Join(cfg.HomeDir, defaultDataDirname)
		} else {
			cfg.DataDir = preCfg.DataDir
		}
		if preCfg.RPCKey == defaultRPCKeyFile {
			cfg.RPCKey = filepath.Join(cfg.HomeDir, "rpc.key")
		} else {
			cfg.RPCKey = preCfg.RPCKey
		}
		if preCfg.RPCCert == defaultRPCCertFile {
			cfg.RPCCert = filepath.Join(cfg.HomeDir, "rpc.cert")
		} else {
			cfg.RPCCert = preCfg.RPCCert
		}
		if preCfg.LogDir == defaultLogDir {
			cfg.LogDir = filepath.Join(cfg.HomeDir, defaultLogDirname)
		} else {
			cfg.LogDir = preCfg.LogDir
		}
	}

	// TODO
	// Create a default config file when one does not exist and the user did
	// not specify an override.
	// TODO


	// Load additional config from file.
	var configFileError error
	parser := newConfigParser(&cfg, flags.Default)
	if !cfg.PrivNet || preCfg.ConfigFile != defaultConfigFile {
		err := flags.NewIniParser(parser).ParseFile(preCfg.ConfigFile)
		if err != nil {
			if _, ok := err.(*os.PathError); !ok {
				fmt.Fprintf(os.Stderr, "Error parsing config "+
					"file: %v\n", err)
				fmt.Fprintln(os.Stderr, usageMessage)
				return nil, nil, err
			}
			configFileError = err
		}
	}

	// Parse command line options again to ensure they take precedence.
	remainingArgs, err := parser.Parse()
	if err != nil {
		if e, ok := err.(*flags.Error); !ok || e.Type != flags.ErrHelp {
			fmt.Fprintln(os.Stderr, usageMessage)
		}
		return nil, nil, err
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
		activeNetParams = &testNetParams
	}
	if cfg.PrivNet {
		numNets++
		// Also disable dns seeding on the private test network.
		activeNetParams = &privNetParams
		cfg.DisableDNSSeed = true
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

	// Append the network type to the data directory so it is "namespaced"
	// per network.  In addition to the block database, there are other
	// pieces of data that are saved to disk such as address manager state.
	// All data is specific to a network, so namespacing the data directory
	// means each individual piece of serialized data does not have to
	// worry about changing names per network and such.
	cfg.DataDir = util.CleanAndExpandPath(cfg.DataDir)
	cfg.DataDir = filepath.Join(cfg.DataDir, activeNetParams.Name)

	// Set logging file if presented
	if !cfg.NoFileLogging {
		// Append the network type to the log directory so it is "namespaced"
		// per network in the same fashion as the data directory.
		cfg.LogDir = util.CleanAndExpandPath(cfg.LogDir)
		cfg.LogDir = filepath.Join(cfg.LogDir, activeNetParams.Name)

		// Initialize log rotation.  After log rotation has been initialized, the
		// logger variables may be used.
		initLogRotator(filepath.Join(cfg.LogDir, defaultLogFilename))
	}

	// Parse, validate, and set debug log level(s).
	if err := parseAndSetDebugLevels(cfg.DebugLevel); err != nil {
		err := fmt.Errorf("%s: %v", funcName, err.Error())
		fmt.Fprintln(os.Stderr, err)
		fmt.Fprintln(os.Stderr, usageMessage)
		return nil, nil, err
	}

	// DebugPrintOrigins
	if cfg.DebugPrintOrigins {
		log.PrintOrigins(true)
	}

	// Warn about missing config file only after all other configuration is
	// done.  This prevents the warning on help messages and invalid
	// options.  Note this should go directly before the return.
	if configFileError != nil {
		log.Warn("missing config file", "error",configFileError)
	}
	return &cfg, remainingArgs, nil
}

// newConfigParser returns a new command line flags parser.
func newConfigParser(cfg *config.Config, options flags.Options) *flags.Parser {
	parser := flags.NewParser(cfg, options)
	return parser
}

// parseAndSetDebugLevels attempts to parse the specified debug level and set
// the levels accordingly.  An appropriate error is returned if anything is
// invalid.
func parseAndSetDebugLevels(debugLevel string) error {

	// When the specified string doesn't have any delimters, treat it as
	// the log level for all subsystems.
	if !strings.Contains(debugLevel, ",") && !strings.Contains(debugLevel, "=") {
		// Validate debug log level.
		lvl, err := log.LvlFromString(debugLevel)
		if err != nil {
			str := "the specified debug level [%v] is invalid"
			return fmt.Errorf(str, debugLevel)
		}
		// Change the logging level for all subsystems.
		glogger.Verbosity(lvl)
		return nil
	}
	// TODO support log for subsystem
	return nil
}

