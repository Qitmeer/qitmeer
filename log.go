// Copyright (c) 2017-2018 The qitmeer developers

package main

import (
	"github.com/HalalChain/qitmeer/node"
	"os"
	"io"
	"github.com/mattn/go-colorable"
	"github.com/jrick/logrotate/rotator"
	"github.com/HalalChain/qitmeer-lib/log"
	"github.com/HalalChain/qitmeer-lib/log/term"
	"path/filepath"
	"fmt"
	"github.com/HalalChain/qitmeer/database"
	"github.com/HalalChain/qitmeer-lib/engine/txscript"
	"github.com/HalalChain/qitmeer/services/blkmgr"
	"github.com/HalalChain/qitmeer/core/blockchain"
	"github.com/HalalChain/qitmeer/services/miner"
	"github.com/HalalChain/qitmeer/core/blockdag"
)

var (
	glogger *log.GlogHandler

	logWrite *logWriter
)

// logWriter implements an io.Writer that outputs to both standard output and
// the write-end pipe of an initialized log rotator.
type logWriter struct{
	// logRotator is one of the logging outputs.  It should be closed on
	// application shutdown.
	logRotator *rotator.Rotator

	// Use for color terminal
	colorableWrite io.Writer
}

func (lw *logWriter) Init() {
	// init a colorful logger if possible
	usecolor := term.IsTty(os.Stdout.Fd()) && os.Getenv("TERM") != "dumb"

	if usecolor {
		lw.colorableWrite = colorable.NewColorableStderr()
	}
}

func (lw *logWriter) Close() {
	if lw.logRotator != nil {
		lw.logRotator.Close()
	}
}

func (lw *logWriter) IsUseColor() bool {
	return lw.colorableWrite != nil
}

func (lw *logWriter) Write(p []byte) (n int, err error) {
	os.Stderr.Write(p)
	if lw.logRotator != nil {
		lw.logRotator.Write(p)
	}

	if lw.colorableWrite != nil {
		lw.colorableWrite.Write(p)
	}

	return len(p), nil
}

func init() {
	// output set to Stderr
	// it's easier to handle when run as a daemon through systemd or supervisord,
	// and Go runtime exceptions are printed to stderr as well.
	logWrite=&logWriter{}
	logWrite.Init()
	glogger = log.NewGlogHandler(log.StreamHandler(io.Writer(logWrite), log.TerminalFormat(logWrite.IsUseColor())))

	log.Root().SetHandler(glogger)

	database.UseLogger(log.New(log.Ctx{"module": "database"}))
	txscript.UseLogger(log.New(log.Ctx{"module": "txscript engine"}))
	blockchainlogger := log.New(log.Ctx{"module": "blockchain"})
	minerlogger := log.New(log.Ctx{"module": "cpu miner"})
	rpclogger := log.New(log.Ctx{"module": "node"})
	blockdaglogger := log.New(log.Ctx{"module": "blockdag"})
	blkmgr.UseLogger(blockchainlogger)
	blockchain.UseLogger(blockchainlogger)
	miner.UseLogger(minerlogger)
	node.UseLogger(rpclogger)
	blockdag.UseLogger(blockdaglogger)
}

// initLogRotator initializes the logging rotater to write logs to logFile and
// create roll files in the same directory.  It must be called before the
// package-global log rotater variables are used.
func initLogRotator(logFile string) {
	logDir, _ := filepath.Split(logFile)
	err := os.MkdirAll(logDir, 0700)
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to create log directory: %v\n", err)
		os.Exit(1)
	}
	r, err := rotator.New(logFile, 10*1024, false, 3)
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to create file rotator: %v\n", err)
		os.Exit(1)
	}

	logWrite.logRotator = r
}
