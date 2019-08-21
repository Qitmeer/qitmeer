// Copyright (c) 2017-2018 The qitmeer developers

package main

import (
	"github.com/Qitmeer/qitmeer/node"
	"github.com/Qitmeer/qitmeer/services/tx"
	"os"
	"io"
	"github.com/mattn/go-colorable"
	"github.com/jrick/logrotate/rotator"
	"github.com/Qitmeer/qitmeer-lib/log"
	"github.com/Qitmeer/qitmeer-lib/log/term"
	"path/filepath"
	"fmt"
	"github.com/Qitmeer/qitmeer/database"
	"github.com/Qitmeer/qitmeer-lib/engine/txscript"
	"github.com/Qitmeer/qitmeer/services/blkmgr"
	"github.com/Qitmeer/qitmeer/core/blockchain"
	"github.com/Qitmeer/qitmeer/services/miner"
	"github.com/Qitmeer/qitmeer/core/blockdag"
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
	if lw.logRotator != nil {
		lw.logRotator.Write(p)
	}

	if lw.colorableWrite != nil {
		lw.colorableWrite.Write(p)
	}else{
		os.Stderr.Write(p)
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
	txscript.UseLogger(log.New(log.Ctx{"module": "txscript"}))
	blkmgr.UseLogger(log.New(log.Ctx{"module": "blkmanager"}))
	blockchain.UseLogger(log.New(log.Ctx{"module": "blockchain"}))
	miner.UseLogger(log.New(log.Ctx{"module": "cpuminer"}))
	node.UseLogger(log.New(log.Ctx{"module": "node"}))
	blockdag.UseLogger(log.New(log.Ctx{"module": "blockdag"}))
	tx.UseLogger(log.New(log.Ctx{"module": "txmanager"}))
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
