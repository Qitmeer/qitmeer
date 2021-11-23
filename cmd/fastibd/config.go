package main

import (
	"fmt"
	"github.com/Qitmeer/qng-core/common/util"
	"github.com/Qitmeer/qng-core/params"
	"os"
	"path/filepath"
	"strings"
)

const (
	defaultDataDirname = "data"
)

var (
	defaultHomeDir  = util.AppDataDir(".", false)
	defaultDataDir  = filepath.Join(defaultHomeDir, defaultDataDirname)
	defaultDbType   = "ffldb"
	defaultDAGType  = "phantom"
	defaultFileName = "blocks.ibd"
)

type Config struct {
	HomeDir string
	DataDir string
	TestNet bool
	MixNet  bool
	PrivNet bool
	DbType  string
	DAGType string

	OutputPath string
	InputPath  string
	DisableBar bool
	EndPoint   string
	ByID       bool
	AidMode    bool
	CPUNum     int
}

func (c *Config) load() error {

	// Create the home directory if it doesn't already exist.
	funcName := "loadConfig"
	err := os.MkdirAll(c.HomeDir, 0700)
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
		return err
	}

	// assign active network params while we're at it
	numNets := 0
	if c.TestNet {
		numNets++
		params.ActiveNetParams = &params.TestNetParam
	}
	if c.PrivNet {
		numNets++
		// Also disable dns seeding on the private test network.
		params.ActiveNetParams = &params.PrivNetParam
	}
	if c.MixNet {
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
		return err
	}

	if err := params.ActiveNetParams.PowConfig.Check(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		return err
	}

	c.DataDir = util.CleanAndExpandPath(c.DataDir)
	c.DataDir = filepath.Join(c.DataDir, params.ActiveNetParams.Name)

	return nil
}

func GetIBDFilePath(path string) (string, error) {
	if len(path) <= 0 {
		return "", fmt.Errorf("Path error")
	}
	if len(path) >= 4 {
		if path[len(path)-4:] == ".ibd" {
			return path, nil
		}
	}
	return strings.TrimRight(strings.TrimRight(path, "/"), "\\") + "/" + defaultFileName, nil
}
